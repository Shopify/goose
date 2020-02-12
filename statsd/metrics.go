package statsd

// Collector represents a metric that can be collected. It knows about the
// metric name and sampling rate, and supports a Collect method to submit a
// metric to statsd.
type Collector struct {
	Name string
	Rate sampleRate // 0 (default value) is interpreted as 100% (1.0)
}

type sampleRate float64

func (s sampleRate) Rate() float64 {
	if s == 0 {
		return 1.0
	}
	return float64(s)
}
