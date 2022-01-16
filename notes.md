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





