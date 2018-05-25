package reiner

import "strings"

// trim 會清理接收到的字串，移除最後無謂的逗點與空白。
func trim(input string) (result string) {
	if len(input) == 0 {
		result = strings.TrimSpace(input)
	} else {
		result = strings.TrimSpace(input[0 : len(input)-2])
	}
	return
}
