package configure

var Configuration *Config

func Configure(config *Config) {
	Configuration = config

	Logger(config)
	Bugsnag(config)
	Metrics(config)
	Profiling(config)
}
