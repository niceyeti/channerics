// semaphore.go is for a counting semaphore.
// Semaphore, like FreeList, is a slightly higher level implementation than native channerics
// and may represent library bloat, so it may be removed in the future.

package channerics

import (
	"context"
	"sync"
)

// Semaphore represents semaphore turnstile semantics implemented using a buffered channel.
type Semaphore struct {
	ch          chan struct{}
	safe_closer sync.Once
}

// NewSemaphore returns a context-sensitive semaphore built on a buffered chan.
// If n == 0, use sync.Mutex or sync.RWMutex instead.
// Also note that an idiomatic local semaphore can be written simply:
//	go func() {
//		sem := make(chan struct{}, 10)
//		defer close(sem)
//		for {
//			data <- workChan
//			select {
//			case ch <- struct{}:
//			case <-context.Done():
//				return
//			}
//			... do work
//			work(data)
//			<-ch
//		}
//	}
//

func NewSemaphore(n int) *Semaphore {
	return &Semaphore{
		ch:          make(chan struct{}, n),
		safe_closer: sync.Once{},
	}
}

// Take 'decrements' the semaphore, blocking until the semaphore is available.
// Taken indicates that take succeeded without context cancellation; if true,
// then Release must be called later, e.g. using 'defer sem.Release()'.
func (sem *Semaphore) Take(ctx context.Context) (taken bool) {
	select {
	case sem.ch <- struct{}{}:
		taken = true
	case <-ctx.Done():
	}

	return
}

// UnsafeCount returns an immediately expired view of the semaphore count,
// since the true count may change immediately afterward at any point.
// Thus the only use-cases for this function are (1) determining when the semaphore
// is 'probably' free, but possibly blocking is still permissible, or (2) you coordinate
// your go-routines using another mechanism to not alter it until you've use it.
func (sem *Semaphore) UnsafeCount() int {
	return cap(sem.ch) - len(sem.ch)
}

// MaybeTake attempts to take (decrement) if it would not block, otherwise returns immediately.
// MaybeTake does not take a context since it always returns immediately.
// Taken indicates that take succeeded without context cancellation; if true,
// then Release must be called later, e.g. using 'defer sem.Release()'.
func (sem *Semaphore) MaybeTake() (taken bool) {
	select {
	case sem.ch <- struct{}{}:
		taken = true
	default:
	}

	return
}

// Release increments the semaphore without blocking, and must be called after calling Take(), its counterpart.
func (sem *Semaphore) Release() {
	select {
	case <-sem.ch:
	default:
	}
	return
}

// Close closes the semaphore's channel, after which other methods will panic if called.
// Close may be called repeatedly from multiple go routines.
func (sem *Semaphore) Close() {
	sem.safe_closer.Do(func() { close(sem.ch) })
	return
}
