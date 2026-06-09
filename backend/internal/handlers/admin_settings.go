package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	configpkg "github.com/liteoj/liteoj/backend/internal/config"
	"github.com/liteoj/liteoj/backend/internal/events"
	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/models"
)

func (h *AdminHandler) GetSettings(c *gin.Context) {
	settings, path, err := configpkg.LoadAdminSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if h.DB != nil {
		settings.Home.Content = h.getHomeContent()
	}
	c.JSON(http.StatusOK, gin.H{
		"path":     path,
		"settings": settings,
	})
}

func (h *AdminHandler) UpdateSettings(c *gin.Context) {
	var req configpkg.AdminSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	path, err := configpkg.SaveAdminSettings(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if h.Live != nil {
		cfg, err := configpkg.LoadFromPath(path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		h.Live.Store(cfg)
		if h.Queue != nil {
			h.Queue.SetJobTimeout(time.Duration(cfg.JudgeMaxWaitSeconds) * time.Second)
		}
	}
	if h.DB != nil {
		if err := h.saveHomeContent(req.Home.Content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":   true,
		"path": path,
	})
}

func (h *AdminHandler) getHomeContent() string {
	var hp models.HomePage
	if err := h.DB.First(&hp, 1).Error; err == nil {
		return hp.Content
	}
	now := time.Now()
	h.DB.Create(&models.HomePage{
		ID:        1,
		Content:   i18n.DefaultHomeMarkdown,
		UpdatedAt: now,
	})
	return i18n.DefaultHomeMarkdown
}

func (h *AdminHandler) saveHomeContent(content string) error {
	var hp struct {
		ID uint
	}
	h.DB.Table("home_pages").FirstOrCreate(&hp, map[string]any{"id": 1})
	if err := h.DB.Table("home_pages").Where("id = ?", 1).Updates(map[string]any{
		"content":    content,
		"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}).Error; err != nil {
		return err
	}
	if h.Broker != nil {
		h.Broker.Publish(events.Event{Type: "home:changed", Data: nil})
	}
	return nil
}
