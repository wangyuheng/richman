package domain

import "context"

type AIService interface {
	CallFunctions(ctx context.Context, content string, ai AI) (*AIMessage, error)
}
