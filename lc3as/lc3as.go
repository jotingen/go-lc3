package lc3as

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
	reReg := regexp.MustCompile(`^R(\d)$`)
	reNum := regexp.MustCompile(`^#(-?\d+)$`)
	for _, line := range assembly {
		instruction := uint16(0)
		items := strings.Split(line, " ")
		op := items[0]
		switch op {

		case "HALT":
			instruction |= 0xF025
			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "AND":
			instruction |= 0x5000

			operand := strings.Split(items[1], ",")

			dr := reReg.FindAllStringSubmatch(operand[0], -1)[0][1]
			drVal, err := strconv.ParseUint(dr, 10, 3)
			if err != nil {
				fmt.Println("Error processing dr ", line)
			}
			instruction |= uint16(drVal) << 9

			sr1 := reReg.FindAllStringSubmatch(operand[1], -1)[0][1]
			sr1Val, err := strconv.ParseUint(sr1, 10, 3)
			if err != nil {
				fmt.Println("Error processing sr1 ", line)
			}
			instruction |= uint16(sr1Val) << 9

			if reReg.MatchString(operand[2]) {
				sr2 := reReg.FindAllStringSubmatch(operand[2], -1)[0][1]
				sr2Val, err := strconv.ParseUint(sr2, 10, 3)
				if err != nil {
					fmt.Println("Error processing sr2 ", line)
				}
				instruction |= uint16(sr2Val)
			} else {
				imm5 := reNum.FindAllStringSubmatch(operand[2], -1)[0][1]
				imm5Val, err := strconv.ParseInt(imm5, 10, 5)
				if err != nil {
					fmt.Println("Error processing imm5 ", line)
				}
				instruction |= 1 << 5
				instruction |= uint16(imm5Val)
			}
			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		case "ADD":
			instruction |= 0x1000

			operand := strings.Split(items[1], ",")

			dr := reReg.FindAllStringSubmatch(operand[0], -1)[0][1]
			drVal, err := strconv.ParseUint(dr, 10, 3)
			if err != nil {
				fmt.Println("Error processing dr ", line)
			}
			instruction |= uint16(drVal) << 9

			sr1 := reReg.FindAllStringSubmatch(operand[1], -1)[0][1]
			sr1Val, err := strconv.ParseUint(sr1, 10, 3)
			if err != nil {
				fmt.Println("Error processing sr1 ", line)
			}
			instruction |= uint16(sr1Val) << 9

			if reReg.MatchString(operand[2]) {
				sr2 := reReg.FindAllStringSubmatch(operand[2], -1)[0][1]
				sr2Val, err := strconv.ParseUint(sr2, 10, 3)
				if err != nil {
					fmt.Println("Error processing sr2 ", line)
				}
				instruction |= uint16(sr2Val)
			} else {
				imm5 := reNum.FindAllStringSubmatch(operand[2], -1)[0][1]
				imm5Val, err := strconv.ParseInt(imm5, 10, 5)
				if err != nil {
					fmt.Println("Error processing imm5 ", line)
				}
				instruction |= 1 << 5
				instruction |= uint16(imm5Val)
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

			operand := strings.Split(items[1], " ")
			instruction |= uint16(table[operand[0]] - offset)

			fmt.Printf("%04x:%04x ; %s\n", currentPC, instruction, line)
			memory[currentPC] = instruction
			currentPC++
			offset++

		}
	}
	return
}
