
# channerics
Channerics is a library of generic chan patterns for golang 1.18+.

## Sources and Credit
Many of these patterns are from around the web and directly extend patterns from [Concurrency in Go](https://www.amazon.com/Concurrency-Go-Tools-Techniques-Developers/dp/1491941197), by Katherine Cox-Buday, to whom most credit should be given.
This library would not exist without her, and makes no claims of originality beyond extending the work and adding tests and patterns, for
community use.

Learning Golang's concurrency model is a terrific way to learn both the language and reusable CSP-style concurrency paradigms in general.
Even if you aren't a golang developer I encourage grabbing a copy of the book and learning CSP style programming.

## Development and Testing
The .devcontainer folder contains the specification for the development container, currently using the golang 1.18 beta image for generics. See the readme in the folder for how to build and develop using the vscode container. With it, you'll have a complete development environment and test stack.

## Contact
I'm actively looking for patterns to add to this library and accept suggestions, just make an issue.

## Golang Channel Pitfalls and Lessons Learned

Prior to writing this library, my understanding of golang channels after reading a few sources was the usual pedestrian 
developer mindset of 'okay, I more or less understand enough to stay out of trouble'. However there were a few
lessons I could not have discovered without writing test code:

### *Write tests*

Writing tests for this repo was the only way to discover the corner cases extended in the remaining lessons-learned below. Per usual code pains, *writing tests is the only way to discover cornercases and sharpen the scope of my function definitions*.

With respect to channels, any function taking context or channels can contain hidden leaky behavior. In short, boilerplate tests should cover conditions like the following:
1) channels that are already closed: will a goroutine needlessly perform some work before it observes this closure?
2) contexts which are already cancelled
3) nil channels
4) various exit conditions: will a go-routine continue non-deterministically generating values after closure/cancellation? promptness: when and how will it respond to cancellation?

### *The impact of `select` non-determinism*

This for-select code fragment occurs throughout golang tutorial blogoverse, often with respect to the 'or-done' pattern:
```
    for _, val := range string[]{"abc", "123"} {
        select {
            case <-done:
                return
            case out <- val:
                fmt.Println(val)
        }
    }
```

Assume that 'done' is already closed before reaching this code fragment. What will be printed? The `select` statement honors each satisfied case with uniform probability, hence this may print nothing, "abc", or even "abc" and "123". The point is simply that 'out' becomes a leaky channel potentially producing values even after 'done' is closed. This code will output n additional values with probability 1/(n+1) after 'done' has been closed.

Variations of the for-select fragment above are so ubiquitous that the risk of leaky exit conditions exist in tons of code bases. Imagine a large pipeline of interconnected channels continuing to perform work on requests long after context cancellation. Further, tests of expected values for `out` (if tests even exist!) will randomly fail and you will have flaky tests in your code base.

The (tedious) solution is to give `done` higher precedence using what I call a 'done-guard' to ensure immediate exit when 'done' is closed:
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

The godocs directly allude to the fact that `time.NewTicker` is an easy source of memory leaks if the returned timer is not stopped. This is because the underlying timer is an OS resource that the user is responsible for releasing; garbage collectors cannot observe when an OS timer is no longer in use. The [NewTicker implementation in this repo](./channels/tick.go) ensures correct deallocation and accepts a `done` channel for cancellation.

## Channel Surfing

I came from a `pthread` and C#/.NET [non-CSP](https://en.wikipedia.org/wiki/Communicating_sequential_processes) concurrency paradigm, and therefore gravitated toward the `sync` package: throw a mutex around it and call it a day. I've also spent dozens of hours reviewing esoteric non-CSP concurrent code to understand the dataflow and the protection of critical sections, which are usually not obvious except to the author. Aside from mixing data and control, .NET languages also support/allow both low-level locking and the higher-level `async` TAP paradigm, despite being dangerously incompatible. This is a disastrous language-design fail (no doubt for backward compatibility), especially in microservices. It inevitably introduces bugs as well as constant review overhead to ensure that code does not mix paradigms, or that third party libraries strictly use one or the other. Enjoy re-reviewing those third party libs on every update if you truly need to ensure some anonymous external developer hasn't broken something!

Golang intentionally included `chan` as a native feature, so don't hesitate to create and (ab)use them like any other garbage-collected data structure. Of course there are caveats and required background knowledge, but such is always the case. With few exceptions, prefer channels over `sync` based implementations because channels improve comprehensibility in terms of cancellation, entrance/exit conditions, and overall, channels simply yield well-defined CSP-style api boundaries. [Free-lists](./channels/freelist.go) and [semaphores](./channels/semaphore.go) exemplify how channels can trivially implement common concurrency patterns.

Channels allow you to compartmentalize three primary concurrent programming responsibilities, which are also guideposts when writing code:
1) data flow: by generalizing the addage 'don't share memory to communicate, share memory by communicating', data flow is clearer, safer, and easily tested.
2) control flow: channel closure, nil channels, and even recursion over a set of channels provide elegant 
control primitives that can be combined to form higher level control mechanisms.
3) encapsulation: combined with go-routines, channels yield clean producer-consumer semantics and composition. The scope of critical regions, variables, and responsibilities are bounded and clear.

One approach is to implement both a channel-based and a sync-based implementation for the same functionality, then compare them. Except in rare cases, the channel-based implementations usually use less code and are more readable. For the 'rare cases', I usually just didn't spot a simpler channel implementation.

By considering the above responsibilities, code-smells and missing coverage become obvious as you write your code. The CSP-paradigm of channels tends to have much less drama and prods developers toward best practices.