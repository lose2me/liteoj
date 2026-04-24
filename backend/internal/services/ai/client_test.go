package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/liteoj/liteoj/backend/internal/config"
)

// TestChat_TruncatedBodySurfacesError reproduces the production failure mode
// that previously masqueraded as "unexpected end of JSON input": the server
// sends a 200 OK with partial JSON and then drops the connection mid-stream.
// The fixed client must return a body-read error rather than silently parsing
// the truncated bytes and confusing the admin with a useless parser message.
func TestChat_TruncatedBodySurfacesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("hijacker not supported")
		}
		conn, bw, err := hj.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		bw.WriteString("HTTP/1.1 200 OK\r\n")
		bw.WriteString("Content-Type: application/json\r\n")
		// Promise more body than we deliver — io.ReadAll will hit EOF early
		// and surface an error.
		bw.WriteString("Content-Length: 4096\r\n\r\n")
		bw.WriteString(`{"choices":[{"message":{"content":"hel`)
		bw.Flush()
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, APIKey: "test", Model: "test", HTTP: &http.Client{}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	text, raw, err := c.Chat(ctx, []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error from truncated body, got nil")
	}
	if !strings.Contains(err.Error(), "读取响应失败") {
		t.Errorf("error should identify read-body cause, got: %v", err)
	}
	if text != "" {
		t.Errorf("expected empty text, got %q", text)
	}
	if !strings.Contains(raw, "hel") {
		t.Errorf("partial raw body should be returned for diagnostics, got %q", raw)
	}
}

// TestChat_HappyPath verifies the normal 200+complete-JSON case still parses
// the assistant content correctly after the signature change.
func TestChat_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"hello world"}}]}`))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, APIKey: "test", Model: "test", HTTP: &http.Client{}}
	text, raw, err := c.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "hello world" {
		t.Errorf("text = %q, want %q", text, "hello world")
	}
	if !strings.Contains(raw, "hello world") {
		t.Errorf("raw should contain the body, got %q", raw)
	}
}

// TestNewFromConfig_NoHardTimeout proves the HTTP client itself imposes no
// hard timeout — only the per-request context drives request lifetime.
// Locks in the fix for the 60s GenAll cutoff that produced "unexpected end
// of JSON input" in production.
func TestNewFromConfig_NoHardTimeout(t *testing.T) {
	c := NewFromConfig(&config.Config{})
	if c.HTTP.Timeout != 0 {
		t.Errorf("HTTP.Timeout = %v, want 0 (context-driven only)", c.HTTP.Timeout)
	}
}

// TestChat_ContextDeadlineHonored confirms the per-request context still
// cuts off slow upstreams, just via context.DeadlineExceeded rather than the
// previous opaque parser failure.
func TestChat_ContextDeadlineHonored(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("hijacker not supported")
		}
		conn, bw, err := hj.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		bw.WriteString("HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\n")
		bw.Flush()
		time.Sleep(2 * time.Second)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, APIKey: "test", Model: "test", HTTP: &http.Client{}}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_, _, err := c.Chat(ctx, []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected context-driven error")
	}
	// Either body-read error or context error — both are honest signals,
	// unlike the old "unexpected end of JSON input".
	if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "读取响应失败") {
		t.Errorf("error should mention context or read-body, got: %v", err)
	}
}
