package cli

import (
	"fmt"
	"os"
	"strings"

	osc52 "github.com/aymanbagabas/go-osc52/v2"
)

func copyTextToClipboard(text string) error {
	seq := osc52.New(text)
	if strings.TrimSpace(os.Getenv("TMUX")) != "" {
		seq = seq.Tmux()
	} else if strings.HasPrefix(strings.ToLower(strings.TrimSpace(os.Getenv("TERM"))), "screen") {
		seq = seq.Screen()
	}
	if _, err := seq.WriteTo(os.Stderr); err != nil {
		return fmt.Errorf("clipboard copy failed: %w", err)
	}
	return nil
}
