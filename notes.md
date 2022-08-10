Building:
go1.17 run -gcflags=-G=3  cmd/generics/main.go
... doesn't seem to work; straightforward example appear to want to compile, then fail on type paramter brackets '['.

Instead use 1.18 beta: /home/jesse/sdk/go1.18beta1/bin/go test

See these vscode settings:
    go.testOnSave
    go.coverOnSave
    go.testFlags

## Generics research
* Proposed spec: https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md
* Generics github issue: https://github.com/golang/go/issues/43651
* Easy: https://qvault.io/golang/how-to-use-golangs-generics/
* https://benjiv.com/golang-generics-introduction/
* Other: https://pdos.csail.mit.edu/6.824/
* https://go101.org/article/channel-use-cases.html

### Generics
1) Interfaces:
* Interfaces can have type parameters, which then just travel inside the definition:
    ```
    type Foo[T comparable] interface {
        Contains(T) bool
        Get() T
        ...
    }
    ```
* Receivers implementing an interface are typed, and the type travels through the definition:
    ```
    Correct:
    func (foo *bar[T any]) SomeMethod(lottaTees []T)
    
    Incorrect:
    func (foo *bar) SomeMethod[T any](bunchaTees []T)
    ```
* Type constraints:
```
Specify types that implement comparison ops <, >, etc.:

 type Ordered interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
        ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
        ~float32 | ~float64 |
        ~string
}
```

## Go Concurrency Rules:

