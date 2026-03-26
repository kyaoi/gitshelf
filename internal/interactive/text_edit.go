package interactive

func ClampTextCursor(value string, cursor int) int {
	length := len([]rune(value))
	if cursor < 0 {
		return 0
	}
	if cursor > length {
		return length
	}
	return cursor
}

func MoveTextCursorLeft(value string, cursor int) int {
	cursor = ClampTextCursor(value, cursor)
	if cursor == 0 {
		return 0
	}
	return cursor - 1
}

func MoveTextCursorRight(value string, cursor int) int {
	cursor = ClampTextCursor(value, cursor)
	length := len([]rune(value))
	if cursor >= length {
		return length
	}
	return cursor + 1
}

func MoveTextCursorStart(value string, cursor int) int {
	_ = value
	_ = cursor
	return 0
}

func MoveTextCursorEnd(value string, cursor int) int {
	_ = cursor
	return len([]rune(value))
}

func InsertRuneAtCursor(value string, cursor int, r rune) (string, int) {
	runes := []rune(value)
	cursor = ClampTextCursor(value, cursor)
	runes = append(runes[:cursor], append([]rune{r}, runes[cursor:]...)...)
	return string(runes), cursor + 1
}

func DeleteRuneBeforeCursor(value string, cursor int) (string, int) {
	runes := []rune(value)
	cursor = ClampTextCursor(value, cursor)
	if cursor == 0 {
		return value, 0
	}
	runes = append(runes[:cursor-1], runes[cursor:]...)
	return string(runes), cursor - 1
}

func DeleteRuneAtCursor(value string, cursor int) (string, int) {
	runes := []rune(value)
	cursor = ClampTextCursor(value, cursor)
	if cursor >= len(runes) {
		return value, cursor
	}
	runes = append(runes[:cursor], runes[cursor+1:]...)
	return string(runes), cursor
}

func RenderTextCursor(value string, cursor int) string {
	runes := []rune(value)
	cursor = ClampTextCursor(value, cursor)
	runes = append(runes[:cursor], append([]rune{'_'}, runes[cursor:]...)...)
	return string(runes)
}
