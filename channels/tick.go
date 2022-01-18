// Copyright 2022 Jesse Waite

// tick.go contains patterns for intervals, ticks, time stuff.

package channerics

import "time"

// NewTicker wraps a ticker with done semantics to ensure safe cleanup/cancellation semantics.
// See the time.Tick docs for background; its timer does not safely stop, which leads to many misuse-cases.
// Note that this method introduces some trivial latency between the timer tick and propagation to the output
// output channel.
func NewTicker(
	done <-chan struct{},
	duration time.Duration,
) chan time.Time {
	tik := time.NewTicker(duration)
	tok := make(chan time.Time)

	go func() {
		defer close(tok)
		defer tik.Stop()

		for {
			select {
			case t := <-tik.C:
				select {
				case tok <- t:
				case <-done:
					return
				}
			case <-done:
				return
			}
		}

	}()
	return tok
}

// Tests:
// 1) happy path
// 2) done closed before NewTicker called
// 3) done closed before tick
// 4) done closed after inner tick but before tick is read
// 5) done closed immediately
// 6) nil done channel: expect not to crash, without done semantics (not a real use case, test anyway)
