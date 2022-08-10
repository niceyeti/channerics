// Copyright 2022 Jesse Waite

// sequences.go contains patterns related to sequence generation, piping, and injection.

package channerics

// Generator streams values via the passed generator until it returns false.
// The closure of done is only detected between calls to generate.
func Generator[T any](
	done <-chan struct{},
	generate func() (T, bool),
) <-chan T {
	ch := make(chan T)

	// It turns out that synchronizing a synchronous func is difficult,
	// except by brute force. This code is verbose because it must give
	// highest precedence to 'done':
	// 0) output no further items after done is closed
	// 1) ensure generate is not called if done is closed
	// 2) check if done afterward, before sending
	// 3) send or done
	go func() {
		defer close(ch)
		for {
			// Since the call to generate is synchronous, we must check if done both before and after.
			select {
			case <-done:
				return
			default:
			}
			val, ok := generate()
			if !ok {
				return
			}
			// Checking done here gives it higher precedence than sending.
			// Otherwise, the sending-select statement will randomly honor done,
			// when done is already closed.
			select {
			case <-done:
				return
			default:
			}

			select {
			case ch <- val:
			case <-done:
				return
			}
		}
	}()

	return ch
}

// Repeater loops over the passed slice.
func Repeater[T any](
	done <-chan struct{},
	seq []T,
) <-chan T {
	ch := make(chan T)

	go func() {
		defer close(ch)

		for {
			for _, val := range seq {
				select {
				case <-done:
					return
				default:
				}

				select {
				case ch <- val:
				case <-done:
					return
				}
			}
		}
	}()

	return ch
}

// Tee streams input values to both returned output channels.
// Output must be read from both channels before subsequent values are sent.
func Tee[T any](
	done <-chan struct{},
	in <-chan T,
) (<-chan T, <-chan T) {
	out1 := make(chan T)
	out2 := make(chan T)

	go func() {
		defer close(out1)
		defer close(out2)

		for v := range OrDone(done, in) {
			var out1, out2 = out1, out2 // intentional shadowing
			for i := 0; i < 2; i++ {
				select {
				case out1 <- v:
					out1 = nil
				case out2 <- v:
					out2 = nil
				case <-done:
					return
				}
			}
		}
	}()

	return out1, out2
}
