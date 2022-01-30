// Copyright 2022 Jesse Waite

// completion.go contains simple patterns related to completion: context, done, and similar patterns.

package channerics

// OrDone streams values from vals until done or vals is closed.
// OrDone is the counterpart to Send which does the same for a writable channel.
// Note the done-guard below, which is predominately for buffered channels;
// a done-guard ensures that done's closure has higher precedence than the vals channel.
func OrDone[T any](
	done <-chan struct{},
	vals <-chan T,
) <-chan T {
	output := make(chan T)

	go func() {
		defer close(output)
		for {
			// Done-guard: this check's 'done' with higher precedence, whereas the
			// select statement below only probabilistically honors done's closure,
			// for example, when vals is a buffered channel with several queued items.
			select {
			case <-done:
				return
			default:
			}

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

// Any returns a single channel that closes when any passed channel is closed.
// Any blocks forever if chans is empty.
// The passed channels' sole purpose should be to communicate closure, not transmit data.
func Any[T any](
	chans ...<-chan T,
) <-chan T {
	switch len(chans) {
	case 0:
		// The recursive base case; not a plausible user call.
		return nil
	case 1:
		return chans[0]
	case 2:
		return eitherDone(chans[0], chans[1])
	}

	done := make(chan T)
	go func() {
		defer close(done)
		select {
		case <-chans[0]:
		case <-chans[1]:
		case <-chans[2]:
		case <-Any(chans[3:]...):
		}
	}()

	return done
}

func eitherDone[T any](
	ch1, ch2 <-chan T,
) <-chan T {
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

// All returns a channel that closes when all the passed channels are closed.
// All waits forever if any channel is nil, and immediately closes if chans is empty.
// The passed channels' sole purpose should be to communicate closure, not transmit data.
func All[T any](
	done chan struct{},
	chans ...<-chan T,
) <-chan T {
	allDone := make(chan T)

	go func() {
		defer close(allDone)
		for _, ch := range chans {
			ok := true
			for ok {
				select {
				case _, ok = <-ch:
				case <-done:
					return
				}
			}
		}
	}()

	return allDone
}
