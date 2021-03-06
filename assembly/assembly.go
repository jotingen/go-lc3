package assembly

//go:generate go-bindata -pkg $GOPACKAGE -o lc3os.go lc3os.asm

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

import (
	"github.com/golang/glog"
)

var (
	table = make(map[string]uint16)
)

func Assemble(assembly []string) (memory []uint16) {
	os, err := Asset("lc3os.asm")
	for i := 0; i < 65536; i++ {
		memory = append(memory, uint16(rand.Uint32()))
	}
	if err != nil {
		fmt.Println("OS code not compiled with program")
	} else {
		assemble(strings.Split(string(os), "\n"), memory)
	}
	assemble(assembly, memory)
	return
}

func assemble(assembly []string, memory []uint16) {
	reSpaces := regexp.MustCompile(`[\s\t]+`)
	reHex := regexp.MustCompile(`^0?x([0-9A-Fa-f]+)$`)
	reDec := regexp.MustCompile(`^#?(-?[0-9A-Fa-f]+)$`)

	//1st pass, remove comments, blank lines, multiple spaces
	reComment := regexp.MustCompile(`;.*$`)
	for i := 0; i < len(assembly); i++ {
		assembly[i] = reComment.ReplaceAllString(assembly[i], "")
		assembly[i] = strings.TrimSpace(assembly[i])
		assembly[i] = reSpaces.ReplaceAllString(assembly[i], " ")
		if assembly[i] == "" {
			assembly = append(assembly[:i], assembly[i+1:]...)
			i--
		}
	}

	//2nd pass, move labels into same line
	reLabel := regexp.MustCompile(`^\w+$`)
	for i := 0; i < len(assembly); i++ {
		if reLabel.MatchString(assembly[i]) {
			switch assembly[i] {
			case ".ORIG":
			case ".FILL":
			case ".BLKW":
			case ".STRINGZ":
			case ".END":
			case "GETC":
			case "OUT":
			case "PUTS":
			case "IN":
			case "PUTSP":
			case "HALT":
			case "ADD":
			case "AND":
			case "BRn", "BRz", "BRp", "BR", "BRzp", "BRnp", "BRnz", "BRnzp":
			case "JMP", "RET":
			case "JSR", "JSRR":
			case "LD":
			case "LDI":
			case "LDR":
			case "NOT":
			case "RTI":
			case "ST":
			case "STI":
			case "STR":
			case "TRAP":
			default:
				assembly[i+1] = strings.Join([]string{assembly[i], assembly[i+1]}, " ")
				assembly = append(assembly[:i], assembly[i+1:]...)
				i--
			}
		}
	}

	//3rd pass, build table
	//Define ORIG to 0x3000 untile overridden
	origin := uint16(0x3000)
	offset := uint16(0)
	for i, line := range assembly {
		items := reSpaces.Split(line, 2)
		switch items[0] {
		case ".ORIG":
			if reHex.MatchString(items[1]) {
				pcHex := reHex.FindAllStringSubmatch(items[1], -1)[0][1]
				pcInt, err := strconv.ParseUint(pcHex, 16, 16)
				if err != nil {
					fmt.Println("Error processing .ORIG ", line)
				}
				origin = uint16(pcInt)
			} else if reDec.MatchString(items[1]) {
				pcDec := reDec.FindAllStringSubmatch(items[1], -1)[0][1]
				pcInt, err := strconv.ParseUint(pcDec, 10, 16)
				if err != nil {
					fmt.Println("Error processing .ORIG ", line)
				}
				origin = uint16(pcInt)
			} else {
				fmt.Println("Error processing .ORIG ", line)
			}

		case ".FILL":
			offset++
		case ".BLKW":
			offset++
		case ".STRINGZ":
			offset++
		case ".END":
		case "GETC":
			offset++
		case "OUT":
			offset++
		case "PUTS":
			offset++
		case "IN":
			offset++
		case "PUTSP":
			offset++
		case "HALT":
			offset++
		case "ADD":
			offset++
		case "AND":
			offset++
		case "BRn", "BRz", "BRp", "BR", "BRzp", "BRnp", "BRnz", "BRnzp":
			offset++
		case "JMP", "RET":
			offset++
		case "JSR", "JSRR":
			offset++
		case "LD":
			offset++
		case "LDI":
			offset++
		case "LDR":
			offset++
		case "NOT":
			offset++
		case "RTI":
			offset++
		case "ST":
			offset++
		case "STI":
			offset++
		case "STR":
			offset++
		case "TRAP":
			offset++
		default:
			//If not a command, assume it was a label
			assembly[i] = items[1]
			table[items[0]] = origin + offset
			offset++
		}
	}

	if glog.V(1) {
		glog.Info("Cleaned Code:")
	}
	for _, line := range assembly {
		if glog.V(1) {
			glog.Info(line)
		}
	}
	fmt.Println()

	if glog.V(1) {
		glog.Info("Table:")
	}
	for key, value := range table {
		if glog.V(1) {
			glog.Infof("%20s:x%04x\n", key, value)
		}
	}
	fmt.Println()

	currentPC := origin
	offset = 0
	for i, line := range assembly {
		if glog.V(1) {
			glog.Infof("Processing %d: %s\n", i, line)
		}
		instruction := uint16(0)
		items := strings.Split(line, " ")
		op := items[0]
		switch op {

		case ".FILL":
			fill := uint16(0)
			if reHex.MatchString(items[1]) {
				fillHex := reHex.FindAllStringSubmatch(items[1], -1)[0][1]
				fillInt, err := strconv.ParseUint(fillHex, 16, 16)
				if err != nil {
					fmt.Println("Error processing ", line)
				}
				fill = uint16(fillInt)
			} else if reDec.MatchString(items[1]) {
				fillDec := reDec.FindAllStringSubmatch(items[1], -1)[0][1]
				fillInt, err := strconv.ParseUint(fillDec, 10, 16)
				if err != nil {
					fmt.Println("Error processing ", line)
				}
				fill = uint16(fillInt)
			} else {
				if _, ok := table[items[1]]; ok {
					fillInt := table[items[1]]
					fill = uint16(fillInt)
				} else {
					fmt.Println("Error processing ", line)
					fmt.Printf("%s not found in lookup table\n", items[1])
				}
			}
			instruction |= fill
			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++
		case ".BLKW":
			operands := strings.Split(items[1], " ")
			count, err := strconv.Atoi(operands[0])
			if err != nil {
				fmt.Println("Error processing .BLKW ", line)
			}
			for r := 0; r < count; r++ {
				val := 0
				if len(operands) >= 2 {
					val, err = strconv.Atoi(operands[1])
					if err != nil {
						fmt.Println("Error processing .BLKW ", line)
					}
				}
				if glog.V(1) {
					glog.Infof("%04x:%04x ; %s\n", currentPC, val, line)
				}
				memory[currentPC] = uint16(val)
				currentPC++
				offset++
			}
		case ".STRINGZ":
			s, err := strconv.Unquote(strings.Join(items[1:], " "))
			if err != nil {
				fmt.Println("Error processing ", items[1])
				fmt.Println(err)
			}
			for _, char := range s {
				if glog.V(1) {
					glog.Infof("%04x:%04x ; %s\n", currentPC, uint16(char), ".STRINGZ "+strconv.Quote(string(char)))
				}
				memory[currentPC] = uint16(char)
				currentPC++
				offset++
			}
		case ".END":
		case "GETC":
			instruction |= 0xF020
			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "OUT":
			instruction |= 0xF021
			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "PUTS":
			instruction |= 0xF022
			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "IN":
			instruction |= 0xF023
			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "PUTSP":
			instruction |= 0xF024
			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "HALT":
			instruction |= 0xF025
			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "AND":
			instruction |= 0x5000

			reAND1 := regexp.MustCompile(`^\s*AND\s+R(\d)\s*,\s*R(\d)\s*,\s*R(\d)`)
			reAND2 := regexp.MustCompile(`^\s*AND\s+R(\d)\s*,\s*R(\d)\s*,\s*#(-?\d+)`)
			if reAND1.MatchString(line) {
				operands := reAND1.FindStringSubmatch(line)

				dr, err := processRegister(operands[1])
				if err != nil {
					fmt.Printf("Error processing DR %d: %s", i, line)
				}

				sr1, err := processRegister(operands[2])
				if err != nil {
					fmt.Printf("Error processing SR1 %d: %s", i, line)
				}

				sr2, err := processRegister(operands[3])
				if err != nil {
					fmt.Printf("Error processing SR2 %d: %s", i, line)
				}

				instruction |= (dr << 9) | (sr1 << 6) | sr2

			} else if reAND2.MatchString(line) {
				operands := reAND2.FindStringSubmatch(line)

				dr, err := processRegister(operands[1])
				if err != nil {
					fmt.Printf("Error processing DR %d: %s", i, line)
				}

				sr1, err := processRegister(operands[2])
				if err != nil {
					fmt.Printf("Error processing SR1 %d: %s", i, line)
				}

				imm5, err := processImm5(operands[3])
				if err != nil {
					fmt.Printf("Error processing IMM5 %d: %s", i, line)
				}

				instruction |= (dr << 9) | (sr1 << 6) | 0x0020 | imm5

			} else {
				fmt.Printf("Error processing line %d: %s", i, line)
			}

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "ADD":
			instruction |= 0x1000

			reADD1 := regexp.MustCompile(`^\s*ADD\s+R(\d)\s*,\s*R(\d)\s*,\s*R(\d)`)
			reADD2 := regexp.MustCompile(`^\s*ADD\s+R(\d)\s*,\s*R(\d)\s*,\s*#(-?\d+)`)
			if reADD1.MatchString(line) {
				operands := reADD1.FindStringSubmatch(line)

				dr, err := processRegister(operands[1])
				if err != nil {
					fmt.Printf("Error processing DR %d: %s", i, line)
				}

				sr1, err := processRegister(operands[2])
				if err != nil {
					fmt.Printf("Error processing SR1 %d: %s", i, line)
				}

				sr2, err := processRegister(operands[3])
				if err != nil {
					fmt.Printf("Error processing SR2 %d: %s", i, line)
				}

				instruction |= (dr << 9) | (sr1 << 6) | sr2

			} else if reADD2.MatchString(line) {
				operands := reADD2.FindStringSubmatch(line)

				dr, err := processRegister(operands[1])
				if err != nil {
					fmt.Printf("Error processing DR %d: %s", i, line)
				}

				sr1, err := processRegister(operands[2])
				if err != nil {
					fmt.Printf("Error processing SR1 %d: %s", i, line)
				}

				imm5, err := processImm5(operands[3])
				if err != nil {
					fmt.Printf("Error processing IMM5 %d: %s", i, line)
				}

				instruction |= (dr << 9) | (sr1 << 6) | 0x0020 | imm5

			} else {
				fmt.Printf("Error processing line %d: %s", i, line)
			}

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "BRn", "BRz", "BRp", "BR", "BRzp", "BRnp", "BRnz", "BRnzp":
			instruction |= 0x0000

			line = replaceLabelAsOffset(line, currentPC)

			reBR := regexp.MustCompile(`^\s*BRn?z?p?\s+#(-?\d+)`)
			operands := reBR.FindStringSubmatch(line)

			if strings.Contains(op, "n") {
				instruction |= 0x0800
			}
			if strings.Contains(op, "z") {
				instruction |= 0x0400
			}
			if strings.Contains(op, "p") {
				instruction |= 0x0200
			}
			if op == "BR" {
				instruction |= 0x0E00
			}

			pcOffset9, err := processOffset9(operands[1])
			if err != nil {
				fmt.Printf("Error processing PCOFFSET9 %d: %s", i, line)
			}

			instruction |= pcOffset9

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "JMP":
			instruction |= 0xC000

			reJMP := regexp.MustCompile(`^\s*JMP\s+R(\d)`)
			operands := reJMP.FindStringSubmatch(line)

			baseR, err := processRegister(operands[1])
			if err != nil {
				fmt.Printf("Error processing BaseR %d: %s", i, line)
			}

			instruction |= baseR << 6

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "RET":
			instruction |= 0xC1C0

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "JSR":
			instruction |= 0x4800

			line = replaceLabelAsOffset(line, currentPC)

			reJSR := regexp.MustCompile(`^\s*JSR\s+#(-?\d+)`)
			operands := reJSR.FindStringSubmatch(line)

			pcOffset11, err := processOffset11(operands[1])
			if err != nil {
				fmt.Printf("Error processing PCOFFSET11 %d: %s", i, line)
			}
			instruction |= pcOffset11

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "JSRR":
			instruction |= 0x4000

			reJSRR := regexp.MustCompile(`^\s*JSRR\s+R(\d)`)
			operands := reJSRR.FindStringSubmatch(line)

			baseR, err := processRegister(operands[1])
			if err != nil {
				fmt.Printf("Error processing BaseR %d: %s", i, line)
			}
			instruction |= baseR << 6

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "LD":
			instruction |= 0x2000

			line = replaceLabelAsOffset(line, currentPC)

			reLD := regexp.MustCompile(`^\s*LD\s+R(\d)\s*,\s*#(-?\d+)`)
			operands := reLD.FindStringSubmatch(line)

			DR, err := processRegister(operands[1])
			if err != nil {
				fmt.Printf("Error processing DR %d: %s", i, line)
			}
			instruction |= DR << 9

			pcOffset9, err := processOffset9(operands[2])
			if err != nil {
				fmt.Printf("Error processing PCOFFSET9 %d: %s", i, line)
			}
			instruction |= pcOffset9

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "LDI":
			instruction |= 0xA000

			line = replaceLabelAsOffset(line, currentPC)

			reLDI := regexp.MustCompile(`^\s*LDI\s+R(\d)\s*,\s*#(-?\d+)`)
			operands := reLDI.FindStringSubmatch(line)

			DR, err := processRegister(operands[1])
			if err != nil {
				fmt.Printf("Error processing DR %d: %s", i, line)
			}
			instruction |= DR << 9

			pcOffset9, err := processOffset9(operands[2])
			if err != nil {
				fmt.Printf("Error processing PCOFFSET9 %d: %s", i, line)
			}
			instruction |= pcOffset9

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "LDR":
			instruction |= 0x6000

			reLDR := regexp.MustCompile(`^\s*LDR\s+R(\d)\s*,\s*R(\d)\s*,\s*#(-?\d+)`)
			operands := reLDR.FindStringSubmatch(line)

			DR, err := processRegister(operands[1])
			if err != nil {
				fmt.Printf("Error processing DR %d: %s", i, line)
			}
			instruction |= DR << 9

			baseR, err := processRegister(operands[2])
			if err != nil {
				fmt.Printf("Error processing baseR %d: %s", i, line)
			}
			instruction |= baseR << 6

			offset6, err := processOffset6(operands[3])
			if err != nil {
				fmt.Printf("Error processing OFFSET6 %d: %s", i, line)
			}
			instruction |= offset6

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "LEA":
			instruction |= 0xE000

			line = replaceLabelAsOffset(line, currentPC)

			reLEA := regexp.MustCompile(`^\s*LEA\s+R(\d)\s*,\s*#(-?\d+)`)
			operands := reLEA.FindStringSubmatch(line)

			DR, err := processRegister(operands[1])
			if err != nil {
				fmt.Printf("Error processing DR %d: %s", i, line)
			}
			instruction |= DR << 9

			pcOffset9, err := processOffset9(operands[2])
			if err != nil {
				fmt.Printf("Error processing PCOFFSET9 %d: %s", i, line)
			}
			instruction |= pcOffset9

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "NOT":
			instruction |= 0x9000

			reNOT := regexp.MustCompile(`^\s*NOT\s+R(\d)\s*,\s*R(\d)`)
			operands := reNOT.FindStringSubmatch(line)

			DR, err := processRegister(operands[1])
			if err != nil {
				fmt.Printf("Error processing DR %d: %s", i, line)
			}
			instruction |= DR << 9

			SR, err := processRegister(operands[2])
			if err != nil {
				fmt.Printf("Error processing SR %d: %s", i, line)
			}
			instruction |= SR << 6

			instruction |= 0x003F

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "RTI":
			instruction |= 0x8000

		case "ST":
			instruction |= 0x3000

			line = replaceLabelAsOffset(line, currentPC)

			reST := regexp.MustCompile(`^\s*ST\s+R(\d)\s*,\s*#(-?\d+)`)
			operands := reST.FindStringSubmatch(line)

			SR, err := processRegister(operands[1])
			if err != nil {
				fmt.Printf("Error processing SR %d: %s", i, line)
			}
			instruction |= SR << 9

			pcOffset9, err := processOffset9(operands[2])
			if err != nil {
				fmt.Printf("Error processing PCOFFSET9 %d: %s", i, line)
			}
			instruction |= pcOffset9

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "STI":
			instruction |= 0xB000

			line = replaceLabelAsOffset(line, currentPC)

			reSTI := regexp.MustCompile(`^\s*STI\s+R(\d)\s*,\s*#(-?\d+)`)
			operands := reSTI.FindStringSubmatch(line)

			SR, err := processRegister(operands[1])
			if err != nil {
				fmt.Printf("Error processing SR %d: %s", i, line)
			}
			instruction |= SR << 9

			pcOffset9, err := processOffset9(operands[2])
			if err != nil {
				fmt.Printf("Error processing PCOFFSET9 %d: %s", i, line)
			}
			instruction |= pcOffset9

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "STR":
			instruction |= 0x7000

			reSTR := regexp.MustCompile(`^\s*STR\s+R(\d)\s*,\s*R(\d)\s*,\s*#(-?\d+)`)
			operands := reSTR.FindStringSubmatch(line)

			SR, err := processRegister(operands[1])
			if err != nil {
				fmt.Printf("Error processing SR %d: %s", i, line)
			}
			instruction |= SR << 9

			baseR, err := processRegister(operands[2])
			if err != nil {
				fmt.Printf("Error processing baseR %d: %s", i, line)
			}
			instruction |= baseR << 6

			offset6, err := processOffset6(operands[3])
			if err != nil {
				fmt.Printf("Error processing OFFSET6 %d: %s", i, line)
			}
			instruction |= offset6

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "TRAP":
			instruction |= 0xF000

			reSTR := regexp.MustCompile(`^\s*TRAP\s+(\w+)`)
			operands := reSTR.FindStringSubmatch(line)

			trapVect8Int, err := strconv.Atoi(operands[1])
			if err != nil {
				fmt.Printf("Error processing TRAP %d: %s", i, line)
			}
			instruction |= uint16(trapVect8Int) & 0x00FF

			if glog.V(1) {
				glog.Infof("%04x:%04x ; %s\n", currentPC, instruction, line)
			}
			memory[currentPC] = instruction
			currentPC++
			offset++

		}

	}
	return
}

