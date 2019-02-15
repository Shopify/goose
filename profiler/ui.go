package profiler

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	httppprof "net/http/pprof"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/google/pprof/driver"
	"github.com/gorilla/mux"
)

func pprofUIHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	f, err := ioutil.TempFile("", "pprof.*.pb")
	if err != nil {
		log(ctx, err).Error("error creating temporary file")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log(ctx, err).WithField("file", f.Name()).Debug("writing pprof to temp file")
	defer removeFile(ctx, f.Name())

	q := r.URL.Query()
	q.Del("debug") // debug doesn't make sense for UI
	r.URL.RawQuery = q.Encode()

	// Record what would normally be served to the user in a file such that it can be passed to the UI renderer
	rec := httptest.NewRecorder()
	profile := vars["profile"]
	if profile == "profile" {
		httppprof.Profile(rec, r)
	} else {
		// Check validity of the profile
		p := pprof.Lookup(profile)
		if p == nil {
			// The /debug/pprof/ui page has links to all the profiles, but some are not supported, so redirect to non-ui.
			log(ctx, nil).WithField("profile", profile).Warn("unknown profile")
			url := strings.Replace(r.URL.Path, "/ui/", "/", 1)
			http.Redirect(w, r, url, http.StatusMovedPermanently)
			return
		}

		httppprof.Handler(profile).ServeHTTP(rec, r)
	}

	if rec.Code != 0 && rec.Code != http.StatusOK {
		log(ctx, err).Error("error recording pprof")
		http.Error(w, rec.Body.String(), rec.Code)
		return
	}
	if _, err = rec.Body.WriteTo(f); err != nil {
		log(ctx, err).Error("error closing file")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := serveUI(ctx, w, r, f.Name()); err != nil {
		log(ctx, err).Warn("error serving pprof ui")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func serveUI(ctx context.Context, w http.ResponseWriter, r *http.Request, inputFile string) error {
	vars := mux.Vars(r)
	return driver.PProf(&driver.Options{
		UI: &dummyUI{ctx: ctx},
		Flagset: &inlineFlagset{
			strings: map[string]string{
				// Meaningless, but force it to behave like a web server
				"http": "localhost:0",
			},
			args: []string{inputFile},
		},
		HTTPServer: func(args *driver.HTTPServerArgs) error {
			path := "/" + vars["view"]
			handler, ok := args.Handlers[path]
			if !ok {
				log(ctx, nil).Warn("unknown view")
				return errors.New("unknown view: " + vars["view"])
			}
			handler.ServeHTTP(w, r)
			return nil
		},
	})
}

func removeFile(ctx context.Context, path string) {
	if err := os.Remove(path); err != nil {
		log(ctx, err).Error("unable to remove file")
	}
}

// inlineFlagset implements the plugin.FlagSet interface.
// Adapted from testFlags in github.com/google/pprof/internal/driver
type inlineFlagset struct {
	strings map[string]string
	args    []string
}

func (f inlineFlagset) AddExtraUsage(_ string) {}

func (inlineFlagset) ExtraUsage() string { return "" }

func (f inlineFlagset) Bool(_ string, d bool, _ string) *bool {
	return &d
}

func (f inlineFlagset) Int(_ string, d int, _ string) *int {
	return &d
}

func (f inlineFlagset) Float64(_ string, d float64, _ string) *float64 {
	return &d
}

func (f inlineFlagset) String(s, d, _ string) *string {
	if t, ok := f.strings[s]; ok {
		return &t
	}
	return &d
}

func (f inlineFlagset) BoolVar(p *bool, _ string, d bool, _ string) {
	*p = d
}

func (f inlineFlagset) IntVar(p *int, _ string, d int, _ string) {
	*p = d
}

func (f inlineFlagset) Float64Var(p *float64, _ string, d float64, _ string) {
	*p = d
}

func (f inlineFlagset) StringVar(p *string, s, d, _ string) {
	if t, ok := f.strings[s]; ok {
		*p = t
	} else {
		*p = d
	}
}

func (f inlineFlagset) StringList(s, d, _ string) *[]*string {
	return &[]*string{}
}

func (f inlineFlagset) Parse(func()) []string {
	return f.args
}

// dummyUI implements a basic UI such that WantBrowser is false and we has a custom logger
type dummyUI struct {
	ctx context.Context
}

func (ui *dummyUI) ReadLine(prompt string) (string, error) {
	return "", nil
}

func (ui *dummyUI) Print(args ...interface{}) {
	log(ui.ctx, nil).Print(args...)
}

func (ui *dummyUI) PrintErr(args ...interface{}) {
	log(ui.ctx, nil).Warn(args...)
}

func (ui *dummyUI) IsTerminal() bool {
	return false
}

func (ui *dummyUI) WantBrowser() bool {
	return false
}

func (ui *dummyUI) SetAutoComplete(func(string) string) {
}
