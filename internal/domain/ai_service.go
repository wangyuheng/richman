package domain

type AIService interface {
	CallFunctions(content string, ai AI) (*AIMessage, error)
}