func processRegister(reg string) (uint16, error) {
	regInt, err := strconv.Atoi(reg)
	if err != nil {
		return 0, err
	}
	return uint16(regInt) & 0x0007, nil
}

func processImm5(imm5 string) (uint16, error) {
	imm5Int, err := strconv.Atoi(imm5)
	if err != nil {
		return 0, err
	}
	return uint16(imm5Int) & 0x001F, nil
}

func processOffset6(offset6 string) (uint16, error) {
	offset6Int, err := strconv.Atoi(offset6)
	if err != nil {
		return 0, err
	}
	return uint16(offset6Int) & 0x003F, nil
}

func processOffset9(offset9 string) (uint16, error) {
	offset9Int, err := strconv.Atoi(offset9)
	if err != nil {
		return 0, err
	}
	return uint16(offset9Int) & 0x01FF, nil
}
func processOffset11(offset11 string) (uint16, error) {
	offset11Int, err := strconv.Atoi(offset11)
	if err != nil {
		return 0, err
	}
	return uint16(offset11Int) & 0x07FF, nil
}

func replaceLabelAsOffset(line string, currentPC uint16) string {
	//fmt.Printf("Got: %s\n", line)
	reLabel := regexp.MustCompile(`\w*$`)
	if reLabel.MatchString(line) {
		label := reLabel.FindAllString(line, 1)
		if _, ok := table[label[0]]; ok {
			line = reLabel.ReplaceAllString(line, fmt.Sprintf("#%d", int16(table[label[0]]-currentPC-1)))
		}
	}
	//fmt.Printf("Created: %s\n", line)
	return line
}

