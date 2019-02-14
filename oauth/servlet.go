package oauth

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Shopify/goose/logger"
	"github.com/Shopify/goose/srvutil"
)

var (
	ErrUnauthenticated = errors.New("user is not authenticated")
	ErrLoginDisabled   = errors.New("login disabled")
)

type Servlet interface {
	srvutil.Servlet
	Middleware(next http.Handler) http.Handler
}

func NewServlet(manager Manager, paths *Paths) Servlet {
	return &servlet{
		manager: manager,
		paths:   paths,
	}
}

type servlet struct {
	manager Manager
	paths   *Paths
}

func (s *servlet) RegisterRouting(r *mux.Router) {
	r.HandleFunc(s.paths.LoginPath, s.loginHandler)
	r.HandleFunc(s.paths.CallbackPath, s.callbackHandler)
}

func (s *servlet) loginEnabled(w http.ResponseWriter, r *http.Request) bool {
	if s.manager != nil {
		return true
	}

	log(r.Context(), nil).Error("login not configured")
	http.Error(w, "Unable to initiate login.", http.StatusInternalServerError)
	return false
}

func (s *servlet) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.authorize(w, r)
		if err != nil {
			log(r.Context(), err).Warn("user is not authorized")
			redirect(r.Context(), w, r, s.paths.LoginURL(r.URL.Path).String())
			return
		}
		r = r.WithContext(logger.WithLoggable(r.Context(), user))

		next.ServeHTTP(w, r)
	})
}

func (s *servlet) authorize(w http.ResponseWriter, r *http.Request) (*User, error) {
	ctx := r.Context()

	if !s.loginEnabled(w, r) {
		return nil, ErrLoginDisabled
	}

	token, ok := s.manager.GetSavedToken(r)
	if !ok {
		return nil, ErrUnauthenticated
	}

	user, err := s.manager.AuthorizeToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *servlet) loginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !s.loginEnabled(w, r) {
		return
	}

	originURL := r.URL.Query().Get("redirect")
	u, err := s.manager.GetLoginURL(ctx, originURL)
	if err != nil {
		log(ctx, err).Error("unable to generate login url")
		http.Error(w, "Unable to initiate login.", http.StatusInternalServerError)
		return
	}

	redirect(ctx, w, r, u)
}

func (s *servlet) callbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !s.loginEnabled(w, r) {
		return
	}

	encState := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if encState == "" || code == "" {
		log(ctx, nil).Warn("parameters missing for OAuth")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	state, err := s.manager.DecodeState(encState)
	if err != nil {
		log(ctx, err).Warn("unable to decode state")
		http.Error(w, "Unable to complete login.", http.StatusForbidden)
		return
	}

	token, err := s.manager.ExchangeCode(r.Context(), code)
	if err != nil {
		log(ctx, err).Warn("unable to exchange oauth code")
		http.Error(w, "Unable to complete login.", http.StatusForbidden)
		return
	}

	user, err := s.manager.AuthorizeToken(ctx, token)
	if err != nil {
		log(ctx, err).Warn("unable to authorize token")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	log(ctx, nil).WithField("user", user.Email).Info("logged as user")
	err = s.manager.SaveToken(w, r, token)
	if err != nil {
		log(ctx, err).Warn("unable to save cookie")
		http.Error(w, "Unable to complete login.", http.StatusForbidden)
		return
	}

	redirectURL := s.paths.RedirectURL(state.Origin)

	redirect(ctx, w, r, redirectURL.String())
}

func redirect(ctx context.Context, w http.ResponseWriter, r *http.Request, redirect string) {
	log(ctx, nil).WithField("redirect", redirect).Info("redirecting user")
	http.Redirect(w, r, redirect, http.StatusFound)
}
