package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/liteoj/liteoj/backend/internal/events"
	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/models"
)

// Join 让当前用户加入题单。如果题单设有密码，必须提供正确密码。
// 已加入时幂等。被管理员踢出后会写入 problem_set_bans，此时拒绝加入
// （admin 自身不受封禁限制）。
func (h *ProblemSetHandler) Join(c *gin.Context) {
	psid, _ := strconv.Atoi(c.Param("id"))
	uid := middleware.CurrentUserID(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.ErrForbidden})
		return
	}
	isAdmin := middleware.CurrentRole(c) == models.RoleAdmin
	access, err := loadProblemSetAccess(h.DB, uint(psid), uid, isAdmin, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !access.Found || (!isAdmin && !access.VisibleToStudent()) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	if access.IsBanned {
		c.JSON(http.StatusForbidden, gin.H{"error": "banned"})
		return
	}
	// 密码校验：仅对非管理员生效。
	if access.HasPassword() && !isAdmin {
		var body struct {
			Password string `json:"password"`
		}
		_ = c.ShouldBindJSON(&body)
		if body.Password != access.ProblemSet.Password {
			c.JSON(http.StatusForbidden, gin.H{"error": "password_incorrect"})
			return
		}
	}
	// 幂等加入：若已存在 membership，直接返回。
	if !access.IsMember {
		m := models.ProblemSetMember{ProblemSetID: uint(psid), UserID: uid, JoinedAt: time.Now()}
		if err := h.DB.Create(&m).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if h.Broker != nil {
			h.Broker.Publish(events.Event{
				Type: "problemset:members:changed",
				Data: map[string]any{"id": uint(psid), "user_id": uid},
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"joined": true})
}
