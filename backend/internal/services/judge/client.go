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

type Cmd struct {
	Args              []string            `json:"args"`
	Env               []string            `json:"env,omitempty"`
	Files             []*File             `json:"files"`
	CPULimit          int64               `json:"cpuLimit,omitempty"`
	ClockLimit        int64               `json:"clockLimit,omitempty"`
	MemoryLimit       int64               `json:"memoryLimit,omitempty"`
	StackLimit        int64               `json:"stackLimit,omitempty"`
	ProcLimit         int64               `json:"procLimit,omitempty"`
	CopyIn            map[string]File     `json:"copyIn,omitempty"`
	CopyOut           []string            `json:"copyOut,omitempty"`
	CopyOutCached     []string            `json:"copyOutCached,omitempty"`
	CopyOutDir        string              `json:"copyOutDir,omitempty"`
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
	// HTTP drives the /run path: keep-alive on, longer timeout, shared across
	// concurrent submissions so TCP/HTTP bookkeeping is amortized.
	HTTP *http.Client
	// ProbeHTTP is for GET /version health checks only. It deliberately opens
	// a fresh TCP connection each time (DisableKeepAlives) and runs on a
	// shorter timeout. This sidesteps an annoying failure mode we saw when
	// go-judge sits behind Windows' `netsh portproxy`: the iphlpsvc relay
	// occasionally poisons pooled sockets so pooled /version calls silently
	// hang until ctx deadline, while a fresh connect always succeeds. /run
	// keeps the pooled client because that path is actively exercised and
	// recovers quickly; /version is low-volume and can eat the handshake cost.
	ProbeHTTP *http.Client
}

func NewClient(base string) *Client {
	return &Client{
		BaseURL: base,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
		ProbeHTTP: &http.Client{
			Timeout: 3 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives: true,
				MaxIdleConns:      -1,
			},
		},
	}
}

func (c *Client) Run(ctx context.Context, cmds []Cmd) ([]Result, error) {
	body := map[string]any{"cmd": cmds}
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/run", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("go-judge: %d %s", resp.StatusCode, string(data))
	}
	var out []Result
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("go-judge decode: %w: %s", err, string(data))
	}
	return out, nil
}

func (c *Client) DeleteFile(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.BaseURL+"/file/"+id, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Version pings go-judge's /version endpoint. Used as a liveness probe and to
// surface the sandbox build string in the admin dashboard. Uses ProbeHTTP
// (no keep-alive) to avoid stale-socket flakes when routed through Windows
// netsh portproxy / iphlpsvc.
func (c *Client) Version(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/version", nil)
	if err != nil {
		return "", err
	}
	resp, err := c.ProbeHTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("go-judge: %d %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return strings.TrimSpace(string(data)), nil
}
