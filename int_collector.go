package collector

import "time"

// Int64MaxSampler reduces a slice of int64 to their maximum
type Int64MaxSampler struct{}

func (md Int64MaxSampler) Reduce(ms ...interface{}) interface{} {
	var max int64
	for _, m := range ms {
		var c = m.(int64)
		if c > max {
			max = c
		}
	}
	return max
}

// IntCollector collects int64s
type IntCollector struct {
	Run func(*IntCollector) int64
}

// Collect runs inside a go routine
func (ic *IntCollector) Collect() Metric {
	return Metric{
		GeneratedAt: time.Now(),
		Value:       ic.Run(ic),
	}
}
