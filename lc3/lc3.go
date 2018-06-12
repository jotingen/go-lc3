package lc3

import (
	"fmt"
	"math/rand"
)

type LC3 struct {
	Reg [8]uint16

	PC uint16

	PSR PSR

	RequestStg1 bool
	RequestStg2 bool

	Temp uint16
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
	op := extract1C(inst, 15, 12)
	switch op {

	case 0x1: //ADD
		dr := extract1C(inst, 11, 9)
		sr1 := extract1C(inst, 8, 6)
		bit5 := extract1C(inst, 5, 5)
		if bit5 == 1 {
			imm5 := extract2C(inst, 4, 0)
			fmt.Printf("  Executing ADD R%d,R%d,#%d\n", dr, sr1, int16(imm5))
			lc3.Reg[dr] = lc3.Reg[sr1] + imm5
		} else {
			sr2 := extract1C(inst, 2, 0)
			fmt.Printf("  Executing ADD R%d,R%d,R%d\n", dr, sr1, sr2)
			lc3.Reg[dr] = lc3.Reg[sr1] + lc3.Reg[sr2]
		}
		lc3.PC++
		lc3.setCC(lc3.Reg[dr])

	case 0x5: //AND
		dr := extract1C(inst, 11, 9)
		sr1 := extract1C(inst, 8, 6)
		bit5 := extract1C(inst, 5, 5)
		if bit5 == 1 {
			imm5 := extract2C(inst, 4, 0)
			fmt.Printf("  Executing AND R%d,R%d,#%d\n", dr, sr1, int16(imm5))
			lc3.Reg[dr] = lc3.Reg[sr1] & imm5
		} else {
			sr2 := extract1C(inst, 2, 0)
			fmt.Printf("  Executing AND R%d,R%d,R%d\n", dr, sr1, sr2)
			lc3.Reg[dr] = lc3.Reg[sr1] & lc3.Reg[sr2]
		}
		lc3.PC++
		lc3.setCC(lc3.Reg[dr])

	case 0x0: //BR
		n := extract1C(inst, 11, 11) == 1
		z := extract1C(inst, 10, 10) == 1
		p := extract1C(inst, 9, 9) == 1
		PCoffset9 := extract2C(inst, 8, 0)
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
		lc3.PC++

	case 0xC: //JMP/RET
		baseR := extract1C(inst, 8, 6)
		fmt.Printf("  Executing JMP R%d\n", baseR)
		lc3.PC = lc3.Reg[baseR]

	case 0x4: //JSR/JSRR
		bit11 := extract1C(inst, 11, 11)
		if bit11 == 1 {

			PCoffset11 := extract2C(inst, 10, 0)
			fmt.Printf("  Executing JSR #%d\n", int16(PCoffset11))
			lc3.Reg[7] = lc3.PC + 1
			lc3.PC += PCoffset11 + 1

		} else {
			baseR := extract2C(inst, 8, 6)
			fmt.Printf("  Executing JSRR R%d\n", baseR)
			lc3.Reg[7] = lc3.PC + 1
			lc3.PC = lc3.Reg[baseR]
		}

	case 0x2: //LD
		dr := extract1C(inst, 11, 9)
		PCoffset9 := extract2C(inst, 8, 0)
		if lc3.RequestStg1 {
			fmt.Printf("  Executing LD R%d #%d\n", dr, int16(PCoffset9))
			lc3.RequestStg1 = false
			lc3.PC++
			lc3.Reg[dr] = data
			lc3.setCC(lc3.Reg[dr])
		} else {
			fmt.Printf("  Requesting LD R%d #%d\n", dr, int16(PCoffset9))
			lc3.RequestStg1 = true
			req := Request{Vld: true, RnW: true, Address: lc3.PC + 1 + PCoffset9}
			fmt.Printf("Sending  PC:x%04x Req:%+v\n", lc3.PC, req)
			return lc3.PC, req, nil
		}

	case 0xA: //LDI
		dr := extract1C(inst, 11, 9)
		PCoffset9 := extract2C(inst, 8, 0)
		if lc3.RequestStg2 {
			fmt.Printf("  Executing LDI R%d #%d\n", dr, int16(PCoffset9))
			lc3.RequestStg2 = false
			lc3.PC++
			lc3.Reg[dr] = data
			lc3.setCC(lc3.Reg[dr])
		} else if lc3.RequestStg1 {
			fmt.Printf("  Requesting LDI R%d #%d\n", dr, int16(PCoffset9))
			lc3.RequestStg1 = false
			lc3.RequestStg2 = true
			req := Request{Vld: true, RnW: true, Address: data}
			fmt.Printf("Sending  PC:x%04x Req:%+v\n", lc3.PC, req)
			return lc3.PC, req, nil
		} else {
			fmt.Printf("  Requesting LDI R%d #%d\n", dr, int16(PCoffset9))
			lc3.RequestStg1 = true
			req := Request{Vld: true, RnW: true, Address: lc3.PC + 1 + PCoffset9}
			fmt.Printf("Sending  PC:x%04x Req:%+v\n", lc3.PC, req)
			return lc3.PC, req, nil
		}

	case 0x6: //LDR
		dr := extract1C(inst, 11, 9)
		baseR := extract1C(inst, 8, 6)
		offset6 := extract2C(inst, 5, 0)
		if lc3.RequestStg1 {
			fmt.Printf("  Executing LDR R%d R%d #%d\n", dr, baseR, int16(offset6))
			lc3.RequestStg1 = false
			lc3.PC++
			lc3.Reg[dr] = data
			lc3.setCC(lc3.Reg[dr])
		} else {
			fmt.Printf("  Requesting LDR R%d R%d #%d\n", dr, baseR, int16(offset6))
			lc3.RequestStg1 = true
			req := Request{Vld: true, RnW: true, Address: lc3.Reg[baseR] + offset6}
			fmt.Printf("Sending  PC:x%04x Req:%+v\n", lc3.PC, req)
			return lc3.PC, req, nil
		}

	case 0xE: //LEA
		dr := extract1C(inst, 11, 9)
		PCoffset9 := extract2C(inst, 8, 0)
		fmt.Printf("  Executing LEA R%d #%d\n", dr, int16(PCoffset9))
		lc3.PC++
		lc3.Reg[dr] = lc3.PC + PCoffset9
		lc3.setCC(lc3.Reg[dr])

	case 0x9: //NOT
		dr := extract1C(inst, 11, 9)
		sr := extract1C(inst, 8, 6)
		fmt.Printf("  Executing NOT R%d R%d\n", dr, sr)
		lc3.PC++
		lc3.Reg[dr] = ^lc3.Reg[sr]
		lc3.setCC(lc3.Reg[dr])

	case 0x8: //RTI

		if !lc3.PSR.Privilege {

			if lc3.RequestStg2 {
				fmt.Printf("  Executing RTI\n")
				lc3.RequestStg2 = false
				lc3.PC = lc3.Temp
				lc3.Reg[6]++
				lc3.PSR.Privilege = extract1C(data, 15, 15) == 1
				lc3.PSR.Priority = uint8(extract1C(data, 10, 8))
				lc3.PSR.N = extract1C(data, 2, 2) == 1
				lc3.PSR.Z = extract1C(data, 1, 1) == 1
				lc3.PSR.P = extract1C(data, 0, 0) == 1
			} else if lc3.RequestStg1 {
				fmt.Printf("  Requesting RTI\n")
				lc3.RequestStg1 = false
				lc3.RequestStg2 = true
				lc3.Temp = data
				lc3.Reg[6]++
				req := Request{Vld: true, RnW: true, Address: lc3.Reg[6]}
				fmt.Printf("Sending  PC:x%04x Req:%+v\n", lc3.PC, req)
				return lc3.PC, req, nil
			} else {
				fmt.Printf("  Requesting RTI\n")
				lc3.RequestStg1 = true
				req := Request{Vld: true, RnW: true, Address: lc3.Reg[6]}
				fmt.Printf("Sending  PC:x%04x Req:%+v\n", lc3.PC, req)
				return lc3.PC, req, nil
			}

		} else {
			//TODO
			//Do nothing for now
		}

	case 0x3: //ST
		sr := extract1C(inst, 11, 9)
		PCoffset9 := extract2C(inst, 8, 0)
		fmt.Printf("  Executing ST R%d #%d\n", sr, int16(PCoffset9))
		lc3.PC++
		req := Request{Vld: true, RnW: false, Address: lc3.PC + PCoffset9, Data: lc3.Reg[sr]}
		fmt.Printf("Sending  PC:x%04x Req:%+v\n", lc3.PC, req)
		return lc3.PC, req, nil

	case 0xB: //STI
		sr := extract1C(inst, 11, 9)
		PCoffset9 := extract2C(inst, 8, 0)
		if lc3.RequestStg1 {
			fmt.Printf("  Executing STI R%d #%d\n", sr, int16(PCoffset9))
			lc3.RequestStg1 = false
			lc3.PC++
			req := Request{Vld: true, RnW: false, Address: data, Data: lc3.Reg[sr]}
			fmt.Printf("Sending  PC:x%04x Req:%+v\n", lc3.PC, req)
			return lc3.PC, req, nil
		} else {
			fmt.Printf("  Requesting STI R%d #%d\n", sr, int16(PCoffset9))
			lc3.RequestStg1 = true
			req := Request{Vld: true, RnW: true, Address: lc3.PC + 1 + PCoffset9}
			fmt.Printf("Sending  PC:x%04x Req:%+v\n", lc3.PC, req)
			return lc3.PC, req, nil
		}

	case 0x7: //STR
		sr := extract1C(inst, 11, 9)
		baseR := extract1C(inst, 8, 6)
		offset6 := extract2C(inst, 5, 0)
		fmt.Printf("  Executing ST R%d R%d #%d\n", sr, baseR, int16(offset6))
		lc3.PC++
		req := Request{Vld: true, RnW: false, Address: lc3.Reg[baseR] + offset6, Data: lc3.Reg[sr]}
		fmt.Printf("Sending  PC:x%04x Req:%+v\n", lc3.PC, req)
		return lc3.PC, req, nil

	case 0xF: //TRAP
		trapvect8 := extract1C(inst, 7, 0)
		if lc3.RequestStg1 {
			fmt.Printf("  Executing TRAP #%d\n", int16(trapvect8))
			lc3.RequestStg1 = false
			lc3.Reg[7] = lc3.PC + 1
			lc3.PC = data
		} else {
			fmt.Printf("  Requesting TRAP #%d\n", int16(trapvect8))
			lc3.RequestStg1 = true
			req := Request{Vld: true, RnW: true, Address: trapvect8}
			fmt.Printf("Sending  PC:x%04x Req:%+v\n", lc3.PC, req)
			return lc3.PC, req, nil
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

func extract1C(inst uint16, hi, lo int) uint16 {
	//fmt.Printf("Inst %04x %d %d ", inst, hi, lo)
	if hi >= 16 || hi < 0 || lo >= 16 || lo < 0 {
		fmt.Println("Argument out of bounds")
	}

	//Build mask
	mask := uint16(0)
	for i := 0; i <= hi-lo; i++ {
		mask = mask << 1
		mask |= 0x0001
	}
	for i := 0; i < lo; i++ {
		mask = mask << 1
	}
	//fmt.Printf("Mask %04x ", mask)

	//Apply mask
	field := inst & mask

	//Shift field down
	field = field >> uint(lo)

	//fmt.Printf("Field %04x\n", field)
	return field
}

func extract2C(inst uint16, hi, lo int) uint16 {
	field := extract1C(inst, hi, lo)

	//fmt.Printf("Field %016b ", field)
	if extract1C(field, hi, hi) == 1 {
		//Build sign extension

		mask := uint16(0)
		for i := 0; i <= 15-hi; i++ {
			mask = mask << 1
			mask |= 0x0001
		}
		mask = mask << uint(hi)
		field = inst | mask

	}
	//fmt.Printf("Field %016b\n", field)

	return field
}
