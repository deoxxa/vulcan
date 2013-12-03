/*
Memory backend, used mostly for testing, but may be extended to become
more useful in the future. In this case it'll need garbage collection.
*/
package backend

import (
	"sync"
	"time"
	"github.com/golang/glog"
	"github.com/mailgun/vulcan/timeutils"
	"github.com/mailgun/minheap"
)

type MemoryBackend struct {
	Hits         map[string]int64
	TimeProvider timeutils.TimeProvider
	expiryTimes  *minheap.MinHeap
	mutex        *sync.Mutex
}

func NewMemoryBackend(timeProvider timeutils.TimeProvider) (*MemoryBackend, error) {
	return &MemoryBackend{
		Hits:         map[string]int64{},
		TimeProvider: timeProvider,
		expiryTimes:  minheap.NewMinHeap(),
		mutex:        &sync.Mutex{},
	}, nil
}

func (b *MemoryBackend) GetCount(key string, period time.Duration) (int64, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.Hits[timeutils.GetHit(b.UtcNow(), key, period)], nil
}

func (b *MemoryBackend) UpdateCount(key string, period time.Duration, increment int64) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

    b.deleteExpiredHits(int(b.UtcNow().Unix()))

    // get hit key
    hit := timeutils.GetHit(b.UtcNow(), key, period)

    // track expiration time
    expiryTime := &minheap.Element{
        Value:    hit,
        Priority: int(b.UtcNow().Add(period).Unix()),
    }
    b.expiryTimes.PushEl(expiryTime)

	b.Hits[hit] += increment

	return nil
}

func (b *MemoryBackend) UtcNow() time.Time {
	return b.TimeProvider.UtcNow()
}


const MaxGcLoopIterations = 100

// log entries
const LogNothingToExpire = "MemB GC: Nothing to expire, earliest is: %s updated=%d, now is %d"
const LogHitHasExpired = "MemB GC: %s lastAccess=%d has expired at %d"

func (b *MemoryBackend) deleteExpiredHits(now int) {
	for i := 0; i < MaxGcLoopIterations; i += 1 {
		if b.expiryTimes.Len() == 0 {
			break
		}
        expiryTime := b.expiryTimes.PeekEl()

        if expiryTime.Priority > now {
			glog.Infof(LogNothingToExpire, expiryTime.Value, expiryTime.Priority, now)
			break
        } else {
            expiryTime := b.expiryTimes.PopEl()
			glog.Infof(LogHitHasExpired, expiryTime.Value, expiryTime.Priority, now)
            delete(b.Hits, expiryTime.Value.(string))
        }
    }
}
