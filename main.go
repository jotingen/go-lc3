package main

import "fmt"

import (
	"github.com/jotingen/go-lc3/lc3"
)

func main() {
	fmt.Println("vim-go")
	lc3 := lc3.LC3{}
	lc3.Init()

	lc3.Step()

	fmt.Printf("%s\n", lc3)
}
