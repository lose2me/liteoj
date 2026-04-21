package models

import "time"

type ProblemSet struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	Title            string     `gorm:"size:255;not null" json:"title"`
	AllowedLangsJSON string     `gorm:"type:text" json:"-"`         // JSON: ["cpp","python"] | "" => 全部允许
	AllowedLangs     []string   `gorm:"-" json:"allowed_langs,omitempty"` // transient: 与 AllowedLangsJSON 互转
	Password         string     `gorm:"size:128" json:"password,omitempty"`
	StartTime        *time.Time `json:"start_time,omitempty"`
	EndTime          *time.Time `json:"end_time,omitempty"`
	// Visible 控制学生是否能在 /problemsets 看到此题单。关闭（false）后：
	// 列表过滤、详情 404、Join 直接 404；admin 不受影响。默认 true，老数据
	// AutoMigrate 新列会填 default。
	Visible bool `gorm:"default:true" json:"visible"`
	// 题单级权限开关。进入题单上下文读题或提交时生效；独立题目页不受影响。
	DisableIdea     bool `gorm:"default:false" json:"disable_idea"`
	DisableSolution bool `gorm:"default:false" json:"disable_solution"`
	DisableAI       bool `gorm:"default:false" json:"disable_ai"`
	CreatedBy       uint       `json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ProblemSetItem struct {
	ID           uint `gorm:"primaryKey" json:"id"`
	ProblemSetID uint `gorm:"index;not null" json:"problemset_id"`
	ProblemID    uint `gorm:"index;not null" json:"problem_id"`
	OrderIndex   int  `gorm:"default:0" json:"order_index"`
}

// ProblemSetMember 表记录用户是否已加入某题单。在题单上下文内做题（即提交
// 带 problemset_id）前，非管理员必须先成为成员。独立题目页不受此约束。
type ProblemSetMember struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ProblemSetID uint      `gorm:"uniqueIndex:uniq_ps_member,priority:1;not null" json:"problemset_id"`
	UserID       uint      `gorm:"uniqueIndex:uniq_ps_member,priority:2;not null" json:"user_id"`
	JoinedAt     time.Time `json:"joined_at"`
}

// ProblemSetBan 记录被管理员从题单踢出的用户。管理员踢出即永久封禁，被封禁
// 的用户不可再次加入该题单（直到管理员显式解除）。独立于 ProblemSetMember：
// 踢出流程会删除 member 行并写入 ban 行。
type ProblemSetBan struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ProblemSetID uint      `gorm:"uniqueIndex:uniq_ps_ban,priority:1;not null" json:"problemset_id"`
	UserID       uint      `gorm:"uniqueIndex:uniq_ps_ban,priority:2;not null" json:"user_id"`
	BannedAt     time.Time `json:"banned_at"`
	BannedBy     uint      `json:"banned_by,omitempty"`
}
