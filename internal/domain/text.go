package domain

import (
	"strings"
	"unicode"
)

func hasVisibleRune(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return unicode.IsGraphic(r) && !unicode.IsSpace(r)
	}) >= 0
}
