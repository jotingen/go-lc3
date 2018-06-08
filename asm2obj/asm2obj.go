package asm2obj

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func Assemble(assembly []string) (pc uint16, memory [65536]uint16) {
	table := make(map[string]int)
	reHex := regexp.MustCompile(`^0?x([0-9A-Fa-f]+)$`)

	//First pass, built table
	offset := 0
	for i, line := range assembly {
		items := strings.Split(line, " ")
		switch items[0] {
		case ".ORIG":
			pcHex := reHex.FindAllStringSubmatch(items[1], -1)[0][1]
			pcInt, err := strconv.ParseUint(pcHex, 16, 16)
			if err != nil {
				fmt.Println("Error processing .ORIG ", table[".ORIG"])
			}
			table[items[0]] = int(pcInt)
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
			//If its a comment, ignore
			//If its empty, ignore
			//If its whitespace, ignore
			//Else its a label, pop off and mark
			split := strings.SplitN(line, " ", 2)
			assembly[i] = split[1]
			table[split[0]] = offset
			offset++
		}
	}

	//If ORIG was not defined, assume x3000
	if _, ok := table[".ORIG"]; !ok {
		table[".ORIG"] = 0x3000
	}

	fmt.Printf("TABLE: %+v\n", table)

	//Process and set PC
	pc = uint16(table[".ORIG"])

	//Second pass
	currentPC := pc
	offset = 0
	for i, line := range assembly {
		instruction := uint16(0)
		items := strings.Split(line, " ")
		op := items[0]
		switch op {

		case ".FILL":
			pcHex := reHex.FindAllStringSubmatch(items[1], -1)[0][1]
			fillInt, err := strconv.ParseUint(pcHex, 16, 16)
			if err != nil {
				fmt.Println("Error processing .FILL ", line)
			}
			instruction |= uint16(fillInt)
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
				if len(operands) >= 2 {
					val, err := strconv.Atoi(operands[1])
					if err != nil {
						fmt.Println("Error processing .BLKW ", line)
					}
					memory[currentPC] = uint16(val)
				}
				currentPC++
				offset++
			}
		case ".STRINGZ":

			s := items[1]
			s = s[1 : len(s)-1] //Remove quotes
			for _, char := range s {
				memory[currentPC] = uint16(char)
				currentPC++
				offset++
			}
		case ".END":
		case "GETC":
			instruction |= 0xF020
			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "OUT":
			instruction |= 0xF021
			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "PUTS":
			instruction |= 0xF022
			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "IN":
			instruction |= 0xF023
			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "PUTSP":
			instruction |= 0xF024
			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "HALT":
			instruction |= 0xF025
			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
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

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
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

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "BRn", "BRz", "BRp", "BR", "BRzp", "BRnp", "BRnz", "BRnzp":
			instruction |= 0x0000

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
				instruction |= 0x0700
			}

			operands := strings.Split(items[1], " ")
			instruction |= uint16(table[operands[0]]-offset-1) & 0x01FF

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "JMP":
			instruction |= 0xC000

			operands := strings.Split(items[1], " ")

			baseR, err := processRegister(operands[1])
			if err != nil {
				fmt.Printf("Error processing BaseR %d: %s", i, line)
			}

			instruction |= baseR << 6

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "RET":
			instruction |= 0xC1C0

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "JSR":
			instruction |= 0x4800

			operands := strings.Split(items[1], " ")

			instruction |= uint16(table[operands[0]]-offset-1) & 0x07FF

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "JSRR":
			instruction |= 0x4000

			operands := strings.Split(items[1], " ")

			baseR, err := processRegister(operands[0])
			if err != nil {
				fmt.Printf("Error processing BaseR %d: %s", i, line)
			}
			instruction |= baseR << 6

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "LD":
			instruction |= 0x2000

			operands := strings.Split(items[1], " ")

			DR, err := processRegister(operands[0])
			if err != nil {
				fmt.Printf("Error processing DR %d: %s", i, line)
			}
			instruction |= DR << 9

			instruction |= uint16(table[operands[0]]-offset-1) & 0x01FF

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "LDI":
			instruction |= 0xA000

			operands := strings.Split(items[1], " ")

			DR, err := processRegister(operands[0])
			if err != nil {
				fmt.Printf("Error processing DR %d: %s", i, line)
			}
			instruction |= DR << 9

			instruction |= uint16(table[operands[1]]-offset-1) & 0x01FF

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "LDR":
			instruction |= 0x6000

			operands := strings.Split(items[1], " ")

			DR, err := processRegister(operands[0])
			if err != nil {
				fmt.Printf("Error processing DR %d: %s", i, line)
			}
			instruction |= DR << 9

			baseR, err := processRegister(operands[1])
			if err != nil {
				fmt.Printf("Error processing baseR %d: %s", i, line)
			}
			instruction |= baseR << 6

			instruction |= uint16(table[operands[2]]-offset-1) & 0x003F

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "LEA":
			instruction |= 0xE000

			operands := strings.Split(items[1], " ")

			DR, err := processRegister(operands[0])
			if err != nil {
				fmt.Printf("Error processing DR %d: %s", i, line)
			}
			instruction |= DR << 9

			instruction |= uint16(table[operands[1]]-offset-1) & 0x01FF

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "NOT":
			instruction |= 0x9000

			operands := strings.Split(items[1], " ")

			DR, err := processRegister(operands[0])
			if err != nil {
				fmt.Printf("Error processing DR %d: %s", i, line)
			}
			instruction |= DR << 9

			SR, err := processRegister(operands[1])
			if err != nil {
				fmt.Printf("Error processing SR %d: %s", i, line)
			}
			instruction |= SR << 6

			instruction |= 0x003F

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "RTI":
			instruction |= 0x8000

		case "ST":
			instruction |= 0x3000

			operands := strings.Split(items[1], " ")

			SR, err := processRegister(operands[0])
			if err != nil {
				fmt.Printf("Error processing SR %d: %s", i, line)
			}
			instruction |= SR << 9

			instruction |= uint16(table[operands[1]]-offset-1) & 0x01FF

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "STI":
			instruction |= 0xB000

			operands := strings.Split(items[1], " ")

			SR, err := processRegister(operands[0])
			if err != nil {
				fmt.Printf("Error processing SR %d: %s", i, line)
			}
			instruction |= SR << 9

			instruction |= uint16(table[operands[1]]-offset-1) & 0x01FF

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "STR":
			instruction |= 0x6000

			operands := strings.Split(items[1], " ")

			SR, err := processRegister(operands[0])
			if err != nil {
				fmt.Printf("Error processing SR %d: %s", i, line)
			}
			instruction |= SR << 9

			baseR, err := processRegister(operands[1])
			if err != nil {
				fmt.Printf("Error processing baseR %d: %s", i, line)
			}
			instruction |= baseR << 6

			instruction |= uint16(table[operands[2]]-offset-1) & 0x003F

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "TRAP":
			instruction |= 0xF000

			trapVect8Int, err := strconv.Atoi(items[1])
			if err != nil {
				fmt.Printf("Error processing TRAP %d: %s", i, line)
			}
			instruction |= uint16(trapVect8Int) & 0x00FF

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
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
