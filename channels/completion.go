// Copyright 2022 Jesse Waite

// completion.go contains simple patterns related to completion: context, done, and similar patterns.

package channerics

/*
Status: 
Attempted to build with `go test -gcflags=-G=3`
Backing up to 1.18 beta: https://go.dev/dl/#go1.18beta1
Note the beta command is in /home/jesse/go.
Try to run it, see what happens...

Installed here: /home/jesse/sdk/go1.18beta1/bin/go
Don't forget to uninstall it once done.
*/

// OrDone streams values from vals until done is closed.
func OrDone[T any](
	done <-chan struct{},
	 vals <-chan T,
) <-chan T {
	output := make(chan T)

	go func(){
		defer close(output)
		for {
			select {
			case v := <-vals:
				select {
				case output <-v:
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
