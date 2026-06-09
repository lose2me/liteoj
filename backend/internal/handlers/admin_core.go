package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/cache"
	"github.com/liteoj/liteoj/backend/internal/control"
	"github.com/liteoj/liteoj/backend/internal/events"
	"github.com/liteoj/liteoj/backend/internal/services/judge"
)

type AdminHandler struct {
	DB     *gorm.DB
	Live   *control.LiveConfig
	Cache  *cache.Cache
	Broker *events.Broker
	Queue  *judge.Queue
}

// publishPS / publishProblem / publishMembers 是 SSE 广播的轻薄包装——让各个
// 写操作 handler 在成功后一行把事件推出去。前端据此对相应页面做重拉。
func (h *AdminHandler) publishPS(psid uint) {
	if h.Broker != nil {
		h.Broker.Publish(events.Event{Type: "problemset:changed", Data: map[string]any{"id": psid}})
	}
}

func (h *AdminHandler) publishProblem(pid uint) {
	if h.Broker != nil {
		h.Broker.Publish(events.Event{Type: "problem:changed", Data: map[string]any{"id": pid}})
	}
}

func (h *AdminHandler) publishMembers(psid uint, uid uint) {
	if h.Broker != nil {
		h.Broker.Publish(events.Event{
			Type: "problemset:members:changed",
			Data: map[string]any{"id": psid, "user_id": uid},
		})
	}
}

type homeReq struct {
	Content string `json:"content"`
}

// UpdateHome 覆盖写入首页单例。Get 路径是公开的 /api/home，这里挂在 admin
// group 下所以只有管理员能改。写入成功后广播 home:changed，让正打开 / 的学
// 生/游客立即看到新内容。
func (h *AdminHandler) UpdateHome(c *gin.Context) {
	var req homeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	var hp struct {
		ID uint
	}
	h.DB.Table("home_pages").FirstOrCreate(&hp, map[string]any{"id": 1})
	if err := h.DB.Table("home_pages").Where("id = ?", 1).Updates(map[string]any{
		"content":    req.Content,
		"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if h.Broker != nil {
		h.Broker.Publish(events.Event{Type: "home:changed", Data: nil})
	}
	c.JSON(200, gin.H{"ok": true})
}
