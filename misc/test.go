package main

func main() {

	done := make(chan struct{})
	ch := make(chan struct{})
	select {
	case ch <- struct{}{}:
	case <-done:
	}

}
