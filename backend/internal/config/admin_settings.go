package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type AdminSettings struct {
	Home  AdminSettingsHome  `json:"home"`
	Judge AdminSettingsJudge `json:"judge"`
	AI    AdminSettingsAI    `json:"ai"`
}

type AdminSettingsHome struct {
	Content string `json:"content"`
}

type AdminSettingsJudge struct {
	BaseURL               string   `json:"base_url"`
	Langs                 []string `json:"langs"`
	DefaultCPUMS          int      `json:"default_cpu_ms"`
	DefaultMemMB          int      `json:"default_mem_mb"`
	QueueWorkers          int      `json:"queue_workers"`
	QueueCap              int      `json:"queue_cap"`
	MaxWaitSeconds        int      `json:"max_wait_seconds"`
	SubmitIntervalSeconds int      `json:"submit_interval_seconds"`
}

type AdminSettingsAI struct {
	Enabled            bool   `json:"enabled"`
	BifrostBaseURL     string `json:"bifrost_base_url"`
	BifrostAPIKey      string `json:"bifrost_api_key"`
	BifrostModel       string `json:"bifrost_model"`
	QueueWorkers       int    `json:"queue_workers"`
	QueueCap           int    `json:"queue_cap"`
	MaxWaitSeconds     int    `json:"max_wait_seconds"`
	PromptWrongAnswer  string `json:"prompt_wrong_answer"`
	PromptOptimize     string `json:"prompt_optimize"`
	PromptTag          string `json:"prompt_tag"`
	PromptGenTitle     string `json:"prompt_gen_title"`
	PromptGenDesc      string `json:"prompt_gen_desc"`
	PromptGenIdea      string `json:"prompt_gen_idea"`
	PromptGenExplain   string `json:"prompt_gen_explain"`
	PromptGenTestcases string `json:"prompt_gen_testcases"`
	PromptGenAll       string `json:"prompt_gen_all"`
}

func LoadAdminSettings() (*AdminSettings, string, error) {
	path := adminSettingsPath()
	s, err := loadAdminSettingsFromPath(path)
	return s, path, err
}

func SaveAdminSettings(settings *AdminSettings) (string, error) {
	path := adminSettingsPath()
	return path, saveAdminSettingsToPath(path, settings)
}

func loadAdminSettingsFromPath(path string) (*AdminSettings, error) {
	t, err := loadTomlConfigFromPath(path)
	if err != nil {
		return nil, err
	}
	applyTomlDefaults(&t)
	return adminSettingsFromToml(t), nil
}

func saveAdminSettingsToPath(path string, settings *AdminSettings) error {
	if settings == nil {
		settings = &AdminSettings{}
	}

	t, err := loadTomlConfigFromPath(path)
	if err != nil {
		return err
	}
	applyTomlDefaults(&t)

	t.Judge.BaseURL = strings.TrimSpace(settings.Judge.BaseURL)
	t.Judge.Langs = normalizeSettingsSlice(settings.Judge.Langs)
	t.Judge.DefaultCPUMS = settings.Judge.DefaultCPUMS
	t.Judge.DefaultMemMB = settings.Judge.DefaultMemMB
	t.Judge.QueueWorkers = settings.Judge.QueueWorkers
	t.Judge.QueueCap = settings.Judge.QueueCap
	t.Judge.MaxWaitSeconds = settings.Judge.MaxWaitSeconds
	t.Judge.SubmitIntervalSeconds = nonNegativeInt(settings.Judge.SubmitIntervalSeconds)

	t.AI.Enabled = settings.AI.Enabled
	t.AI.BifrostBaseURL = strings.TrimSpace(settings.AI.BifrostBaseURL)
	t.AI.BifrostAPIKey = strings.TrimSpace(settings.AI.BifrostAPIKey)
	t.AI.BifrostModel = strings.TrimSpace(settings.AI.BifrostModel)
	t.AI.QueueWorkers = settings.AI.QueueWorkers
	t.AI.QueueCap = settings.AI.QueueCap
	t.AI.MaxWaitSeconds = settings.AI.MaxWaitSeconds
	t.AI.PromptWrongAnswer = settings.AI.PromptWrongAnswer
	t.AI.PromptOptimize = settings.AI.PromptOptimize
	t.AI.PromptTag = settings.AI.PromptTag
	t.AI.PromptGenTitle = settings.AI.PromptGenTitle
	t.AI.PromptGenDesc = settings.AI.PromptGenDesc
	t.AI.PromptGenIdea = settings.AI.PromptGenIdea
	t.AI.PromptGenExplain = settings.AI.PromptGenExplain
	t.AI.PromptGenTestcases = settings.AI.PromptGenTestcases
	t.AI.PromptGenAll = settings.AI.PromptGenAll

	data, err := toml.Marshal(t)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}

