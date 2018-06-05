package lc3

import (
	"fmt"
	"math/rand"
)

type LC3 struct {
	Reg [16]uint16

	PC uint16

	Memory [65536]uint16
}

func (lc3 *LC3) Init() {
	for i := range lc3.Reg {
		lc3.Reg[i] = uint16(rand.Intn(65536))
	}
	lc3.PC = 0x3000
}

func (lc3 *LC3) Step() {
	fmt.Println("vim-go")
}

func (lc3 LC3) String() (s string) {
	s += "Reg:\n"
	for i, r := range lc3.Reg {
		s += fmt.Sprintf("%02d: %016b x%04x\n", i, r, r)
	}
	s += "\n"

	s += fmt.Sprintf("PC: %016b x%04x d%d\n", lc3.PC, lc3.PC, lc3.PC)
	s += "\n"
	return s
}
