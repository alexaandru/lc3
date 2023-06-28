package vm

import "fmt"

type BadTrapError word

func (e BadTrapError) Error() string {
	return fmt.Sprintf("bad trap %d", e)
}

type BadOpError word

func (e BadOpError) Error() string {
	return fmt.Sprintf("bad op %d", e)
}
