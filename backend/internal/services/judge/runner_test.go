package judge

import (
	"testing"

	"github.com/liteoj/liteoj/backend/internal/models"
)

func TestCompareOutputIgnoresTrailingWhitespace(t *testing.T) {
	got := "hello world \r\n\r\n"
	expected := "hello world\n"

	if verdict := compareOutput(got, expected); verdict != models.VerdictAC {
		t.Fatalf("expected AC for trailing-whitespace-only diff, got %s", verdict)
	}
}

func TestCompareOutputRejectsRealDifference(t *testing.T) {
	got := "hello world!"
	expected := "hello world"

	if verdict := compareOutput(got, expected); verdict != models.VerdictWA {
		t.Fatalf("expected WA for content diff, got %s", verdict)
	}
}
