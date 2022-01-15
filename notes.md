Building:
go1.17 run -gcflags=-G=3  cmd/generics/main.go
... doesn't seem to work; straightforward example appear to want to compile, then fail on type paramter brackets '['.

Instead use 1.18 beta: /home/jesse/sdk/go1.18beta1/bin/go test



See these vscode settings:
    go.testOnSave
    go.coverOnSave
    go.testFlags


Module: a collection of packages which are released together
Package: a collection of source files in the same directory, by convention of the same name as the package.
Module path: declared in go.mod, e.g. `module example.com`, and where the module can be downloaded with `go download`.
Import path: a packages relative path, prefixed by the module path.
Go Path:
    * Use `go env GOPATH` and `go help GOPATH`



