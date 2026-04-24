package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/events"
	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/models"
)

// HomeHandler 负责公开首页 markdown 的读/写。写操作通过 AdminHandler.UpdateHome
// 暴露；这里只管读 + 首次访问时 lazy 初始化默认内容。
type HomeHandler struct {
	DB     *gorm.DB
	Broker *events.Broker
}

// ensureSingleton 确保 id=1 的单例行存在——首次启动或 seed 缺失时兜底创建。
func (h *HomeHandler) ensureSingleton() {
	var hp models.HomePage
	if err := h.DB.First(&hp, 1).Error; err == nil {
		return
	}
	h.DB.Create(&models.HomePage{ID: 1, Content: i18n.DefaultHomeMarkdown, UpdatedAt: time.Now()})
}

func (h *HomeHandler) Get(c *gin.Context) {
	h.ensureSingleton()
	var hp models.HomePage
	h.DB.First(&hp, 1)
	c.JSON(http.StatusOK, gin.H{"content": hp.Content, "updated_at": hp.UpdatedAt})
}
