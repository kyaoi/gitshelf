package cli

import (
	"os"
	"time"

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
	return uiColor(text, "1;38;5;45")
}

func uiPrimary(text string) string {
	return uiColor(text, "1;38;5;255")
}

func uiMuted(text string) string {
	return uiColor(text, "38;5;244")
}

func uiShortID(shortID string) string {
	return uiColor("["+shortID+"]", "38;5;109")
}

func uiKind(kind shelf.Kind) string {
	return uiColor(string(kind), "38;5;81")
}

func uiStatus(status shelf.Status) string {
	code := "38;5;250"
	switch status {
	case "open":
		code = "38;5;51"
	case "in_progress":
		code = "38;5;220"
	case "blocked":
		code = "38;5;203"
	case "done":
		code = "38;5;78"
	case "cancelled":
		code = "38;5;245"
	}
	return uiColor(string(status), code)
}

func uiLinkType(linkType shelf.LinkType) string {
	return uiColor(string(linkType), "38;5;141")
}

func uiDue(dueOn string) string {
	if dueOn == "" {
		return uiMuted("-")
	}
	today := time.Now().Local().Format("2006-01-02")
	switch {
	case dueOn < today:
		return uiColor(dueOn, "38;5;203")
	case dueOn == today:
		return uiColor(dueOn, "38;5;220")
	default:
		return uiColor(dueOn, "38;5;114")
	}
}
