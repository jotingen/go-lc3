.ORIG x3000

;Initialize Frame Counter
AND R6, R6, #0

;Clear Buffer
JSR CLEARBUFFER

;Randomize Buffer
JSR RANDOMIZEBUFFER

;Load Buffer
JSR LOADBUFFER

REPEAT
	;Increment Frame Counter
	ADD R6, R6, #1

	;Clear Buffer
	JSR CLEARBUFFER

	;Life
	JSR LIFE

	;Load Buffer
	JSR LOADBUFFER

	;Just Repeat
	BR REPEAT
	;ADD R0,R6,#-1
	;BRzn REPEAT


HALT

BUFFER  .FILL 0x8200 ;buffered display address
DISPLAY .FILL 0xC000 ;display address
PIXELS  .FILL 0x3E00 ;15872 pixels

;;;; LIFE ;;;;
; 
; Step through life
;

LIFE

	;Save registers
	ST R1, LIFE_SAVE_R1
	ST R2, LIFE_SAVE_R2
	ST R3, LIFE_SAVE_R3
	ST R4, LIFE_SAVE_R4
	ST R5, LIFE_SAVE_R5
	ST R6, LIFE_SAVE_R6
	ST R7, LIFE_SAVE_R7
                
	;Load buffer display into display space
	LD R3,DISPLAY ;Display address stored in R3
	LD R4,BUFFER  ;Buffer address stored in R4
	LD R5,PIXELS  ;Total number of pixels stored in R5
	ADD R5,R4,R5  ;End buffer address stored in R5

LIFE_START   	
        ;R0 is the neighbor count
	;R1 is the neighbor being checked
	;R2 is scratch for generating the neighbor address
	;R3 is the current cell address in the display
	;R4 is the new cell address in the buffer

	;Initialize R0 to be the neighbor count
	AND R0, R0, #0

        ;Check x,y offsets, where 0,0 is top left
        ;Check -1,-1
	LD R2, LIFE_N128
        ADD R2, R2, R3
        LDR R1, R2, #-1  ;-129
        BRnp LIFE_NO_N1_N1
        ADD R0,R0,#1
        LIFE_NO_N1_N1

        ;Check  0,-1
        LDR R1, R3, #-1
        BRnp LIFE_NO_0_N1
        ADD R0,R0,#1
        LIFE_NO_0_N1

        ;Check  1,-1
	LD R2, LIFE_128
        ADD R2, R2, R3
        LDR R1, R2, #-1  ; 127
        BRnp LIFE_NO_1_N1
        ADD R0,R0,#1
        LIFE_NO_1_N1

        ;Check -1, 0
	LD R2, LIFE_N128
        ADD R2, R2, R3
        LDR R1, R2, #0  ;-128
        BRnp LIFE_NO_N1_0
        ADD R0,R0,#1
        LIFE_NO_N1_0

        ;Check  1, 0
	LD R2, LIFE_128
        ADD R2, R2, R3
        LDR R1, R2, #0  ; 128
        BRnp LIFE_NO_1_0
        ADD R0,R0,#1
        LIFE_NO_1_0

        ;Check -1, 1
	LD R2, LIFE_N128
        ADD R2, R2, R3
        LDR R1, R2, #1  ;-127
        BRnp LIFE_NO_N1_1
        ADD R0,R0,#1
        LIFE_NO_N1_1

        ;Check  0, 1
        LDR R1, R3, #1
        BRnp LIFE_NO_0_1
        ADD R0,R0,#1
        LIFE_NO_0_1

        ;Check  1, 1
	LD R2, LIFE_128
        ADD R2, R2, R3
        LDR R1, R2, #1  ; 129
        BRnp LIFE_NO_1_1
        ADD R0,R0,#1
        LIFE_NO_1_1

        ;R1 is now the cell being checked
        LDR R1, R3, #0
        BRnp LIFE_NOT_ALIVE
        	;Cell is alive
        	ADD R2,R0,#-2
        	BRn LIFE_DONE ;Cell had less than 2 neighbors, dies
        	ADD R2,R0,#-3
        	BRp LIFE_DONE ;Cell had more than 3 neighbors, dies

        	;Send new cell to buffer       
        	AND R2,R2,#0
        	STR R2,R4,#0 
        	BR LIFE_DONE

        LIFE_NOT_ALIVE
        	;Cell is not alive
        	ADD R2,R0,#-3
        	BRnp LIFE_DONE ;Cell had more or less than 3 neighbors, stays dead

        	;Send new cell to buffer       
        	AND R2,R2,#0
        	STR R2,R4,#0 

	LIFE_DONE

	;Increment display address
	ADD R3,R3,#1 
	;Increment buffer address
	ADD R4,R4,#1 

	;Determine if we are at the last pixel
	NOT R2,R4    ;Subtract current address from max address
	ADD R2,R2,#1
	ADD R2,R2,R5
	BRp LIFE_START    ;Repeat until current address > max address

	;Restore registers
	LD R1, LIFE_SAVE_R1
	LD R2, LIFE_SAVE_R2
	LD R3, LIFE_SAVE_R3
	LD R4, LIFE_SAVE_R4
	LD R5, LIFE_SAVE_R5
	LD R6, LIFE_SAVE_R6
	LD R7, LIFE_SAVE_R7

	RET     

	LIFE_128  .FILL   0x0080 ;  128
	LIFE_N128 .FILL   0xFF80 ; -128

	; Used to save and restore registers
	LIFE_SAVE_R0 .FILL x0000
	LIFE_SAVE_R1 .FILL x0000
	LIFE_SAVE_R2 .FILL x0000
	LIFE_SAVE_R3 .FILL x0000
	LIFE_SAVE_R4 .FILL x0000
	LIFE_SAVE_R5 .FILL x0000
	LIFE_SAVE_R6 .FILL x0000
	LIFE_SAVE_R7 .FILL x0000