func loadTomlConfigFromPath(path string) (tomlConfig, error) {
	var t tomlConfig
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return t, nil
		}
		return t, err
	}
	if err := toml.Unmarshal(data, &t); err != nil {
		return t, err
	}
	return t, nil
}

func adminSettingsFromToml(t tomlConfig) *AdminSettings {
	return &AdminSettings{
		Judge: AdminSettingsJudge{
			BaseURL:               t.Judge.BaseURL,
			Langs:                 normalizeSettingsSlice(t.Judge.Langs),
			DefaultCPUMS:          t.Judge.DefaultCPUMS,
			DefaultMemMB:          t.Judge.DefaultMemMB,
			QueueWorkers:          t.Judge.QueueWorkers,
			QueueCap:              t.Judge.QueueCap,
			MaxWaitSeconds:        t.Judge.MaxWaitSeconds,
			SubmitIntervalSeconds: t.Judge.SubmitIntervalSeconds,
		},
		AI: AdminSettingsAI{
			Enabled:            t.AI.Enabled,
			BifrostBaseURL:     t.AI.BifrostBaseURL,
			BifrostAPIKey:      t.AI.BifrostAPIKey,
			BifrostModel:       t.AI.BifrostModel,
			QueueWorkers:       t.AI.QueueWorkers,
			QueueCap:           t.AI.QueueCap,
			MaxWaitSeconds:     t.AI.MaxWaitSeconds,
			PromptWrongAnswer:  t.AI.PromptWrongAnswer,
			PromptOptimize:     t.AI.PromptOptimize,
			PromptTag:          t.AI.PromptTag,
			PromptGenTitle:     t.AI.PromptGenTitle,
			PromptGenDesc:      t.AI.PromptGenDesc,
			PromptGenIdea:      t.AI.PromptGenIdea,
			PromptGenExplain:   t.AI.PromptGenExplain,
			PromptGenTestcases: t.AI.PromptGenTestcases,
			PromptGenAll:       t.AI.PromptGenAll,
		},
	}
}

func applyTomlDefaults(t *tomlConfig) {
	t.App.Port = or(t.App.Port, "8080")
	t.App.Mode = or(t.App.Mode, "dev")

	t.DB.Driver = or(t.DB.Driver, "sqlite")
	t.DB.DSN = or(t.DB.DSN, "./data/liteoj.db")

	t.JWT.Secret = or(t.JWT.Secret, "change-me")
	t.JWT.TTLHours = orInt(t.JWT.TTLHours, 24)

	t.AdminInit.Username = or(t.AdminInit.Username, "admin")
	t.AdminInit.Password = or(t.AdminInit.Password, "admin123")
	t.AdminInit.Name = or(t.AdminInit.Name, "超级管理员")

	t.Judge.BaseURL = or(t.Judge.BaseURL, "http://127.0.0.1:5050")
	t.Judge.Langs = orSlice(t.Judge.Langs, []string{"c", "cpp", "java", "python"})
	t.Judge.DefaultCPUMS = orInt(t.Judge.DefaultCPUMS, 1000)
	t.Judge.DefaultMemMB = orInt(t.Judge.DefaultMemMB, 256)
	t.Judge.QueueWorkers = orInt(t.Judge.QueueWorkers, 1)
	t.Judge.QueueCap = orInt(t.Judge.QueueCap, 256)
	t.Judge.MaxWaitSeconds = orInt(t.Judge.MaxWaitSeconds, 120)
	t.Judge.SubmitIntervalSeconds = nonNegativeInt(t.Judge.SubmitIntervalSeconds)

	t.AI.BifrostModel = or(t.AI.BifrostModel, "deepseek-chat")
	t.AI.QueueWorkers = orInt(t.AI.QueueWorkers, 2)
	t.AI.QueueCap = orInt(t.AI.QueueCap, 32)
	t.AI.MaxWaitSeconds = orInt(t.AI.MaxWaitSeconds, 180)

	t.Upload.Dir = or(t.Upload.Dir, "./data/uploads")
}

func adminSettingsPath() string {
	if p := resolveConfigPath(); p != "" {
		return p
	}
	if p := os.Getenv("LITEOJ_CONFIG"); p != "" {
		return p
	}
	return "config.toml"
}

func normalizeSettingsSlice(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		v := strings.TrimSpace(item)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
