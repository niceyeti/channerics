// Copyright 2022 Jesse Waite

// completion.go contains simple patterns related to completion: context, done, and similar patterns.

package channerics

// OrDone streams values from vals until done or vals is closed.
// OrDone is the counterpart to Send which does the same for a writable channel.
func OrDone[T any](
	done <-chan struct{},
	vals <-chan T,
) <-chan T {
	output := make(chan T)

	go func() {
		defer close(output)
		for {
			select {
			case v, ok := <-vals:
				if !ok {
					return
				}
				select {
				case output <- v:
				case <-done:
					return
				}
			case <-done:
				return
			}
		}
	}()

	return output
}

// Any returns a single channel that is closed when any passed channel is closed.
func Any[T any](chans ...<-chan T) (done <-chan T) {
	switch len(chans) {
	case 0:
		// 0 is the recursive base case, not a plausible user call.
		done = nil
	case 1:
		done = chans[0]
	case 2:
		done = eitherDone(chans[0], chans[1])
	case 3:
		prefix := eitherDone(chans[0], chans[1])
		remainder := Any(chans[2:]...)
		done = eitherDone(prefix, remainder)
	}
	return
}

func eitherDone[T any](ch1, ch2 <-chan T) <-chan T {
	done := make(chan T)

	go func() {
		defer close(done)
		select {
		case <-ch1:
		case <-ch2:
		}
	}()

	return done
}

// All returns a channel that is closed when all the passed channels are closed.
func All[T any](chans ...<-chan T) (done <-chan T) {
	switch len(chans) {
	case 0:
		// 0 is the recursive base case, not a plausible user call.
		done = nil
	case 1:
		done = chans[0]
	case 2:
		done = eitherDone(chans[0], chans[1])
	case 3:
		prefix := eitherDone(chans[0], chans[1])
		remainder := Any(chans[2:]...)
		done = eitherDone(prefix, remainder)
	}
	return
}
