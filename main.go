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
	var err error

	var r lc3.Request
	var m uint16

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
		"HALT",
	}
	pc, memory := lc3as.Assemble(assembly)

	lc3 := lc3.LC3{}
	lc3.Init(pc)

	//Spoof some test instructions
	//memory[0x3000] = 0x103F //ADD R0,R0,#31
	//memory[0x3001] = 0x1001 //ADD R0,R0,R1
	//memory[0x3002] = 0x54A0 //AND R2,R2,#0
	//memory[0x3003] = 0x0E10 //BR (x3003 + x10)
	//memory[0x3013] = 0x56E0 //AND R3,R3,#0

	cycles := 0
	fmt.Printf("\n%s\n", lc3)
	timeStart := time.Now()
	for memory[pc] != 0xF025 { //Breakout on HALT instruction

		pc, r, err = lc3.Step(memory[pc], m)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
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

}
