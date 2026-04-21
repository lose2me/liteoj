package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/models"
)

type AdminStatsHandler struct {
	DB *gorm.DB
}

// Overview returns the admin dashboard numbers in a single payload, including
// online user count (users whose last_seen_at is within the last 5 minutes).
func (h *AdminStatsHandler) Overview(c *gin.Context) {
	var problems, problemsets, users, submissions, online int64
	h.DB.Model(&models.Problem{}).Count(&problems)
	h.DB.Model(&models.ProblemSet{}).Count(&problemsets)
	h.DB.Model(&models.User{}).Count(&users)
	h.DB.Model(&models.Submission{}).Count(&submissions)
	cutoff := time.Now().Add(-5 * time.Minute)
	h.DB.Model(&models.User{}).Where("last_seen_at IS NOT NULL AND last_seen_at >= ?", cutoff).Count(&online)
	c.JSON(http.StatusOK, gin.H{
		"problems":     problems,
		"problemsets":  problemsets,
		"users":        users,
		"submissions":  submissions,
		"online_users": online,
	})
}

// OnlineUsers returns the list of users recently seen. Useful for the admin
// dashboard presence widget.
func (h *AdminStatsHandler) OnlineUsers(c *gin.Context) {
	cutoff := time.Now().Add(-5 * time.Minute)
	var users []models.User
	h.DB.Where("last_seen_at IS NOT NULL AND last_seen_at >= ?", cutoff).
		Order("last_seen_at DESC").Find(&users)
	type item struct {
		ID         uint      `json:"id"`
		Username   string    `json:"username"`
		Name       string    `json:"name"`
		Role       string    `json:"role"`
		LastSeenAt time.Time `json:"last_seen_at"`
	}
	out := make([]item, 0, len(users))
	for _, u := range users {
		if u.LastSeenAt == nil {
			continue
		}
		out = append(out, item{
			ID: u.ID, Username: u.Username, Name: u.Name, Role: string(u.Role),
			LastSeenAt: *u.LastSeenAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}
