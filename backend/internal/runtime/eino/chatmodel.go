package eino

import (
	"context"
	"fmt"
	"strings"

	einoclaude "github.com/cloudwego/eino-ext/components/model/claude"
	einoopenai "github.com/cloudwego/eino-ext/components/model/openai"
	einomodel "github.com/cloudwego/eino/components/model"

	"github.com/insmtx/Leros/backend/config"
	"github.com/insmtx/Leros/backend/types"
)

// NewChatModel 根据 provider 映射创建 Eino 对话模型。
func NewChatModel(ctx context.Context, cfg *config.LLMConfig) (einomodel.ToolCallingChatModel, error) {
	if cfg == nil {
		return nil, fmt.Errorf("llm config is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("llm api key is required")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, fmt.Errorf("llm model is required")
	}

	provider := types.LLMProviderType(strings.TrimSpace(cfg.Provider))
	switch provider {
	case types.LLMProviderAnthropic:
		return newClaudeChatModel(ctx, cfg)
	default:
		return newOpenAICompatibleChatModel(ctx, cfg)
	}
}

func newOpenAICompatibleChatModel(ctx context.Context, cfg *config.LLMConfig) (einomodel.ToolCallingChatModel, error) {
	chatModel, err := einoopenai.NewChatModel(ctx, &einoopenai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("create eino openai chat model: %w", err)
	}

	return chatModel, nil
}

func newClaudeChatModel(ctx context.Context, cfg *config.LLMConfig) (einomodel.ToolCallingChatModel, error) {
	var baseURL *string
	if strings.TrimSpace(cfg.BaseURL) != "" {
		baseURL = &cfg.BaseURL
	}
	chatModel, err := einoclaude.NewChatModel(ctx, &einoclaude.Config{
		APIKey:    cfg.APIKey,
		BaseURL:   baseURL,
		Model:     cfg.Model,
		MaxTokens: 4096,
	})
	if err != nil {
		return nil, fmt.Errorf("create eino claude chat model: %w", err)
	}
	return chatModel, nil
}
