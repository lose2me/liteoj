package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/liteoj/liteoj/backend/internal/config"
	"github.com/liteoj/liteoj/backend/internal/models"
)

func TestAnalyzeWrongAnswer_AdminForceAddsOverride(t *testing.T) {
	var req chatReq
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"ok\":true,\"reason\":\"\",\"explanation\":\"ok\"}"}}]}`))
	}))
	defer srv.Close()

	p := NewPrompts(&config.Config{
		AIEnabled:      true,
		BifrostBaseURL: srv.URL,
		BifrostAPIKey:  "test",
		BifrostModel:   "test",
		AIPromptWA:     "base prompt",
	}, &Client{BaseURL: srv.URL, APIKey: "test", Model: "test", HTTP: &http.Client{}})

	_, prompt, _, err := p.AnalyzeWrongAnswer(context.Background(),
		&models.Problem{Title: "A", Description: "B"},
		&models.Submission{Language: "cpp", Code: "int main(){return 0;}", Verdict: "WA"},
		"[]", true)
	if err != nil {
		t.Fatalf("AnalyzeWrongAnswer force admin: %v", err)
	}
	if len(req.Messages) == 0 {
		t.Fatalf("expected outbound messages")
	}
	sys := req.Messages[0].Content
	if !strings.Contains(sys, "管理员强制解析模式") {
		t.Fatalf("system prompt should include admin-force override, got: %s", sys)
	}
	if !strings.Contains(prompt, "管理员强制解析模式") {
		t.Fatalf("audit prompt should include admin-force override")
	}
}
