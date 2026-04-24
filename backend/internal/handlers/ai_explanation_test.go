package handlers

import "testing"

func TestSanitizeAdminAIExplanation_HidesLegacyEnvelopeForAdmin(t *testing.T) {
	raw := `{"ok":false,"reason":"随机字符或占位变量名","explanation":""}`
	if !isLegacyAnalyzeEnvelope(raw) {
		t.Fatalf("expected legacy analyze envelope to be detected")
	}
	if got := sanitizeAdminAIExplanation(raw, true); got != "" {
		t.Fatalf("admin view should hide legacy analyze envelope, got %q", got)
	}
	if got := sanitizeAdminAIExplanation(raw, false); got != raw {
		t.Fatalf("non-admin view should keep original text, got %q", got)
	}
}
