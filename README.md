
# channerics
Channerics is a library of generic chan patterns for golang 1.18+.

## Sources and Credit
Many of these patterns are from around the web and also directly extend patterns from [Concurrency in Go](https://www.amazon.com/Concurrency-Go-Tools-Techniques-Developers/dp/1491941197), by Katherine Cox-Buday, to whom most credit should be given. This library would not exist without her, and makes no claims of originality beyond extending the work and adding tests and patterns, for the community to use.

Learning Golang's concurrency model is a terrific way to learn both the language and reusable CSP-style concurrency paradigms in general. Even if you aren't a golang developer I encourage grabbing a copy of the book.

## Development and Testing
The .devcontainer folder contains the specification for the development container, currently using the golang 1.18 beta image for generics. See the readme in the folder for how to build and develop using the vscode container. With it, you'll have a complete development environment and test stack.

## Contact
I'm actively looking for patterns to add to this library and accept suggestions, just make an issue.