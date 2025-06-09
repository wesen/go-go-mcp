package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

// SamplingProvider defines the interface for requesting sampling from clients
// This is typically implemented by the client or a bridge component
type SamplingProvider interface {
	// RequestSampling sends a sampling request to the client
	RequestSampling(ctx context.Context, request protocol.CreateMessageRequest) (*protocol.CreateMessageResponse, error)
}

// SamplingServer provides functionality for servers to request sampling from clients
// In a real implementation, this would typically be injected with a SamplingProvider
type SamplingServer struct {
	provider SamplingProvider
}

// NewSamplingServer creates a new sampling server with a provider
func NewSamplingServer(provider SamplingProvider) *SamplingServer {
	return &SamplingServer{
		provider: provider,
	}
}

// RequestSampling sends a sampling request through the provider
func (s *SamplingServer) RequestSampling(ctx context.Context, request protocol.CreateMessageRequest) (*protocol.CreateMessageResponse, error) {
	if s.provider == nil {
		return nil, fmt.Errorf("no sampling provider available")
	}

	return s.provider.RequestSampling(ctx, request)
}

// SamplingRequestBuilder helps build sampling requests
type SamplingRequestBuilder struct {
	request protocol.CreateMessageRequest
}

// NewSamplingRequest creates a new sampling request builder
func NewSamplingRequest() *SamplingRequestBuilder {
	return &SamplingRequestBuilder{
		request: protocol.CreateMessageRequest{
			Messages: make([]protocol.Message, 0),
		},
	}
}

// AddMessage adds a message to the conversation
func (b *SamplingRequestBuilder) AddMessage(role string, content protocol.MessageContent) *SamplingRequestBuilder {
	b.request.Messages = append(b.request.Messages, protocol.Message{
		Role:    role,
		Content: content,
	})
	return b
}

// AddTextMessage adds a text message to the conversation
func (b *SamplingRequestBuilder) AddTextMessage(role, text string) *SamplingRequestBuilder {
	return b.AddMessage(role, protocol.MessageContent{
		Type: "text",
		Text: text,
	})
}

// AddImageMessage adds an image message to the conversation
func (b *SamplingRequestBuilder) AddImageMessage(role, data, mimeType string) *SamplingRequestBuilder {
	return b.AddMessage(role, protocol.MessageContent{
		Type:     "image",
		Data:     data,
		MimeType: mimeType,
	})
}

// WithSystemPrompt sets the system prompt
func (b *SamplingRequestBuilder) WithSystemPrompt(prompt string) *SamplingRequestBuilder {
	b.request.SystemPrompt = prompt
	return b
}

// WithMaxTokens sets the maximum number of tokens
func (b *SamplingRequestBuilder) WithMaxTokens(maxTokens int) *SamplingRequestBuilder {
	b.request.MaxTokens = maxTokens
	return b
}

// WithTemperature sets the sampling temperature
func (b *SamplingRequestBuilder) WithTemperature(temperature float64) *SamplingRequestBuilder {
	b.request.Temperature = &temperature
	return b
}

// WithModelPreferences sets model preferences
func (b *SamplingRequestBuilder) WithModelPreferences(prefs *protocol.ModelPreferences) *SamplingRequestBuilder {
	b.request.ModelPreferences = prefs
	return b
}

// WithModelHint adds a model hint
func (b *SamplingRequestBuilder) WithModelHint(modelName string) *SamplingRequestBuilder {
	if b.request.ModelPreferences == nil {
		b.request.ModelPreferences = &protocol.ModelPreferences{}
	}
	b.request.ModelPreferences.Hints = append(b.request.ModelPreferences.Hints, protocol.ModelHint{
		Name: modelName,
	})
	return b
}

// WithIncludeContext sets context inclusion
func (b *SamplingRequestBuilder) WithIncludeContext(includeContext string) *SamplingRequestBuilder {
	b.request.IncludeContext = includeContext
	return b
}

// WithStopSequences sets stop sequences
func (b *SamplingRequestBuilder) WithStopSequences(sequences []string) *SamplingRequestBuilder {
	b.request.StopSequences = sequences
	return b
}

// WithMetadata sets additional metadata
func (b *SamplingRequestBuilder) WithMetadata(metadata map[string]interface{}) *SamplingRequestBuilder {
	b.request.Metadata = metadata
	return b
}

// Build returns the constructed sampling request
func (b *SamplingRequestBuilder) Build() protocol.CreateMessageRequest {
	return b.request
}

// Helper functions
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal %T: %v", v, err))
	}
	return data
}

var requestIDCounter int64

func generateRequestID() int64 {
	requestIDCounter++
	return requestIDCounter
}
