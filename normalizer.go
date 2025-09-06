package main

import "strings"

const (
	leadingChars  = `([{'"`
	trailingChars = `.,;:!?'")]}>»›`
)

func TrimWrapperPunctuation(s string) string {
	for len(s) > 1 {
		originalLength := len(s)

		if strings.HasSuffix(s, "...") {
			break
		}

		if (strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")")) ||
			(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")) ||
			(strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) {
			if strings.Contains(s, "://") {
				s = s[1 : len(s)-1]
				continue
			}
		}

		s = strings.TrimPrefix(s, "'")
		s = strings.TrimPrefix(s, `"`)

		if len(s) > 0 {
			lastChar := s[len(s)-1:]
			if strings.Contains(trailingChars, lastChar) {
				if lastChar == ")" && strings.Contains(s, "(") {
					// do nothing
				} else {
					s = s[:len(s)-1]
				}
			}
		}

		if len(s) == originalLength {
			break
		}
	}
	return s
}
