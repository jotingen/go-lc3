package lc3

import (
	"fmt"
	"math/rand"
	"time"
)

import (
	"github.com/golang/glog"
)

import (
	"github.com/jotingen/go-lc3/assembly"
)

type Memory [65536]uint16

type LC3 struct {
	Reg [8]uint16

	PC uint16

	PSR PSR

	Memory []uint16

	TimerStarted bool
	TimerStart   time.Time
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
		s += "Super "
	} else {
		raw |= 0x8000
		s += "User  "
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
	return fmt.Sprintf("PSR:%04x (%s)", raw, s)
}

func (lc3 *LC3) Init(pc uint16, m []uint16) {
	for i := range lc3.Reg {
		lc3.Reg[i] = uint16(rand.Intn(65536))
	}
	lc3.PSR.Privilege = false
	lc3.PC = pc
	lc3.Memory = m
	//fmt.Printf("Set PC: %04x\n", lc3.PC)
}

func (lc3 *LC3) Step() (uint16, error) {
	//fmt.Printf("Recieved Inst:x%04x Data:x%04x\n", inst, data)
	inst := lc3.Memory[lc3.PC]

	//For now, if the instruction is unrecognized, panic
	op := assembly.Extract1C(inst, 15, 12)
	switch op {

	case 0x1: //ADD
		dr := assembly.Extract1C(inst, 11, 9)
		sr1 := assembly.Extract1C(inst, 8, 6)
		bit5 := assembly.Extract1C(inst, 5, 5)
		if bit5 == 1 {
			imm5 := assembly.Extract2C(inst, 4, 0)
			if glog.V(2) {
				glog.Infof("0x%04x: ADD R%d,R%d,#%d\n", lc3.PC, dr, sr1, int16(imm5))
			}
			lc3.Reg[dr] = lc3.Reg[sr1] + imm5
		} else {
			sr2 := assembly.Extract1C(inst, 2, 0)
			if glog.V(2) {
				glog.Infof("0x%04x: ADD R%d,R%d,R%d\n", lc3.PC, dr, sr1, sr2)
			}
			lc3.Reg[dr] = lc3.Reg[sr1] + lc3.Reg[sr2]
		}
		lc3.PC++
		lc3.setCC(lc3.Reg[dr])

	case 0x5: //AND
		dr := assembly.Extract1C(inst, 11, 9)
		sr1 := assembly.Extract1C(inst, 8, 6)
		bit5 := assembly.Extract1C(inst, 5, 5)
		if bit5 == 1 {
			imm5 := assembly.Extract2C(inst, 4, 0)
			if glog.V(2) {
				glog.Infof("0x%04x: AND R%d,R%d,#%d\n", lc3.PC, dr, sr1, int16(imm5))
			}
			lc3.Reg[dr] = lc3.Reg[sr1] & imm5
		} else {
			sr2 := assembly.Extract1C(inst, 2, 0)
			if glog.V(2) {
				glog.Infof("0x%04x: AND R%d,R%d,R%d\n", lc3.PC, dr, sr1, sr2)
			}
			lc3.Reg[dr] = lc3.Reg[sr1] & lc3.Reg[sr2]
		}
		lc3.PC++
		lc3.setCC(lc3.Reg[dr])

	case 0x0: //BR
		n := assembly.Extract1C(inst, 11, 11) == 1
		z := assembly.Extract1C(inst, 10, 10) == 1
		p := assembly.Extract1C(inst, 9, 9) == 1
		PCoffset9 := assembly.Extract2C(inst, 8, 0)

		brString := fmt.Sprintf("0x%04x: BR", lc3.PC)
		if n {
			brString += fmt.Sprintf("n")
		}
		if z {
			brString += fmt.Sprintf("z")
		}
		if p {
			brString += fmt.Sprintf("p")
		}
		brString += fmt.Sprintf(" #%d\n", int16(PCoffset9))
		if glog.V(2) {
			glog.Info(brString)
		}
		if (n && lc3.PSR.N) || (z && lc3.PSR.Z) || (p && lc3.PSR.P) {
			lc3.PC += PCoffset9
		}
		lc3.PC++

	case 0xC: //JMP/RET
		baseR := assembly.Extract1C(inst, 8, 6)
		if glog.V(2) {
			glog.Infof("0x%04x: JMP R%d\n", lc3.PC, baseR)
		}
		lc3.PC = lc3.Reg[baseR]

	case 0x4: //JSR/JSRR
		bit11 := assembly.Extract1C(inst, 11, 11)
		if bit11 == 1 {

			PCoffset11 := assembly.Extract2C(inst, 10, 0)
			if glog.V(2) {
				glog.Infof("0x%04x: JSR #%d\n", lc3.PC, int16(PCoffset11))
			}
			lc3.Reg[7] = lc3.PC + 1
			lc3.PC += PCoffset11 + 1

		} else {
			baseR := assembly.Extract2C(inst, 8, 6)
			if glog.V(2) {
				glog.Infof("0x%04x: JSRR R%d\n", lc3.PC, baseR)
			}
			lc3.Reg[7] = lc3.PC + 1
			lc3.PC = lc3.Reg[baseR]
		}

	case 0x2: //LD
		dr := assembly.Extract1C(inst, 11, 9)
		PCoffset9 := assembly.Extract2C(inst, 8, 0)
		if glog.V(2) {
			glog.Infof("0x%04x: LD R%d #%d\n", lc3.PC, dr, int16(PCoffset9))
		}
		lc3.PC++
		lc3.Reg[dr] = lc3.Memory[lc3.PC+PCoffset9]
		lc3.setCC(lc3.Reg[dr])

	case 0xA: //LDI
		dr := assembly.Extract1C(inst, 11, 9)
		PCoffset9 := assembly.Extract2C(inst, 8, 0)
		if glog.V(2) {
			glog.Infof("0x%04x: LDI R%d #%d\n", lc3.PC, dr, int16(PCoffset9))
		}
		lc3.PC++
		lc3.Reg[dr] = lc3.Memory[lc3.Memory[lc3.PC+PCoffset9]]
		lc3.setCC(lc3.Reg[dr])

	case 0x6: //LDR
		dr := assembly.Extract1C(inst, 11, 9)
		baseR := assembly.Extract1C(inst, 8, 6)
		offset6 := assembly.Extract2C(inst, 5, 0)
		if glog.V(2) {
			glog.Infof("0x%04x: LDR R%d R%d #%d\n", lc3.PC, dr, baseR, int16(offset6))
		}
		lc3.PC++
		lc3.Reg[dr] = lc3.Memory[lc3.Reg[baseR]+offset6]
		lc3.setCC(lc3.Reg[dr])

	case 0xE: //LEA
		dr := assembly.Extract1C(inst, 11, 9)
		PCoffset9 := assembly.Extract2C(inst, 8, 0)
		if glog.V(2) {
			glog.Infof("0x%04x: LEA R%d #%d\n", lc3.PC, dr, int16(PCoffset9))
		}
		lc3.PC++
		lc3.Reg[dr] = lc3.PC + PCoffset9
		lc3.setCC(lc3.Reg[dr])

	case 0x9: //NOT
		dr := assembly.Extract1C(inst, 11, 9)
		sr := assembly.Extract1C(inst, 8, 6)
		if glog.V(2) {
			glog.Infof("0x%04x: NOT R%d R%d\n", lc3.PC, dr, sr)
		}
		lc3.PC++
		lc3.Reg[dr] = ^lc3.Reg[sr]
		lc3.setCC(lc3.Reg[dr])

	case 0x8: //RTI

		if !lc3.PSR.Privilege {

			if glog.V(2) {
				glog.Infof("0x%04x: RTI\n", lc3.PC)
			}
			lc3.PC = lc3.Memory[lc3.Reg[6]]
			lc3.Reg[6]++
			lc3.PSR.Privilege = assembly.Extract1C(lc3.Memory[lc3.Reg[6]], 15, 15) == 0
			lc3.PSR.Priority = uint8(assembly.Extract1C(lc3.Memory[lc3.Reg[6]], 10, 8))
			lc3.PSR.N = assembly.Extract1C(lc3.Memory[lc3.Reg[6]], 2, 2) == 1
			lc3.PSR.Z = assembly.Extract1C(lc3.Memory[lc3.Reg[6]], 1, 1) == 1
			lc3.PSR.P = assembly.Extract1C(lc3.Memory[lc3.Reg[6]], 0, 0) == 1
			lc3.Reg[6]++

		} else {
			//TODO
			//Do nothing for now
		}

	case 0x3: //ST
		sr := assembly.Extract1C(inst, 11, 9)
		PCoffset9 := assembly.Extract2C(inst, 8, 0)
		if glog.V(2) {
			glog.Infof("0x%04x: ST R%d #%d\n", lc3.PC, sr, int16(PCoffset9))
		}
		lc3.PC++
		lc3.Memory[lc3.PC+PCoffset9] = lc3.Reg[sr]

	case 0xB: //STI
		sr := assembly.Extract1C(inst, 11, 9)
		PCoffset9 := assembly.Extract2C(inst, 8, 0)
		if glog.V(2) {
			glog.Infof("0x%04x: STI R%d #%d\n", lc3.PC, sr, int16(PCoffset9))
		}
		lc3.PC++
		lc3.Memory[lc3.Memory[lc3.PC+PCoffset9]] = lc3.Reg[sr]

	case 0x7: //STR
		sr := assembly.Extract1C(inst, 11, 9)
		baseR := assembly.Extract1C(inst, 8, 6)
		offset6 := assembly.Extract2C(inst, 5, 0)
		if glog.V(2) {
			glog.Infof("0x%04x: ST R%d R%d #%d\n", lc3.PC, sr, baseR, int16(offset6))
		}
		lc3.PC++
		lc3.Memory[lc3.Reg[baseR]+offset6] = lc3.Reg[sr]

	case 0xF: //TRAP
		trapvect8 := assembly.Extract1C(inst, 7, 0)
		if glog.V(2) {
			glog.Infof("0x%04x: TRAP #%d\n", lc3.PC, int16(trapvect8))
		}
		lc3.Reg[7] = lc3.PC + 1
		lc3.PC = lc3.Memory[trapvect8]

	default:
		return lc3.PC, fmt.Errorf("Op not recognized: x%x", op)

	}

	//Timer Registers
	if lc3.Memory[0xFE0A] != 0 {
		if (lc3.Memory[0xFE08]&0x8000)>>15 == 0 {
			if lc3.TimerStarted {
				elapsedMilliseconds := time.Now().Sub(lc3.TimerStart)

				if elapsedMilliseconds >= (time.Duration(lc3.Memory[0xFE0A]) * time.Millisecond) {

					lc3.Memory[0xFE08] = 0x8000
				} else {
					lc3.Memory[0xFE08] = 0x0000
				}
			} else {
				lc3.TimerStart = time.Now()
				lc3.TimerStarted = true
				lc3.Memory[0xFE08] = 0x0000
			}

		} else {
			lc3.Memory[0xFE08] = 0x0000
		}
	} else {
		lc3.Memory[0xFE08] = 0x0000
	}

	//Update clock register
	time := time.Now()
	//CLK1
	lc3.Memory[0xFE0C] = uint16(uint64(time.Nanosecond()) / 1e6)
	//CLK2
	lc3.Memory[0xFE0E] = uint16(uint64(time.Unix()) & 0xFFFF)
	//CLK3
	lc3.Memory[0xFE10] = uint16((uint64(time.Unix()) & 0xFFFF0000) >> 16)

	//Increment MCC
	lc3.Memory[0xFFFF]++

	return lc3.PC, nil
}

