package util

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	letters = "abcdefghijklmnopqrstuvwxyz"
)

// RemoveNotAllowedCharacters returns provider containing only alphanumeric characters or hyphens
func RemoveNotAllowedCharacters(provider string) string {
	for _, char := range strings.ToLower(provider) {
		if !strings.ContainsRune(letters, char) {
			provider = strings.ReplaceAll(provider, string(char), "")
		}
	}
	return provider
}

// StartWithLetter returns given name but starting with letter
func StartWithLetter(name string) string {
	if len(name) == 0 {
		return "c"
	} else if !unicode.IsLetter(rune(name[0])) {
		return fmt.Sprintf("c-%.9s", name)
	}
	return name
}
