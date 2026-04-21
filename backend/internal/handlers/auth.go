package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/auth"
	"github.com/liteoj/liteoj/backend/internal/config"
	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/models"
)

type AuthHandler struct {
	DB *gorm.DB
	C  *config.Config
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrBadRequest})
		return
	}
	var u models.User
	if err := h.DB.Where("username = ?", req.Username).First(&u).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.ErrBadCredentials})
		return
	}
	if !auth.CheckPassword(u.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.ErrBadCredentials})
		return
	}
	tok, err := auth.Issue(h.C.JWTSecret, h.C.JWTTTL(), &u)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.ErrIssueToken})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": tok,
		"user": gin.H{
			"id": u.ID, "username": u.Username, "name": u.Name, "role": u.Role,
		},
	})
}

type changePwdReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var req changePwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrBadPassword})
		return
	}
	var u models.User
	if err := h.DB.First(&u, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrUserNotFound})
		return
	}
	if !auth.CheckPassword(u.PasswordHash, req.OldPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrOldPassword})
		return
	}
	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.ErrHashFailed})
		return
	}
	if err := h.DB.Model(&u).Update("password_hash", hash).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.ErrUpdateFailed})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AuthHandler) Me(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var u models.User
	if err := h.DB.First(&u, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": u.ID, "username": u.Username, "name": u.Name, "role": u.Role,
	})
}
