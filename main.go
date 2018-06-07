package main

import (
	"fmt"
	"os"
	"time"
)

import (
	"github.com/jotingen/go-lc3/lc3"
	"github.com/jotingen/go-lc3/lc3as"
)

var (
	memory [65536]uint16
)

func main() {
	fmt.Println("vim-go")
	assembly := []string{
		".ORIG x3000",
		"AND R0,R0,#0",
		"AND R1,R1,#0",
		"AND R2,R2,#0",
		"AND R3,R3,#0",
		"AND R4,R4,#0",
		"AND R5,R5,#0",
		"AND R6,R6,#0",
		"AND R7,R7,#0",
		"ADD R1,R1,#1",
		"ADD R0,R0,R1",
		"ADD R0,R0,R1",
		"BRnzp SKIP",
		"ADD R2,R0,R1",
		"SKIP ADD R3,R0,R1",
		"ADD R4,R0,R1",
		"ADD R7,R7,#-5",
		"LOOP ADD R7,R7,#1",
		"BRn LOOP",
		"HALT",
	}
	pc, memory := processAssembly(assembly)

	err := run(pc, memory)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func processAssembly(assembly []string) (pc uint16, memory [65536]uint16) {
	return lc3as.Assemble(assembly)
}

func run(pc uint16, memory [65536]uint16) (err error) {
	var r lc3.Request
	var m uint16

	lc3 := lc3.LC3{}
	lc3.Init(pc)

	fmt.Printf("\n%s\n", lc3)

	cycles := 0
	timeStart := time.Now()
	for memory[pc] != 0xF025 { //Breakout on HALT instruction

		pc, r, err = lc3.Step(memory[pc], m)
		if err != nil {
			fmt.Println(err)
			return err
		}
		if r.Vld {
			if r.RnW {
				m = memory[r.Address]
			} else {
				memory[r.Address] = r.Data
			}
		}

		cycles++

	}
	timeEnd := time.Now()

	fmt.Printf("\n%s\n", lc3)

	nanosecondsPerCycle := float64(timeEnd.Sub(timeStart)) / float64(cycles)
	secondsPerCycle := float64(nanosecondsPerCycle) / 1000.0 / 1000.0 / 1000.0
	hertz := 1 / secondsPerCycle
	fmt.Printf("%dcycles/%s = %1.0fHz\n", cycles, timeEnd.Sub(timeStart), hertz)

	return nil
}
