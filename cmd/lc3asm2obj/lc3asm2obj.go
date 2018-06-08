package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
)

import (
	"github.com/jotingen/go-lc3/asm2obj"
)

var (
	memory [65536]uint16
)

func main() {
	//	assembly := []string{
	//		".ORIG x3000",
	//		"AND R0,R0,#0",
	//		"AND R1,R1,#0",
	//		"AND R2,R2,#0",
	//		"AND R3,R3,#0",
	//		"AND R4,R4,#0",
	//		"AND R5,R5,#0",
	//		"AND R6,R6,#0",
	//		"AND R7,R7,#0",
	//		"ADD R1,R1,#1",
	//		"ADD R0,R0,R1",
	//		"ADD R0,R0,R1",
	//		"BRnzp SKIP",
	//		"ADD R2,R0,R1",
	//		"SKIP ADD R3,R0,R1",
	//		"ADD R4,R0,R1",
	//		"ADD R7,R7,#-5",
	//		"LOOP ADD R7,R7,#1",
	//		"BRn LOOP",
	//		"HALT",
	//	}
	_, memory := processAssembly("test")

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

}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func processAssembly(file string) (pc uint16, memory [65536]uint16) {
	assembly, err := readLines(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return asm2obj.Assemble(assembly)
}

func dumpASCII(file string, memory [65536]uint16) {
	obj, err := os.Create(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer obj.Close()

	for i := range memory {
		obj.WriteString(fmt.Sprintf("%04x:%04x\n", i, memory[i]))
	}
	obj.Sync()
}
