.ORIG x3000

;Initialize stack
LD R5, STACK
LD R6, STACK

;Clear Buffer
JSR CLEARBUFFER

;Randomize Buffer
JSR RANDOMIZEBUFFER

;Load Buffer
JSR LOADBUFFER

REPEAT
	;Clear Buffer
	JSR CLEARBUFFER

	;Life
	JSR LIFE

	;Load Buffer
	JSR LOADBUFFER

	;Just Repeat
	BR REPEAT


HALT

STACK   .FILL 0x5000 ;make a stack here, Room for 20480 items until we hit the display buffer
BUFFER  .FILL 0x8200 ;buffered display address
DISPLAY .FILL 0xC000 ;display address
PIXELS  .FILL 0x3E00 ;15872 pixels

;;;; LIFE ;;;;
; 
; Step through life
;

LIFE

	;Load buffer display into display space
	LD R2,DISPLAY ;Display address stored in R2
	LD R3,BUFFER  ;Buffer address stored in R3
	LD R4,PIXELS  ;Total number of pixels stored in R4
	ADD R4,R3,R4  ;End buffer address stored in R4

LIFE_START   	
        ;R0 is the neighbor count
	;R1 is the neighbor being checked
	;R2 is the current cell address in the display
	;R3 is the new cell address in the buffer

	;Initialize R0 to be the neighbor count
	AND R0, R0, #0

        ;Check x,y offsets, where 0,0 is top left
        ;Check -1,-1
	LD R1, LIFE_N128
        ADD R1, R1, R2
        LDR R1, R1, #-1  ;-129
        BRnp LIFE_NO_N1_N1
        ADD R0,R0,#1
        LIFE_NO_N1_N1

        ;Check  0,-1
        LDR R1, R2, #-1
        BRnp LIFE_NO_0_N1
        ADD R0,R0,#1
        LIFE_NO_0_N1

        ;Check  1,-1
	LD R1, LIFE_128
        ADD R1, R1, R2
        LDR R1, R1, #-1  ; 127
        BRnp LIFE_NO_1_N1
        ADD R0,R0,#1
        LIFE_NO_1_N1

        ;Check -1, 0
	LD R1, LIFE_N128
        ADD R1, R1, R2
        LDR R1, R1, #0  ;-128
        BRnp LIFE_NO_N1_0
        ADD R0,R0,#1
        LIFE_NO_N1_0

        ;Check  1, 0
	LD R1, LIFE_128
        ADD R1, R1, R2
        LDR R1, R1, #0  ; 128
        BRnp LIFE_NO_1_0
        ADD R0,R0,#1
        LIFE_NO_1_0

        ;Check -1, 1
	LD R1, LIFE_N128
        ADD R1, R1, R2
        LDR R1, R1, #1  ;-127
        BRnp LIFE_NO_N1_1
        ADD R0,R0,#1
        LIFE_NO_N1_1

        ;Check  0, 1
        LDR R1, R2, #1
        BRnp LIFE_NO_0_1
        ADD R0,R0,#1
        LIFE_NO_0_1

        ;Check  1, 1
	LD R1, LIFE_128
        ADD R1, R1, R2
        LDR R1, R1, #1  ; 129
        BRnp LIFE_NO_1_1
        ADD R0,R0,#1
        LIFE_NO_1_1

        ;R1 is now the cell being checked
        LDR R1, R2, #0
        BRnp LIFE_NOT_ALIVE
        	;Cell is alive
        	ADD R1,R0,#-2
        	BRn LIFE_DONE ;Cell had less than 2 neighbors, dies
        	ADD R1,R0,#-3
        	BRp LIFE_DONE ;Cell had more than 3 neighbors, dies

        	;Send new cell to buffer       
        	AND R1,R1,#0
        	STR R1,R3,#0 
        	BR LIFE_DONE

        LIFE_NOT_ALIVE
        	;Cell is not alive
        	ADD R1,R0,#-3
        	BRnp LIFE_DONE ;Cell had more or less than 3 neighbors, stays dead

        	;Send new cell to buffer       
        	AND R1,R1,#0
        	STR R1,R3,#0 

	LIFE_DONE

	;Increment display address
	ADD R2,R2,#1 
	;Increment buffer address
	ADD R3,R3,#1 

	;Determine if we are at the last pixel
	NOT R1,R3    ;Subtract current address from max address
	ADD R1,R1,#1
	ADD R1,R1,R4
	BRp LIFE_START    ;Repeat until current address > max address

	RET     

	LIFE_128  .FILL   0x0080 ;  128
	LIFE_N128 .FILL   0xFF80 ; -128


