package vm

type word uint16

func (w word) op() word {
	return w >> 12
}

func (w word) immediate() bool {
	return (w>>5)&1 == 1
}

func (w word) r1() word {
	return (w >> 6) & 7
}

func (w word) r2() word {
	return w & 7
}

func (w word) val2() word {
	return (w & 0x1F).extendSign(5)
}

func (w word) dst() word {
	return (w >> 9) & 7
}

func (w word) pcOfs() word {
	return (w & 0x1FF).extendSign(9)
}

func (w word) extendSign(bitCount byte) word {
	if (w>>(bitCount-1))&1 > 0 {
		w |= (MaxWord << bitCount)
	}

	return w
}

func mkWord(op word, dst, src1, imm, src2imm byte) word {
	return op<<12 | word(dst&7)<<9 | word(src1&7)<<6 | word(imm&1)<<5 | word(src2imm&0x1F)
}

func loByte(w word) []byte {
	return []byte{byte(w)}
}

func twoByte(w word) []byte {
	return []byte{byte(w >> 8), byte(w)}
}