Based on the [language spec](https://go.dev/ref/spec#Close), this table summarizes behavior for actions at left on the channel types/states at top.

|         | Buffered | Unbuffered | Nil | Closed |
|---------|----------|------------|-----|-------|
| *read <-ch* | Value (1) immediate if items buffered (2) after send if empty or unbuffered (3) drains buffered values before closing. | Block; returns before send returns | Block forever | Default val (with optional `ok` false)
| *send ch<-* | (1) Block if full; (2) send occurs before receive, if len < cap | Returns after receive completes | Block forever | Panic (even if send occurs before closure!)
| *close* | Chan returns remaining items, then `ok == false`. | Channel closed. | Panic | Panic |
| *len* | 0 | Count of buffered items | 0 | 0 (probably unsafe) |
| *cap* | 0 | Capacity of the channel | 0 | 0 (probably unsafe) |

Some dev rules can be derived directly from this table:
* Since sending requires knowledge of the channel state, any goroutine sending on a channel is responsible for closing it. Hence the channel should also be created in the same context as the worker goroutine.
* Len and cap are always zero for unbuffered channels, even if there is a pending send, since by definition unbuffered means `cap == 0`.
* For unbuffered channels, send returns after read completes.
* There are no 'occurs before' rules for channel closure; i.e. say a goroutine is guaranteed to send before an unbuffered channel is closed; it will still panic from the awaiting send. The same applies to buffered channels **at capacity**, but otherwise their values can be drained after closure (though this drainage issue with buffered channels always seems like a code smell).

Much of these come from the language specification and [The Go Memory Model](https://go.dev/ref/mem).
It is tempting to claim "if you need to know these in depth, you are being too clever",
but in the course of writing this library I learned otherwise. It was critical to memorize and understand these rules in depth to ensure certain exit conditions and invariants that could (will) affect system behavior. 

One hazard is understanding the exit conditions by which composed channels (via workers and select statements) exits via 'done' closure; in many cases, pending values may be received before some top-level channel receive-operation returns `ok == false`. In short, the closure of done may not be an immediate effect in a system, but may need to propagate. When select statements are nested using intermediary channels, a hierarchy of these relationships introduces the non-deterministic pseudo-random behavior of `select` into a system, whereby some pending values may be sent or a `done` closure honored.

### Package initialization
* If package A imports package B, B's init() functions run to completion before those of A.

### Go routines
* The go statement that starts a new goroutine happens before the goroutine's execution begins. 

### Defer
* Use defer for channel closure. This is just basic defer usage, but guarantees closure to prevent goroutine and context leaks, even if a panic occurs.

### Sync
* A single call of f() from once.Do(f) returns before any other call of once.Do(f) returns; all other calls block until the first call completes.

### Channels
Most of these properties establish the ordering of go routines, when referring to 'before' and 'after' relationships. In the context of a single goroutine... you might be too clever.
* The closing of a channel happens before a receive that returns a zero value because the channel is closed.
* A send on a channel happens before the corresponding receive from that channel completes. 
* Buffered channels
    * Except in special cases, prefer unbuffered channels since unbuffered properties are more comprehensible and make overall system behavior more easily verified.
    * The kth receive on a channel with capacity C happens before the k+Cth send from that channel completes. This implies that a full buffered channel has the same behavior as an unbuffered channel: a receive occurs before a send completes. For example, if `cap` is 0 (unbuffered) the first receive (k=1) occurs before the k+0=0+1=1st send completes. It is a perilous property to rely on, since it requires one to know when the buffer is full, which seems 'too clever'. One exception is [using a buffered channel as a semaphore]().
* Unbuffered channels
    * A receive from an unbuffered channel happens before the send on that channel completes. This can be used to establish an ordering between go routines.

### Caching and memory
Don't use double locking. At the end of The Go Memory Model, the author asserts that the for-loop below may never exit, by way of never observing the write to `done`:

    ```
    var a string
    var done bool

    func setup() {
        a = "hello, world"
        done = true
    }

    func main() {
        go setup()
        for !done {
        }
        print(a)
    }
    ```
The author's primary point is that there is no guarantee that observing the write to `done` implies that `a` has been written, since this ordering is only guaranteed within the context of a single goroutine; otherwise the compiler may order re-order the write operations however it wants, without some synchronization mechanism. 

The assertion about the for-loop not exiting is completely unrelated, and has to do with the possibility that an architecture may registerize `done` on one cpu (the one running the for-loop) without subsequently reaching beyond its cached value to read the new value written by another cpu; i.e., nothing enforces that the compiler should make this value consistent across cpus, similar (but not related) to how `volatile` in C enforces that the compiler does not optimize externally-modifiable memory in a manner that would lead to inconsistency. Thus the assertion is about hardware, not mere language specification. But the above code is obviously dangerous, and should utilize synchronization mechanisms to rule out this possibility, if not simply to make the code comprehensible.

But it leads to this conclusion, the same as traditional C, that When reading golang code make no assumptions about:
* the ordering of writes across goroutine boundaries
* the consistency of memory locations (variable names/addresses) across goroutine boundaries
Note this just repeats the golang concurrency mantra: "share memory by communicating, don't communicate by sharing memory". For developers, each goroutine can be seen as a cpu-context that ultimately makes no promises about the consistency of its variables w.r.t. other goroutines, except through golang synchronization mechanisms.

## Dev container notes
Microsoft image tags: https://mcr.microsoft.com/v2/vscode/devcontainers/go/tags/list

Dev container files definitions:
* https://github.com/microsoft/vscode-dev-containers/tree/v0.195.0/containers/go

Primary source: 
* https://github.com/golang/vscode-go/tree/master/.vscode


  1) Edit and then build .devcontainer/base.Dockerfile:
     docker build -f base.Dockerfile -t godev:latest .
  2) Ensure that .devcontainer/Dockerfile FROM command points to the (1) image: `FROM godev`
  3) Open this folder in vscode, then select "reopen in container".
     It should automatically find this container/image and be off and running.
  
Source:
* https://benmatselby.dev/post/vscode-dev-containers/


## Idioms: Mini Patterns