;;;; CLEAR BUFFER ;;;;
; 
; Clear the buffer space
;

CLEARBUFFER

	;Save registers
	ADD R6,R6,#1
	STR R0, R6, #0
	ADD R6,R6,#1
	STR R1, R6, #0
	ADD R6,R6,#1
	STR R2, R6, #0
	ADD R6,R6,#1
	STR R3, R6, #0
	ADD R6,R6,#1
	STR R4, R6, #0
	ADD R6,R6,#1
	STR R5, R6, #0

	LD R4,BUFFER  ;Buffer address stored in R4
	LD R5,PIXELS  ;Total number of pixels stored in R5
	ADD R5,R4,R5  ;End buffer address stored in R5

	AND R0,R0,#0
	NOT R0,R0

	CL_START
		;Send clear pixel       
		STR R0,R4,#0 

		;Increment display address
		ADD R4,R4,#1 

		;Determine if we are at the last pixel
		NOT R2,R4    ;Subtract current address from max address
		ADD R2,R2,#1
		ADD R2,R2,R5
		BRp CL_START    ;Repeat until current address > max address

	;Restore registers
	LDR R5, R6, #0
	ADD R6,R6,#-1
	LDR R4, R6, #0
	ADD R6,R6,#-1
	LDR R3, R6, #0
	ADD R6,R6,#-1
	LDR R2, R6, #0
	ADD R6,R6,#-1
	LDR R1, R6, #0
	ADD R6,R6,#-1
	LDR R0, R6, #0
	ADD R6,R6,#-1

	RET     



;;;; RANDOMIZE BUFFER ;;;;
; 
; Randomize the buffer space
;

RANDOMIZEBUFFER

	;Save registers
	ST R1, RB_SAVE_R1
	ST R2, RB_SAVE_R2
	ST R3, RB_SAVE_R3
	ST R4, RB_SAVE_R4
	ST R5, RB_SAVE_R5
	ST R6, RB_SAVE_R6
	ST R7, RB_SAVE_R7

	LD R4,BUFFER  ;Buffer address stored in R4
	LD R5,PIXELS  ;Total number of pixels stored in R5
	ADD R5,R4,R5  ;End buffer address stored in R5

	LD  R0,CLK2 ;Load clock to "randomly" seed LFSR

	RB_START
		;Determine whether or not to create cell based on mask of LFSR
		LD R1,MASK
		AND R1,R1,R0
		BRz MASK_WAS_0
		AND R1,R1,#0
		NOT R1,R1
		MASK_WAS_0

		;Send pixel       
		STR R1,R4,#0 

		;Update LFSR for next pixel
		JSR LFSR     

		;Increment display address
		ADD R4,R4,#1 

		;Determine if we are at the last pixel
		NOT R2,R4    ;Subtract current address from max address
		ADD R2,R2,#1
		ADD R2,R2,R5
		BRp RB_START    ;Repeat until current address > max address

	;Restore registers
	LD R1, RB_SAVE_R1
	LD R2, RB_SAVE_R2
	LD R3, RB_SAVE_R3
	LD R4, RB_SAVE_R4
	LD R5, RB_SAVE_R5
	LD R6, RB_SAVE_R6
	LD R7, RB_SAVE_R7

	RET     

	MASK .FILL 0x0010
	CLK2	.FILL 0xFE0E		; clock 1 register

	; Used to save and restore registers
	RB_SAVE_R0 .FILL x0000
	RB_SAVE_R1 .FILL x0000
	RB_SAVE_R2 .FILL x0000
	RB_SAVE_R3 .FILL x0000
	RB_SAVE_R4 .FILL x0000
	RB_SAVE_R5 .FILL x0000
	RB_SAVE_R6 .FILL x0000
	RB_SAVE_R7 .FILL x0000


