// Copyright 2022 Jesse Waite

// tick.go contains patterns for intervals, ticks, time stuff.

package channerics

import "time"

// NewTicker wraps a ticker with done semantics to ensure safe cleanup/cancellation.
// See the time.Tick docs for background; timer.Tick does not safely stop, which
// leads to many misuse-cases. Note that this method introduces some trivial latency
// between the timer tick and propagation to the output channel. Also, the closure
// of done may yield one remaining pending-tick depending on whether the inner select
// sending the tick selects the pending-sending case or the done case. However the time
// value is always guaranteed to be prior to the when close was called, t_tick <= t_closure.
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
