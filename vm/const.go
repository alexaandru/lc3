package vm

const (
	MaxWord = 1<<16 - 1
	MemMax  = MaxWord + 1
	halt    = word(0xF025)
)

const (
	BR word = iota
	ADD
	LD
	ST
	JSR
	AND
	LDR
	STR
	_ // RTI
	NOT
	LDI
	STI
	JMP
	_ // RES
	LEA
	TRAP
)

const (
	POS word = 1 << iota
	ZERO
	NEG
)

const (
	TrapGetc word = iota + 0x20
	TrapOut
	TrapPuts
	TrapIn
	TrapPutsp
	TrapHalt
)

const (
	KbdStatus word = 0xFE00
	KbdData   word = 0xFE02
)
