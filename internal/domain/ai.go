package domain

type AI struct {
	Introduction string
	Functions    []AIFunction
}

type AIReq struct {
	Model     string       `json:"model"`
	Messages  []AIMessage  `json:"messages"`
	Functions []AIFunction `json:"functions"`
}

type AIFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  AIParameter `json:"parameters"`
}

type AIParameter struct {
	Type       string                `json:"type"`
	Properties map[string]AIProperty `json:"properties"`
	Required   *[]string             `json:"required,omitempty"`
}

type AIProperty struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Enum        *[]string `json:"enum,omitempty"`
}

type AIChoices struct {
	Index        int       `json:"index"`
	Message      AIMessage `json:"message"`
	FinishReason string    `json:"finish_reason"`
}

type AIResp struct {
	Id      string      `json:"id"`
	Object  string      `json:"object"`
	Created int         `json:"created"`
	Model   string      `json:"model"`
	Choices []AIChoices `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type AIMessage struct {
	Role         string          `json:"role"`
	Content      string          `json:"content"`
	FunctionCall *AIFunctionCall `json:"function_call,omitempty"`
}

type AIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}
