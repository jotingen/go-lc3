.ORIG x0200
LD R0,DISPLAY
AND R1,R1,#0
AND R2,R2,#0
AND R3,R3,#0
AND R4,R4,#0
AND R5,R5,#0
AND R6,R6,#0
AND R7,R7,#0
STR R1,R0,#0
STR R2,R0,#1
STR R3,R0,#2
STR R4,R0,#3
STR R5,R0,#4
STR R6,R0,#5
STR R7,R0,#6
REPEAT
NOT R1,R1,#0
NOT R2,R2,#0
NOT R3,R3,#0
NOT R4,R4,#0
NOT R5,R5,#0
NOT R6,R6,#0
NOT R7,R7,#0
STR R1,R0,#0
STR R2,R0,#1
STR R3,R0,#2
STR R4,R0,#3
STR R5,R0,#4
STR R6,R0,#5
STR R7,R0,#6
BR REPEAT
HALT
DISPLAY .FILL 0xC000