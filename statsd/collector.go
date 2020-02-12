package statsd

// collector represents a metric that can be collected.
type collector struct {
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
