package test_for_generics

// Because I am lazy... this file will either compile or fail if a
// go compiler version does not support generics, which can be obnoxious
// when switching between environments:
// "go build canary.go" will fail when the environment does not support generics.
// This file will go away once go 1.18 is officially released.

import "fmt"

func print[T any](item T) {
	fmt.Println(item)
}

func test() {
	print(123)
}
