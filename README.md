# channerics
Channerics is a personal generic chan pattern library for golang 1.18+ with 100% unit-test code coverage.
I wrote this to get acquainted with generics and concurrency prior to the release of generics in 1.18 and I use it in several non-production apps. Anything in this library is free to use/copy. 

Note that by the time Go 1.19 is released, mature pattern libraries will almost certainly exist. You should find and evaluate them instead.

## Sources and Credit

Many of these patterns are from around the web or directly extend patterns from [Concurrency in Go](https://www.amazon.com/Concurrency-Go-Tools-Techniques-Developers/dp/1491941197), by Katherine Cox-Buday, to whom most credit should be given.
This library would not exist without her, and makes no claims of originality beyond extending the work and adding tests and patterns, for
community use.

Teiva Harsanyi's ["100 Go mistakes"](https://www.manning.com/books/100-go-mistakes-and-how-to-avoid-them) contains two chapters on concurrency best-practices and gotchas, making it a good code-quality addendum to Cox-Buday's book.

Grabbing these books and learning about the CSP paradigm is a huge developer asset, even for non-golang developers.

## Development and Testing
The .devcontainer folder contains the specification for the development container, currently using the golang 1.18 beta image for generics. See the readme in the folder for how to build and develop using the vscode container. It provides a complete development environment and test stack.

## Golang Channel Pitfalls and Lessons Learned

This library taught me a few lessons that I could not have discovered without writing tests. The following topics are not complete, just personal notes.

### *Write tests*

Writing tests was the only way to discover the corner cases extended in the remaining lessons-learned below. Per usual code pains, *writing tests is the only way to discover cornercases and sharpen the scope of my function definitions*.

With respect to channels, any function taking context or channels can contain hidden leaky behavior. In short, boilerplate tests should cover conditions like the following:
1) channels that are already closed: will a goroutine needlessly perform some work before it observes closure?
2) contexts which are already cancelled
3) nil channels
4) various exit conditions: will a go-routine continue non-deterministically generating values after closure/cancellation? In terms of promptness, when and how will it respond to cancellation?

The granular behavior of concurrency patterns in Go carry system-level implications for distributed systems:
1) Sender requirements: how will senders behave when blocked?
2) Post-conditions: how will the system as a whole behave on cancellation? Will there be any leaky behavior, for example along code paths through one or more select statements?
3) Utilization: in environments like Kubernetes, a ResourceQuota may limit your process to specific slices of cpu time. Assumptions may have differed under development, limiting performance in production (see Harsanyi's book for an explanation).

### *The impact of `select` non-determinism*

This for-select code fragment occurs throughout golang tutorial blogoverse, often with respect to the 'or-done' pattern:
```
    for _, val := range string[]{"abc", "123"} {
        select {
            case <-done:
                return
            case out <- val:
                fmt.Printf("%s ", val)
        }
    }
```

Assume that 'done' is already closed before reaching this code fragment. What will be printed? Since the `select` statement honors each satisfied case with uniform probability, all of the following are possible:
* *nothing*
* abc
* abc 123

The point is that 'out' becomes a leaky channel potentially producing values even after 'done' is closed. The code will output n additional values with probability 1/(n+1) after 'done' closes.

Variations of the for-select fragment above are so common that leaky exit conditions exist in many codebases. Imagine a large pipeline of interconnected channels continuing to perform work on requests long after context cancellation. Further, tests of expected values for `out` (if tests even exist!) could randomly fail and you will have flaky tests in your code base.

A solution is to give `done` higher precedence using what I call a 'done-guard' to ensure immediate exit when 'done' is closed:
```
    for _, val := range string[]{"abc", "123"} {
        // ensure we exit immediately if 'done' is closed
        select {
            case <-done:
                return
            default:
        }
        // send val or exit if 'done' is closed
        select {
            case <-done:
                return
            case out <- val:
                fmt.Println(val)
        }
    }
```

There are many variations of this problem. Many 'orDone' code examples in wide-circulation contain the exact flaw described above and may not exit as immediately as one might expect.

## NewTicker likes to memory leak

The godocs refer to the fact that `time.NewTicker` is an easy source of memory leaks if the returned timer is not stopped. The underlying timer is an OS resource that the user is responsible for releasing, and garbage collectors cannot observe when an OS timer is no longer in use. The [NewTicker implementation in this repo](./channels/tick.go) ensures correct deallocation and accepts a `done` channel for cancellation.

## Channel Surfing

Coming from a traditional pthread and C#/.NET non-[CSP](https://en.wikipedia.org/wiki/Communicating_sequential_processes) concurrency paradigm, I gravitated toward the `sync` package: throw a mutex around a resource and call it a day. However, its software lifecycle typically requires dozens of hours reviewing esoteric non-CSP concurrent code to understand the dataflow and the protection of critical sections, which are usually non-obvious except to the code author. Aside from mixing data and control, .NET languages also support both low-level locking and threading, as well as the higher-level `async` TAP paradigm, despite being dangerously incompatible. The former utilizes OS threads directly, whereas the latter demuxes async routines over threads, such that combining them can lead to issues like thread starvation that go undetected until production (!). This is a disastrous language flaw (no doubt for backward compatibility), especially in microservices. It inevitably introduces bugs as well as constant review overhead to ensure that code does not mix paradigms, nor that third party libraries strictly use one paradigm or the other. Re-reviewing third party libs on every release is not a task anyone should desire!

Golang intentionally included `chan` as a native feature, so don't hesitate to create, use, and abuse them like any other garbage-collected primitive. Of course they require background knowledge, but such is always the case. In general, prefer channels over `sync` based implementations because channels improve comprehensibility in terms of cancellation, entrance/exit conditions, and overall, channels simply yield well-defined CSP-style api boundaries. [Free-lists](./channels/freelist.go) and [semaphores](./channels/semaphore.go) exemplify how channels can trivially implement common concurrency patterns. Mutexes and other sync primitives still have utility, so one just needs to evaluate chans vs sync per use-case. Teiva Harsanyi's "100 Go Mistakes" has additional info.

Channels help to compartmentalize three primary concurrent programming responsibilities, which are also guideposts when writing code:
1) data flow: by generalizing the addage 'don't share memory to communicate, share memory by communicating', data flow is clearer, safer, and easily tested.
2) control flow: channel closure, nil channels, and even recursion over a set of channels provide elegant 
control primitives that can be combined to form higher level control mechanisms.
3) encapsulation: combined with go-routines, channels yield clean producer-consumer semantics and composition. The scope of critical regions, variables, and responsibilities are bounded and clear. The value of clear ownership across api boundaries is something that any developer can appreciate.

One approach is to implement both a channel-based and a sync-based implementation for the same functionality, then compare them. Except in rare cases, the channel-based implementations usually use less code and are more readable. For the 'rare cases', I usually just didn't spot a simpler channel implementation.

By considering the above responsibilities, code-smells and missing coverage become obvious as you write your code. The CSP-paradigm of channels tends to have much less drama and prods developers toward best practices.