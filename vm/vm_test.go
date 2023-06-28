package vm

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"golang.org/x/term"
)

func TestExecute(t *testing.T) {
	testCases := []struct {
		op                      word
		dst, src1, imm, src2imm byte
		cpu, expCpu             cpu
	}{
		{
			ADD, 0, 1, 0, 2,
			cpu{},
			cpu{cond: ZERO},
		},
		{
			ADD, 0, 1, 0, 2,
			cpu{data: [8]word{0, 1, 2}},
			cpu{data: [8]word{3, 1, 2}, cond: POS},
		},
		{
			ADD, 7, 5, 0, 6,
			cpu{data: [8]word{0, 0, 0, 0, 0, 15, 100, 0}},
			cpu{data: [8]word{0, 0, 0, 0, 0, 15, 100, 115}, cond: POS},
		},
		{
			ADD, 0, 1, 1, 15,
			cpu{data: [8]word{0, 100, 200}},
			cpu{data: [8]word{115, 100, 200}, cond: POS},
		},
		{
			ADD, 0, 1, 1, 31,
			cpu{data: [8]word{0, 100, 200}},
			cpu{data: [8]word{99, 100, 200}, cond: POS},
		},
		{
			AND, 0, 1, 0, 2,
			cpu{},
			cpu{cond: ZERO},
		},
		{
			AND, 0, 1, 0, 2,
			cpu{data: [8]word{0, 1, 2}},
			cpu{data: [8]word{0, 1, 2}, cond: ZERO},
		},
		{
			AND, 7, 5, 0, 6,
			cpu{data: [8]word{0, 0, 0, 0, 0, 15, 100, 0}},
			cpu{data: [8]word{0, 0, 0, 0, 0, 15, 100, 15 & 100}, cond: POS},
		},
		{
			AND, 0, 1, 1, 15,
			cpu{data: [8]word{0, 100, 200}},
			cpu{data: [8]word{15 & 100, 100, 200}, cond: POS},
		},
		{
			AND, 0, 1, 1, 31,
			cpu{data: [8]word{0, 100, 200}},
			cpu{data: [8]word{100 & word(31).extendSign(5), 100, 200}, cond: POS},
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			vm := New(nil)
			*vm.cpu = tc.cpu
			instr := mkWord(tc.op, tc.dst, tc.src1, tc.imm, tc.src2imm)

			t.Log("instruction:", instr, "immediate:", instr.immediate(), "val1:", vm.Val1(instr), "val2:",
				fmt.Sprintf("%016b", vm.Val2(instr)))

			if vm.execute(instr); *vm.cpu != tc.expCpu {
				t.Fatalf("Expected %v got %v", tc.expCpu, *vm.cpu)
			}
		})
	}
}

var intgTestCases = []struct {
	programFile string
	expCpu      [8]word
}{
	{
		"test00.obj",
		[8]word{0x44c2, 0x22c2, 0x9204, 0x6784, 0x0, 0x2200, 0x0, 0x3e9},
	},
	{
		"test01.obj",
		[8]word{0x0, 0xd163, 0xe6ee, 0x0, 0xea75, 0xba4b, 0x0, 0x13e9},
	},

	{
		"test02.obj",
		[8]word{0x0, 0x0, 0x0, 0x0, 0x0, 0xff1, 0xa042, 0x23e9},
	},
	{
		"test03.obj",
		[8]word{0xc574, 0x0, 0x1640, 0x1640, 0x103f, 0x2403, 0x3440, 0x33e9},
	},
	{
		"test04.obj",
		[8]word{0xfffe, 0x0, 0xfffe, 0x6402, 0xf025, 0x6401, 0x0, 0x43e9},
	},
	{
		"test05.obj",
		[8]word{0x648b, 0x52e9, 0xffff, 0x11a2, 0x0, 0xbed0, 0x0, 0x53e9},
	},
	{
		"test06.obj",
		[8]word{0x0, 0x0, 0x0, 0x0, 0xcd83, 0x0, 0x53a0, 0x63e9},
	},
	{
		"test07.obj",
		[8]word{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x73e9},
	},
	{
		"test08.obj",
		[8]word{0x0, 0x12c0, 0x0, 0x0, 0x0, 0x1204, 0x0, 0x83e9},
	},
	{
		"test09.obj",
		[8]word{0x4504, 0x2282, 0x2282, 0x0, 0x2284, 0xffff, 0x0, 0x93e9},
	},
}

func TestIntegration(t *testing.T) {
	for _, tc := range intgTestCases {
		t.Run("", func(t *testing.T) {
			vm := New(term.NewTerminal(os.Stdout, ""))
			vm.LoadImageFromFile(filepath.Join("testdata", tc.programFile))

			go func() {
				for {
					vm.keysBuffer <- 'a'
					// Let's not keep the kb loop waiting...
					// FIXME: OTOH, this does look like a bug...
					// Why do we need to do this if we don't use Traps??
					time.Sleep(time.Millisecond)
				}
			}()

			_ = vm.Run()

			if !reflect.DeepEqual(tc.expCpu, vm.data) {
				t.Fatalf("Expected:\n\t%+#v, got\n\t%+#v\n", tc.expCpu, vm.data)
			}
		})
	}
}

func BenchmarkIntegration(b *testing.B) {
	for _, tc := range intgTestCases {
		b.Run("", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				b.StopTimer()
				vm := New(term.NewTerminal(os.Stdout, ""))
				vm.LoadImageFromFile(filepath.Join("testdata", tc.programFile))

				go func() {
					for {
						vm.keysBuffer <- 'a'
						time.Sleep(time.Millisecond)
					}
				}()

				b.StartTimer()
				_ = vm.Run()
			}
		})
	}
}
