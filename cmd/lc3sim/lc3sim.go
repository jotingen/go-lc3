package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

import (
	"github.com/jotingen/go-lc3/asm2obj"
	"github.com/jotingen/go-lc3/lc3"
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
		"LD R5,LOOP",
		"NOT R4,R4",
		"ADD R4,R4,#1",
		"HALT",
	}
	pc, memory := processAssembly(assembly)

	ascii, err := os.Create("mem.obj.ascii")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer ascii.Close()

	for i := range memory {
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, memory[i])
		ascii.Write(b)
	}
	ascii.Sync()

	obj, err := os.Create("mem.obj")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer obj.Close()

	for i := range memory {
		obj.WriteString(fmt.Sprintf("%04x:%04x\n", i, memory[i]))
	}
	obj.Sync()

	err = run(pc, memory)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func processAssembly(assembly []string) (pc uint16, memory [65536]uint16) {
	return asm2obj.Assemble(assembly)
}

func run(pc uint16, memory [65536]uint16) (err error) {
	var r lc3.Request
	var m uint16

	lc3 := lc3.LC3{}
	lc3.Init(pc)

	fmt.Printf("\n%s\n", lc3)

	cycles := 0
	timeStart := time.Now()
	for pc != 0x0025 { //Breakout when PC goes to HALT address

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
