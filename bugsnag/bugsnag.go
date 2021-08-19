package bugsnag

import (
	bugsnaggo "github.com/bugsnag/bugsnag-go/v2"
)

var (
	def = newBugsnagger(&bugsnaggo.Config, bugsnaggo.Notify)
)

// TabWriter defines the interface to be implemented by the different client objects that want custom decorated bugsnag tabs
type TabWriter interface {
	CreateBugsnagTab() Tab
}

// These define the static APIs which can be accessed directly from the bugsnag package

func Notify(err error, rawData ...interface{}) {
	def.Notify(err, rawData...)
}

func AutoNotify(rawData ...interface{}) {
	def.AutoNotify(rawData...)
}

func AutoRecover(rawData ...interface{}) {
	def.AutoRecover(rawData...)
}

func Setup(apiKey string, commit string, env string, packages []string) {
	def.Setup(apiKey, commit, env, packages)
}

func Configured() bool {
	return bugsnaggo.Config.APIKey != ""
}
