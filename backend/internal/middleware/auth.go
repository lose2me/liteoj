package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/auth"
	"github.com/liteoj/liteoj/backend/internal/config"
	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/models"
)

const (
	CtxUserID   = "uid"
	CtxUsername = "usr"
	CtxRole     = "role"
)

// lastSeenThrottle avoids hitting the DB on every authed request: we only
// update last_seen_at when the cached ts is older than throttleWindow.
type lastSeenThrottle struct {
	mu sync.Mutex
	m  map[uint]time.Time
}

var seen = lastSeenThrottle{m: map[uint]time.Time{}}

const throttleWindow = 60 * time.Second

func Auth(c *config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		raw := ctx.GetHeader("Authorization")
		if raw == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.ErrMissingToken})
			return
		}
		parts := strings.SplitN(raw, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.ErrBadAuthHeader})
			return
		}
		claims, err := auth.Parse(c.JWTSecret, parts[1])
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.ErrInvalidToken})
			return
		}
		if !applyClaims(ctx, claims, db) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.ErrInvalidToken})
			return
		}
		ctx.Next()
	}
}

// OptionalAuth 同 Auth 但缺省 / 无效 token 不 abort，只是不写入 user ctx。
// 用于公开页面（如 /problems 列表）：匿名能拿到数据，登录用户能额外看到
// my_status 等与身份相关的字段。
func OptionalAuth(c *config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		raw := ctx.GetHeader("Authorization")
		if raw == "" {
			ctx.Next()
			return
		}
		parts := strings.SplitN(raw, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			ctx.Next()
			return
		}
		claims, err := auth.Parse(c.JWTSecret, parts[1])
		if err != nil {
			ctx.Next()
			return
		}
		applyClaims(ctx, claims, db)
		ctx.Next()
	}
}

// applyClaims resolves the token's user id against the database and writes the
// current authoritative user identity into gin ctx. Returning false means the
// token refers to a user that no longer exists.
func applyClaims(ctx *gin.Context, claims *auth.Claims, db *gorm.DB) bool {
	if db == nil {
		ctx.Set(CtxUserID, claims.UserID)
		ctx.Set(CtxUsername, claims.Username)
		ctx.Set(CtxRole, claims.Role)
		return true
	}
	var u models.User
	if err := db.Select("id", "username", "role").First(&u, claims.UserID).Error; err != nil {
		return false
	}
	ctx.Set(CtxUserID, u.ID)
	ctx.Set(CtxUsername, u.Username)
	ctx.Set(CtxRole, u.Role)
	now := time.Now()
	seen.mu.Lock()
	last := seen.m[u.ID]
	update := now.Sub(last) > throttleWindow
	if update {
		seen.m[u.ID] = now
	}
	seen.mu.Unlock()
	if update {
		db.Model(&models.User{}).Where("id = ?", u.ID).Update("last_seen_at", now)
	}
	return true
}

func AdminOnly() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		r, _ := ctx.Get(CtxRole)
		if r != models.RoleAdmin {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": i18n.ErrAdminOnly})
			return
		}
		ctx.Next()
	}
}

func CurrentUserID(ctx *gin.Context) uint {
	v, _ := ctx.Get(CtxUserID)
	id, _ := v.(uint)
	return id
}

func CurrentUsername(ctx *gin.Context) string {
	v, _ := ctx.Get(CtxUsername)
	name, _ := v.(string)
	return name
}

func CurrentRole(ctx *gin.Context) models.Role {
	v, _ := ctx.Get(CtxRole)
	r, _ := v.(models.Role)
	return r
}
