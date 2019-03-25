package bugsnag

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	bugsnaggo "github.com/bugsnag/bugsnag-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
)

var (
	mSnagger *bugsnagger
)

func init() {
	mNotifier := func(err error, rawData ...interface{}) error {
		panic("Notifier function is called")
	}

	mConfig := &bugsnaggo.Configuration{APIKey: "key"}
	mSnagger = NewBugsnagger(mConfig, mNotifier)
}

func TestSetup(t *testing.T) {
	var (
		apiKey = "apiTest"
		commit = "commitSHA"
		env    = "envTest"
		pack   = "packageTest"
	)

	Setup(apiKey, commit, env, []string{pack})
	require.Equal(t, apiKey, bugsnaggo.Config.APIKey)
	require.Equal(t, commit, bugsnaggo.Config.AppVersion)
	require.Equal(t, env, bugsnaggo.Config.ReleaseStage)
	require.Equal(t, pack, bugsnaggo.Config.ProjectPackages[0])
	require.Equal(t, "main*", bugsnaggo.Config.ProjectPackages[1])
	require.Equal(t, "github.com/Shopify/goose/bugsnag", bugsnaggo.Config.ProjectPackages[2])
	require.True(t, strings.Contains(bugsnaggo.Config.ProjectPackages[3], "github.com/Shopify/goose/*"))
	require.True(t, bugsnaggo.Config.Synchronous)
}

func TestNotify(t *testing.T) {
	require.NotPanics(t, func() { mSnagger.Notify(nil) })
	require.Panics(t, func() { mSnagger.Notify(errors.New("err")) })
}

func TestBuildDataError_cause(t *testing.T) {
	dataList, err := mSnagger.buildData(errors.New("err"))
	require.Equal(t, "err", err.Error())
	require.Equal(t, "err", newBugsnagData(dataList).getBugsnaggoErrorClass())
	require.True(t, newBugsnagData(dataList).hasErrTab())
	require.Equal(t, "", newBugsnagData(dataList).getContext())
}

func TestBuildDataError_wrappedCause(t *testing.T) {
	err := errors.Wrap(errors.Wrap(errors.New("err"), "msg1"), "msg2")
	dataList, err := mSnagger.buildData(err)
	require.Equal(t, "err", err.Error())
	require.Equal(t, "err", newBugsnagData(dataList).getBugsnaggoErrorClass())
	require.True(t, newBugsnagData(dataList).hasErrTab())
	require.Equal(t, "msg2: msg1", newBugsnagData(dataList).getContext())
}

func TestBuildDataError_urlError(t *testing.T) {
	// url error
	err := errors.Wrap(&url.Error{Op: "GET", URL: "lol.com", Err: errors.New("err")}, "url fail")
	dataList, err := mSnagger.buildData(err)
	require.Equal(t, "err", err.Error())
	require.Equal(t, "GET lol.com: err", newBugsnagData(dataList).getBugsnaggoErrorClass())
	require.True(t, newBugsnagData(dataList).hasErrTab())
	require.Equal(t, "url fail: GET lol.com", newBugsnagData(dataList).getContext())
}

func TestBuildDataError_http2Stream(t *testing.T) {
	// http2 stream error with code
	err := errors.Wrap(&http2.StreamError{Code: 1}, "http2 fail")
	dataList, err := mSnagger.buildData(err)
	require.Equal(t, "stream error: PROTOCOL_ERROR", err.Error())
	require.Equal(t, "stream error: stream ID 0; PROTOCOL_ERROR", newBugsnagData(dataList).getBugsnaggoErrorClass())
	require.True(t, newBugsnagData(dataList).hasErrTab())
	require.Equal(t, "http2 fail: stream error", newBugsnagData(dataList).getContext())
}

func TestBuildDataString(t *testing.T) {
	dataList, err := mSnagger.buildData(errors.New("err"), "database connection")
	require.Equal(t, "err", err.Error())
	require.True(t, newBugsnagData(dataList).hasErrTab())
	require.Equal(t, "database connection", newBugsnagData(dataList).getContext())
}

