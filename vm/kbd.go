package vm

import (
	"os"
	"time"
)

func (vm *VM) KbdLoop() {
	buf, ticker := []byte{0}, time.NewTicker(5*time.Millisecond)
	for range ticker.C {
		if n, err := os.Stdin.Read(buf); err != nil || n == 0 {
			continue
		}

		vm.keysBuffer <- buf[0]
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
