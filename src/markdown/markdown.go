package markdown

import "strings"

func EnsureFormatting(text string) string {
	numDelimiters := strings.Count(text, "```")
	numSingleDelimiters := strings.Count(strings.Replace(text, "```", "", -1), "`")

	if (numDelimiters % 2) == 1 {
		text += "```"
	}
	if (numSingleDelimiters % 2) == 1 {
		text += "`"
	}

	return text
}
