.ORIG x3000

;Initialize register values for test
LD R4,DISPLAY ;Address stored in R4
LD R5,PIXELS  ;Total number of pixels stored in R5
ADD R5,R4,R5  ;End address stored in R5

AND R0,R0,#0
ADD R0,R0,#9 ;LFSR

START
	STR R0,R4,#0 ;Send pixel       
	JSR LFSR     ;Update LFSR for next pixel
	ADD R4,R4,#1 ;Increment display address
        JSR START
END
	
;JSR XOR

HALT

DISPLAY .FILL 0xC000 ;starting address
PIXELS  .FILL 0x3E00 ;15872 pixels

;;;; LSFR ;;;;
; 
; R0 = (R0<<1)+(R0[15] ^ R0[14] ^ R0[12] ^ R0[3])
;
; IN:  R0 
; OUT: R0 

LFSR
	;Save registers
	STI R1, SAVE_R1
	STI R2, SAVE_R2
	STI R3, SAVE_R3
	STI R4, SAVE_R4
	STI R5, SAVE_R5
	STI R6, SAVE_R6
	STI R7, SAVE_R7

	;R3 = R0
	ADD R3,R0,#0

	;Bitmask for 15
	LD R1,LFSR_BITMASK_15
	AND R1,R1,R3
	BRz BIT15_WAS_0
	AND R1,R1,#0
	ADD R1,R1,#1
	BIT15_WAS_0
	
	;Bitmask for 14
	LD R2,LFSR_BITMASK_14
	AND R2,R2,R3
	BRz BIT14_WAS_0
	AND R2,R2,#0
	ADD R2,R2,#1
	BIT14_WAS_0
	
	JSR XOR

	ADD R1,R0,#0

	;Bitmask for 12
	LD R2,LFSR_BITMASK_12
	AND R2,R2,R3
	BRz BIT12_WAS_0
	AND R2,R2,#0
	ADD R2,R2,#1
	BIT12_WAS_0
	
	JSR XOR

	ADD R1,R0,#0

	;Bitmask for 3
	LD R2,LFSR_BITMASK_03
	AND R2,R2,R3
	BRz BIT03_WAS_0
	AND R2,R2,#0
	ADD R2,R2,#1
	BIT03_WAS_0

	JSR XOR

	;R3 = R3 << 1
	ADD R3,R3,R3

	;Add shifted register to feedback
	ADD R0,R0,R3

	;Restore registers
	LDI R1, SAVE_R1
	LDI R2, SAVE_R2
	LDI R3, SAVE_R3
	LDI R4, SAVE_R4
	LDI R5, SAVE_R5
	LDI R6, SAVE_R6
	LDI R7, SAVE_R7

	RET     
	LFSR_BITMASK_15 .FILL x8000
	LFSR_BITMASK_14 .FILL x4000
	LFSR_BITMASK_12 .FILL x1000
	LFSR_BITMASK_03 .FILL x0008

;;;; XOR ;;;;
; 
; R0 = R1 ^ R2
;
; IN:  R1 
; IN:  R2 
; OUT: R0 

XOR
	;Save registers
	STI R3, SAVE_R3
	STI R4, SAVE_R4
	STI R5, SAVE_R5

	AND R3,R1,R2
	NOT R3,R3
	AND R4,R1,R3
	NOT R4,R4
	AND R5,R2,R3
	NOT R5,R5
	AND R0,R4,R5
	NOT R0,R0

	;Restore registers
	LDI R3, SAVE_R3
	LDI R4, SAVE_R4
	LDI R5, SAVE_R5

	RET     

; Used to save and restore registers
SAVE_R0 .FILL x3500
SAVE_R1 .FILL x3501
SAVE_R2 .FILL x3502
SAVE_R3 .FILL x3503
SAVE_R4 .FILL x3504
SAVE_R5 .FILL x3505
SAVE_R6 .FILL x3506
SAVE_R7 .FILL x3507
