package metrics

import (
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

const (
	DefaultImplementationEnvVar = "STATSD_IMPLEMENTATION"
	DefaultStatsdEndpoint       = "STATSD_ADDR"
	DefaultTagsEnvVar           = "STATSD_DEFAULT_TAGS"
)

var (
	ErrUnknownImplementation = errors.New("unknown metrics implementation")
	ErrEnvVarMissing         = errors.New("environment variable missing")
)

var (
	defaultBackend     = NewNullBackend()
	defaultBackendLock = sync.RWMutex{}
)

func DefaultBackend() Backend {
	defaultBackendLock.RLock()
	defer defaultBackendLock.RUnlock()

	return defaultBackend
}

// SetDefaultBackend replaces the current backend with the given Backend.
func SetDefaultBackend(backend Backend) {
	defaultBackendLock.Lock()
	defer defaultBackendLock.Unlock()

	defaultBackend = backend
}

// NewBackendFromEnv returns the appropriate Backend
func NewBackendFromEnv() (Backend, error) {
	impl := os.Getenv(DefaultImplementationEnvVar)

	switch strings.ToLower(impl) {
	case "datadog":
		addr := os.Getenv(DefaultStatsdEndpoint)
		if addr == "" {
			return nil, errors.Wrap(ErrEnvVarMissing, DefaultStatsdEndpoint)
		}
		return NewDatadogBackend(addr)
	case "log":
		return NewLogrusBackend(logrus.StandardLogger(), logrus.DebugLevel), nil
	case "null", "":
		return NewNullBackend(), nil
	default:
		return nil, errors.Wrap(ErrUnknownImplementation, impl)
	}
}

func ConfigureDefaultBackend(prefix string) (Backend, error) {
	c, err := NewBackendFromEnv()
	if err != nil {
		return nil, err
	}
	c = BackendWithDefaultWrappers(c, prefix)
	SetDefaultBackend(c)
	return c, nil
}

// BackendWithDefaultWrappers wraps the Backend with the standard wrappers
func BackendWithDefaultWrappers(backend Backend, prefix string) Backend {
	backend = NewPrefixWrapper(backend, prefix)
	backend = NewDefaultTagsWrapper(backend)
	backend = NewContextWrapper(backend)
	return backend
}

func NewDefaultTagsWrapper(backend Backend) Backend {
	return NewTagsWrapper(backend, TagsFromEnv(DefaultTagsEnvVar))
}
