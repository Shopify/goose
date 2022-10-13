package srvutil

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/tomb.v2"

	"github.com/Shopify/goose/logger"
	"github.com/Shopify/goose/safely"
)

const (
	keepAlivePeriod = 3 * time.Minute
)

// Server wraps an http.Server to make it runnable and stoppable
// If its tomb dies, the server will be stopped
type Server interface {
	safely.Runnable
	Addr() *net.TCPAddr
}

func NewServer(t *tomb.Tomb, bind string, servlet Servlet) Server {
	return NewServerFromFactory(t, servlet, func(handler http.Handler) http.Server {
		return http.Server{ //nolint:gosec
			Addr:    bind,
			Handler: handler,
		}
	})
}

type ServerFactory func(handler http.Handler) http.Server

func NewServerFromFactory(t *tomb.Tomb, servlet Servlet, factory ServerFactory) Server {
	router := mux.NewRouter()
	servlet.RegisterRouting(router)

	return &server{
		server:   factory(router),
		haveAddr: make(chan struct{}),
		tomb:     t,
	}
}

type server struct {
	server http.Server
	tomb   *tomb.Tomb

	haveAddr chan struct{}
	addr     *net.TCPAddr
}

func (c *server) Tomb() *tomb.Tomb {
	return c.tomb
}

func (c *server) Addr() *net.TCPAddr {
	<-c.haveAddr
	return c.addr
}

func (c *server) Run() error {
	ctx := logger.WithField(context.Background(), "bind", c.server.Addr)

	log(ctx, nil).Info("starting server")

	ln, err := net.Listen("tcp", c.server.Addr)
	if err != nil {
		return err
	}

	c.addr = ln.Addr().(*net.TCPAddr)
	close(c.haveAddr)

	ctx = logger.WithField(ctx, "addr", c.addr.String())

	log(ctx, nil).Info("started server")
	defer log(ctx, nil).Debug("stopped server")

	listener := stoppableKeepaliveListener{
		TCPListener: ln.(*net.TCPListener),
		tomb:        c.tomb,
	}

	shutdown := make(chan error)
	go func() {
		<-c.tomb.Dying()
		log(ctx, c.tomb.Err()).Info("shutting down server")

		// Call Shutdown to allow in-flight requests to gracefully complete.
		ctx := context.Background()
		shutdown <- c.server.Shutdown(ctx)
	}()

	if err := c.server.Serve(listener); err != http.ErrServerClosed {
		return err
	}

	log(ctx, nil).Debug("waiting for server to complete shutdown")
	return <-shutdown
}

type stoppableKeepaliveListener struct {
	*net.TCPListener

	tomb *tomb.Tomb
}

func (ln stoppableKeepaliveListener) Accept() (net.Conn, error) {
	for {
		if !ln.tomb.Alive() {
			return nil, http.ErrServerClosed
		}

		if err := ln.SetDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
			return nil, err
		}

		tc, err := ln.AcceptTCP()

		if err != nil {
			netErr, ok := err.(net.Error)

			// If this is a timeout, then continue to wait for new connections
			if ok && netErr.Timeout() && netErr.Temporary() { //nolint:staticcheck
				continue
			} else {
				return nil, err
			}
		}

		if err := tc.SetKeepAlive(true); err != nil {
			return nil, err
		}
		if err := tc.SetKeepAlivePeriod(keepAlivePeriod); err != nil {
			return nil, err
		}
		return tc, nil
	}
}
