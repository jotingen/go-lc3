package main

import (
	//"encoding/binary"
	"fmt"
	//"os"
	"time"
)

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

import (
	"github.com/jotingen/go-lc3/asm2obj"
	"github.com/jotingen/go-lc3/lc3"
)

const (
	//Video output is memory-mapped from address location xC000 to xFDFF.
	//The video display is 128 by 124 pixels (15,872 pixels total) and
	//the coordinate system starts from (0,0) at the top left corner of the display
	SCALE = 3
	X     = 128 * SCALE
	Y     = 124 * SCALE
)

var (
	memory [65536]uint16
)

func main() {
	assembly := []string{
		".ORIG x0200",
		"LD R0,DISPLAY",
		"AND R1,R1,#0",
		"AND R2,R2,#0",
		"AND R3,R3,#0",
		"AND R4,R4,#0",
		"AND R5,R5,#0",
		"AND R6,R6,#0",
		"AND R7,R7,#0",
		"STR R1,R0,#0",
		"STR R2,R0,#1",
		"STR R3,R0,#2",
		"STR R4,R0,#3",
		"STR R5,R0,#4",
		"STR R6,R0,#5",
		"STR R7,R0,#6",
		"REPEAT",
		"NOT R1,R1,#0",
		"NOT R2,R2,#0",
		"NOT R3,R3,#0",
		"NOT R4,R4,#0",
		"NOT R5,R5,#0",
		"NOT R6,R6,#0",
		"NOT R7,R7,#0",
		"STR R1,R0,#0",
		"STR R2,R0,#1",
		"STR R3,R0,#2",
		"STR R4,R0,#3",
		"STR R5,R0,#4",
		"STR R6,R0,#5",
		"STR R7,R0,#6",
		"BR REPEAT",
		//"ADD R1,R1,#1",
		//"ADD R1,R2,R3",
		//"ADD R1,R2,R3",
		//"BRnzp SKIP",
		//"ADD R2,R0,R1",
		//"SKIP ADD R3,R0,R1",
		//"ADD R4,R0,R1",
		//"ADD R7,R7,#-5",
		//"LOOP ADD R7,R7,#1",
		//"BRn LOOP",
		//"LD R5,LOOP",
		//"NOT R4,R4",
		//"ADD R4,R4,#1",
		"HALT",
		"DISPLAY .FILL 0xC000",
	}
	memory = processAssembly(assembly)

	pixelgl.Run(run)

}

func processAssembly(assembly []string) (memory [65536]uint16) {
	return asm2obj.Assemble(assembly)
}

func run() {
	r := lc3.Request{}
	data := uint16(0x0000)
	pc := uint16(0x0200)

	cfg := pixelgl.WindowConfig{
		Title:  "Life",
		Bounds: pixel.R(0, 0, X, Y),
		VSync:  true,
	}
	lc3 := lc3.LC3{}
	lc3.Init(pc)
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	imd := imdraw.New(nil)

	fmt.Printf("\n%s\n", lc3)

	cycles := 0
	timeStart := time.Now()
	for { //Breakout when PC reads HALT address
		//Clean Display
		imd.Reset()
		imd.Clear()
		win.Clear(colornames.White)

		//Step through CPU
		pc, r, err = lc3.Step(memory[pc], data)
		if err != nil {
			panic(err)
		}

		//Process memory requests
		if r.Vld {
			if r.RnW {
				if r.Address == 0x0025 {
					break
				}
				data = memory[r.Address]
			} else {
				memory[r.Address] = r.Data
			}
		}

		//Update display
		for y := 0; y < Y/SCALE; y++ {
			for x := 0; x < X/SCALE; x++ {
				addr := 0xC000 + y*0x0080 + x
				imd.Color = pixel.RGB(
					float64((memory[addr]&0x7C00)>>10)/32,
					float64((memory[addr]&0x0380)>>5)/32,
					float64((memory[addr]&0x001F)>>0)/32,
				)
				imd.Push(pixel.V(float64(x*SCALE), float64((Y/SCALE-y-1)*SCALE)))
				imd.Push(pixel.V(float64(x*SCALE+SCALE), float64((Y/SCALE-y-1)*SCALE+SCALE)))
				imd.Rectangle(0)

				//if x < 7 && y == 0 {
				//	fmt.Printf("%d:%d 0x%04x %3.1f:%3.1f:%3.1f %3.1f:%3.1f %3.1f:%3.1f\n", x, y, addr,
				//		float64((memory[addr]&0x7C00)>>10)/32,
				//		float64((memory[addr]&0x0380)>>5)/32,
				//		float64((memory[addr]&0x001F)>>0)/32,
				//		float64(x*SCALE), float64((Y/SCALE-y-1)*SCALE),
				//		float64(x*SCALE+SCALE), float64((Y/SCALE-y-1)*SCALE+SCALE),
				//	)
				//}
			}
		}
		imd.Draw(win)
		win.Update()

		//Update cycle counter
		cycles++
	}
	timeEnd := time.Now()

	fmt.Printf("\n%s\n", lc3)

	nanosecondsPerCycle := float64(timeEnd.Sub(timeStart)) / float64(cycles)
	secondsPerCycle := float64(nanosecondsPerCycle) / 1000.0 / 1000.0 / 1000.0
	hertz := 1 / secondsPerCycle
	fmt.Printf("%dcycles/%s = %1.0fHz\n", cycles, timeEnd.Sub(timeStart), hertz)

	//Just loop so i can see the display
	for {
	}
}
