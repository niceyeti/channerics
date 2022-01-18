// WorkerFn is a nil'able function returning a T item.
type WorkerFn func[T any]() T

// FanOut takes a generator function 
// TBD: the only reason for the complex generator pattern is that it allows
// me to break when workerFact returns nil. This is verbose, and I'm looking
// for simpler ways of taking some worker-builder funcs.
func FanOut[T any](next func() WorkerFn[T]) ([]<-chan T) {
	builder := next(); 
	for builder != nil { 
		go func {
			builder()
		}()

		builder = workerFact()
	}
}

func FanOut[T any](next func() chan T]) ([]<-chan T) {
	ch := next(); 
	for ch != nil { 
		

		builder = workerFact()
	}
}



func FanOut[T any]() {
	outputs := make([]chan T, n)
	for i := 0; i < n; i++ {
		outputs[i] = workerChan(i)
	}

	return outputs
}




/*
An interesting case of a potential bad generic pattern discovered in the course
of implementing FanOut, above. Say you want to pass in a generator to some function
that returns T's; instead you pass a function that returns functions that return T's,
and thus can break when you receive nil. This seems like complectifying; when implementing
this with generics, the type of the generator func must match the function to which it is 
passed, which can work, but is not very readable.

package main

import "fmt"

func foo[T any](gen func() WorkerFn[T]) {
	for {
		if fn := gen(); fn != nil {
			fmt.Println(fn())
		} else {
			return
		}
	}
}

type WorkerFn[T any] func() T

func main() {
	i := 0
	generator := func() WorkerFn[int] {
		i++
		if i < 5 {
			return WorkerFn[int](func() int {
				return i
			})
		}
		return nil
	}

	foo(generator)
}






*/
