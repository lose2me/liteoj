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
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		parts := strings.SplitN(raw, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "bad auth header"})
			return
		}
		claims, err := auth.Parse(c.JWTSecret, parts[1])
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		applyClaims(ctx, claims, db)
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

// applyClaims 把 claims 写入 ctx 并跑 last_seen 节流更新（与 Auth 共享）。
func applyClaims(ctx *gin.Context, claims *auth.Claims, db *gorm.DB) {
	ctx.Set(CtxUserID, claims.UserID)
	ctx.Set(CtxUsername, claims.Username)
	ctx.Set(CtxRole, claims.Role)
	if db == nil {
		return
	}
	now := time.Now()
	seen.mu.Lock()
	last := seen.m[claims.UserID]
	update := now.Sub(last) > throttleWindow
	if update {
		seen.m[claims.UserID] = now
	}
	seen.mu.Unlock()
	if update {
		db.Model(&models.User{}).Where("id = ?", claims.UserID).Update("last_seen_at", now)
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		r, _ := ctx.Get(CtxRole)
		if r != models.RoleAdmin {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin only"})
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
