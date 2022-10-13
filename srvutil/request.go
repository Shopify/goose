package srvutil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/Shopify/goose/logger"
	"github.com/Shopify/goose/statsd"
)

var log = logger.New("srvutil")

const (
	UserEmailHeaderKey = "x-galaxy-user-email"
	UserEmailKey       = "email"
	UUIDHeaderKey      = "x-request-id"
	PathKey            = "path"
	RouteKey           = "route"
)

// Use route and vars to add to context
// Example:
// /hello/{name}, with name=world
// route: /hello/@name
// route_name: world
func buildRouteContext(r *http.Request) context.Context {
	ctx := r.Context()
	// Always add the path to the log fields,
	// but not to the tags since Datadog doesn't like tags with high cardinality.
	ctx = logger.WithField(ctx, PathKey, r.URL.Path)

	route := mux.CurrentRoute(r)
	if route == nil {
		log(ctx, nil).Debug("unable to get current route")
		return ctx
	}

	tpl, err := route.GetPathTemplate()
	if err != nil {
		log(ctx, err).Error("unable to get the route's template")
		return ctx
	}

	tpl, err = replaceMatchableParts(tpl)
	if err != nil {
		log(ctx, err).Error("unable to parse the route's template")
		return ctx
	}

	// Adds the route as log field and tag.
	ctx = statsd.WithTag(ctx, RouteKey, tpl)

	// Adds the mux variables as log fields only, but not tags.
	fields := logrus.Fields{RouteKey: tpl}
	for k, v := range mux.Vars(r) {
		fields[fmt.Sprintf("%s_%s", RouteKey, k)] = v
	}
	ctx = logger.WithFields(ctx, fields)

	return ctx
}

func BuildContext(r *http.Request) (context.Context, string) {
	ctx := buildRouteContext(r)

	// If caller specifies a request ID, use that instead of generating one
	var id string
	if id = r.Header.Get(UUIDHeaderKey); id == "" {
		ctx, id = logger.WithUUID(ctx)
	} else {
		ctx = logger.WithField(ctx, logger.UUIDKey, id)
	}

	if email := r.Header.Get(UserEmailHeaderKey); email != "" {
		ctx = logger.WithField(ctx, UserEmailKey, email)
	}

	return ctx, id
}

// RequestContextMiddleware can be used with github.com/gorilla/mux:Router.Use or wrapping a Handler
func RequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, id := BuildContext(r)
		r = r.WithContext(ctx)
		w.Header().Set(UUIDHeaderKey, id)

		next.ServeHTTP(w, r)
	})
}

// replaceMatchableParts replaces mux templates "/{name:[a-z]+}" with "/hello/@name" to be tag-friendly
// Works with:
//
//	/{name}
//	/{name:[a-z]+}
//	/{name:(?:[a-z]{2}){2}}
func replaceMatchableParts(tpl string) (string, error) {
	const (
		modePath = iota
		modeName
		modeMatcher
	)

	var result strings.Builder
	mode := modePath
	nestCount := 0

	for _, char := range tpl {
		switch mode {
		case modePath:
			switch char {
			case '{':
				mode = modeName
				result.WriteRune('@')
			case '}':
				return "", errors.New("unexpected closing curly brace")
			default:
				result.WriteRune(char)
			}
		case modeName:
			switch char {
			case ':':
				mode = modeMatcher
			case '{':
				return "", errors.New("unexpected opening curly brace")
			case '}':
				mode = modePath
			default:
				result.WriteRune(char)
			}
		case modeMatcher:
			switch char {
			case '{':
				nestCount++
			case '}':
				if nestCount == 0 {
					mode = modePath
				} else {
					nestCount--
				}
			}
		}
	}
	if mode != modePath {
		return "", errors.New("unbalanced curly braces")
	}
	return result.String(), nil
}
