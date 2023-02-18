package analysis

import (
	"unicode"
)

func parseName(name string) tokenList {
	tokens := tokenList{}
	t := token{}
	braces := 0

	for _, ch := range name {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
			if !t.IsEmpty() {

				tokens.Push(t)
				t.Text = ""
			}
			if ch == '(' || ch == '[' {
				braces++
				t.InBraces = true
			} else if (ch == ')' || ch == ']') && braces > 0 {
				braces--
				if braces == 0 {
					t.InBraces = false
				}
			}

		} else {
			t.Push(ch)
		}
	}

	if !t.IsEmpty() {
		tokens.Push(t)
	}

	return tokens
}
