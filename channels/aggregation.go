// Copyright 2022 Jesse Waite

// aggregation.go contains patterns for combining channels: fan-in, etc.

package channerics

import "sync"

// Merge merges multiple channels into a single output chan, also often called
// 'Fan In'. The returned channel closes if done is closed; it also closes if
// all inputs are closed and all outputs are drained.
func Merge[T any](
	done <-chan struct{},
	inputs ...<-chan T,
) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup

	multiplex := func(in <-chan T) {
		defer wg.Done()
		for v := range OrDone(done, in) {
			select {
			case out <- v:
			case <-done:
				return
			}
		}
	}

	// Multiplex the inputs.
	wg.Add(len(inputs))
	for _, in := range inputs {
		go multiplex(in)
	}

	// Await done or closure of all inputs.
	closer := func() {
		wg.Wait()
		close(out)
	}
	go closer()

	return out
}

/*
// FanOut takes a T chan generator, wraps each generated channel with OrDone,
// and returns the slice of chans. This is little more than a convenience function
// for creating a slice of channels that abide the or-done pattern, with the caller
// responsible for each chan's behavior. There doesn't seem to be a more helpful
// FanOut method, since most of the fan-out pattern is implemented on the user side.
// TBD: think of a more useful FanOut method. This isn't even really a FanOut
// since there are no go statements below. Write an example.
func FanOut[T any](
	done <-chan struct{},
	next func() chan T,
) (outputs []<-chan T) {
	ch := next()
	for ch != nil {
		orDone := OrDone(done, ch)
		outputs = append(outputs, orDone)
		ch = next()
	}

	return
}
*/
