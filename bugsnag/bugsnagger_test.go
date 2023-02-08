package bugsnag

import (
	"flag"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	bugsnaggo "github.com/bugsnag/bugsnag-go/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
)

var (
	mSnagger   *bugsnagger
	projectDir string
)

func init() {
	mNotifier := func(err error, rawData ...interface{}) error {
		panic("Notifier function is called")
	}

	mConfig := &bugsnaggo.Configuration{APIKey: "key"}
	mSnagger = newBugsnagger(mConfig, mNotifier)

	_, file, _, _ := runtime.Caller(0)
	if !strings.HasSuffix(file, "goose/bugsnag/bugsnagger_test.go") {
		panic("unable to determine project dir")
	}
	projectDir = path.Dir(path.Dir(file))
}

func TestMain(m *testing.M) {
	if !flag.Parsed() {
		flag.Parse()
	}

	// This test is incompatible with paniconexit0
	panicFlag := flag.CommandLine.Lookup("test.paniconexit0")
	if panicFlag != nil {
		err := panicFlag.Value.Set("false")
		if err != nil {
			panic(err)
		}
	}

	code := m.Run()
	os.Exit(code)
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
	require.Equal(t, "github.com/Shopify/goose/v2/bugsnag", bugsnaggo.Config.ProjectPackages[2])
	require.True(t, strings.Contains(bugsnaggo.Config.ProjectPackages[3], projectDir+"/*"))
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
	urlError := &url.Error{Op: "GET", URL: "lol.com", Err: errors.New("err")}
	err := errors.Wrap(urlError, "url fail")
	dataList, err := mSnagger.buildData(err)
	require.Equal(t, "err", err.Error())
	require.Equal(t, urlError.Error(), newBugsnagData(dataList).getBugsnaggoErrorClass())
	require.True(t, newBugsnagData(dataList).hasErrTab())
	require.Equal(t, "url fail: "+urlError.Error(), newBugsnagData(dataList).getContext()+": err")
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
		Label: "custom",
		Rows:  Rows{"key": ct.val},
	}
}

func TestBuildDataTabWriter(t *testing.T) {
	dataList, err := mSnagger.buildData(errors.New("err"), &customTab{"val"})
	require.Equal(t, "err", err.Error())
	require.Equal(t, Rows{"key": "val"}, newBugsnagData(dataList).getTab("custom"))
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
	snagger := newBugsnagger(errConfig, notifier)

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
