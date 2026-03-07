package interactive

import (
	"errors"
	"os"

	"golang.org/x/term"
)

var ErrNonTTY = errors.New("interactive mode requires a TTY")

func IsTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}
