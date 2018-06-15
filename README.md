# go-lc3

LC-3 implemented in Go

Instruction set defined in Appendix A at http://highered.mheducation.com/sites/0072467509/student_view0/appendices_a__b__c__d____e.html
Additional festures defined at http://people.cs.georgetown.edu/~squier/Teaching/HardwareFundamentals/LC3-trunk/docs/LC3-uArch-extended.html

Specifications
These are cobbled together from sources online, in an attempt to keep compatability with whatever sample programs are available
(+) Indicates features that seem to be above and beyond what is described in the textbook associated with the LC3
(-) Indicates a currently unimplemented feature

Memory Map:
x0000 - x00FF  Trap Vector Table
x0100 - x01FF  Interrupt Vector Table
x0200 - x2FFF  Operating System
x3000 - xBFFF  User code & stack
xC000 - xFDFF  Video memory
               Video display can also be accessed via memory addresses xC000 to xFDFF. 
               The display is 128x124 pixels (15,872 pixels)
               The coordinate system starts from (0,0) at the top left corner of the display.
               Since each row is 128 pixels long, in order to find the location exactly one row 
               below a given location, add x0080 to it. 
               As a general rule, this is the formula to find the memory location associated with a given (row, col):
               addr = xC000 + row\*x0080 + col
               Each VRAM location represents one pixel, which means that the value it contains must be 
               formatted as a pixel would be (i.e. RGB format):
               [15]    - Unused
               [14:10] - Red
               [9:5]   - Green
               [4:0]   - Blue
xFE00 - xFFFF  Device register addresses


Devices And Registers:
xFE00  - KBSR  Keyboard Status Register: when KBSR[15] is 1, the keyboard has received a new character.
xFE02  - KBDR  Keyboard Data Register: when a new character is available, KBSR[7:0] contains the ASCII value of the typed character.
xFE04  - DSR   Display Status Register: when DSR[15] is 1, the display is ready to receive a new character to display.
xFE06  - DDR   Display Data Register: when the display is ready, the display will print the ASCII character contained in DDR[7:0].
xFE08 +  TMR   Timer Register: TMR[15] is 1 if the timer has gone off, and 0 otherwise.
xFE0A +  TMI   Timer Interval Register: the number of milliseconds between timer ticks. 
               Setting this to 0 disables the timer, and setting it to 1 or more sets the timer.
               The emulator will use the host's clock to control updates to TMR
xFE0C +  CLK1  Precision unit of Unix Epoch time, 1/(2^15) = ~30.5 microseconds
xFE0E +  CLK2  Unix Epoch Time, bits [15:0]
xFE10 +  CLK3  Unix Epoch Time, bits [31:16]
               Provides a real world time value, based off of the unix epoch
xFE12 +- MPR   Memory Protection Register
               Defines if memory range can be accessed in user mode
               1:User access allowed 0:Only Superuser access allowed
               [0]  - x0000-x0FFF
               [1]  - x1000-x1FFF
               [2]  - x2000-x2FFF
               ...
               [15] - xF000-xFFFF
xFE14 +  VCR   Video Control Register
               Sync bit for the video display.  The user program can set bit [15] to 1 when it is done writing to 
               the video memory.  The display controller will buffer the memory, and then set this bit back to 0 when done capturing.
               The program can poll this bit and start updating the video memory once it clears.
               If the display sees this bit as cleared, it will wait until the next frame to try again.
               The goal is to avoid tearing.
xFFFE  - MCR   Machine Control Register
               [15] - Clock Enable
               [14] - Timer Interrupt Enable
               [13:0] - cycle interval between timer interrupts
xFFFF +  MCC   Machine Cycle Counter
               Value is incremented at every clock cycle
