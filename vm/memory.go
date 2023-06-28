package vm

type memory [MemMax]word

func (m *memory) MemWrite(addr, val word) {
	(*m)[addr] = val
}

func (m *memory) MemRead(addr word) word {
	if addr == KbdData {
		m[KbdStatus] &= 0x7FFF
	}

	return m[addr]
}
