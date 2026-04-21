package models

import "time"

// HomePage 是"未登录首页"（路由 /）的 markdown 内容——始终单例（id=1）。
// admin 在后台概览页编辑；学生/游客访问 / 时渲染。默认文案由 seed 初始化。
type HomePage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Content   string    `gorm:"type:text" json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
}
