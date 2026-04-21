package models

import "time"

// Problem is the canonical problem record students and admins see. The AI
// "import" textarea that feeds the generation flows is intentionally NOT
// stored here — it lives only in the admin form as transient input, so the
// DB stays the source of truth for the structured fields alone.
type Problem struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Title           string    `gorm:"size:255;not null" json:"title"`
	Description     string    `gorm:"type:text" json:"description"`
	Difficulty      string    `gorm:"size:32" json:"difficulty"`
	TimeLimitMS     int       `gorm:"default:1000" json:"time_limit_ms"`
	MemoryLimitMB   int       `gorm:"default:256" json:"memory_limit_mb"`
	TagsJSON        string    `gorm:"type:text" json:"tags_json"`
	SolutionMD      string    `gorm:"type:text" json:"solution_md"`
	SolutionIdeaMD  string    `gorm:"type:text" json:"solution_idea_md"`
	CreatedBy       uint      `json:"created_by"`
	Visible         bool      `gorm:"default:true" json:"visible"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Testcase struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	ProblemID      uint   `gorm:"index;not null" json:"problem_id"`
	Input          string `gorm:"type:text" json:"input"`
	ExpectedOutput string `gorm:"type:text" json:"expected_output"`
	OrderIndex     int    `gorm:"default:0" json:"order_index"`
}
