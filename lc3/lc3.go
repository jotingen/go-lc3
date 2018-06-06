package lc3

import (
	"fmt"
	"math/rand"
)

type LC3 struct {
	Reg [8]uint16

	PC uint16

	PSR PSR

	Memory [65536]uint16
}

type PSR struct {
	Privilege bool
	Priority  uint8
	N         bool
	Z         bool
	P         bool
}

func (psr PSR) String() (s string) {
	raw := uint16(0)

	if psr.Privilege {
		raw |= 0x8000
		s += "User  "
	} else {
		s += "Super "
	}

	raw &= (uint16(psr.Priority) & 0x0007) << 8
	s += fmt.Sprintf("%d ", psr.Priority)

	if psr.P {
		raw |= 0x0004
		s += "P"
	} else {
		s += "p"
	}
	if psr.Z {
		raw |= 0x0002
		s += "Z"
	} else {
		s += "z"
	}
	if psr.N {
		raw |= 0x0001
		s += "N"
	} else {
		s += "n"
	}
	return fmt.Sprintf("PSR: %016b x%04x (%s)", raw, raw, s)
}

type Request struct {
	Vld     bool
	RnW     bool
	Address uint16
	Data    uint16
}

func (lc3 *LC3) Init(pc uint16) {
	for i := range lc3.Reg {
		lc3.Reg[i] = uint16(rand.Intn(65536))
	}
	lc3.PC = pc
	fmt.Printf("Set PC: %04x\n", lc3.PC)
}

func (lc3 *LC3) Step(inst uint16, data uint16) (uint16, Request, error) {
	fmt.Printf("Recieved Inst:x%04x Data:x%04x\n", inst, data)

	//For now, if the instruction is unrecognized, panic
	op := inst >> 12
	switch op {

	case 0x1: //ADD
		dr := (inst & 0x0E00) >> 9
		sr1 := (inst & 0x01C0) >> 6
		bit5 := (inst & 0x0020) >> 5
		if bit5 == 1 {
			imm5 := inst & 0x001F
			if (imm5&0x10)>>4 == 1 {
				imm5 = imm5 | 0xFFE0
			}
			fmt.Printf("  Executing ADD R%d,R%d,#%d\n", dr, sr1, int16(imm5))
			lc3.Reg[dr] = lc3.Reg[sr1] + imm5
		} else {
			sr2 := inst & 0x0007
			fmt.Printf("  Executing ADD R%d,R%d,R%d\n", dr, sr1, sr2)
			lc3.Reg[dr] = lc3.Reg[sr1] + lc3.Reg[sr2]
		}
		lc3.PC++
		lc3.setCC(lc3.Reg[dr])

	case 0x5: //AND
		dr := (inst & 0x0E00) >> 9
		sr1 := (inst & 0x01C0) >> 6
		bit5 := (inst & 0x0020) >> 5
		if bit5 == 1 {
			imm5 := inst & 0x001F
			if (imm5&0x10)>>4 == 1 {
				imm5 = imm5 | 0xFFE0
			}
			fmt.Printf("  Executing AND R%d,R%d,#%d\n", dr, sr1, int16(imm5))
			lc3.Reg[dr] = lc3.Reg[sr1] & imm5
		} else {
			sr2 := inst & 0x0007
			fmt.Printf("  Executing AND R%d,R%d,R%d\n", dr, sr1, sr2)
			lc3.Reg[dr] = lc3.Reg[sr1] & lc3.Reg[sr2]
		}
		lc3.PC++
		lc3.setCC(lc3.Reg[dr])

	case 0x0: //BR
		n := (inst&0x0800)>>11 == 1
		z := (inst&0x0400)>>10 == 1
		p := (inst&0x0200)>>9 == 1
		PCoffset9 := inst & 0x01FF
		if (PCoffset9&0x100)>>8 == 1 {
			PCoffset9 = PCoffset9 | 0xFE00
		}
		fmt.Print("  Executing BR")
		if n {
			fmt.Print("n")
		}
		if z {
			fmt.Print("z")
		}
		if p {
			fmt.Print("p")
		}
		fmt.Printf(" #%d\n", int16(PCoffset9))
		if (n && lc3.PSR.N) || (z && lc3.PSR.Z) || (p && lc3.PSR.P) {
			lc3.PC += PCoffset9
		}

	default:
		return lc3.PC, Request{}, fmt.Errorf("Op not recognized: x%x", op)

	}

	fmt.Printf("Sending  PC:x%04x Req:%+v\n", lc3.PC, Request{})
	return lc3.PC, Request{}, nil
}

func (lc3 LC3) String() (s string) {
	for i, r := range lc3.Reg {
		s += fmt.Sprintf("R%d:  %016b x%04x %d\n", i, r, r, int16(r))
	}
	s += "\n"

	s += fmt.Sprintf("PC:  %016b x%04x %d\n", lc3.PC, lc3.PC, lc3.PC)
	s += "\n"

	s += fmt.Sprintf("%s\n", lc3.PSR)

	return s
}

func isPositive(data uint16) bool {
	return int16(data) > 0
}

func isZero(data uint16) bool {
	return data == 0
}

func isNegative(data uint16) bool {
	return int16(data) < 0
}

func (lc3 *LC3) setCC(data uint16) {
	lc3.PSR.N = isNegative(data)
	lc3.PSR.Z = isZero(data)
	lc3.PSR.P = isPositive(data)
}
