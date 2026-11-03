package qahub

import (
	"strings"
	"unicode"
)

func trimInvisible(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.In(r, unicode.Cc, unicode.Cf)
	})
}
