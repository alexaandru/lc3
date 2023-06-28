package vm

import "os"

func (vm *VM) LoadImageFromFile(fname string) {
	buf, err := os.ReadFile(fname)
	if err != nil {
		println("Cannot load image:", err.Error())
		os.Exit(2)
	}

	vm.pc = word(buf[0])<<8 | word(buf[1])
	target := vm.memory[int(vm.pc) : int(vm.pc)+len(buf)/2-1]
	for i := 2; i < len(buf); i += 2 {
		target[i/2-1] = word(buf[i])<<8 | word(buf[i+1])
	}
}
