package collector

import (
	"container/list"
	"time"
)

// RetentionCleaner takes care of removing metrics which have exeeded their lifetime
type RetentionCleaner struct {
	Interval time.Duration
}

// Cleanup checks the collectors metrics, dropping everythings from the front older than the given interval
func (rc RetentionCleaner) Cleanup(l *list.List) {
	if l.Len() == 0 {
		return
	}

	// drop entries older than retention policy
	var refDate = time.Now().Add(rc.Interval * -1)
	for {
		if l.Len() == 0 {
			break
		}

		im, _ := l.Front().Value.(Metric)
		if im.GeneratedAt.Before(refDate) {
			l.Remove(l.Front())
		} else {
			break
		}
	}
}
