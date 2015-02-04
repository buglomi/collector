package collector

import "time"

// IntMetric is just a arbitrary int metric storage
type IntMetric struct {
	Time  time.Time
	Value int64
}

// GeneratedAt returns the Metrics generation timestamp
func (im IntMetric) GeneratedAt() time.Time {
	return im.Time
}

// IntCollector collects ints
type IntCollector struct {
	Run func(*IntCollector) int64
}

func NewIntCollector(runner func(*IntCollector) int64) *IntCollector {
	return &IntCollector{
		Run: runner,
	}
}

// Collect runs inside a go routine
func (ic *IntCollector) Collect() Metric {
	return IntMetric{
		Time:  time.Now(),
		Value: ic.Run(ic),
	}
}
