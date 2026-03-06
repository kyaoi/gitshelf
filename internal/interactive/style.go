package interactive

import "os"

func uiEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("CLICOLOR_FORCE") == "1" {
		return true
	}
	return IsTTY()
}

func uiColor(text string, code string) string {
	if !uiEnabled() {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func uiPrompt(text string) string {
	return uiColor(text, "1;36")
}

func uiHelp(text string) string {
	return uiColor(text, "2")
}

func uiSelected(text string) string {
	return uiColor(text, "1;37")
}

func uiSearch(text string) string {
	return uiColor(text, "33")
}

func uiPreviewHeader(text string) string {
	return uiColor(text, "1;35")
}
