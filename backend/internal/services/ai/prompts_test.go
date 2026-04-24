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

func writeChatResponse(t *testing.T, w http.ResponseWriter, content string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"choices": []map[string]any{
			{"message": map[string]any{"role": "assistant", "content": content}},
		},
	}); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}

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

func TestGenDesc_DropsEmptyInputSampleSection(t *testing.T) {
	fence := "```"
	content := "## 题目描述\n\n输出一行 Hello World。\n\n## 输入格式\n\n本题无输入。\n\n## 输出格式\n\n输出固定字符串。\n\n" +
		"## 输入 #1\n\n" + fence + "\n" + fence + "\n\n" +
		"## 输出 #1\n\n" + fence + "\nHello World\n" + fence + "\n\n" +
		"## 数据范围\n\n无。"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeChatResponse(t, w, content)
	}))
	defer srv.Close()

	p := NewPrompts(&config.Config{
		AIEnabled:       true,
		BifrostBaseURL:  srv.URL,
		BifrostAPIKey:   "test",
		BifrostModel:    "test",
		AIPromptGenDesc: "prompt",
	}, &Client{BaseURL: srv.URL, APIKey: "test", Model: "test", HTTP: &http.Client{}})

	text, _, _, err := p.GenDesc(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("GenDesc: %v", err)
	}
	if strings.Contains(text, "## 输入 #1") {
		t.Fatalf("empty sample input section should be removed, got: %s", text)
	}
	if !strings.Contains(text, "## 输出 #1") {
		t.Fatalf("sample output section should remain, got: %s", text)
	}
}

func TestGenAll_NormalizesDescriptionAndTestcases(t *testing.T) {
	fence := "```"
	payload := map[string]any{
		"title": "Hello World",
		"description": "## 题目描述\n\n输出一行 Hello World。\n\n## 输入格式\n\n本题无输入。\n\n## 输出格式\n\n输出固定字符串。\n\n" +
			"## 输入 #1\n\n" + fence + "\n" + fence + "\n\n" +
			"## 输出 #1\n\n" + fence + "\nHello World\n" + fence + "\n\n" +
			"## 数据范围\n\n无。",
		"solution_idea_md": "## 算法分析\n\n直接输出。\n\n## 实现要点\n\n- 输出固定字符串\n\n## 复杂度分析\n\n时间复杂度 $O(1)$，空间复杂度 $O(1)$",
		"solution_md":      "## 题目分析\n\n直接输出。\n\n## 算法与做法\n\n调用输出语句即可。\n\n## 参考实现\n\n```cpp\n#include <bits/stdc++.h>\nusing namespace std;\nint main(){ cout << \"Hello World\\n\"; }\n```\n\n## 复杂度分析\n\n时间复杂度 $O(1)$，空间复杂度 $O(1)$",
		"testcases": []map[string]string{
			{"input": "", "expected_output": "Hello World\n"},
			{"input": fence + "\n" + fence, "expected_output": fence + "\nHello World\n" + fence},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeChatResponse(t, w, string(body))
	}))
	defer srv.Close()

	p := NewPrompts(&config.Config{
		AIEnabled:      true,
		BifrostBaseURL: srv.URL,
		BifrostAPIKey:  "test",
		BifrostModel:   "test",
		AIPromptGenAll: "prompt",
	}, &Client{BaseURL: srv.URL, APIKey: "test", Model: "test", HTTP: &http.Client{}})

	got, _, _, err := p.GenAll(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("GenAll: %v", err)
	}
	if strings.Contains(got.Description, "## 输入 #1") {
		t.Fatalf("empty sample input section should be removed, got: %s", got.Description)
	}
	if len(got.Testcases) != 1 {
		t.Fatalf("expected deduplicated testcase list, got %#v", got.Testcases)
	}
	if got.Testcases[0].Input != "" || got.Testcases[0].ExpectedOutput != "Hello World" {
		t.Fatalf("unexpected testcase normalization result: %#v", got.Testcases[0])
	}
}
