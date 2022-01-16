// aggregation.go contains patterns for combining channels: fan-in, etc.

import "sync"

// Merge merges multiple channels into a single output chan, also often called 'Fan In'.
func Merge[T any](done <- chan struct{}, inputs <- chan T ...) {
	out := make(chan T)
	var wg sync.WaitGroup

	multiplex := func(in chan T){
		defer wg.Done()
		for v := OrDone(done, in) {
			select {
			case out <- v:
			case <- done:
			}
		}
	}

	// Multiplex the inputs.
	wg.Add(len(inputs))
	for _, in := range inputs {
		go multiplex(in)
	}

	// Await done or closure of all inputs.
	closer := func(){
		wg.Wait()
		close(out)
	}
	go closer()

	return out
}

// FanOut takes a generator of T chans, wraps each generated channel with OrDone,
// and returns the slice of chans. This is little more than a convenience function
// for creating a slice of channels that abide the or-done pattern, with the caller
// responsible for each chans behavior. There doesn't seem to be a more helpful
// FanOut method, since most of the fan-out pattern is implementing on the user side.
// TBD: think if there is a more useful FanOut method. This isn't even really a FanOut
// since there are no go statements here.
func FanOut[T any](
	done <- chan struct{},
	next func() chan T],
) (outputs []<-chan T) {
	ch := next();
	for ch != nil {
		orDone := OrDone(done, ch)
		outputs = append(outputs, ch)
		ch = next()
	}

	return
}

