package profiler

import (
	"net/http"
	httppprof "net/http/pprof"
	"strings"

	"github.com/gorilla/mux"

	"github.com/Shopify/goose/srvutil"
)

// NewServlet returns a Servlet which can serve pprof requests
//
// /debug/pprof/ is the homepage
// /debug/pprof/ui/ is the same homepage, but where all links point to a UI instead of raw downloads
// /debug/pprof/*/ are the raw handlers
// /debug/pprof/ui/*/ are the same handlers, but with a rendered UI
func NewServlet() srvutil.Servlet {
	s := &pprofServlet{}
	return srvutil.PrefixServlet(s, "/debug/pprof")
}

type pprofServlet struct{}

func (s *pprofServlet) RegisterRouting(r *mux.Router) {
	r.StrictSlash(true)

	r.HandleFunc("/cmdline", httppprof.Cmdline)
	r.HandleFunc("/profile", httppprof.Profile)
	r.HandleFunc("/symbol", httppprof.Symbol)
	r.HandleFunc("/trace", httppprof.Trace)

	r.HandleFunc("/ui/{profile}/{view}", pprofUIHandler)
	r.HandleFunc("/ui/{profile}/", pprofUIHandler)

	// Serve the same index as /pprof/, but because the path is /ui/, all links will be relative to that,
	// translating to /pprof/ui/{profile}
	r.HandleFunc("/ui/", func(w http.ResponseWriter, r *http.Request) {
		// But we have to hide it from the Index, otherwise it will look for a "ui" profile
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "ui/")
		httppprof.Index(w, r)
	})

	r.HandleFunc("/{profiler}", httppprof.Index)
	r.HandleFunc("/", httppprof.Index)
}
