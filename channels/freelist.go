// freelist.go is for free-list.
// Free-lists are a trivial implementation and may be removed in the future.

package channerics

import "errors"

// FreeList maintains a list of items for reuse as a strategy to reduce memory
// memory allocations. It can be implemented using a simple buffered channel.
// The idea comes from Effective Go, but there may be other sources.
type FreeList[T any] interface {
	// Get an item from the free-list.
	Get() (T, bool)
	// Return an item to the free-list.
	Put(T) bool
}

type freeList[T any] struct {
	ch   chan T
	newt func() T
}

var ErrInvalidSize = errors.New("size must be greater than zero")

// NewFreeList returns a free list of the given size, calling createFn when T's are created.
func NewFreeList[T any](
	size int,
	createFn func() T,
) (FreeList[T], error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}

	return &freeList[T]{
		ch:   make(chan T, size),
		newt: createFn,
	}, nil
}

// Put returns an item to the free-list, if not at capacity.
func (fl *freeList[T]) Put(t T) bool {
	select {
	case fl.ch <- t:
		return true
	default:
		// do nothing, drop t on the floor
	}
	return false
}

// Get returns an item from the free-list or allocates a new one if none are free.
func (fl *freeList[T]) Get() (t T, isNew bool) {
	select {
	case t = <-fl.ch:
	default:
		t = fl.newt()
		isNew = true
	}

	return
}
