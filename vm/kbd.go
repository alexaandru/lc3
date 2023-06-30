package vm

import (
	"os"
	"time"
)

func hackArrowKeysToWASM(p []byte) []byte {
	if len(p) < 3 || p[0] != '\x1b' || p[1] != '[' {
		return p
	}

	switch p[2] {
	case 'D': // left
		return []byte{'a'}
	case 'C': // right
		return []byte{'d'}
	case 'B': // down
		return []byte{'s'}
	case 'A': // up
		return []byte{'w'}
	default:
		return p
	}
}

func (vm *VM) KbdLoop() {
	ticker := time.NewTicker(5 * time.Millisecond)
	for range ticker.C {
		buf := make([]byte, 4)
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			continue
		}

		for _, b := range hackArrowKeysToWASM(buf[:n]) {
			vm.keysBuffer <- b
		}
	}
}

func (vm *VM) processKbd() {
	if !vm.running {
		return
	}

	kbsr := vm.MemRead(KbdStatus)
	if ready := kbsr&0x8000 == 0; ready && len(vm.keysBuffer) > 0 {
		vm.MemWrite(KbdStatus, kbsr|0x8000)
		vm.MemWrite(KbdData, word(<-vm.keysBuffer))
	}
}