func TestBuildDataHttpRequest(t *testing.T) {
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: "lol.com",
		},
	}
	dataList, err := mSnagger.buildData(errors.New("err"), req)
	require.Equal(t, "err", err.Error())
	require.Equal(t, "lol.com", newBugsnagData(dataList).getContext())
}

func TestBuildDataLogEntry(t *testing.T) {
	entry := log.NewEntry(&log.Logger{})
	entry = entry.WithFields(log.Fields{
		"key":    "val",
		"shopID": 12345,
	})
	entry.Message = "should be ignored"
	err := errors.Wrap(errors.New("inner error"), "err error")
	dataList, err := mSnagger.buildData(err, entry)
	require.Equal(t, "inner error", err.Error())
	require.Equal(t, "inner error", newBugsnagData(dataList).getBugsnaggoErrorClass())
	require.Equal(t, "err error", newBugsnagData(dataList).getContext())
	require.Equal(t, Rows{"key": "val", "shopID": 12345}, newBugsnagData(dataList).getLogTab())
	require.Equal(t, "12345", newBugsnagData(dataList).getUser())
}

func TestBuildDataLogFields(t *testing.T) {
	dataList, err := mSnagger.buildData(errors.New("err"), log.Fields{
		"key":    "val",
		"shopID": 12345,
	})
	require.Equal(t, "err", err.Error())
	require.Equal(t, "err", newBugsnagData(dataList).getBugsnaggoErrorClass())
	require.Equal(t, "", newBugsnagData(dataList).getContext())
	require.Equal(t, Rows{"key": "val", "shopID": 12345}, newBugsnagData(dataList).getLogTab())
	require.Equal(t, "12345", newBugsnagData(dataList).getUser())
}

func TestBuildDataTab(t *testing.T) {
	tab := Tab{
		Label: "test",
		Rows: map[string]interface{}{
			"err": int(1),
		},
	}
	dataList, err := mSnagger.buildData(errors.New("err"), tab)
	require.Equal(t, "err", err.Error())
	require.Equal(t, Rows{"err": 1}, newBugsnagData(dataList).getTab("test"))
}

type customTab struct {
	val string
}

func (ct *customTab) CreateBugsnagTab() Tab {
	return Tab{
		Label: "test",
		Rows:  Rows{"key": ct.val},
	}
}

func TestBuildDataTabWriter(t *testing.T) {
	dataList, err := mSnagger.buildData(errors.New("err"), &customTab{"val"})
	require.Equal(t, "err", err.Error())
	require.Equal(t, Rows{"key": "val"}, newBugsnagData(dataList).getTab("test"))
}

func TestBuildDataRegularData(t *testing.T) {
	dataList, err := mSnagger.buildData(errors.New("err"), 100)
	require.Equal(t, "err", err.Error())
	require.True(t, newBugsnagData(dataList).hasValue(100))
}

func TestBuildDataCustomErrorClass(t *testing.T) {
	cause := errors.New("err")
	err := WithErrorClass(cause, "CUSTOM")
	dataList, returnedCause := mSnagger.buildData(err)
	require.Equal(t, cause, returnedCause)
	require.Equal(t, "CUSTOM", newBugsnagData(dataList).getBugsnaggoErrorClass())
}

func TestAutoRecover(t *testing.T) {
	done := make(chan interface{})
	// when autoRecover is called inside a go-routine, we should no longer receive a panic
	// in the main process.
	notifier := func(err error, rawData ...interface{}) (e error) {
		defer func() {
			close(done)
		}()
		require.Equal(t, "oops (string)", err.Error())
		require.True(t, newBugsnagData(rawData).hasErrTab())
		return nil
	}

	errConfig := &bugsnaggo.Configuration{APIKey: "errAPI"}
	snagger := NewBugsnagger(errConfig, notifier)

	go func() {
		defer snagger.AutoRecover()
		panic("oops")
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("test timed out")
	}
}
