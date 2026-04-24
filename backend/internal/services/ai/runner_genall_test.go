package ai

import (
	"testing"

	"github.com/liteoj/liteoj/backend/internal/models"
)

func TestBuildMissingTestcaseRows_AppendsOnlyNewCases(t *testing.T) {
	existing := []models.Testcase{
		{ProblemID: 7, Input: "", ExpectedOutput: "Hello World\n", OrderIndex: 1},
		{ProblemID: 7, Input: "1 2\n", ExpectedOutput: "3\n", OrderIndex: 2},
	}
	generated := []GeneratedTestcase{
		{Input: "", ExpectedOutput: "Hello World"},
		{Input: "1 2\r\n", ExpectedOutput: "3"},
		{Input: "2 3", ExpectedOutput: "5"},
	}

	rows := buildMissingTestcaseRows(7, existing, generated)
	if len(rows) != 1 {
		t.Fatalf("expected exactly one new testcase, got %#v", rows)
	}
	if rows[0].ProblemID != 7 {
		t.Fatalf("problem id mismatch: %#v", rows[0])
	}
	if rows[0].Input != "2 3" || rows[0].ExpectedOutput != "5" {
		t.Fatalf("unexpected testcase content: %#v", rows[0])
	}
	if rows[0].OrderIndex != 3 {
		t.Fatalf("expected appended order index 3, got %#v", rows[0])
	}
}
