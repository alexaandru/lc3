package main

import (
	"os"
	"os/signal"
	"syscall"

	Vm "github.com/alexaandru/lc3/vm"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

func main() {
	fd := int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fatal(err.Error())
	}

	termios, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		fatal(err.Error())
	}

	// I want raw mode but I also want Ctrl-C to work.
	termios.Iflag &^= unix.IGNBRK
	termios.Iflag |= unix.BRKINT
	termios.Lflag |= unix.ISIG

	if err := unix.IoctlSetTermios(fd, unix.TCSETS, termios); err != nil {
		fatal(err.Error())
	}

	trm := term.NewTerminal(os.Stdin, "")
	defer term.Restore(fd, oldState) //nolint:errcheck // not much we can do about

	vm := Vm.New(trm)

	ch := make(chan os.Signal, 4)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		print("Caught signal: ", unix.SignalName((<-ch).(syscall.Signal)),
			// TODO: I mean, it's convenient to press a key to exit,
			// but I would like to know why that happens!?! :-D
			"\r\nPress any key to exit.\r\n")
		vm.Stop()
	}()

	vm.WriteString("\x1b[?1049hBooting LC-3 VM 💾\nQuit with Ctrl-C\n\x1b[?25l")
	defer vm.WriteString("\x1b[?1049l\x1b[?25h\x1b[0m")

	vm.LoadImageFromFile(os.Args[1])

	go vm.KbdLoop()

	if err := vm.Run(); err != nil {
		vm.WriteString("error: " + err.Error())
	}
}

func fatal(msg string) {
	println(msg)
	os.Exit(1)
}

func init() {
	if len(os.Args) < 2 {
		fatal("Usage: " + os.Args[0] + " file1 [file2 [...]]")
	}
}
