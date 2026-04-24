package models

import "time"

// Problem is the canonical problem record students and admins see. The AI
// "import" textarea that feeds the generation flows is intentionally NOT
// stored here — it lives only in the admin form as transient input, so the
// DB stays the source of truth for the structured fields alone.
type Problem struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Title          string    `gorm:"size:255;not null" json:"title"`
	Description    string    `gorm:"type:text" json:"description"`
	Difficulty     string    `gorm:"size:32;index:idx_problems_difficulty;index:idx_problems_visible_difficulty,priority:2" json:"difficulty"`
	TimeLimitMS    int       `gorm:"default:1000" json:"time_limit_ms"`
	MemoryLimitMB  int       `gorm:"default:256" json:"memory_limit_mb"`
	SolutionMD     string    `gorm:"type:text" json:"solution_md"`
	SolutionIdeaMD string    `gorm:"type:text" json:"solution_idea_md"`
	CreatedBy      uint      `json:"created_by"`
	Visible        bool      `gorm:"default:true;index:idx_problems_visible;index:idx_problems_visible_difficulty,priority:1" json:"visible"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ProblemTag normalizes the many-to-many relation between problems and tags.
// The composite primary key guarantees one tag can only be attached once per
// problem, while the secondary (tag_id, problem_id) index accelerates the
// problem list's tag filter.
type ProblemTag struct {
	ProblemID uint `gorm:"primaryKey;autoIncrement:false;index:idx_problem_tags_tag_problem,priority:2;not null" json:"problem_id"`
	TagID     uint `gorm:"primaryKey;autoIncrement:false;index:idx_problem_tags_tag_problem,priority:1;not null" json:"tag_id"`
}

type Testcase struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	ProblemID      uint   `gorm:"index;not null" json:"problem_id"`
	Input          string `gorm:"type:text" json:"input"`
	ExpectedOutput string `gorm:"type:text" json:"expected_output"`
	OrderIndex     int    `gorm:"default:0" json:"order_index"`
}
