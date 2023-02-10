package srvutil

import (
	"net/http"

	"github.com/Shopify/goose/v2/httperrors"
)

func Fallible(fn func() error) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := fn()
		if err != nil {
			if he, ok := err.(*httperrors.HttpError); ok {
				he.Write(w)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}
