// Copyright 2022 Jesse Waite

package channerics

// AsType takes an channel of interfaces and convert it to any type.
// Reflection is not used: users must ensure the interface is of type T or this will panic.
func AsType[T any](
	done chan struct{},
	vals chan interface{},
) <-chan T {
	ch := make(chan T)

	go func() {
		defer close(ch)

		for v := range vals {
			select {
			case ch <- v.(T):
			case <-done:
			}
		}
	}()

	return ch
}

// Adapter returns a channel of vals converted from vals channel using conversionFn,
// which must be a fast, non-blocking function.
func Adapter[T1 any, T2 any](
	done chan struct{},
	vals <-chan T1,
	convertFn func(T1) T2,
) <-chan T2 {
	out := make(chan T2)

	go func() {
		defer close(out)
		for val := range OrDone(done, vals) {
			select {
			case out <- convertFn(val):
			case <-done:
			}
		}
	}()

	return out
}
