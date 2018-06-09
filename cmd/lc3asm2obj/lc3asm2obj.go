package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"strings"
)

import (
	"github.com/jotingen/go-lc3/asm2obj"
)

var (
	ascii bool
	out   string
	in    string

	memory [65536]uint16
)

func main() {

	flag.BoolVar(&ascii, "ascii", false, "Print out program in ascii")
	flag.StringVar(&out, "o", "", "Print to custom file")
	flag.StringVar(&in, "i", "", "Input assembly file")
	flag.Parse()

	fmt.Println(in, out, ascii)
	_, memory := processAssembly(in)

	if ascii {
		if out == "" {
			if strings.HasSuffix(in, ".asm") {
				out = in[:len(in)-len(".asm")] + ".obj.ascii"
			} else {
				out = in + ".obj.ascii"
			}
		}
		obj, err := os.Create(out)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer obj.Close()

		for i := range memory {
			b := make([]byte, 2)
			binary.BigEndian.PutUint16(b, memory[i])
			obj.Write(b)
		}
		obj.Sync()
	} else {
		if out == "" {
			if strings.HasSuffix(in, ".asm") {
				out = in[:len(in)-len(".asm")] + ".obj"
			} else {
				out = in + ".obj"
			}
		}
		obj, err := os.Create(out)
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

}

func processAssembly(file string) (pc uint16, memory [65536]uint16) {
	assembly, err := readLines(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return asm2obj.Assemble(assembly)
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
