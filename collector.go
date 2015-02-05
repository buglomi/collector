// Package collector contains utilities to collect arbitrary historic metrics.
//
// Metrics are always stored in memory in a double linked list and discarded once
// the retention interval of its agent is reached.
package collector

import (
	"container/list"
	"sync"
	"time"
)

// Metric contain an arbitrary value and a timestamp
type Metric struct {
	GeneratedAt time.Time
	Value       interface{}
	downsampled bool
}

// Collector is a generic interface for anything which collects metrics. Metrics can be anything,
// the collector only contains methods necessary for scheduling as well as retention cleanup and compaction
type Collector interface {
	// Collect is executed by the agent. It's job is to add a new metric to the metrics list
	Collect() Metric
}

// DownSampler reduces many metrics into a single metric
type DownSampler interface {
	Reduce(...interface{}) interface{}
}

// MetricStorage stores metrics in a double linked list to support downsampling and easy removal
type MetricStorage struct {
	mu          *sync.Mutex
	downSampler DownSampler
	*list.List
}

func newMetricStorage(d DownSampler) *MetricStorage {
	return &MetricStorage{
		mu:          &sync.Mutex{},
		List:        list.New(),
		downSampler: d,
	}
}

// Agent executes arbitrary collectors and enforces the retention policy
type Agent struct {
	// Interval defines the duration between metric collections. Applies to all registered Collectors
	Interval time.Duration
	// RetentionInterval defines how long metrics are kept until they are discarded. Applies to all registered Collectors
	RetentionInterval time.Duration
	// Collectors are executed every interval. All they do is produce metrics
	Collectors map[string]Collector
	// Metrics contain all generated metrics, by collector name
	Metrics          map[string]*MetricStorage
	retentionCleaner RetentionCleaner
	Compactor
}

// NewAgent initializes a new Agent with given interval & retentionInterval
func NewAgent(interval, retentionInterval time.Duration) *Agent {
	return &Agent{
		Interval:          interval,
		RetentionInterval: retentionInterval,
		Collectors:        map[string]Collector{},
		Metrics:           map[string]*MetricStorage{},
		retentionCleaner:  RetentionCleaner{retentionInterval},
		Compactor: Compactor{
			HighResolutionInterval: 0,
			DownSamplingInterval:   0,
		},
	}
}

// Add adds a new collector to the agent, and initializes all dependencies
func (a *Agent) Add(name string, c Collector, d DownSampler) {
	a.Collectors[name] = c
	a.Metrics[name] = newMetricStorage(d)
}

func (a *Agent) process(name string) {
	var s = a.Metrics[name]
	s.mu.Lock()

	s.PushBack(a.Collectors[name].Collect())
	a.retentionCleaner.Cleanup(s.List)

	if s.downSampler != nil {
		a.Compactor.Reduce(s.List, a.Interval, s.downSampler)
	}

	s.mu.Unlock()
}

// Run executes all registered collectors every Interval, inside their own go routine
func (a *Agent) Run() {
	for {
		select {
		case <-time.After(a.Interval):
			for name := range a.Collectors {
				go a.process(name)
			}
		}
	}
}
