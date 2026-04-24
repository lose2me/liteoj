package handlers

import (
	"encoding/json"
	"strings"
)

type legacyAnalyzeEnvelope struct {
	OK          *bool   `json:"ok"`
	Reason      *string `json:"reason"`
	Explanation *string `json:"explanation"`
}

func isLegacyAnalyzeEnvelope(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(raw, "{") {
		return false
	}
	var parsed legacyAnalyzeEnvelope
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return false
	}
	return parsed.OK != nil && parsed.Reason != nil && parsed.Explanation != nil
}

func sanitizeAdminAIExplanation(explanation string, isAdmin bool) string {
	if isAdmin && isLegacyAnalyzeEnvelope(explanation) {
		return ""
	}
	return explanation
}
