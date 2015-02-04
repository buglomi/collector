// Package collector contains utilities to collect arbitrary historic metrics.
//
// Metrics are always stored in memory in a double linked list and discarded once
// the retention interval of its agent is reached.
package collector

import (
	"sync"
	"time"
)
import "container/list"

// Metric defines our minimum requirement for every metric: it must have a timestamp
// The timestamp is used by RetentionCleaner to drop outdated metrics
type Metric interface {
	GeneratedAt() time.Time
}

// Collector is a generic interface for anything which collects metrics. Metrics can be anything,
// the collector only contains methods necessary for scheduling as well as retention cleanup and compaction
type Collector interface {
	// Collect is executed by the agent. It's job is to add a new metric to the metrics list
	Collect() Metric
}

type MetricStorage struct {
	sync.Mutex
	*list.List
}

func NewMetricStorage() *MetricStorage {
	return &MetricStorage{
		Mutex: sync.Mutex{},
		List:  list.New(),
	}
}

// Agent executes arbitrary collectors and enforces the retention policy
type Agent struct {
	// Interval defines the duration between metric collections. Applies to all registered Collectors
	Interval time.Duration
	// RetentionInterval defines how long metrics are kept until they are discarded.
	RetentionInterval time.Duration
	// Collectors are executed every interval. All they do is produce metrics
	Collectors map[string]Collector
	// Metrics contain all generated metrics, by collector name
	Metrics          map[string]*list.List
	locks            map[string]sync.Mutex
	retentionCleaner RetentionCleaner
}

func NewAgent(interval, retentionInterval time.Duration) *Agent {
	return &Agent{
		Interval:          interval,
		RetentionInterval: retentionInterval,
		Collectors:        map[string]Collector{},
		Metrics:           map[string]*list.List{},
		locks:             map[string]sync.Mutex{},
		retentionCleaner:  RetentionCleaner{rt},
	}
}

func (a *Agent) Add(name string, c Collector) {
	a.Collectors[name] = c
	a.Metrics[name] = list.New()
	a.locks[name] = sync.Mutex{}
}

func (a *Agent) process(name string) {
	var m = a.locks[name]
	m.Lock()

	a.Metrics[name].PushBack(a.Collectors[name].Collect())
	a.retentionCleaner.Cleanup(a.Metrics[name])

	m.Unlock()
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
