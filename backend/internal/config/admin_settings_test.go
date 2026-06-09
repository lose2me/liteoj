package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func TestLoadAdminSettingsAppliesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("[judge]\nbase_url = \"http://judge.local\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	got, err := loadAdminSettingsFromPath(path)
	if err != nil {
		t.Fatalf("load admin settings: %v", err)
	}

	if got.Judge.BaseURL != "http://judge.local" {
		t.Fatalf("unexpected judge base url: %s", got.Judge.BaseURL)
	}
	if got.Judge.DefaultCPUMS != 1000 || got.Judge.DefaultMemMB != 256 {
		t.Fatalf("unexpected judge defaults: %+v", got.Judge)
	}
	if got.AI.BifrostModel != "deepseek-chat" || got.AI.QueueWorkers != 2 {
		t.Fatalf("unexpected ai defaults: %+v", got.AI)
	}
}

func TestSaveAdminSettingsPreservesUneditedSections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	original := []byte(`
[db]
driver = "postgres"
dsn = "postgres://liteoj"

[jwt]
secret = "secret-1"
ttl_hours = 72

[admin_init]
username = "root"
password = "secret-admin"
name = "Root Admin"

[admin_danger]
secondary_password = "danger-secret"

[judge]
base_url = "http://old-judge"
langs = ["c", "cpp"]
default_cpu_ms = 1000
default_mem_mb = 256
queue_workers = 1
queue_cap = 256
max_wait_seconds = 120
submit_interval_seconds = 20

[ai]
enabled = false
bifrost_base_url = ""
bifrost_api_key = ""
bifrost_model = "deepseek-chat"
queue_workers = 2
queue_cap = 32
max_wait_seconds = 180
prompt_gen_testcases = "old testcase prompt"
`)
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	settings, err := loadAdminSettingsFromPath(path)
	if err != nil {
		t.Fatalf("load admin settings: %v", err)
	}
	settings.Judge.BaseURL = "http://new-judge"
	settings.Judge.Langs = []string{"cpp", "python", "python"}
	settings.Judge.SubmitIntervalSeconds = 20
	settings.AI.Enabled = true
	settings.AI.BifrostBaseURL = "https://api.example.com"
	settings.AI.PromptGenTestcases = "new testcase prompt"

	if err := saveAdminSettingsToPath(path, settings); err != nil {
		t.Fatalf("save admin settings: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}
	var got tomlConfig
	if err := toml.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal saved config: %v", err)
	}

	if got.DB.Driver != "postgres" || got.DB.DSN != "postgres://liteoj" {
		t.Fatalf("db section changed unexpectedly: %+v", got.DB)
	}
	if got.JWT.Secret != "secret-1" || got.JWT.TTLHours != 72 {
		t.Fatalf("jwt section changed unexpectedly: %+v", got.JWT)
	}
	if got.AdminInit.Username != "root" || got.AdminInit.Password != "secret-admin" || got.AdminInit.Name != "Root Admin" {
		t.Fatalf("admin_init section changed unexpectedly: %+v", got.AdminInit)
	}
	if got.AdminDanger.SecondaryPassword != "danger-secret" {
		t.Fatalf("admin_danger section changed unexpectedly: %+v", got.AdminDanger)
	}
	if got.Judge.BaseURL != "http://new-judge" {
		t.Fatalf("judge base url not updated: %s", got.Judge.BaseURL)
	}
	if len(got.Judge.Langs) != 2 || got.Judge.Langs[0] != "cpp" || got.Judge.Langs[1] != "python" {
		t.Fatalf("judge langs not normalized: %#v", got.Judge.Langs)
	}
	if got.Judge.SubmitIntervalSeconds != 20 {
		t.Fatalf("judge submit interval not updated: %d", got.Judge.SubmitIntervalSeconds)
	}
	if !got.AI.Enabled || got.AI.BifrostBaseURL != "https://api.example.com" {
		t.Fatalf("ai section not updated: %+v", got.AI)
	}
	if got.AI.PromptGenTestcases != "new testcase prompt" {
		t.Fatalf("ai testcase prompt not updated: %+v", got.AI)
	}
	text := string(data)
	if !strings.Contains(text, "submit_interval_seconds = 20") {
		t.Fatalf("saved config missing submit_interval_seconds: %s", text)
	}
	if strings.Contains(text, "submit_limit_per_minute") {
		t.Fatalf("saved config should not contain submit_limit_per_minute: %s", text)
	}
}
