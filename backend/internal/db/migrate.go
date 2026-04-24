package db

import (
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/models"
)

// EnsureSchema creates the current fixed schema and required indexes.
func EnsureSchema(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.User{},
		&models.TagGroup{},
		&models.Tag{},
		&models.Problem{},
		&models.ProblemTag{},
		&models.Testcase{},
		&models.ProblemSet{},
		&models.ProblemSetItem{},
		&models.ProblemSetMember{},
		&models.ProblemSetBan{},
		&models.Submission{},
		&models.AITask{},
		&models.HomePage{},
	); err != nil {
		return err
	}
	return ensurePerformanceIndexes(db)
}

// ensurePerformanceIndexes adds the indexes the current fixed schema relies on.
func ensurePerformanceIndexes(db *gorm.DB) error {
	for _, sql := range []string{
		`CREATE INDEX IF NOT EXISTS idx_problems_visible ON problems (visible)`,
		`CREATE INDEX IF NOT EXISTS idx_problems_difficulty ON problems (difficulty)`,
		`CREATE INDEX IF NOT EXISTS idx_problems_visible_difficulty ON problems (visible, difficulty)`,
		`CREATE INDEX IF NOT EXISTS idx_problem_tags_problem ON problem_tags (problem_id, tag_id)`,
		`CREATE INDEX IF NOT EXISTS idx_problem_tags_tag_problem ON problem_tags (tag_id, problem_id)`,
		`CREATE INDEX IF NOT EXISTS idx_problem_sets_visible ON problem_sets (visible)`,
		`CREATE INDEX IF NOT EXISTS idx_ps_members_user ON problem_set_members (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ps_bans_user ON problem_set_bans (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ai_tasks_subject ON ai_tasks (subject)`,
		`CREATE INDEX IF NOT EXISTS idx_status_subject ON ai_tasks (status, subject)`,
		`CREATE INDEX IF NOT EXISTS idx_status_kind_subject ON ai_tasks (status, kind, subject)`,
	} {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}