Interesting channel idioms:
1) Use channel nility to eliminate cases in a select, since nil channels block forever:
```
    for {
        select {
            case ch1 <- val:
                // set ch1 to nil to eliminate it from consideration
                ch1 = nil
            case ch2 <- val:
                ch2 = nil
            ...
            // return/break conditions
        }
    }
```
2) Recursively processing (in this case, combining) a slice of channels. Note how the select statement reduces resource consumption and recursive depth. Note: the done chan is excluded to keep example simple:
```
    func any[T any](inputs []<-chan T) <-chan T {
        switch (len(inputs)) {
            case 0:
                return nil // or some other base case
            case 1:
                return inputs[0]
            case 2:
                return either(inputs[0], inputs[1])
        }

        out := make(chan T)
        go func() {
            defer close(out)
            select {
                case <-inputs[0]:
                case <-inputs[1]:
                case <-inputs[2]:
                case <-any(inputs[3:])
            }
        }()
        
        return out
    }
```
3) Use `default` in select statement to make it non-blocking. Input will be ready only if it is ready, output will only be sent the first time that would otherwise result in a block:
```
// Read-if-ready:
select {
    case <-input:
    default:
}
// continue doing other things

// Send once, but don't block:
select {
    case heartbeat <- struct{}{}:
    default:
}
// continue other work

// "Is this channel blocked?" This example demonstrates IfValue semantics (for brevity, it does not include 'done' semantics, but could be extended to do so):
var isBlocked, ok bool
var val T
select {
case val, ok = <-in:
    isBlocked = false
    // handle ok and val...
default:
    isBlocked = true
}
```
This technique is used with buffered channels to implement free-lists.

4) Generics and struct/receivers/interfaces
* TODO: fill this in. It is an important problem that I have yet to see cearly explained.

5) It is okay to not close channels explicitly, and expect them to be closed when they go out of scope for garbage collection.
However, this applies mostly to tests, otherwise channel closure should be seen as an explicit best-practice, if not simply to show you thought about it.
```
Note that it is only necessary to close a channel if the receiver is looking for a close. Closing the channel is a control signal on the channel indicating that no more data follows. 
```

## Library Side Effects, Gotchas
* TODO: track and fill these out. By doing so I can probably fix or eliminate weird cases, but others will be library choices that will be important to highlight.
* Be wary of exit and channel-closure patterns. For instance, both Merge and OrDone have these properties:
    * They close immediately when 'done' is closed.
    * They do not close immediately when their input(s) channels are closed; instead, output must be drained before these funcs enter a select stmt that includes input channel-closure detection as one of its exit conditions.
* All never closes if any of its input channels is nil. I'm on the fence about this, but it depends on its foreseen usage.
    * Would most users find this useful or hazardous?
        * The case for nil:
            * I can pass potentially nil channels if I expect them to be 
* By the definition of the language specification, `select` will select from multiple available options with a (pseudo) uniform random distribution. But this violates `done` semantics in an unexpected way, whereby a for-select may run forever (sans calculus limits :P). When done is closed, the following loop may execute twice: first by randomly selecting the vals channel to send on, then selecting the `done` case on the second iteration:
```
ch := make(chan struct{}, 10)
done := make(chan struct{})
close(done)

go func(){
    defer close(ch)
    for {
        select {
            case ch <- struct{}:
            case <-done:
                return
        }
    }
}()
```
In fact, the probability of n-iterations is 1 / 2^n. Select'ing over cases that include a `done` channel is extremely prevalent, but **this pattern allows for non-deterministic exit conditions in a system**, within the language specification. The `select` statement is effectively leaky, and may yield values long after a `done` channel is closed. A `done` guard can be added as a mechanism to give `done` higher precedence, similar to this per the above:
```
    for {
        // a done-guard
        select {
            case <-done:
                return
            default:
        }

        select {
            case ch <- struct{}:
            case <-done:
                return
        }
    }
```

This program demonstrates the issue (playground version: https://go.dev/play/p/6cgTqPOqTB-):
```
package main

import (
	"fmt"
)

func test() {
	n := 10
	ch := make(chan struct{}, n)
	defer close(ch)
	done := make(chan struct{})

	close(done)
	// 'done' is now closed. So we exit the select stmt immediately, right? Nope...
	for i := 0; i < n; i++ {
		select {
		case ch <- struct{}{}:
		case <-done:
			fmt.Println("Executed i times: ", i+1)
			return
		}
	}

	// deploy a pretend-listener, so the compiler doesn't remove our chan...
	go func() {
		for {
			_, ok := <-ch
			if !ok {
				return
			}
		}
	}()
}

func main() {
	for i := 0; i < 20; i++ {
		test()
	}
}
```