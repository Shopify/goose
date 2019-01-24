package srvutil_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/Shopify/goose/srvutil"
)

func ExampleEnvVarHeaderMiddleware() {
	if os.Getenv("HOSTNAME") == "" {
		// Not all systems set the HOSTNAME env var
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatal(err)
		}

		err = os.Setenv("HOSTNAME", hostname)
		if err != nil {
			log.Fatal(err)
		}
	}

	r := mux.NewRouter()
	r.Use(srvutil.EnvVarHeaderMiddleware(map[string]string{
		"HOSTNAME": "X-Hostname",
	}))
}

func TestEnvVarHeaderMiddleware(t *testing.T) {
	err := os.Setenv("GOOSE_TEST", "foo")
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Use(srvutil.EnvVarHeaderMiddleware(map[string]string{
		"GOOSE_TEST": "X-Goose-Test",
	}))

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		log.Fatal(err)
	}

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "foo", w.Header().Get("X-Goose-Test"))
}
