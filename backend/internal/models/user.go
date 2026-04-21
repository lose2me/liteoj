package models

import "time"

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleStudent Role = "student"
)

type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Name         string     `gorm:"size:128" json:"name"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Role         Role       `gorm:"size:16;not null;default:student" json:"role"`
	LastSeenAt   *time.Time `gorm:"index" json:"last_seen_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