func (lc3 LC3) String() (s string) {
	for i, r := range lc3.Reg {
		s += fmt.Sprintf("R%d:%04x ", i, r)
	}
	s += "\n"

	s += fmt.Sprintf("PC:%04x %s\n", lc3.PC, lc3.PSR)

	s += fmt.Sprintf("KBSR:%04x KBDR:%04x\n", lc3.Memory[0xFE00], lc3.Memory[0xFE02])
	s += fmt.Sprintf(" DSR:%04x  DDR:%04x\n", lc3.Memory[0xFE04], lc3.Memory[0xFE06])
	s += fmt.Sprintf(" TMR:%04x  TMI:%04x\n", lc3.Memory[0xFE08], lc3.Memory[0xFE0A])
	s += fmt.Sprintf("CLK1:%04x CLK2:%04x CLK3:%04x (%s)\n", lc3.Memory[0xFE0C], lc3.Memory[0xFE0E], lc3.Memory[0xFE10],
		time.Unix(int64(uint32(lc3.Memory[0xFE0E])|(uint32(lc3.Memory[0xFE10])<<16)), int64(lc3.Memory[0xFE0C])*1e6))
	s += fmt.Sprintf(" MPR:%04x\n", lc3.Memory[0xFE12])
	s += fmt.Sprintf(" VCR:%04x\n", lc3.Memory[0xFE14])
	s += fmt.Sprintf(" MCR:%04x\n", lc3.Memory[0xFFFE])
	s += fmt.Sprintf(" MCC:%04x\n", lc3.Memory[0xFFFF])

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
