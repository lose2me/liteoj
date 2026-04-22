package models

type TagGroup struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"uniqueIndex;size:64;not null" json:"name"`
	OrderIndex int    `gorm:"default:0" json:"order_index"`
}

type Tag struct {
	ID uint `gorm:"primaryKey" json:"id"`
	// uniq_group_name 保证同一 group 内标签名唯一；防 AI 打标签时把字典外或重复名
	// 写进来（runner 会以 group+name 反查 tag id，重名会挑到错误的那条）。
	GroupID    uint   `gorm:"index;index:uniq_group_name,unique,priority:1;not null" json:"group_id"`
	Name       string `gorm:"size:64;index:uniq_group_name,unique,priority:2;not null" json:"name"`
	OrderIndex int    `gorm:"default:0" json:"order_index"`
}
