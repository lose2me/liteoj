package judge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/liteoj/liteoj/backend/internal/models"
)

// TestcaseResult is a per-case record rendered to the client.
type TestcaseResult struct {
	Index    int    `json:"index"`
	Verdict  string `json:"verdict"`
	TimeMS   int    `json:"time_ms"`
	MemoryKB int    `json:"memory_kb"`
	Message  string `json:"message,omitempty"`
}

// RunnerInput is the unit of work a judge worker consumes.
type RunnerInput struct {
	Lang       string
	Code       string
	Testcases  []models.Testcase
	CPULimitMS int
	MemLimitMB int
}

type RunnerOutput struct {
	Verdict       string
	Message       string
	TimeMS        int
	MemoryKB      int
	CaseResults   []TestcaseResult
	CaseResultRaw string // JSON of CaseResults for storage
}

// Runner wraps a go-judge client and turns submissions into verdicts.
type Runner struct {
	Client *Client
}

func NewRunner(client *Client) *Runner { return &Runner{Client: client} }

// Judge compiles (if needed) and runs each testcase. Output comparison ignores
// trailing whitespace on each line and trailing empty lines.
func (r *Runner) Judge(ctx context.Context, in RunnerInput) (*RunnerOutput, error) {
	lang, ok := Languages[in.Lang]
	if !ok {
		return &RunnerOutput{Verdict: models.VerdictCE, Message: "unsupported language"}, nil
	}

	cpuNS := int64(in.CPULimitMS) * 1_000_000
	memBytes := int64(in.MemLimitMB) * 1024 * 1024
	if cpuNS == 0 {
		cpuNS = 1_000_000_000
	}
	if memBytes == 0 {
		memBytes = 256 * 1024 * 1024
	}

	var binFileID string
	if lang.Compile != nil {
		cmds := []Cmd{{
			Args: lang.Compile,
			Env:  append([]string{"PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin"}, lang.Env...),
			Files: []*File{
				{Content: ""},
				{Name: "stdout", Max: 10240},
				{Name: "stderr", Max: 10240},
			},
			CPULimit:      10_000_000_000,
			MemoryLimit:   512 * 1024 * 1024,
			ProcLimit:     50,
			CopyIn:        map[string]File{lang.Src: {Content: in.Code}},
			CopyOutCached: []string{lang.CompileOut},
		}}
		results, err := r.Client.Run(ctx, cmds)
		if err != nil {
			return nil, err
		}
		if len(results) != 1 {
			return nil, fmt.Errorf("compile: unexpected result count %d", len(results))
		}
		res := results[0]
		if res.Status != "Accepted" {
			msg := res.Files["stderr"]
			if msg == "" {
				msg = res.Error
			}
			return &RunnerOutput{Verdict: models.VerdictCE, Message: truncate(msg, 4096)}, nil
		}
		binFileID = res.FileIDs[lang.CompileOut]
		defer func() {
			if binFileID != "" {
				_ = r.Client.DeleteFile(context.Background(), binFileID)
			}
		}()
	}

	// Per-testcase execution.
	cases := make([]TestcaseResult, 0, len(in.Testcases))
	overall := models.VerdictAC
	var overallMsg string
	var maxTime, maxMem int

	for idx, tc := range in.Testcases {
		copyIn := map[string]File{}
		if binFileID != "" {
			copyIn[lang.CompileOut] = File{FileID: binFileID}
		} else {
			copyIn[lang.Src] = File{Content: in.Code}
		}
		runArgs := lang.Run
		cmd := Cmd{
			Args:        runArgs,
			Env:         append([]string{"PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin"}, lang.Env...),
			Files:       []*File{{Content: tc.Input}, {Name: "stdout", Max: 1 << 20}, {Name: "stderr", Max: 1 << 16}},
			CPULimit:    cpuNS,
			ClockLimit:  cpuNS * 3,
			MemoryLimit: memBytes,
			ProcLimit:   50,
			CopyIn:      copyIn,
		}
		results, err := r.Client.Run(ctx, []Cmd{cmd})
		if err != nil {
			return nil, err
		}
		res := results[0]
		tcTimeMS := int(res.Time / 1_000_000)
		tcMemKB := int(res.Memory / 1024)
		if tcTimeMS > maxTime {
			maxTime = tcTimeMS
		}
		if tcMemKB > maxMem {
			maxMem = tcMemKB
		}

		verdict := mapStatus(res.Status, res.ExitStatus)
		msg := ""
		if verdict == models.VerdictAC {
			verdict = compareOutput(res.Files["stdout"], tc.ExpectedOutput)
		}
		if verdict != models.VerdictAC {
			if stderr := strings.TrimSpace(res.Files["stderr"]); stderr != "" {
				msg = truncate(stderr, 2048)
			} else if res.Error != "" {
				msg = res.Error
			}
		}
		cases = append(cases, TestcaseResult{
			Index: idx + 1, Verdict: verdict,
			TimeMS: tcTimeMS, MemoryKB: tcMemKB,
			Message: msg,
		})
		if verdict != models.VerdictAC && overall == models.VerdictAC {
			overall = verdict
			overallMsg = msg
		}
	}

	raw, _ := json.Marshal(cases)
	return &RunnerOutput{
		Verdict:       overall,
		Message:       overallMsg,
		TimeMS:        maxTime,
		MemoryKB:      maxMem,
		CaseResults:   cases,
		CaseResultRaw: string(raw),
	}, nil
}

// mapStatus translates a go-judge `Status` field into our verdict enum. The
// split mirrors the user-facing definitions:
//   Signalled          → RE  (segfault / div-by-zero / stack overflow / OOB)
//   Nonzero Exit       → SE  (non-signal abnormal exit)
//   Dangerous Syscall  → SE  (sandbox caught forbidden syscall)
//   Internal Error / ""→ UKE (judge machine itself malfunctioned)
// Output Limit Exceeded maps to OLE (was previously collapsed into WA).
func mapStatus(status string, exitStatus int) string {
	switch status {
	case "Accepted":
		return models.VerdictAC
	case "Time Limit Exceeded":
		return models.VerdictTLE
	case "Memory Limit Exceeded":
		return models.VerdictMLE
	case "Output Limit Exceeded":
		return models.VerdictOLE
	case "Signalled":
		return models.VerdictRE
	case "Nonzero Exit Status", "Dangerous Syscall":
		return models.VerdictSE
	case "Internal Error", "":
		return models.VerdictUKE
	default:
		if exitStatus != 0 {
			return models.VerdictRE
		}
		return models.VerdictUKE
	}
}

// compareOutput decides AC / PE / WA by two tiers of equality:
//   byte-exact      → AC
//   normalized      → PE (trailing whitespace / trailing empty lines only)
//   otherwise       → WA
// Called only when the sandbox reported Accepted; a failing sandbox status
// short-circuits before we get here.
func compareOutput(got, expected string) string {
	if got == expected {
		return models.VerdictAC
	}
	if normalize(got) == normalize(expected) {
		return models.VerdictPE
	}
	return models.VerdictWA
}

func normalize(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t\r")
	}
	// drop trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "... (truncated)"
}
