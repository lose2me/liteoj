package models

type TagGroup struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"uniqueIndex;size:64;not null" json:"name"`
	OrderIndex int    `gorm:"default:0" json:"order_index"`
}

type Tag struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	GroupID    uint   `gorm:"index;not null" json:"group_id"`
	Name       string `gorm:"size:64;not null" json:"name"`
	OrderIndex int    `gorm:"default:0" json:"order_index"`
}
