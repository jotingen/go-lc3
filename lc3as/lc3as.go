package lc3as

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func Assemble(assembly []string) (pc uint16, memory [65536]uint16) {
	table := make(map[string]string)

	//First pass, built table
	for _, line := range assembly {
		items := strings.Split(line, " ")
		switch items[0] {
		case ".ORIG":
			table[items[0]] = items[1]
		}
	}

	//If ORIG was not defined, assume x3000
	if _, ok := table[".ORIG"]; !ok {
		table[".ORIG"] = "x3000"
	}

	//Process and set PC
	reHex := regexp.MustCompile(`^0?x([0-9A-Fa-f]+)$`)
	pcHex := reHex.FindAllStringSubmatch(table[".ORIG"], -1)[0][1]
	pcUint, err := strconv.ParseUint(pcHex, 16, 16)
	if err != nil {
		fmt.Println("Error processing .ORIG ", table[".ORIG"])
	}
	pc = uint16(pcUint)

	//Second pass
	currentPC := pc
	reReg := regexp.MustCompile(`^R(\d)$`)
	reNum := regexp.MustCompile(`^#(-?\d+)$`)
	for _, line := range assembly {
		instruction := uint16(0)
		items := strings.Split(line, " ")
		op := items[0]
		switch op {

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
		}
	}
	return
}
