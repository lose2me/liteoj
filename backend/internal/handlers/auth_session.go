package handlers

import (
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/models"
)

const loginVersionBumpExpr = "COALESCE(login_version, 0) + 1"

func bumpUserLoginVersion(tx *gorm.DB, userID uint) error {
	return tx.Model(&models.User{}).
		Where("id = ?", userID).
		Update("login_version", gorm.Expr(loginVersionBumpExpr)).
		Error
}

func loadUserTokenState(tx *gorm.DB, userID uint) (*models.User, error) {
	var u models.User
	if err := tx.Select("id", "username", "name", "role", "login_version").First(&u, userID).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func updateUserPassword(tx *gorm.DB, userID uint, hash string) error {
	return tx.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"password_hash": hash,
			"login_version": gorm.Expr(loginVersionBumpExpr),
		}).
		Error
}
