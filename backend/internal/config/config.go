package config

import (
	"log"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Config is the flat, read-only runtime configuration surface handlers consume.
// It's produced from a nested TOML file (or defaults) at Load time.
type Config struct {
	AppPort string
	AppMode string

	DBDriver string
	DBDSN    string

	JWTSecret   string
	JWTTTLHours int

	AdminInitUsername string
	AdminInitPassword string
	AdminInitName     string
	// AdminDangerSecondaryPassword gates destructive admin-only actions such as
	// clearing all business data while preserving the tag dictionary.
	AdminDangerSecondaryPassword string

	JudgeBaseURL      string
	JudgeLangs        []string
	JudgeDefaultCPU   int
	JudgeDefaultMem   int
	JudgeQueueWorkers int
	JudgeQueueCap     int
	// JudgeMaxWaitSeconds 是 judge worker 从取出 job 到落最终 verdict 的
	// 上限秒数。主要作用是 go-judge 不可达时让卡死的提交在可控时间内落 SE，
	// 避免 queue_workers=1 + 长 TCP 超时造成后续提交被无限期堵。默认 120；
	// 判题机通畅时几乎用不到这个上限。
	JudgeMaxWaitSeconds int

	BifrostBaseURL string
	BifrostAPIKey  string
	BifrostModel   string
	AIEnabled      bool
	// AIQueueWorkers 控制同时有多少个 AI 任务并发调用上游模型。
	AIQueueWorkers int
	AIQueueCap     int
	// AIMaxWaitSeconds 是后端等待一次 AI 调用的上限秒数，超过即把 context
	// 取消、任务判失败。覆盖所有 kind（analyze / optimize / tag / gen_*），
	// 方便把需要长推理的机型（DeepSeek-V3 等）统一放宽到 5~10 分钟。
	AIMaxWaitSeconds int
	AIPromptWA       string
	AIPromptOpt      string
	AIPromptTag      string
	// Problem authoring prompts — all sourced from raw content pasted into
	// the ProblemEdit "详细" field. Prompts must come from config.toml; the
	// code ships no defaults so the admin owns the format/style/heading
	// conventions end-to-end.
	AIPromptGenTitle   string
	AIPromptGenDesc    string
	AIPromptGenIdea    string
	AIPromptGenExplain string
	AIPromptGenAll     string

	UploadDir string
}

// tomlConfig is the on-disk TOML structure. We keep it separate from Config so
// that call sites see a flat surface while the file uses readable sections.
type tomlConfig struct {
	App struct {
		Port string `toml:"port"`
		Mode string `toml:"mode"`
	} `toml:"app"`

	DB struct {
		Driver string `toml:"driver"`
		DSN    string `toml:"dsn"`
	} `toml:"db"`

	JWT struct {
		Secret   string `toml:"secret"`
		TTLHours int    `toml:"ttl_hours"`
	} `toml:"jwt"`

	AdminInit struct {
		Username string `toml:"username"`
		Password string `toml:"password"`
		Name     string `toml:"name"`
	} `toml:"admin_init"`

	AdminDanger struct {
		SecondaryPassword string `toml:"secondary_password"`
	} `toml:"admin_danger"`

	Judge struct {
		BaseURL        string   `toml:"base_url"`
		Langs          []string `toml:"langs"`
		DefaultCPUMS   int      `toml:"default_cpu_ms"`
		DefaultMemMB   int      `toml:"default_mem_mb"`
		QueueWorkers   int      `toml:"queue_workers"`
		QueueCap       int      `toml:"queue_cap"`
		MaxWaitSeconds int      `toml:"max_wait_seconds"`
	} `toml:"judge"`

	AI struct {
		Enabled           bool   `toml:"enabled"`
		BifrostBaseURL    string `toml:"bifrost_base_url"`
		BifrostAPIKey     string `toml:"bifrost_api_key"`
		BifrostModel      string `toml:"bifrost_model"`
		QueueWorkers      int    `toml:"queue_workers"`
		QueueCap          int    `toml:"queue_cap"`
		MaxWaitSeconds    int    `toml:"max_wait_seconds"`
		PromptWrongAnswer string `toml:"prompt_wrong_answer"`
		PromptOptimize    string `toml:"prompt_optimize"`
		PromptTag         string `toml:"prompt_tag"`
		PromptGenTitle    string `toml:"prompt_gen_title"`
		PromptGenDesc     string `toml:"prompt_gen_desc"`
		PromptGenIdea     string `toml:"prompt_gen_idea"`
		PromptGenExplain  string `toml:"prompt_gen_explain"`
		PromptGenAll      string `toml:"prompt_gen_all"`
	} `toml:"ai"`

	Upload struct {
		Dir string `toml:"dir"`
	} `toml:"upload"`
}

// Load reads config.toml (or the file at LITEOJ_CONFIG) and fills in defaults
// for any missing keys. A missing file is not fatal.
func Load() *Config {
	path := resolveConfigPath()
	var t tomlConfig
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := toml.Unmarshal(data, &t); err != nil {
				log.Fatalf("config: parse %s: %v", path, err)
			}
			log.Printf("config: loaded %s", path)
		} else {
			log.Printf("config: %s unreadable (%v), using defaults", path, err)
		}
	} else {
		log.Printf("config: no config.toml found, using defaults")
	}

	return &Config{
		AppPort:                      or(t.App.Port, "8080"),
		AppMode:                      or(t.App.Mode, "dev"),
		DBDriver:                     or(t.DB.Driver, "sqlite"),
		DBDSN:                        or(t.DB.DSN, "./data/liteoj.db"),
		JWTSecret:                    or(t.JWT.Secret, "change-me"),
		JWTTTLHours:                  orInt(t.JWT.TTLHours, 24),
		AdminInitUsername:            or(t.AdminInit.Username, "admin"),
		AdminInitPassword:            or(t.AdminInit.Password, "admin123"),
		AdminInitName:                or(t.AdminInit.Name, "超级管理员"),
		AdminDangerSecondaryPassword: t.AdminDanger.SecondaryPassword,
		JudgeBaseURL:                 or(t.Judge.BaseURL, "http://127.0.0.1:5050"),
		JudgeLangs:                   orSlice(t.Judge.Langs, []string{"c", "cpp", "java", "python"}),
		JudgeDefaultCPU:              orInt(t.Judge.DefaultCPUMS, 1000),
		JudgeDefaultMem:              orInt(t.Judge.DefaultMemMB, 256),
		JudgeQueueWorkers:            orInt(t.Judge.QueueWorkers, 1),
		JudgeQueueCap:                orInt(t.Judge.QueueCap, 256),
		JudgeMaxWaitSeconds:          orInt(t.Judge.MaxWaitSeconds, 120),
		BifrostBaseURL:               t.AI.BifrostBaseURL,
		BifrostAPIKey:                t.AI.BifrostAPIKey,
		BifrostModel:                 or(t.AI.BifrostModel, "deepseek-chat"),
		AIEnabled:                    t.AI.Enabled,
		AIQueueWorkers:               orInt(t.AI.QueueWorkers, 2),
		AIQueueCap:                   orInt(t.AI.QueueCap, 32),
		// 默认 180s：沿用旧 kindTimeout 里最长的 GenAll 预算。旧机型 / 短
		// 推理只需要一个兜底上限；DeepSeek-V3 这类慢思考机型在 config.toml
		// 里把它调到 600 即可。
		AIMaxWaitSeconds:   orInt(t.AI.MaxWaitSeconds, 180),
		AIPromptWA:         t.AI.PromptWrongAnswer,
		AIPromptOpt:        t.AI.PromptOptimize,
		AIPromptTag:        t.AI.PromptTag,
		AIPromptGenTitle:   t.AI.PromptGenTitle,
		AIPromptGenDesc:    t.AI.PromptGenDesc,
		AIPromptGenIdea:    t.AI.PromptGenIdea,
		AIPromptGenExplain: t.AI.PromptGenExplain,
		AIPromptGenAll:     t.AI.PromptGenAll,
		UploadDir:          or(t.Upload.Dir, "./data/uploads"),
	}
}

func (c *Config) JWTTTL() time.Duration {
	return time.Duration(c.JWTTTLHours) * time.Hour
}

func or(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func orInt(v, fallback int) int {
	if v == 0 {
		return fallback
	}
	return v
}

func orSlice(v, fallback []string) []string {
	if len(v) == 0 {
		return fallback
	}
	return v
}

// resolveConfigPath looks at $LITEOJ_CONFIG, then ./config.toml, then up to
// two parents. Running `go run ./cmd/liteoj` from backend/ still finds the
// project-root config.toml.
func resolveConfigPath() string {
	if p := os.Getenv("LITEOJ_CONFIG"); p != "" {
		return p
	}
	candidates := []string{"config.toml", "../config.toml", "../../config.toml"}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}
