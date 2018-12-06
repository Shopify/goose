package srvutil

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Servlet represents a component that can handle HTTP requests
// It it responsible for registering on the Router.
//
// Simple Servlets can be created with FuncServlet or HandlerServlet,
// but a common pattern is to have a complex struct that will initialize dependencies.
//
// Servlets can also be modified and modified using PrefixServlet, UseServlet, and CombineServlets.
//
// See examples.
type Servlet interface {
	RegisterRouting(r *mux.Router)
}

type InlineServlet func(r *mux.Router)

func (m InlineServlet) RegisterRouting(r *mux.Router) {
	m(r)
}

// FuncServlet creates a Servlet from a function.
func FuncServlet(path string, handler http.HandlerFunc) Servlet {
	return InlineServlet(func(r *mux.Router) {
		r.HandleFunc(path, handler)
	})
}

// HandlerServlet creates a Servlet from a Handler, which implements ServeHTTP.
func HandlerServlet(path string, handler http.Handler) Servlet {
	return InlineServlet(func(r *mux.Router) {
		r.Handle(path, handler)
	})
}

// PrefixServlet registers a whole Servlet under a path prefix.
func PrefixServlet(s Servlet, path string) Servlet {
	return InlineServlet(func(r *mux.Router) {
		s.RegisterRouting(r.PathPrefix(path).Subrouter())
	})
}

// UseServlet applies a middleware to a whole Servlet.
// Great for applying authentication layers.
func UseServlet(s Servlet, mwf ...mux.MiddlewareFunc) Servlet {
	return InlineServlet(func(r *mux.Router) {
		r = r.NewRoute().Subrouter()
		r.Use(mwf...)
		s.RegisterRouting(r)
	})
}

// CombineServlets combines all Servlets into one, without other modifications.
// Great for splitting the dependency managements into smaller servlets.
func CombineServlets(servlets ...Servlet) Servlet {
	return InlineServlet(func(r *mux.Router) {
		for _, s := range servlets {
			s.RegisterRouting(r)
		}
	})
}