;;;; CLEAR BUFFER ;;;;
; 
; Clear the buffer space
;

CLEARBUFFER

	;Save registers
	ST R1, CL_SAVE_R1
	ST R2, CL_SAVE_R2
	ST R3, CL_SAVE_R3
	ST R4, CL_SAVE_R4
	ST R5, CL_SAVE_R5
	ST R6, CL_SAVE_R6
	ST R7, CL_SAVE_R7

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
	LD R1, CL_SAVE_R1
	LD R2, CL_SAVE_R2
	LD R3, CL_SAVE_R3
	LD R4, CL_SAVE_R4
	LD R5, CL_SAVE_R5
	LD R6, CL_SAVE_R6
	LD R7, CL_SAVE_R7

	RET     

	; Used to save and restore registers
	CL_SAVE_R0 .FILL x0000
	CL_SAVE_R1 .FILL x0000
	CL_SAVE_R2 .FILL x0000
	CL_SAVE_R3 .FILL x0000
	CL_SAVE_R4 .FILL x0000
	CL_SAVE_R5 .FILL x0000
	CL_SAVE_R6 .FILL x0000
	CL_SAVE_R7 .FILL x0000


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
; R0 = R1 ^ R2
;
; IN:  R1 
; IN:  R2 
; OUT: R0 

XOR
	;Save registers
	ST R1, XOR_SAVE_R1
	ST R2, XOR_SAVE_R2
	ST R3, XOR_SAVE_R3
	ST R4, XOR_SAVE_R4
	ST R5, XOR_SAVE_R5
	ST R6, XOR_SAVE_R6
	ST R7, XOR_SAVE_R7

	AND R3,R1,R2
	NOT R3,R3
	AND R4,R1,R3
	NOT R4,R4
	AND R5,R2,R3
	NOT R5,R5
	AND R0,R4,R5
	NOT R0,R0

	;Restore registers
	LD R1, XOR_SAVE_R1
	LD R2, XOR_SAVE_R2
	LD R3, XOR_SAVE_R3
	LD R4, XOR_SAVE_R4
	LD R5, XOR_SAVE_R5
	LD R6, XOR_SAVE_R6
	LD R7, XOR_SAVE_R7

	RET     

	; Used to save and restore registers
	XOR_SAVE_R0 .FILL x0000
	XOR_SAVE_R1 .FILL x0000
	XOR_SAVE_R2 .FILL x0000
	XOR_SAVE_R3 .FILL x0000
	XOR_SAVE_R4 .FILL x0000
	XOR_SAVE_R5 .FILL x0000
	XOR_SAVE_R6 .FILL x0000
	XOR_SAVE_R7 .FILL x0000
