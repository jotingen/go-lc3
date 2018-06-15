package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"
)

import (
	"github.com/dustin/go-humanize"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/golang/glog"
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
	in string

	memory []uint16
)

func main() {
	flag.StringVar(&in, "i", "", "Input assembly file")
	flag.Lookup("log_dir").Value.Set(".")
	flag.Parse()

	memory = processAssembly(in)

	pixelgl.Run(run)

	glog.Flush()
}

func processAssembly(file string) (memory []uint16) {
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

func run() {
	pc := uint16(0x0200)

	cfg := pixelgl.WindowConfig{
		Title:  "L3C",
		Bounds: pixel.R(0, 0, X, Y),
		VSync:  true,
	}
	lc3 := lc3.LC3{}
	lc3.Init(pc, memory)
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	imd := imdraw.New(nil)

	fmt.Printf("\n%s\n", lc3)

	go func() {
		cycles := 0
		timeStart := time.Now()
		for { //Breakout when PC reads HALT address
			//Step through CPU
			pc, err = lc3.Step()
			if err != nil {
				panic(err)
			}

			//Process memory requests
			if pc == memory[0x0025] {
				break
			}

			//Update cycle counter
			cycles++

			if cycles%10000000 == 0 {
				glog.Flush()
			}
		}
		timeEnd := time.Now()
		glog.Flush()

		fmt.Printf("\n%s\n", lc3)

		nanosecondsPerCycle := float64(timeEnd.Sub(timeStart)) / float64(cycles)
		secondsPerCycle := float64(nanosecondsPerCycle) / 1000.0 / 1000.0 / 1000.0
		hertz := 1 / secondsPerCycle
		siVal, siPrefix := humanize.ComputeSI(hertz)
		fmt.Printf("%dcycles/%s = %4.2f%sHz\n", cycles, timeEnd.Sub(timeStart), siVal, siPrefix)
	}()

	for !win.Closed() {
		if (memory[0xFE14]&0x8000)>>15 == 1 {
			if glog.V(2) {
				glog.Info("Updating display")
			}

			//Clean Display
			imd.Reset()
			imd.Clear()
			win.Clear(colornames.White)

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
			memory[0xFE14] = 0x7FFF
		}

	}
	////Just loop so i can see the display
	//for {
	//}
}
