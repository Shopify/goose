// Package logrussplunk allows hooking into logrus such that all logs will go
// to splunk via the Splunk HTTP Event Collector (HEC).
package logrussplunk

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/tomb.v2"

	"github.com/Shopify/goose/logger"
)

type jsonMap map[string]interface{}

// Hook can be registered as a `logrus.Hook`.
type Hook struct {
	client *http.Client
	config Configuration
	events chan jsonMap
	tomb   tomb.Tomb
}

const eventsChannelBufferSize = 1000

var log = logger.New("logrussplunk")

// ErrMissingSplunkHost is returned by `NewSplunkHook` if the given configuration
// has no splunk host
var ErrMissingSplunkHost = errors.New("missing splunk host")

// ErrEventsBufferOverflow is returned by `Fire` if the we log and event but have
// yet to process all previous events.
var ErrEventsBufferOverflow = errors.New("events buffer overflow")

// ErrSplunkSendFailed indicates that the hook failed to submit an error to Splunk.
type ErrSplunkSendFailed struct {
	err error
}

func (e ErrSplunkSendFailed) Error() string {
	return "failed to send error to splunk: " + e.err.Error()
}

// closingByteBuffer wraps a `bytes.Buffer` with a `Close` operation so it can be
// used as an `io.ReadCloser` (for the body of an HTTP request)
type closingByteBuffer struct {
	bytes.Buffer
}

func (cb *closingByteBuffer) Close() (err error) {
	return nil
}

// NewSplunkHook initializes a logrus hook which sends logs to the configured
// Splunk HEC. The returned object should be registered with a log via `AddHook()`
func NewSplunkHook(config Configuration) (*Hook, error) {
	if config.SplunkHECHost == "" {
		return nil, ErrMissingSplunkHost
	}

	if config.HostName == "" {
		var err error
		if config.HostName, err = os.Hostname(); err != nil {
			return nil, err
		}
	}

	hook := &Hook{
		client: &http.Client{},
		config: config,
		events: make(chan jsonMap, eventsChannelBufferSize),
	}
	return hook, nil
}

// Tomb implements the `genmain.Component` interface
func (hook *Hook) Tomb() *tomb.Tomb {
	return &hook.tomb
}

// Fire forwards a log entry to Bugsnag.
func (hook *Hook) Fire(entry *logrus.Entry) error {
	event := jsonMap{
		"time":       entry.Time.Unix(),
		"host":       hook.config.HostName,
		"sourcetype": "logrus",
		"source":     "logrus",
		"event": jsonMap{
			"tool":    hook.config.ToolName,
			"message": entry.Message,
			"level":   entry.Level.String(),
			"fields":  entry.Data,
		},
	}

	select {
	case hook.events <- event:
		return nil
	default:
		return ErrEventsBufferOverflow
	}
}

// Run is part of the `genmain.Component` interface
func (hook *Hook) Run() error {
	for {
		batch := hook.nextBatch()
		body := closingByteBuffer{}

		for _, event := range batch {
			eventJSON, err := json.Marshal(event)
			if err != nil {
				return err
			}

			body.Write(eventJSON)
			body.WriteByte('\n')
		}

		if err := hook.submitToSplunk(&body); err != nil {
			// Assume a transient error, put events back on channel so they can be sent in
			// the next iteration
			for _, event := range batch {
				select {
				case hook.events <- event:
				default:
				}
			}

			// TODO investigate what errors are not transient and need to be handled
			log(nil, err).Warn("failed to submit log events to splunk")

			// TODO add exponential backoff
		}

		select {
		case <-hook.tomb.Dying():
			return hook.tomb.Err()
		default:
		}
	}
}

func (hook *Hook) submitToSplunk(body *closingByteBuffer) error {
	request, err := http.NewRequest("POST", hook.config.SplunkHECHost, body)
	if err != nil {
		return err
	}

	request.ContentLength = int64(body.Len())
	request.SetBasicAuth(hook.config.BasicAuthUser, hook.config.BasicAuthPass)
	resp, err := hook.client.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return ErrSplunkSendFailed{err: errors.New(resp.Status)}
	}

	return nil
}

func (hook *Hook) nextBatch() []jsonMap {
	var eventsBatch []jsonMap
	for {
		select {
		case event := <-hook.events:
			eventsBatch = append(eventsBatch, event)
		default:
			if len(eventsBatch) == 0 {
				<-time.After(1 * time.Second)
				continue
			}
			return eventsBatch
		}
	}
}

// Levels enumerates the log levels on which this hook will receive `Fire` events from logrus.
func (hook *Hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}
