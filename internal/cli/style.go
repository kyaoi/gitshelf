package cli

import (
	"os"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"golang.org/x/term"
)

func uiEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("CLICOLOR_FORCE") == "1" {
		return true
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func uiColor(text string, code string) string {
	if !uiEnabled() {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func uiHeading(text string) string {
	return uiColor(text, "1;36")
}

func uiPrimary(text string) string {
	return uiColor(text, "1;37")
}

func uiMuted(text string) string {
	return uiColor(text, "2")
}

func uiShortID(shortID string) string {
	return uiColor("["+shortID+"]", "2;37")
}

func uiKind(kind shelf.Kind) string {
	return uiColor(string(kind), "36")
}

func uiStatus(status shelf.Status) string {
	code := "37"
	switch status {
	case "open":
		code = "36"
	case "in_progress":
		code = "33"
	case "blocked":
		code = "31"
	case "done":
		code = "32"
	case "cancelled":
		code = "90"
	}
	return uiColor(string(status), code)
}

func uiLinkType(linkType shelf.LinkType) string {
	return uiColor(string(linkType), "35")
}

func uiDue(dueOn string) string {
	if dueOn == "" {
		return uiMuted("-")
	}
	return uiColor(dueOn, "35")
}
