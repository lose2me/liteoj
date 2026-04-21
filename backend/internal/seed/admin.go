package seed

import (
	"errors"
	"log"

	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/auth"
	"github.com/liteoj/liteoj/backend/internal/config"
	"github.com/liteoj/liteoj/backend/internal/models"
)

// EnsureAdmin creates an admin from config if no admin exists.
func EnsureAdmin(db *gorm.DB, c *config.Config) error {
	var count int64
	if err := db.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if c.AdminInitUsername == "" || c.AdminInitPassword == "" {
		return errors.New("seed: ADMIN_INIT_USERNAME / ADMIN_INIT_PASSWORD must be set for first boot")
	}
	hash, err := auth.HashPassword(c.AdminInitPassword)
	if err != nil {
		return err
	}
	u := &models.User{
		Username:     c.AdminInitUsername,
		Name:         c.AdminInitName,
		PasswordHash: hash,
		Role:         models.RoleAdmin,
	}
	if err := db.Create(u).Error; err != nil {
		return err
	}
	log.Printf("seed: created initial admin %q", u.Username)
	return nil
}
