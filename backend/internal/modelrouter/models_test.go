package modelrouter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/Leros/backend/types"
)

func TestNewOpenAIModelsResponseUsesProtocolShape(t *testing.T) {
	models := []*types.LLMModel{
		newModelRouterTestModel(1, "main", "gpt-4.1", true),
		newModelRouterTestModel(1, "fast", "gpt-4o", false),
		newModelRouterTestModel(1, "duplicate", "gpt-4o", false),
		newModelRouterTestModel(1, "blank", " ", false),
	}
	resp := newOpenAIModelsResponse(models)

	if resp.Object != "list" {
		t.Fatalf("expected object list, got %q", resp.Object)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 unique models, got %d: %+v", len(resp.Data), resp.Data)
	}
	if len(resp.Models) != len(resp.Data) {
		t.Fatalf("expected models alias to match data, got %d and %d", len(resp.Models), len(resp.Data))
	}
	for _, model := range resp.Data {
		if model.Object != "model" {
			t.Fatalf("expected model object, got %q", model.Object)
		}
		if model.Created != 0 {
			t.Fatalf("expected created 0, got %d", model.Created)
		}
		if model.OwnedBy != "" {
			t.Fatalf("expected empty owned_by, got %q", model.OwnedBy)
		}
		if model.ID != "gpt-4.1" && model.ID != "gpt-4o" {
			t.Fatalf("unexpected model id %q", model.ID)
		}
	}
}

func TestListModelsWithoutDatabaseReturnsEmptyList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	v1 := r.Group("/v1")
	RegisterRoutes(v1, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp openAIModelsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Object != "list" || len(resp.Data) != 0 || len(resp.Models) != 0 {
		t.Fatalf("expected empty list response, got %+v", resp)
	}
}

func newModelRouterTestModel(orgID uint, code string, modelName string, isDefault bool) *types.LLMModel {
	return &types.LLMModel{
		OrgID:           orgID,
		Code:            code,
		Name:            code,
		Provider:        string(types.LLMProviderOpenAI),
		ModelName:       modelName,
		BaseURL:         "https://api.openai.com/v1",
		BaseURLHasV1:    true,
		APIKeyEncrypted: "encrypted-key",
		APIKeyMasked:    "sk-***",
		MaxTokens:       4096,
		Temperature:     0.7,
		TimeoutSec:      120,
		Status:          string(types.LLMModelStatusActive),
		IsDefault:       isDefault,
	}
}
