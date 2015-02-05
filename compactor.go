package collector

import (
	"container/list"
	"time"
)

// Compactor identifies elements after the high resolution interval and samples the values down
type Compactor struct {
	HighResolutionInterval time.Duration
	DownSamplingInterval   time.Duration
}

// findDownSamplingStart returns first element after HighResolutionInterval
func (c Compactor) findDownSamplingStart(l *list.List) *list.Element {
	var e = l.Back()
	var m = e.Value.(Metric)
	var highResEnd = m.GeneratedAt.Add(c.HighResolutionInterval * -1)

	for {
		m = e.Value.(Metric)
		if m.GeneratedAt.Before(highResEnd) || m.GeneratedAt.Equal(highResEnd) {
			break
		}
		e = e.Prev()
		if e == nil {
			return nil
		}
	}
	return e
}

// findNextSamplingGroup returns the next group of metrics ready for compaction
func (c Compactor) findNextSamplingGroup(l *list.List) []*list.Element {
	var e = l.Back()
	var m = e.Value.(Metric)

	var highResEnd = m.GeneratedAt.Add(c.HighResolutionInterval * -1)
	var sampleEnd = highResEnd.Add(c.DownSamplingInterval * -1)
	e = c.findDownSamplingStart(l)
	if e == nil {
		return nil
	}

	// sampleEnd , ... , downSampleRef
	var group []*list.Element
	for {
		m = e.Value.(Metric)
		if m.downsampled {
			break
		}
		if m.GeneratedAt.Before(sampleEnd) {
			break
		}

		group = append(group, e)
		e = e.Prev()
		if e == nil {
			break
		}
	}
	return group
}

// Reduce replaces many metrics with their downsampled equivalent
func (c Compactor) Reduce(l *list.List, interval time.Duration, sampler DownSampler) {
	if l.Len() == 0 {
		return
	}

	var group = c.findNextSamplingGroup(l)
	var expect = int(c.DownSamplingInterval / interval)
	if len(group) < expect {
		return
	}

	var values = make([]interface{}, len(group))
	for i, e := range group {
		values[i] = e.Value.(Metric).Value
	}

	var m = Metric{
		Value:       sampler.Reduce(values...),
		GeneratedAt: group[len(group)-1].Value.(Metric).GeneratedAt,
		downsampled: true,
	}
	l.InsertAfter(m, group[0])

	for _, e := range group {
		l.Remove(e)
	}
}
