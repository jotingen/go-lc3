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
	term "github.com/nsf/termbox-go"
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

	//SCALE is the scqling factor fo the window
	SCALE = 3
	//X is the width in pixels of the window, after scaling
	X = 128 * SCALE
	//Y is the height in pixels of the window, after scaling
	Y = 124 * SCALE
)

var (
	in string

	memory []uint16

	targetSpeed float64 = 2000000 //Hz

	cycles int64 = 0

	fps = time.Tick(time.Second / 120)
)

//LC3 is a wrapper around the lc3 model to allow for additional methods to be added
type LC3 struct {
	*lc3.LC3
}

func main() {
	flag.StringVar(&in, "i", "", "Input assembly file")
	err := flag.Lookup("log_dir").Value.Set(".")
	if err != nil {
		panic(err)
	}
	flag.Parse()

	memory = processAssembly(in)
	glog.Infof("%+v", memory)

	pixelgl.Run(run)

	glog.Flush()
}

func reset() {
	err := term.Sync()
	if err != nil {
		panic(err)
	}
}

func processAssembly(file string) (memory []uint16) {
	assembly, err := readLines(file)
	if err != nil {
		panic(err)
	}
	return asm2obj.Assemble(assembly)
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func run() {
	pc := uint16(0x0200)

	//Create Terminal
	err := term.Init()
	if err != nil {
		panic(err)
	}
	defer term.Close()

	//Create Window
	cfg := pixelgl.WindowConfig{
		Title:  "L3C",
		Bounds: pixel.R(0, 0, X, Y),
	}

	lc3 := LC3{&lc3.LC3{}}
	lc3.Init(pc, memory)
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	glog.Infof("\n%s", lc3)

	//L3Sim
	go lc3.sim(win)

	//Keyboard
	go keyboard(win)

	//Terminal window
	go terminal(win, &lc3)

	//Display window
	display(win)
}

//L3Sim
func (lc3 *LC3) sim(win *pixelgl.Window) {

	for { //Breakout when PC reads HALT address

		//Step through CPU
		pc, err := lc3.Step()
		if err != nil {
			panic(err)
		}
		if glog.V(3) {
			glog.Infof("\n%s", lc3)
		}

		//Process memory requests
		if pc == memory[0x0025] {
			break
		}

		//Console
		//When DSR[15] is 0, there is a character ready to print
		if (memory[0xFE04]&0x8000)>>15 == 0 {
			//Print character in DDR[7:0]
			fmt.Printf("%c", rune(uint8(memory[0xFE06])))
			if glog.V(1) {
				glog.Info("Printed char", rune(uint8(memory[0xFE06])))
			}
			//Set DSR[15] to 1 once printed
			memory[0xFE04] = memory[0xFE04] | 0x8000
		}

		//Update cycle counter
		cycles++

		if cycles%10000000 == 0 {
			glog.Flush()
		}

	}
	glog.Flush()

	glog.Infof("\n%s", lc3)

	win.SetClosed(true)
}

//Keyboard
func keyboard(win *pixelgl.Window) {

	//Listen from terminal
	go func() {
	keyPressListenerLoop:
		for !win.Closed() {
			switch ev := term.PollEvent(); ev.Type {
			case term.EventKey:
				switch ev.Key {
				case term.KeyEsc:
					win.SetClosed(true)
					break keyPressListenerLoop
				case term.KeyEnter:
					reset()

					//When KBSR[15] is 0, ready for new keyboard input
					if (memory[0xFE00]&0x8000)>>15 == 0 {
						//Set character into KBDR[7:0]
						memory[0xFE02] = 0x0000 | uint16(uint8('\n'))
						//SET KBSR[15]
						memory[0xFE00] = 0x8000
						if glog.V(1) {
							glog.Info("Recieved key \\n")
						}
					}
				default:
					reset()

					//When KBSR[15] is 0, ready for new keyboard input
					if (memory[0xFE00]&0x8000)>>15 == 0 {
						//Set character into KBDR[7:0]
						memory[0xFE02] = 0x0000 | uint16(uint8(ev.Ch))
						//SET KBSR[15]
						memory[0xFE00] = 0x8000
						if glog.V(1) {
							glog.Info("Recieved key", ev.Ch)
						}
					}
				}
			case term.EventError:
				panic(ev.Err)
			}

			<-fps
		}
	}()

	//Listen from window
	go func() {
		for !win.Closed() {
			if win.JustPressed(pixelgl.KeyEscape) {
				win.SetClosed(true)
			}
			if win.JustPressed(pixelgl.KeyEnter) {
				//When KBSR[15] is 0, ready for new keyboard input
				if (memory[0xFE00]&0x8000)>>15 == 0 {
					//Set character into KBDR[7:0]
					memory[0xFE02] = 0x0000 | uint16(uint8('\n'))
					//SET KBSR[15]
					memory[0xFE00] = 0x8000
					if glog.V(1) {
						glog.Info("Recieved key \\n")
					}
				}
			}
			s := win.Typed()
			if s != "" {
				//When KBSR[15] is 0, ready for new keyboard input
				if (memory[0xFE00]&0x8000)>>15 == 0 {
					//Set character into KBDR[7:0]
					memory[0xFE02] = 0x0000 | uint16(uint8([]rune(s)[0]))
					//SET KBSR[15]
					memory[0xFE00] = 0x8000
					if glog.V(1) {
						glog.Info("Recieved key", []rune(s)[0])
					}
				}
			}

			<-fps
		}
	}()

}

//Terminal window
func terminal(win *pixelgl.Window, lc3 *LC3) {

	cyclesConsoleRefresh := cycles
	timeConsoleRefresh := time.Now()
	for !win.Closed() {
		//Update display

		err := term.Sync()
		if err != nil {
			panic(err)
		}

		termWidth, _ := term.Size()
		buffer := term.CellBuffer()

		sCycles := fmt.Sprintf("%d", cycles)
		for i, c := range sCycles {
			buffer[i].Ch = c
		}

		timeEnd := time.Now()
		nanosecondsPerCycle := float64(timeEnd.Sub(timeConsoleRefresh)) / float64(cycles-cyclesConsoleRefresh)
		secondsPerCycle := nanosecondsPerCycle / 1000.0 / 1000.0 / 1000.0
		hertz := 1 / secondsPerCycle
		siVal, siPrefix := humanize.ComputeSI(hertz)
		sHertz := fmt.Sprintf("%2.0f%sHz\n", siVal, siPrefix)
		for i, c := range sHertz {
			buffer[i+10].Ch = c
		}

		for r := 0; r < 8; r++ {
			for i, c := range fmt.Sprintf("R%d:%04X", r, lc3.Reg[r]) {
				buffer[termWidth*1+r*8+i].Ch = c
			}
		}

		for i, c := range fmt.Sprintf("PC:%04x %s\n", lc3.PC, lc3.PSR) {
			buffer[termWidth*2+i].Ch = c
		}

		err = term.Flush()
		if err != nil {
			panic(err)
		}

		timeConsoleRefresh = time.Now()
		cyclesConsoleRefresh = cycles

		<-fps
	}
}

//Display window
func display(win *pixelgl.Window) {

	imd := imdraw.New(nil)
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
				}
			}
			imd.Draw(win)

			memory[0xFE14] = 0x7FFF
		}
		win.Update()

		<-fps
	}
}
