package handlers

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/models"
)

func TestResetAutoIncrementSequencesSQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Problem{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if err := db.Create(&models.Problem{Title: "old-1"}).Error; err != nil {
		t.Fatalf("insert old-1: %v", err)
	}
	if err := db.Create(&models.Problem{Title: "old-2"}).Error; err != nil {
		t.Fatalf("insert old-2: %v", err)
	}
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Problem{}).Error; err != nil {
		t.Fatalf("delete all: %v", err)
	}
	if err := resetAutoIncrementSequences(db, "sqlite", []string{"problems"}); err != nil {
		t.Fatalf("reset sequence: %v", err)
	}

	p := models.Problem{Title: "fresh"}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("insert fresh: %v", err)
	}
	if p.ID != 1 {
		t.Fatalf("expected problem id 1 after reset, got %d", p.ID)
	}
}