func Dissassemble(inst uint16) string {

	op := Extract1C(inst, 15, 12)
	switch op {

	case 0x1: //ADD
		dr := Extract1C(inst, 11, 9)
		sr1 := Extract1C(inst, 8, 6)
		bit5 := Extract1C(inst, 5, 5)
		if bit5 == 1 {
			imm5 := Extract2C(inst, 4, 0)
			return fmt.Sprintf("ADD R%d,R%d,#%d\n", dr, sr1, int16(imm5))
		} else {
			sr2 := Extract1C(inst, 2, 0)
			return fmt.Sprintf("ADD R%d,R%d,R%d\n", dr, sr1, sr2)
		}

	case 0x5: //AND
		dr := Extract1C(inst, 11, 9)
		sr1 := Extract1C(inst, 8, 6)
		bit5 := Extract1C(inst, 5, 5)
		if bit5 == 1 {
			imm5 := Extract2C(inst, 4, 0)
			return fmt.Sprintf("AND R%d,R%d,#%d\n", dr, sr1, int16(imm5))
		} else {
			sr2 := Extract1C(inst, 2, 0)
			return fmt.Sprintf("AND R%d,R%d,R%d\n", dr, sr1, sr2)
		}

	case 0x0: //BR
		n := Extract1C(inst, 11, 11) == 1
		z := Extract1C(inst, 10, 10) == 1
		p := Extract1C(inst, 9, 9) == 1
		PCoffset9 := Extract2C(inst, 8, 0)

		brString := fmt.Sprintf("BR")
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
		return brString

	case 0xC: //JMP/RET
		baseR := Extract1C(inst, 8, 6)
		return fmt.Sprintf("JMP R%d\n", baseR)

	case 0x4: //JSR/JSRR
		bit11 := Extract1C(inst, 11, 11)
		if bit11 == 1 {

			PCoffset11 := Extract2C(inst, 10, 0)
			return fmt.Sprintf("JSR #%d\n", int16(PCoffset11))

		} else {
			baseR := Extract2C(inst, 8, 6)
			return fmt.Sprintf("JSRR R%d\n", baseR)
		}

	case 0x2: //LD
		dr := Extract1C(inst, 11, 9)
		PCoffset9 := Extract2C(inst, 8, 0)
		return fmt.Sprintf("LD R%d #%d\n", dr, int16(PCoffset9))

	case 0xA: //LDI
		dr := Extract1C(inst, 11, 9)
		PCoffset9 := Extract2C(inst, 8, 0)
		return fmt.Sprintf("LDI R%d #%d\n", dr, int16(PCoffset9))

	case 0x6: //LDR
		dr := Extract1C(inst, 11, 9)
		baseR := Extract1C(inst, 8, 6)
		offset6 := Extract2C(inst, 5, 0)
		return fmt.Sprintf("LDR R%d R%d #%d\n", dr, baseR, int16(offset6))

	case 0xE: //LEA
		dr := Extract1C(inst, 11, 9)
		PCoffset9 := Extract2C(inst, 8, 0)
		return fmt.Sprintf("LEA R%d #%d\n", dr, int16(PCoffset9))

	case 0x9: //NOT
		dr := Extract1C(inst, 11, 9)
		sr := Extract1C(inst, 8, 6)
		return fmt.Sprintf("NOT R%d R%d\n", dr, sr)

	case 0x8: //RTI

		return fmt.Sprintf("RTI\n")

	case 0x3: //ST
		sr := Extract1C(inst, 11, 9)
		PCoffset9 := Extract2C(inst, 8, 0)
		return fmt.Sprintf("ST R%d #%d\n", sr, int16(PCoffset9))

	case 0xB: //STI
		sr := Extract1C(inst, 11, 9)
		PCoffset9 := Extract2C(inst, 8, 0)
		return fmt.Sprintf("STI R%d #%d\n", sr, int16(PCoffset9))

	case 0x7: //STR
		sr := Extract1C(inst, 11, 9)
		baseR := Extract1C(inst, 8, 6)
		offset6 := Extract2C(inst, 5, 0)
		return fmt.Sprintf("ST R%d R%d #%d\n", sr, baseR, int16(offset6))

	case 0xF: //TRAP
		trapvect8 := Extract1C(inst, 7, 0)
		return fmt.Sprintf("TRAP #%d\n", int16(trapvect8))

	default:
		return "Unknown"

	}

}

func Extract1C(inst uint16, hi, lo int) uint16 {
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

func Extract2C(inst uint16, hi, lo int) uint16 {
	field := Extract1C(inst, hi, lo)

	//fmt.Printf("Field %016b ", field)
	if Extract1C(field, hi, hi) == 1 {
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