;;;; LOAD BUFFER ;;;;
; 
; Loads the buffer for the display
;

LOADBUFFER

	;Save registers
	ST R1, LB_SAVE_R1
	ST R2, LB_SAVE_R2
	ST R3, LB_SAVE_R3
	ST R4, LB_SAVE_R4
	ST R5, LB_SAVE_R5
	ST R6, LB_SAVE_R6
	ST R7, LB_SAVE_R7
                
	;Load buffer display into display space
	LD R3,DISPLAY ;Display address stored in R3
	LD R4,BUFFER  ;Buffer address stored in R4
	LD R5,PIXELS  ;Total number of pixels stored in R5
	ADD R5,R4,R5  ;End buffer address stored in R5

LB_START   	
	LDR R2,R4,#0  
	STR R2,R3,#0  
	ADD R3,R3,#1
	ADD R4,R4,#1
	
	;Determine if we are at the last pixel
	NOT R2,R4    ;Subtract current address from max address
	ADD R2,R2,#1
	ADD R2,R2,R5
	BRp LB_START    ;Repeat until current address > max address

	;Set VCR to indicate buffer is ready
	LD  R1,VCR
	LD  R2,VCR_MASK
	STR R2,R1,#0	

	;Wait until display indicates buffer is taken
	LB_POLL_VCR
		LDI R1,VCR
		LD  R2,VCR_MASK
		NOT R2,R2
		AND R2,R2,R1
		BRnp LB_POLL_VCR

	;Restore registers
	LD R1, LB_SAVE_R1
	LD R2, LB_SAVE_R2
	LD R3, LB_SAVE_R3
	LD R4, LB_SAVE_R4
	LD R5, LB_SAVE_R5
	LD R6, LB_SAVE_R6
	LD R7, LB_SAVE_R7

	RET     

	VCR	.FILL xFE14		; video control register
	VCR_MASK .FILL 0x8000

	; Used to save and restore registers
	LB_SAVE_R0 .FILL x0000
	LB_SAVE_R1 .FILL x0000
	LB_SAVE_R2 .FILL x0000
	LB_SAVE_R3 .FILL x0000
	LB_SAVE_R4 .FILL x0000
	LB_SAVE_R5 .FILL x0000
	LB_SAVE_R6 .FILL x0000
	LB_SAVE_R7 .FILL x0000

;;;; LSFR ;;;;
; 
; R0 = (R0<<1)+(R0[15] ^ R0[14] ^ R0[12] ^ R0[3])
;
; IN:  R0 
; OUT: R0 

