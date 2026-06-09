package control

import (
	"strings"
	"time"

	"github.com/liteoj/liteoj/backend/internal/config"
	"github.com/liteoj/liteoj/backend/internal/services/ai"
	"github.com/liteoj/liteoj/backend/internal/services/judge"
)

// LiveConfig owns the runtime config pointer plus the few long-lived clients
// that need manual refresh when config.toml changes.
type LiveConfig struct {
	Runtime     *config.RuntimeConfig
	JudgeClient *judge.Client
	AIClient    *ai.Client
	AIRunner    *ai.Runner
}

func NewLiveConfig(cfg *config.Config, judgeClient *judge.Client, aiClient *ai.Client, aiRunner *ai.Runner) *LiveConfig {
	return &LiveConfig{
		Runtime:     config.NewRuntimeConfig(cfg),
		JudgeClient: judgeClient,
		AIClient:    aiClient,
		AIRunner:    aiRunner,
	}
}

func (l *LiveConfig) Current() *config.Config {
	if l == nil || l.Runtime == nil {
		return nil
	}
	return l.Runtime.Current()
}

func (l *LiveConfig) Store(cfg *config.Config) {
	if l == nil || cfg == nil {
		return
	}
	if l.Runtime != nil {
		l.Runtime.Store(cfg)
	}
	if l.JudgeClient != nil {
		l.JudgeClient.BaseURL = strings.TrimSpace(cfg.JudgeBaseURL)
	}
	if l.AIClient != nil {
		l.AIClient.BaseURL = strings.TrimSpace(cfg.BifrostBaseURL)
		l.AIClient.APIKey = strings.TrimSpace(cfg.BifrostAPIKey)
		l.AIClient.Model = strings.TrimSpace(cfg.BifrostModel)
	}
	if l.AIRunner != nil {
		l.AIRunner.SetMaxWait(time.Duration(cfg.AIMaxWaitSeconds) * time.Second)
	}
}
