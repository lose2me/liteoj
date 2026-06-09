package models

import "time"

type ProblemSet struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	Title            string     `gorm:"size:255;not null" json:"title"`
	AllowedLangsJSON string     `gorm:"type:text" json:"-"`               // JSON: ["cpp","python"] | "" => 全部允许
	AllowedLangs     []string   `gorm:"-" json:"allowed_langs,omitempty"` // transient: 与 AllowedLangsJSON 互转
	Password         string     `gorm:"size:128" json:"password,omitempty"`
	StartTime        *time.Time `json:"start_time,omitempty"`
	EndTime          *time.Time `json:"end_time,omitempty"`
	// Visible 控制学生是否能在 /problemsets 看到此题单。关闭（false）后：
	// 列表过滤、详情 404、Join 直接 404；admin 不受影响。默认 true，老数据
	// AutoMigrate 新列会填 default。
	Visible bool `gorm:"default:true;index:idx_problem_sets_visible" json:"visible"`
	// 学生端白名单开关。学生在独立题目页默认看不到思路 / 题解 / AI，
	// 只有进入显式启用这些能力的题单上下文时才开放。底层沿用历史列名
	// disable_*，这样旧数据会自然“反转”为白名单语义，无需额外迁移。
	EnableIdea     bool `gorm:"column:disable_idea;default:false" json:"enable_idea"`
	EnableSolution bool `gorm:"column:disable_solution;default:false" json:"enable_solution"`
	EnableAI       bool `gorm:"column:disable_ai;default:false" json:"enable_ai"`
	// EnableBonus 控制题单是否启用“每日加分”功能。关闭时学生侧不展示加分列，
	// 后台也不允许维护该题单的每日分数表。默认 false。
	EnableBonus bool      `gorm:"default:false;index:idx_problem_sets_bonus" json:"enable_bonus"`
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProblemSetItem struct {
	ID uint `gorm:"primaryKey" json:"id"`
	// uniq_ps_problem 保证同一题单里同一题只挂一行，防止重复加题（前端拖拽 + 后端
	// 批量更新都会经过 SetProblemSetItems，这个约束是最后一道防线）。
	// idx_ps_order 加速题单视图按 order_index 排序。
	ProblemSetID uint `gorm:"index;index:uniq_ps_problem,unique,priority:1;index:idx_ps_order,priority:1;not null" json:"problemset_id"`
	ProblemID    uint `gorm:"index;index:uniq_ps_problem,unique,priority:2;not null" json:"problem_id"`
	OrderIndex   int  `gorm:"default:0;index:idx_ps_order,priority:2" json:"order_index"`
}

// ProblemSetMember 表记录用户是否已加入某题单。在题单上下文内做题（即提交
// 带 problemset_id）前，非管理员必须先成为成员。独立题目页不受此约束。
type ProblemSetMember struct {
	ID           uint `gorm:"primaryKey" json:"id"`
	ProblemSetID uint `gorm:"uniqueIndex:uniq_ps_member,priority:1;not null" json:"problemset_id"`
	// 额外 user_id 索引支撑 /problemsets 列表里"拉我加入过哪些题单"这类反向查。
	UserID   uint      `gorm:"index:idx_ps_members_user;uniqueIndex:uniq_ps_member,priority:2;not null" json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
}

// ProblemSetBan 记录被管理员从题单踢出的用户。管理员踢出即永久封禁，被封禁
// 的用户不可再次加入该题单（直到管理员显式解除）。独立于 ProblemSetMember：
// 踢出流程会删除 member 行并写入 ban 行。
type ProblemSetBan struct {
	ID           uint `gorm:"primaryKey" json:"id"`
	ProblemSetID uint `gorm:"uniqueIndex:uniq_ps_ban,priority:1;not null" json:"problemset_id"`
	// 同 ProblemSetMember：当前用户是否被某些题单踢出，需要按 user_id 反查。
	UserID   uint      `gorm:"index:idx_ps_bans_user;uniqueIndex:uniq_ps_ban,priority:2;not null" json:"user_id"`
	BannedAt time.Time `json:"banned_at"`
	BannedBy uint      `json:"banned_by,omitempty"`
}

// ProblemSetDailyBonus 是题单的“每日一张表”加分记录。每行表示某个学生在某天
// 于该题单下获得的分数；(problem_set_id, user_id, score_date) 唯一。
// score_date 使用 YYYY-MM-DD 字符串，避免时区/数据库 date 类型差异。
type ProblemSetDailyBonus struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ProblemSetID uint      `gorm:"index:idx_ps_bonus_ps_date,priority:1;uniqueIndex:uniq_ps_bonus,priority:1;not null" json:"problemset_id"`
	UserID       uint      `gorm:"index:idx_ps_bonus_user;uniqueIndex:uniq_ps_bonus,priority:2;not null" json:"user_id"`
	ScoreDate    string    `gorm:"size:10;index:idx_ps_bonus_ps_date,priority:2;uniqueIndex:uniq_ps_bonus,priority:3;not null" json:"score_date"`
	Score        int       `gorm:"default:0;not null" json:"score"`
	CreatedBy    uint      `json:"created_by,omitempty"`
	UpdatedBy    uint      `json:"updated_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
