package common

import "strings"

func Trim(content string) string {
	res := strings.ReplaceAll(content, " ", " ")

	res = strings.TrimSpace(content)
	res = strings.Trim(content, "\r")
	res = strings.Trim(content, "\n")
	res = strings.TrimSpace(content)
	return res
}
