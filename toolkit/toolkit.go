package toolkit

import "strings"

// Trim trims the input string by removing the last unnecessary comma and the trailing space.
func Trim(input string) (result string) {
	if len(input) == 0 {
		result = strings.TrimSpace(input)
	} else {
		result = strings.TrimSpace(input[0 : len(input)-2])
	}
	return
}
