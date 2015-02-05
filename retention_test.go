package collector

import (
	"container/list"
	"testing"
	"time"
)

func intCollectorTestRunner(ic *IntCollector) int64 {
	return time.Now().Unix()
}

func TestRetentionCleanerCleanupPartial(t *testing.T) {
	t.Parallel()

	var l = list.New()
	var c = &IntCollector{intCollectorTestRunner}
	l.PushBack(c.Collect())
	time.Sleep(time.Millisecond)
	l.PushBack(c.Collect())
	time.Sleep(time.Millisecond)

	RetentionCleaner{time.Millisecond * 2}.Cleanup(l)
	if l.Len() != 1 {
		t.Fatalf("Expected RetentionCleaner to remove 1 item, removed %d", 2-l.Len())
	}
}

func TestRetentionCleanerCleanupComplete(t *testing.T) {
	t.Parallel()

	var l = list.New()
	var c = &IntCollector{intCollectorTestRunner}
	l.PushBack(c.Collect())
	l.PushBack(c.Collect())
	time.Sleep(time.Millisecond * 2)

	RetentionCleaner{time.Millisecond}.Cleanup(l)
	if l.Len() != 0 {
		t.Fatalf("Expected RetentionCleaner to remove 2 item, removed %d", 2-l.Len())
	}
}
