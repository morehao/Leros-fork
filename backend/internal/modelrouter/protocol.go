// Package modelrouter provides Worker built-in LLM model routing capability.
package modelrouter

import (
	"fmt"
	"strings"
)

// Protocol represents LLM API protocol type.
type Protocol string

const (
	// ProtocolOpenAIChat OpenAI Chat Completions protocol.
	ProtocolOpenAIChat Protocol = "openai_chat"
	// ProtocolOpenAIResponses OpenAI Responses protocol.
	ProtocolOpenAIResponses Protocol = "openai_responses"
	// ProtocolAnthropicMessages Anthropic Messages protocol.
	ProtocolAnthropicMessages Protocol = "anthropic_messages"
)

// ProtocolFromPath determines entry protocol from request path.
func ProtocolFromPath(path string) (Protocol, error) {
	switch {
	case strings.HasSuffix(path, "/chat/completions"):
		return ProtocolOpenAIChat, nil
	case strings.HasSuffix(path, "/messages"):
		return ProtocolAnthropicMessages, nil
	case strings.HasSuffix(path, "/responses"):
		return ProtocolOpenAIResponses, nil
	default:
		return "", fmt.Errorf("unsupported path: %s", path)
	}
}

// DefaultProtocolForProvider returns the default upstream protocol for a provider.
func DefaultProtocolForProvider(provider string) Protocol {
	switch strings.ToLower(provider) {
	case "anthropic":
		return ProtocolAnthropicMessages
	default:
		return ProtocolOpenAIChat
	}
}

// UpstreamConfig describes the complete upstream forwarding configuration.
type UpstreamConfig struct {
	ModelName    string
	Provider     string
	BaseURL      string
	BaseURLHasV1 bool
	APIKey       string
	Protocol     Protocol
	MaxTokens    int
	Temperature  float64
	TimeoutSec   int
}
