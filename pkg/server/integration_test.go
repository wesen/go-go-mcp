package server

import (
	"context"
	"testing"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

// MockSamplingProvider for testing
type MockSamplingProvider struct {
	lastRequest *protocol.CreateMessageRequest
	response    *protocol.CreateMessageResponse
	err         error
}

func (m *MockSamplingProvider) RequestSampling(ctx context.Context, request protocol.CreateMessageRequest) (*protocol.CreateMessageResponse, error) {
	m.lastRequest = &request
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func TestSamplingServerIntegration(t *testing.T) {
	t.Run("Successful sampling request", func(t *testing.T) {
		// Create mock provider
		mockProvider := &MockSamplingProvider{
			response: &protocol.CreateMessageResponse{
				Role: "assistant",
				Content: protocol.MessageContent{
					Type: "text",
					Text: "Test response",
				},
				Model:      "test-model",
				StopReason: "endTurn",
			},
		}

		// Create sampling server
		samplingServer := NewSamplingServer(mockProvider)

		// Build a test request
		request := NewSamplingRequest().
			AddTextMessage("user", "Hello, world!").
			WithSystemPrompt("You are helpful").
			WithMaxTokens(100).
			WithTemperature(0.7).
			WithModelHint("claude-3").
			Build()

		// Make the sampling request
		ctx := context.Background()
		response, err := samplingServer.RequestSampling(ctx, request)

		// Verify no error
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Verify response
		if response == nil {
			t.Fatal("Expected response, got nil")
		}

		if response.Content.Text != "Test response" {
			t.Errorf("Expected 'Test response', got '%s'", response.Content.Text)
		}

		if response.Model != "test-model" {
			t.Errorf("Expected 'test-model', got '%s'", response.Model)
		}

		// Verify the request was passed correctly to the provider
		if mockProvider.lastRequest == nil {
			t.Fatal("Expected request to be captured by mock provider")
		}

		if len(mockProvider.lastRequest.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(mockProvider.lastRequest.Messages))
		}

		if mockProvider.lastRequest.Messages[0].Content.Text != "Hello, world!" {
			t.Errorf("Expected 'Hello, world!', got '%s'", mockProvider.lastRequest.Messages[0].Content.Text)
		}

		if mockProvider.lastRequest.SystemPrompt != "You are helpful" {
			t.Errorf("Expected 'You are helpful', got '%s'", mockProvider.lastRequest.SystemPrompt)
		}

		if mockProvider.lastRequest.MaxTokens != 100 {
			t.Errorf("Expected max tokens 100, got %d", mockProvider.lastRequest.MaxTokens)
		}

		if mockProvider.lastRequest.Temperature == nil || *mockProvider.lastRequest.Temperature != 0.7 {
			t.Errorf("Expected temperature 0.7, got %v", mockProvider.lastRequest.Temperature)
		}
	})

	t.Run("Error handling", func(t *testing.T) {
		// Create mock provider with error
		mockProvider := &MockSamplingProvider{
			err: &protocol.Error{
				Code:    -32603,
				Message: "Internal error",
			},
		}

		// Create sampling server
		samplingServer := NewSamplingServer(mockProvider)

		// Build a test request
		request := NewSamplingRequest().
			AddTextMessage("user", "Hello, world!").
			Build()

		// Make the sampling request
		ctx := context.Background()
		response, err := samplingServer.RequestSampling(ctx, request)

		// Verify error is returned
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if response != nil {
			t.Errorf("Expected nil response on error, got %v", response)
		}
	})

	t.Run("No provider error", func(t *testing.T) {
		// Create sampling server without provider
		samplingServer := NewSamplingServer(nil)

		// Build a test request
		request := NewSamplingRequest().
			AddTextMessage("user", "Hello, world!").
			Build()

		// Make the sampling request
		ctx := context.Background()
		response, err := samplingServer.RequestSampling(ctx, request)

		// Verify error is returned
		if err == nil {
			t.Fatal("Expected error for nil provider, got nil")
		}

		if response != nil {
			t.Errorf("Expected nil response on error, got %v", response)
		}

		expectedError := "no sampling provider available"
		if err.Error() != expectedError {
			t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
		}
	})
}

func TestComplexSamplingScenarios(t *testing.T) {
	t.Run("Multi-turn conversation", func(t *testing.T) {
		mockProvider := &MockSamplingProvider{
			response: &protocol.CreateMessageResponse{
				Role: "assistant",
				Content: protocol.MessageContent{
					Type: "text",
					Text: "I understand the conversation context.",
				},
				Model:      "claude-3",
				StopReason: "endTurn",
			},
		}

		samplingServer := NewSamplingServer(mockProvider)

		request := NewSamplingRequest().
			AddTextMessage("user", "Hello").
			AddTextMessage("assistant", "Hi there!").
			AddTextMessage("user", "Can you help me?").
			WithSystemPrompt("You are a helpful assistant").
			WithIncludeContext("thisServer").
			WithMaxTokens(200).
			Build()

		ctx := context.Background()
		response, err := samplingServer.RequestSampling(ctx, request)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if response == nil {
			t.Fatal("Expected response")
		}

		// Verify conversation was preserved
		if len(mockProvider.lastRequest.Messages) != 3 {
			t.Errorf("Expected 3 messages, got %d", len(mockProvider.lastRequest.Messages))
		}

		if mockProvider.lastRequest.IncludeContext != "thisServer" {
			t.Errorf("Expected includeContext 'thisServer', got '%s'", mockProvider.lastRequest.IncludeContext)
		}
	})

	t.Run("Image message handling", func(t *testing.T) {
		mockProvider := &MockSamplingProvider{
			response: &protocol.CreateMessageResponse{
				Role: "assistant",
				Content: protocol.MessageContent{
					Type: "text",
					Text: "I can see the image you provided.",
				},
				Model:      "claude-3-vision",
				StopReason: "endTurn",
			},
		}

		samplingServer := NewSamplingServer(mockProvider)

		request := NewSamplingRequest().
			AddImageMessage("user", "iVBORw0KGgoAAAANSUhEUgAA...", "image/png").
			AddTextMessage("user", "What do you see in this image?").
			WithMaxTokens(300).
			Build()

		ctx := context.Background()
		response, err := samplingServer.RequestSampling(ctx, request)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if response == nil {
			t.Fatal("Expected response")
		}

		// Verify image message was handled
		if len(mockProvider.lastRequest.Messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(mockProvider.lastRequest.Messages))
		}

		imageMsg := mockProvider.lastRequest.Messages[0]
		if imageMsg.Content.Type != "image" {
			t.Errorf("Expected image type, got %s", imageMsg.Content.Type)
		}

		if imageMsg.Content.MimeType != "image/png" {
			t.Errorf("Expected image/png mime type, got %s", imageMsg.Content.MimeType)
		}
	})
}
