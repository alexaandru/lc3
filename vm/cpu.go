package vm

type cpu struct {
	data [8]word
	pc   word
	cond word
}

func (r *cpu) R(i word) word {
	return r.data[i]
}

func (r *cpu) Val0(instr word) word {
	return r.R(instr.dst())
}

func (r *cpu) Val1(instr word) word {
	return r.R(instr.r1())
}

func (r *cpu) Val2(instr word) word {
	if instr.immediate() {
		return instr.val2()
	}

	return r.R(instr.r2())
}

func (r *cpu) Set(instr word, fn func(word, word) word) {
	r.SetR(instr.dst(), fn(r.Val1(instr), r.Val2(instr)))
}

func (r *cpu) SetR(i, data word) {
	r.data[i] = data
	r.updateFlag(i)
}

func (r *cpu) updateFlag(i word) {
	switch val := r.R(i); {
	case val == 0:
		r.cond = ZERO
	case val>>15 == 1:
		r.cond = NEG
	default:
		r.cond = POS
	}
}
