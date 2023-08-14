package common

import (
	"context"
	"strings"
)

func Trim(content string) string {
	res := strings.ReplaceAll(content, " ", " ")

	res = strings.TrimSpace(content)
	res = strings.Trim(content, "\r")
	res = strings.Trim(content, "\n")
	res = strings.TrimSpace(content)
	return res
}

const (
	CurrentUserID string = "CURRENT_USER_ID"
)

func GetCurrentUserID(ctx context.Context) string {
	if uid, ok := ctx.Value(CurrentUserID).(string); ok {
		return uid
	}
	return ""
}