LFSR
	;Save registers
	ST R1, LFSR_SAVE_R1
	ST R2, LFSR_SAVE_R2
	ST R3, LFSR_SAVE_R3
	ST R4, LFSR_SAVE_R4
	ST R5, LFSR_SAVE_R5
	ST R6, LFSR_SAVE_R6
	ST R7, LFSR_SAVE_R7

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
	
	ADD R5,R6,#0
	STR R1,R6,#0
	STR R2,R6,#1
	ADD R6,R6,#2 ;Inputs
	ADD R6,R6,#1 ;Return
	JSR XOR
	LDR R0,R6,#-1 ;Return
	ADD R6,R6,#-1
	ADD R6,R6,#-2 ;Inputs

	ADD R1,R0,#0

	;Bitmask for 12
	LD R2,LFSR_BITMASK_12
	AND R2,R2,R3
	BRz BIT12_WAS_0
	AND R2,R2,#0
	ADD R2,R2,#1
	BIT12_WAS_0
	
	STR R1,R6,#0
	STR R2,R6,#1
	ADD R6,R6,#2 ;Inputs
	ADD R6,R6,#1 ;Return
	JSR XOR
	LDR R0,R6,#-1 ;Return
	ADD R6,R6,#-1
	ADD R6,R6,#-2 ;Inputs

	ADD R1,R0,#0

	;Bitmask for 3
	LD R2,LFSR_BITMASK_03
	AND R2,R2,R3
	BRz BIT03_WAS_0
	AND R2,R2,#0
	ADD R2,R2,#1
	BIT03_WAS_0

	STR R1,R6,#0
	STR R2,R6,#1
	ADD R6,R6,#2 ;Inputs
	ADD R6,R6,#1 ;Return
	JSR XOR
	LDR R0,R6,#-1 ;Return
	ADD R6,R6,#-1
	ADD R6,R6,#-2 ;Inputs

	;R3 = R3 << 1
	ADD R3,R3,R3

	;Add shifted register to feedback
	ADD R0,R0,R3

	;Restore registers
	LD R1, LFSR_SAVE_R1
	LD R2, LFSR_SAVE_R2
	LD R3, LFSR_SAVE_R3
	LD R4, LFSR_SAVE_R4
	LD R5, LFSR_SAVE_R5
	LD R6, LFSR_SAVE_R6
	LD R7, LFSR_SAVE_R7

	RET     

	LFSR_BITMASK_15 .FILL x8000
	LFSR_BITMASK_14 .FILL x4000
	LFSR_BITMASK_12 .FILL x1000
	LFSR_BITMASK_03 .FILL x0008

	; Used to save and restore registers
	LFSR_SAVE_R0 .FILL x0000
	LFSR_SAVE_R1 .FILL x0000
	LFSR_SAVE_R2 .FILL x0000
	LFSR_SAVE_R3 .FILL x0000
	LFSR_SAVE_R4 .FILL x0000
	LFSR_SAVE_R5 .FILL x0000
	LFSR_SAVE_R6 .FILL x0000
	LFSR_SAVE_R7 .FILL x0000

;;;; XOR ;;;;
; 
; X = A ^ B
; X = !(!(A & !(A & B)) & !(B & !(A & B)))
;
; IN:  A 
; IN:  B 
; OUT: X 

XOR
	;Save callers frame pointer
	STR R5,R6,#0
	ADD R6,R6,#1

	;Set current frame pointer
	ADD R5,R6,#0

	;Save registers 0-5
	STR R0,R6,#0
	STR R1,R6,#1
	STR R2,R6,#2
	STR R3,R6,#3
	STR R4,R6,#4
	ADD R6,R6,#5

	; clear registers to help debug
	AND R0,R0,#0
	AND R1,R1,#0
	AND R2,R2,#0
	AND R3,R3,#0
	AND R4,R4,#0

	;Load A to R0
	LDR R0,R5,#-4

	;Load B to R1
	LDR R1,R5,#-3

	;Put !(A & B) onto R2
	
	AND R2,R0,R1
	NOT R2,R2

	;Put !(A & !(A & B)) onto R0
	AND R0,R0,R2
	NOT R0,R0

	;Put !(B & !(A & B)) onto R1
	AND R1,R1,R2
	NOT R1,R1

	;Put !(!(A & !(A & B)) & !(B & !(A & B))) onto R0
	AND R0,R0,R1
	NOT R0,R0

	;Save result to callers return
	STR R0,R5,#-2


	;Restore registers
	LDR R4,R6,#-1
	LDR R3,R6,#-2
	LDR R2,R6,#-3
	LDR R1,R6,#-4
	LDR R0,R6,#-5
	ADD R6,R6,#-5

	;Restore callers frame pointer
	LDR R5,R6,#-1
	ADD R6,R6,#-1

	RET     
