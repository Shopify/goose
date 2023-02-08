package metrics

import (
	"github.com/Shopify/goose/v2/statsd"
)

// Note that statsd's default backend is the nullBackend, which doesn't do anything.
// For these metrics to work, use statsd.SetBackend.
var (
	GenMainRun      = &statsd.Timer{Name: "genmain.run"}
	GenMainShutdown = &statsd.Timer{Name: "genmain.shutdown"}

	HTTPRequest = &statsd.Timer{Name: "http.request"}

	ShellCommandRun = &statsd.Timer{Name: "shell.command.run"}
)
