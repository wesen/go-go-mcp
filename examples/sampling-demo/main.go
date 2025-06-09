package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/server"
	"github.com/go-go-golems/go-go-mcp/pkg/transport"
	"github.com/go-go-golems/go-go-mcp/pkg/transport/stdio"
	"github.com/rs/zerolog"
)

// SamplingDemoServer demonstrates the use of sampling in MCP
type SamplingDemoServer struct {
	*server.Server
	sampling *server.SamplingServer
}

// NewSamplingDemoServer creates a new demo server with sampling capabilities
func NewSamplingDemoServer(logger zerolog.Logger, transport transport.Transport, samplingProvider server.SamplingProvider) *SamplingDemoServer {
	baseServer := server.NewServer(logger, transport)
	
	demo := &SamplingDemoServer{
		Server:   baseServer,
		sampling: server.NewSamplingServer(samplingProvider),
	}

	// Register our custom tools
	demo.registerTools()
	
	return demo
}

func (s *SamplingDemoServer) registerTools() {
	// Register tools that use sampling to analyze text
	_ = &AnalyzeTextTool{sampling: s.sampling}
	_ = &SummarizerTool{sampling: s.sampling}
	_ = &ConversationTool{sampling: s.sampling}
	// Note: These would need to be properly registered with the tool provider
	// For now, this is just demonstrating the structure
}

// AnalyzeTextTool demonstrates a tool that uses sampling
type AnalyzeTextTool struct {
	sampling *server.SamplingServer
}

// Execute analyzes text using LLM sampling
func (t *AnalyzeTextTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	text, ok := args["text"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'text' parameter")
	}

	analysisType, ok := args["analysis_type"].(string)
	if !ok {
		analysisType = "sentiment" // default
	}

	// Build a sampling request
	request := server.NewSamplingRequest().
		AddTextMessage("user", fmt.Sprintf("Please analyze the following text for %s:\n\n%s", analysisType, text)).
		WithSystemPrompt("You are a helpful text analysis assistant. Provide clear, concise analysis.").
		WithMaxTokens(500).
		WithModelHint("claude-3").
		WithIncludeContext("none").
		WithTemperature(0.3).
		Build()

	// Request sampling from the client
	response, err := t.sampling.RequestSampling(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("sampling failed: %w", err)
	}

	return map[string]interface{}{
		"analysis":    response.Content.Text,
		"model_used":  response.Model,
		"stop_reason": response.StopReason,
	}, nil
}

// SummarizerTool demonstrates another sampling use case
type SummarizerTool struct {
	sampling *server.SamplingServer
}

func (t *SummarizerTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'content' parameter")
	}

	length, ok := args["length"].(string)
	if !ok {
		length = "medium" // default: short, medium, long
	}

	var prompt string
	var maxTokens int

	switch length {
	case "short":
		prompt = "Provide a brief 1-2 sentence summary of the following content:"
		maxTokens = 100
	case "long":
		prompt = "Provide a detailed summary with key points of the following content:"
		maxTokens = 800
	default: // medium
		prompt = "Provide a medium-length summary of the following content:"
		maxTokens = 300
	}

	request := server.NewSamplingRequest().
		AddTextMessage("user", fmt.Sprintf("%s\n\n%s", prompt, content)).
		WithSystemPrompt("You are a skilled summarization assistant. Focus on the most important information.").
		WithMaxTokens(maxTokens).
		WithModelHint("claude-3-sonnet").
		WithIncludeContext("none").
		WithTemperature(0.2).
		Build()

	response, err := t.sampling.RequestSampling(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("sampling failed: %w", err)
	}

	return map[string]interface{}{
		"summary":     response.Content.Text,
		"length":      length,
		"model_used":  response.Model,
		"stop_reason": response.StopReason,
	}, nil
}

// ConversationTool demonstrates multi-turn sampling
type ConversationTool struct {
	sampling *server.SamplingServer
}

func (t *ConversationTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	messages, ok := args["messages"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'messages' parameter")
	}

	builder := server.NewSamplingRequest().
		WithSystemPrompt("You are a helpful assistant engaged in a conversation.").
		WithMaxTokens(500).
		WithModelHint("claude-3").
		WithIncludeContext("thisServer").
		WithTemperature(0.7)

	// Add conversation history
	for _, msg := range messages {
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}

		role, ok := msgMap["role"].(string)
		if !ok {
			continue
		}

		text, ok := msgMap["text"].(string)
		if !ok {
			continue
		}

		builder.AddTextMessage(role, text)
	}

	request := builder.Build()
	
	response, err := t.sampling.RequestSampling(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("sampling failed: %w", err)
	}

	return map[string]interface{}{
		"response":    response.Content.Text,
		"model_used":  response.Model,
		"stop_reason": response.StopReason,
	}, nil
}

// MockSamplingProvider provides a mock implementation for demonstration
type MockSamplingProvider struct{}

func (m *MockSamplingProvider) RequestSampling(ctx context.Context, request protocol.CreateMessageRequest) (*protocol.CreateMessageResponse, error) {
	// This is a mock implementation - in real usage, this would be provided by the MCP client
	return &protocol.CreateMessageResponse{
		Role: "assistant",
		Content: protocol.MessageContent{
			Type: "text",
			Text: "This is a mock response from the sampling provider. In a real implementation, this would come from the MCP client's LLM.",
		},
		Model:      "mock-model",
		StopReason: "endTurn",
	}, nil
}

func main() {
	// Set up logging
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// Use stdio transport for MCP
	transport, err := stdio.NewStdioTransport(
		transport.WithLogger(logger),
	)
	if err != nil {
		log.Fatal("Failed to create transport:", err)
	}

	// Create a mock sampling provider for demonstration
	samplingProvider := &MockSamplingProvider{}

	// Create the demo server
	demoServer := NewSamplingDemoServer(logger, transport, samplingProvider)

	// Start the server
	ctx := context.Background()
	if err := demoServer.Start(ctx); err != nil {
		log.Fatal("Failed to start demo server:", err)
	}
}
