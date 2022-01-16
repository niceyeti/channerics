package main

import "fmt"

func print[T any](item T) {
	fmt.Println(item)
}

func main() {
	print(123)
}