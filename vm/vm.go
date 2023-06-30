package vm

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"golang.org/x/term"
)

type VM struct {
	*memory
	*cpu
	err        error
	keysBuffer chan byte
	*term.Terminal
	running bool
}

func New(trm *term.Terminal) *VM {
	return &VM{
		running: true,
		memory:  &memory{}, cpu: &cpu{cond: ZERO},
		keysBuffer: make(chan byte, 10), Terminal: trm,
	}
}

func (vm *VM) Run() error {
	for vm.next() {
		vm.processKbd()
		vm.execute(vm.popInstr())
	}

	return vm.err
}

func (vm *VM) next() bool {
	return vm.running
}

func (vm *VM) Stop() {
	vm.running = false
}

func (vm *VM) execute(instr word) {
	switch op := instr.op(); op {
	case BR: // branch
		if condFlag := instr.dst(); condFlag&vm.cond > 0 {
			vm.pc += instr.pcOfs()
		}
	case ADD: // add
		vm.Set(instr, func(a, b word) word { return a + b })
	case LD: // load
		pcOfs := instr.pcOfs()
		vm.SetR(instr.dst(), vm.MemRead(vm.pc+pcOfs))
	case JSR: // jump register
		vm.data[7] = vm.pc
		if longFlag := (instr >> 11) & 1; longFlag > 0 {
			longPcOfs := (instr & 0x7FF).extendSign(11)
			vm.pc += longPcOfs // JSR
		} else {
			vm.pc = vm.Val1(instr) // JSRR
		}
	case AND: // bitwise and
		vm.Set(instr, func(a, b word) word { return a & b })
	case LDR: // load register
		ofs := (instr & 0x3F).extendSign(6)
		vm.SetR(instr.dst(), vm.MemRead(vm.Val1(instr)+ofs))
	case NOT: // bitwise not
		vm.SetR(instr.dst(), ^vm.Val1(instr))
	case LDI: // load indirect
		pcOfs := instr.pcOfs()
		vm.SetR(instr.dst(), vm.MemRead(vm.MemRead(vm.pc+pcOfs)))
	case ST: // store
		pcOfs := instr.pcOfs()
		vm.MemWrite(vm.pc+pcOfs, vm.Val0(instr))
	case STI: // store indirect
		pcOfs := instr.pcOfs()
		vm.MemWrite(vm.MemRead(vm.pc+pcOfs), vm.Val0(instr))
	case STR: // store register
		ofs := (instr & 0x3F).extendSign(6)
		vm.MemWrite(vm.Val1(instr)+ofs, vm.Val0(instr))
	case JMP: // jump
		vm.pc = vm.Val1(instr)
	case LEA: // load effective address
		pcOfs := instr.pcOfs()
		vm.SetR(instr.dst(), vm.pc+pcOfs)
	case TRAP: // execute trap
		vm.data[7] = vm.pc

		switch trap := instr & 0xFF; trap {
		case TrapGetc: // get character from keyboard, not echoed onto the terminal
			for len(vm.keysBuffer) == 0 && vm.running {
				time.Sleep(time.Millisecond)
			}

			vm.SetR(0, word(<-vm.keysBuffer))
		case TrapOut: // output a character
			if _, vm.err = vm.Write([]byte{byte(vm.R(0))}); vm.err != nil {
				vm.Stop()
			}
		case TrapIn: // get character from keyboard, echoed onto the terminal
			vm.WriteString("Enter a character: ")

			c := <-vm.keysBuffer

			vm.WriteString(string(rune(c)))
			vm.SetR(instr.dst(), word(c))
		case TrapPuts: // output a byte string
			vm.puts(loByte)
		case TrapPutsp: // output a word string
			vm.puts(twoByte)
		case TrapHalt: // halt the program
			// vm.WriteString("HALT\n")
			vm.Stop()
		default:
			vm.err = BadTrapError(trap)
			vm.Stop()
		}
	default:
		vm.err = BadOpError(op)
		vm.Stop()
	}
}

func (vm *VM) popInstr() word {
	w := vm.MemRead(vm.pc)
	vm.pc++

	return w
}

var hackTopBottomLine = 0

func dec(s string) string {
	return fmt.Sprintf("\x1b(0%s\x1b(B", s)
}

func hackDrawingBox(p []byte) []byte {
	if bytes.Contains(p, []byte{'|'}) {
		return []byte(strings.ReplaceAll(string(p), "|", "\x1b(0x\x1b(B"))
	} else if bytes.Contains(p, []byte{'+'}) {
		s := strings.ReplaceAll(string(p), "-", dec("q"))
		if hackTopBottomLine%2 == 0 {
			s = strings.Replace(s, "+", dec("l"), 1)
			s = strings.Replace(s, "+", dec("k"), 1)
		} else {
			s = strings.Replace(s, "+", dec("m"), 1)
			s = strings.Replace(s, "+", dec("j"), 1)
		}
		hackTopBottomLine++

		return []byte(s)
	} else {
		return p
	}
}

func (vm *VM) puts(fn func(word) []byte) {
	address, buf := vm.R(0), []byte{}
	for ok, c, i := true, word(0), word(0); ok; ok, i = (c != 0), i+1 {
		c = vm.MemRead(address + i)
		buf = append(buf, fn(c)...)
	}

	buf = hackDrawingBox(buf)

	if _, vm.err = vm.Write(buf); vm.err != nil {
		vm.Stop()
	}
}

func (vm *VM) WriteString(s string) {
	if _, vm.err = vm.Write([]byte(s)); vm.err != nil {
		vm.Stop()
	}
}
