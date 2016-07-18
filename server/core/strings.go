package core

import (
	"strings"
	"unicode"
)

// StringSplitQuoteAware splits a string on the separator but not
func StringSplitQuoteAware(line string, separator rune) []string {
	lastQuote := rune(0)
	f := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		default:
			return c == separator
		}
	}
	return strings.FieldsFunc(line, f)
}
