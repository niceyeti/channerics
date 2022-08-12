// Copyright Jesse Waite 2022

package channerics

import "sync"

// Broadcast returns n channels that repeat the data of the input channel.
// Each item read from input is sent to every output channel in parallel, and
// all output values must be consumed before subsequent values can be sent to input.
// Thus broadcast() behaves as a single unbuffered channel coupling. To avoid the
// problem of consumers-blocking-senders, one could make @input a buffered channel
// of size 1 and discard subsequent sends when the channel would block due to slow readers.
func Broadcast[T any](
	done <-chan struct{},
	input <-chan T,
	n int,
) (outputs []<-chan T) {
	outChans := make([]chan T, n)
	for i := 0; i < n; i++ {
		outChans[i] = make(chan T)
		outputs = append(outputs, outChans[i])
	}

	broadcast := func() {
		defer func() {
			for _, outChan := range outChans {
				close(outChan)
			}
		}()

		wg := sync.WaitGroup{}
		for item := range OrDone(done, input) {
			wg.Add(len(outChans))
			for _, outChan := range outChans {
				go func(item T, outChan chan T) {
					defer wg.Done()
					select {
					case outChan <- item:
					case <-done:
					}
				}(item, outChan)
			}
			wg.Wait()
		}
	}
	go broadcast()

	return
}
