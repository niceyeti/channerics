## Standards
1) no external dependencies and only 'pure' language package dependencies (no ioutil, net/http, etc.).
2) 100% code coverage

## Stage one
1) The following patterns: OrDone, Fan Out/In, Cond, Tee, Bridge, Repeater, Generator, Or, Merge, Cancellation, Batcher
2) Tests:
    * unit
    * race-detection

## Stage two
1) First round cleanup and repo re-org
2) Other/advanced patterns: token bucket (rate limiting), buck wild stuff, edge/liveness, etc.
    * Patterns for system behavior:
        * Device signals and interrupts
        * Graceful shutdown patterns for servers (SIGINT)
        * Watch patterns for processes? E.g. use ptrace to monitor syscalls and queue stats.
3) Pop the hood for a while: build options, advanced settings, internal chan considerations/optimizations
4) Advanced change-propagation library? See old observerbl notes for reqs.

## Stage three
1) Concurrency at scale: context, healing, observability methods/tools/approaches/diagnostics.

## Stage When-Go-1.18-Is-Released
When go 1.18 is released there will be full support for exported generics.
* Replace all instances of "func generic" with "func ".
* Uninstall beta: /home/jesse/sdk/go1.18beta1/bin/go
* Update dockerfile and lock on 1.18; or, use Microsoft's vscode container if they publish a 1.18 version.
    * currently pointing at golang:latest