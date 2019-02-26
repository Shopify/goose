package srvutil

import (
	"context"
	"errors"
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

// ErrStopped indicates the server was stopped by a request to the Tomb.
var ErrStopped = errors.New("stopped")

// Server wraps an http.Server to make it runnable and stoppable
// If its tomb dies, the server will be stopped
type Server interface {
	safely.Runnable
	Addr() *net.TCPAddr
}

func NewServer(t *tomb.Tomb, bind string, servlet Servlet) Server {
	router := mux.NewRouter()
	servlet.RegisterRouting(router)

	return &server{
		Server: http.Server{
			Addr:    bind,
			Handler: router,
		},
		haveAddr: make(chan struct{}),
		tomb:     t,
	}
}

type server struct {
	http.Server
	tomb *tomb.Tomb

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
	ctx := logger.WithField(context.Background(), "bind", c.Server.Addr)

	log(ctx, nil).Info("starting server")

	ln, err := net.Listen("tcp", c.Server.Addr)
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
	return c.Serve(listener)
}

type stoppableKeepaliveListener struct {
	*net.TCPListener

	tomb *tomb.Tomb
}

func (ln stoppableKeepaliveListener) Accept() (net.Conn, error) {
	for {
		select {
		case <-ln.tomb.Dying():
			return nil, ErrStopped
		default:
		}

		if err := ln.SetDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
			return nil, err
		}

		tc, err := ln.AcceptTCP()

		if err != nil {
			netErr, ok := err.(net.Error)

			//If this is a timeout, then continue to wait for new connections
			if ok && netErr.Timeout() && netErr.Temporary() {
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
