package util

import (
	"fmt"
	"strings"
	"unicode"
)

// RemoveNotAllowedCharacters returns string containing only alphanumeric characters or hyphens
func RemoveNotAllowedCharacters(str string) string {
	for _, char := range strings.ToLower(str) {
		if !unicode.IsLetter(char) {
			str = strings.ReplaceAll(str, string(char), "")
		}
	}
	return str
}

// StartWithLetter returns given string but starting with letter
func StartWithLetter(str string) string {
	if len(str) == 0 {
		return "c"
	} else if !unicode.IsLetter(rune(str[0])) {
		return fmt.Sprintf("c-%.9s", str)
	}
	return str
}
