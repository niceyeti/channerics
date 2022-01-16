// Copyright 2022 Jesse Waite

// sequences.go contains patterns related to sequence generation, piping, and injection.

package channerics

// Repeater streams values from the passed generator.
func Generator[T any](
	done <-chan struct{},
	generate func() T,
) <-chan T {
	ch := make(chan T)

	go func() {
		defer close(ch)

		for {
			select {
			case ch <- generate():
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

		for _, val := range seq {
			select {
			case ch <- val:
			case <-done:
				return
			}
		}
	}()

	return ch
}

// Tee streams input values to both returned output channels.
// Output must be read from both channels before the next value is sent.
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
			var out1, out2 = out1, out2
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
