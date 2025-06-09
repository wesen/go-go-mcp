package server

import (
	"testing"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

func TestSamplingRequestBuilder(t *testing.T) {
	t.Run("Basic text message", func(t *testing.T) {
		request := NewSamplingRequest().
			AddTextMessage("user", "Hello, world!").
			WithMaxTokens(100).
			Build()

		if len(request.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(request.Messages))
		}

		msg := request.Messages[0]
		if msg.Role != "user" {
			t.Errorf("Expected role 'user', got '%s'", msg.Role)
		}

		if msg.Content.Type != "text" {
			t.Errorf("Expected content type 'text', got '%s'", msg.Content.Type)
		}

		if msg.Content.Text != "Hello, world!" {
			t.Errorf("Expected text 'Hello, world!', got '%s'", msg.Content.Text)
		}

		if request.MaxTokens != 100 {
			t.Errorf("Expected max tokens 100, got %d", request.MaxTokens)
		}
	})

	t.Run("Image message", func(t *testing.T) {
		request := NewSamplingRequest().
			AddImageMessage("user", "base64data", "image/png").
			Build()

		if len(request.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(request.Messages))
		}

		msg := request.Messages[0]
		if msg.Content.Type != "image" {
			t.Errorf("Expected content type 'image', got '%s'", msg.Content.Type)
		}

		if msg.Content.Data != "base64data" {
			t.Errorf("Expected data 'base64data', got '%s'", msg.Content.Data)
		}

		if msg.Content.MimeType != "image/png" {
			t.Errorf("Expected mime type 'image/png', got '%s'", msg.Content.MimeType)
		}
	})

	t.Run("Model preferences", func(t *testing.T) {
		request := NewSamplingRequest().
			WithModelHint("claude-3-sonnet").
			WithModelHint("gpt-4").
			Build()

		if request.ModelPreferences == nil {
			t.Fatal("Expected model preferences to be set")
		}

		if len(request.ModelPreferences.Hints) != 2 {
			t.Errorf("Expected 2 model hints, got %d", len(request.ModelPreferences.Hints))
		}

		if request.ModelPreferences.Hints[0].Name != "claude-3-sonnet" {
			t.Errorf("Expected first hint 'claude-3-sonnet', got '%s'", request.ModelPreferences.Hints[0].Name)
		}

		if request.ModelPreferences.Hints[1].Name != "gpt-4" {
			t.Errorf("Expected second hint 'gpt-4', got '%s'", request.ModelPreferences.Hints[1].Name)
		}
	})

	t.Run("Complex request", func(t *testing.T) {
		temperature := 0.7
		request := NewSamplingRequest().
			AddTextMessage("user", "Analyze this text").
			AddTextMessage("assistant", "I'll help you analyze it.").
			AddTextMessage("user", "Great! Here's the text to analyze...").
			WithSystemPrompt("You are a helpful analysis assistant").
			WithMaxTokens(500).
			WithTemperature(temperature).
			WithIncludeContext("thisServer").
			WithStopSequences([]string{"[END]", "---"}).
			WithMetadata(map[string]interface{}{
				"task": "analysis",
				"priority": "high",
			}).
			Build()

		// Check messages
		if len(request.Messages) != 3 {
			t.Errorf("Expected 3 messages, got %d", len(request.Messages))
		}

		// Check system prompt
		if request.SystemPrompt != "You are a helpful analysis assistant" {
			t.Errorf("Unexpected system prompt: %s", request.SystemPrompt)
		}

		// Check max tokens
		if request.MaxTokens != 500 {
			t.Errorf("Expected max tokens 500, got %d", request.MaxTokens)
		}

		// Check temperature
		if request.Temperature == nil || *request.Temperature != 0.7 {
			t.Errorf("Expected temperature 0.7, got %v", request.Temperature)
		}

		// Check include context
		if request.IncludeContext != "thisServer" {
			t.Errorf("Expected include context 'thisServer', got '%s'", request.IncludeContext)
		}

		// Check stop sequences
		if len(request.StopSequences) != 2 {
			t.Errorf("Expected 2 stop sequences, got %d", len(request.StopSequences))
		}

		// Check metadata
		if request.Metadata == nil {
			t.Fatal("Expected metadata to be set")
		}

		if request.Metadata["task"] != "analysis" {
			t.Errorf("Expected metadata task 'analysis', got %v", request.Metadata["task"])
		}

		if request.Metadata["priority"] != "high" {
			t.Errorf("Expected metadata priority 'high', got %v", request.Metadata["priority"])
		}
	})
}

func TestCreateMessageRequest_Validation(t *testing.T) {
	t.Run("Valid request", func(t *testing.T) {
		request := protocol.CreateMessageRequest{
			Messages: []protocol.Message{
				{
					Role: "user",
					Content: protocol.MessageContent{
						Type: "text",
						Text: "Hello",
					},
				},
			},
			MaxTokens: 100,
		}

		// Basic validation - ensure fields are properly set
		if len(request.Messages) == 0 {
			t.Error("Expected at least one message")
		}

		if request.MaxTokens <= 0 {
			t.Error("Expected max tokens to be positive")
		}
	})

	t.Run("Temperature validation", func(t *testing.T) {
		temp := 0.5
		request := protocol.CreateMessageRequest{
			Messages: []protocol.Message{
				{
					Role: "user",
					Content: protocol.MessageContent{
						Type: "text",
						Text: "Hello",
					},
				},
			},
			MaxTokens:   100,
			Temperature: &temp,
		}

		if request.Temperature == nil || *request.Temperature != 0.5 {
			t.Errorf("Expected temperature 0.5, got %v", request.Temperature)
		}
	})
}
