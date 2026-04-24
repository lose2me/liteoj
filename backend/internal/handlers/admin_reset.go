package handlers

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/auth"
	"github.com/liteoj/liteoj/backend/internal/events"
	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/models"
)

type resetDataReq struct {
	SecondaryPassword string `json:"secondary_password"`
}

// ResetData clears all business data except the tag dictionary and home-page
// announcement, then recreates the default admin from config.toml as a fresh
// row. No existing user/session is preserved.
func (h *AdminHandler) ResetData(c *gin.Context) {
	var req resetDataReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrBadRequest})
		return
	}

	want := strings.TrimSpace(h.C.AdminDangerSecondaryPassword)
	if want == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrDangerSecondaryPasswordMissing})
		return
	}
	if subtle.ConstantTimeCompare([]byte(req.SecondaryPassword), []byte(want)) != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": i18n.ErrDangerSecondaryPasswordWrong})
		return
	}

	hash, err := auth.HashPassword(h.C.AdminInitPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	seqTables := []string{
		"ai_tasks",
		"submissions",
		"problem_set_bans",
		"problem_set_members",
		"problem_set_items",
		"problem_sets",
		"testcases",
		"problems",
		"users",
	}

	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		for _, model := range []any{
			&models.AITask{},
			&models.Submission{},
			&models.ProblemSetBan{},
			&models.ProblemSetMember{},
			&models.ProblemSetItem{},
			&models.ProblemSet{},
			&models.Testcase{},
			&models.ProblemTag{},
			&models.Problem{},
			&models.User{},
		} {
			if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(model).Error; err != nil {
				return err
			}
		}
		if err := resetAutoIncrementSequences(tx, h.C.DBDriver, seqTables); err != nil {
			return err
		}
		admin := &models.User{
			Username:     h.C.AdminInitUsername,
			Name:         h.C.AdminInitName,
			PasswordHash: hash,
			Role:         models.RoleAdmin,
		}
		return tx.Create(admin).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uploadWarning := ""
	if err := clearUploadDir(h.C.UploadDir); err != nil {
		uploadWarning = err.Error()
	}

	if h.Cache != nil {
		h.Cache.Invalidate("problems:")
	}
	if h.Broker != nil {
		h.Broker.Publish(events.Event{Type: "problem:changed", Data: map[string]any{"id": 0}})
		h.Broker.Publish(events.Event{Type: "problemset:changed", Data: map[string]any{"id": 0}})
		h.Broker.Publish(events.Event{Type: "ai:tasks:changed", Data: nil})
	}

	resp := gin.H{"ok": true}
	if uploadWarning != "" {
		resp["warning"] = i18n.WarnDangerResetUploadCleanupFailed(uploadWarning)
	}
	c.JSON(http.StatusOK, resp)
}

func clearUploadDir(dir string) error {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if filepath.Dir(abs) == abs {
		return errors.New(i18n.ErrDangerRefuseClearRoot)
	}
	base := strings.ToLower(filepath.Base(abs))
	if !strings.Contains(base, "upload") {
		return errors.New(i18n.ErrDangerRefuseClearSuspiciousDir(abs))
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(abs, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func resetAutoIncrementSequences(tx *gorm.DB, driver string, tables []string) error {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", "sqlite":
		for _, table := range tables {
			if err := tx.Exec("DELETE FROM sqlite_sequence WHERE name = ?", table).Error; err != nil {
				return err
			}
		}
		return nil
	case "postgres":
		for _, table := range tables {
			sql := fmt.Sprintf(
				"SELECT setval(pg_get_serial_sequence('%s', 'id'), COALESCE((SELECT MAX(id) FROM %s), 1), EXISTS(SELECT 1 FROM %s))",
				table, table, table,
			)
			if err := tx.Exec(sql).Error; err != nil {
				return err
			}
		}
		return nil
	default:
		return nil
	}
}
