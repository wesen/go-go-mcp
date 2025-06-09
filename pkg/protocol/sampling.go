package protocol

// Message represents a message in a conversation
type Message struct {
	Role    string         `json:"role"` // "user" or "assistant"
	Content MessageContent `json:"content"`
	Model   string         `json:"model,omitempty"`
}

// MessageContent represents different types of content in a message
type MessageContent struct {
	Type     string `json:"type"`           // "text" or "image"
	Text     string `json:"text,omitempty"` // For text content
	Data     string `json:"data,omitempty"` // Base64 encoded for image content
	MimeType string `json:"mimeType,omitempty"`
}

// ModelPreferences represents preferences for model selection
type ModelPreferences struct {
	Hints                []ModelHint `json:"hints,omitempty"`
	CostPriority         float64     `json:"costPriority,omitempty"`
	SpeedPriority        float64     `json:"speedPriority,omitempty"`
	IntelligencePriority float64     `json:"intelligencePriority,omitempty"`
}

// ModelHint represents a suggested model name
type ModelHint struct {
	Name string `json:"name,omitempty"`
}

// CreateMessageRequest represents a request to create a message
type CreateMessageRequest struct {
	Messages         []Message                  `json:"messages"`
	ModelPreferences *ModelPreferences          `json:"modelPreferences,omitempty"`
	SystemPrompt     string                     `json:"systemPrompt,omitempty"`
	IncludeContext   string                     `json:"includeContext,omitempty"` // "none", "thisServer", "allServers"
	Temperature      *float64                   `json:"temperature,omitempty"`
	MaxTokens        int                        `json:"maxTokens"`
	StopSequences    []string                   `json:"stopSequences,omitempty"`
	Metadata         map[string]interface{}     `json:"metadata,omitempty"`
}

// CreateMessageResponse represents the response to a create message request
type CreateMessageResponse struct {
	Role       string         `json:"role"`
	Content    MessageContent `json:"content"`
	Model      string         `json:"model,omitempty"`
	StopReason string         `json:"stopReason,omitempty"`
}

// SamplingClient represents the interface for clients that support sampling
type SamplingClient interface {
	// CreateMessage requests the client to create a message using LLM sampling
	CreateMessage(request CreateMessageRequest) (*CreateMessageResponse, error)
}
