package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/liteoj/liteoj/backend/internal/auth"
	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/models"
)

// BulkCreateUsers upserts a batch of users pasted from the admin UI as three
// aligned text columns (name / username / password). The frontend guarantees
// equal line counts before enabling the submit button, but we still defensively
// reject misaligned payloads here so an API-level caller can't corrupt the
// mapping.
//
// Semantics match the retired Excel-import path it replaces: existing usernames
// are treated as UPDATE (overwrite name + password_hash); new ones are INSERT
// with role=student. Failures are collected per row and returned alongside
// the created/updated counters — the UI surfaces them inline.
func (h *AdminHandler) BulkCreateUsers(c *gin.Context) {
	type bulkReq struct {
		Users []struct {
			Name     string `json:"name"`
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"users"`
	}
	var req bulkReq
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Users) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrBadRequest})
		return
	}

	var created, updated int
	var failures []gin.H
	for i, u := range req.Users {
		row := i + 1
		username := strings.TrimSpace(u.Username)
		name := strings.TrimSpace(u.Name)
		pwd := u.Password // password: do NOT trim, users may want leading/trailing intentionally
		if username == "" || pwd == "" {
			failures = append(failures, gin.H{"row": row, "error": i18n.ErrBulkRowEmpty})
			continue
		}
		hash, err := auth.HashPassword(pwd)
		if err != nil {
			failures = append(failures, gin.H{"row": row, "error": err.Error()})
			continue
		}
		var existing models.User
		tx := h.DB.Where("username = ?", username).First(&existing)
		if tx.Error == nil {
			h.DB.Model(&existing).Updates(map[string]any{"name": name, "password_hash": hash})
			updated++
			continue
		}
		if err := h.DB.Create(&models.User{
			Username:     username,
			Name:         name,
			PasswordHash: hash,
			Role:         models.RoleStudent,
		}).Error; err != nil {
			failures = append(failures, gin.H{"row": row, "error": err.Error()})
			continue
		}
		created++
	}
	c.JSON(http.StatusOK, gin.H{
		"created":  created,
		"updated":  updated,
		"failures": failures,
		"summary":  i18n.BulkSummary(created, updated, len(failures)),
	})
}
