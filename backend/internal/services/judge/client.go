package judge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/liteoj/liteoj/backend/internal/i18n"
)

// Minimal go-judge client. Reference: https://github.com/criyle/go-judge
//
// The /run endpoint accepts one or more "Cmd" specs and returns a result array
// of the same length. We only model the subset we need.

type File struct {
	Src     string `json:"src,omitempty"`
	Content string `json:"content,omitempty"`
	FileID  string `json:"fileId,omitempty"`
	Name    string `json:"name,omitempty"`
	Max     int64  `json:"max,omitempty"`
	Pipe    bool   `json:"pipe,omitempty"`
}

// MarshalJSON keeps go-judge's file union shape valid.
// In particular, an empty stdin must still be encoded as {"content":""}
// rather than {}. The plain `omitempty` tags lose that distinction.
func (f File) MarshalJSON() ([]byte, error) {
	var out map[string]any
	switch {
	case f.Src != "":
		out = map[string]any{"src": f.Src}
	case f.FileID != "":
		out = map[string]any{"fileId": f.FileID}
	case f.Name != "":
		out = map[string]any{"name": f.Name}
		if f.Max != 0 {
			out["max"] = f.Max
		}
		if f.Pipe {
			out["pipe"] = true
		}
	default:
		out = map[string]any{"content": f.Content}
	}
	return json.Marshal(out)
}

type Cmd struct {
	Args          []string        `json:"args"`
	Env           []string        `json:"env,omitempty"`
	Files         []*File         `json:"files"`
	CPULimit      int64           `json:"cpuLimit,omitempty"`
	ClockLimit    int64           `json:"clockLimit,omitempty"`
	MemoryLimit   int64           `json:"memoryLimit,omitempty"`
	StackLimit    int64           `json:"stackLimit,omitempty"`
	ProcLimit     int64           `json:"procLimit,omitempty"`
	CopyIn        map[string]File `json:"copyIn,omitempty"`
	CopyOut       []string        `json:"copyOut,omitempty"`
	CopyOutCached []string        `json:"copyOutCached,omitempty"`
	CopyOutDir    string          `json:"copyOutDir,omitempty"`
}

type Result struct {
	Status     string            `json:"status"`
	ExitStatus int               `json:"exitStatus"`
	Error      string            `json:"error"`
	Time       int64             `json:"time"`   // ns
	Memory     int64             `json:"memory"` // bytes
	RunTime    int64             `json:"runTime"`
	Files      map[string]string `json:"files"`
	FileIDs    map[string]string `json:"fileIds"`
}

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func NewClient(base string) *Client {
	return &Client{
		BaseURL: base,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Run(ctx context.Context, cmds []Cmd) ([]Result, error) {
	body := map[string]any{"cmd": cmds}
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	status, data, err := c.do(ctx, http.MethodPost, "/run", buf)
	if err != nil {
		return nil, err
	}
	if status/100 != 2 {
		return nil, fmt.Errorf("%s", i18n.ErrGoJudgeStatus(status, string(data)))
	}
	var out []Result
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("%s", i18n.ErrGoJudgeDecode(err, string(data)))
	}
	return out, nil
}

func (c *Client) DeleteFile(ctx context.Context, id string) error {
	status, data, err := c.do(ctx, http.MethodDelete, "/file/"+id, nil)
	if err != nil {
		return err
	}
	if status/100 != 2 {
		return fmt.Errorf("go-judge: %d %s", status, strings.TrimSpace(string(data)))
	}
	return nil
}

func (c *Client) do(ctx context.Context, method, path string, body []byte) (int, []byte, error) {
	url := c.BaseURL + path
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return 0, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data, nil
}
