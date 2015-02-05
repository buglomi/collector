package collector

import (
	"container/list"
	"testing"
	"time"
)

func testCompact(l *list.List) int {
	var eCount = l.Len()
	var c = Compactor{
		HighResolutionInterval: time.Millisecond * 500,
		DownSamplingInterval:   time.Millisecond * 500,
	}
	c.Reduce(l, time.Millisecond*250, Int64MaxSampler{})
	return eCount - l.Len()
}

func TestCompactorReduceEmpty(t *testing.T) {
	if testCompact(list.New()) != 0 {
		t.Fatalf("Expected empty list to stay empty")
	}
}

func TestCompactorReduceAllHighResolution(t *testing.T) {
	var l = list.New()
	l.PushBack(Metric{
		GeneratedAt: time.Date(2015, 02, 02, 11, 59, 59, 750000000, time.UTC),
		Value:       int64(0),
	})
	l.PushBack(Metric{
		GeneratedAt: time.Date(2015, 02, 02, 12, 00, 00, 000000000, time.UTC),
		Value:       int64(0),
	})
	if testCompact(l) != 0 {
		t.Fatalf("Expected high res list to stay untouched")
	}
}

func TestCompactorReduceSamplingGroup(t *testing.T) {
	var l = list.New()
	l.PushBack(Metric{
		GeneratedAt: time.Date(2015, 02, 02, 11, 59, 59, 250000000, time.UTC),
		Value:       int64(1),
	})
	l.PushBack(Metric{
		GeneratedAt: time.Date(2015, 02, 02, 11, 59, 59, 500000000, time.UTC),
		Value:       int64(2),
	})
	l.PushBack(Metric{
		GeneratedAt: time.Date(2015, 02, 02, 11, 59, 59, 750000000, time.UTC),
		Value:       int64(3),
	})
	l.PushBack(Metric{
		GeneratedAt: time.Date(2015, 02, 02, 12, 00, 00, 000000000, time.UTC),
		Value:       int64(4),
	})

	if x := testCompact(l); x != 1 {
		t.Fatalf("Expected single compaction, got: %d", x)
	}

	if l.Front().Value.(Metric).Value.(int64) != 2 {
		t.Fatalf("Wrong downsampling")
	}
}

func TestCompactorReduceSamplingGroupReSample(t *testing.T) {
	var l = list.New()
	l.PushBack(Metric{
		GeneratedAt: time.Date(2015, 02, 02, 11, 59, 59, 250000000, time.UTC),
		Value:       int64(1),
		downsampled: true,
	})
	l.PushBack(Metric{
		GeneratedAt: time.Date(2015, 02, 02, 11, 59, 59, 500000000, time.UTC),
		Value:       int64(2),
	})
	l.PushBack(Metric{
		GeneratedAt: time.Date(2015, 02, 02, 11, 59, 59, 750000000, time.UTC),
		Value:       int64(3),
	})
	l.PushBack(Metric{
		GeneratedAt: time.Date(2015, 02, 02, 12, 00, 00, 000000000, time.UTC),
		Value:       int64(4),
	})

	if x := testCompact(l); x != 0 {
		t.Fatalf("Expected no compaction, got: %d", x)
	}
}
