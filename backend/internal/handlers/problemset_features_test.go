package handlers

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/models"
)

func TestLoadStudentFeatureFlags(t *testing.T) {
	db := openProblemsetFeatureTestDB(t)
	u := models.User{Username: "stu1", Name: "学生甲", PasswordHash: "x", Role: models.RoleStudent}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	ps := models.ProblemSet{
		Title:          "Set A",
		EnableIdea:     true,
		EnableSolution: false,
		EnableAI:       true,
	}
	if err := db.Create(&ps).Error; err != nil {
		t.Fatalf("create problemset: %v", err)
	}
	if err := db.Create(&models.ProblemSetMember{
		ProblemSetID: ps.ID,
		UserID:       u.ID,
	}).Error; err != nil {
		t.Fatalf("create member: %v", err)
	}

	admin := loadStudentFeatureFlags(db, 0, true, nil, nil)
	if !admin.Idea || !admin.Solution || !admin.AI {
		t.Fatalf("admin should bypass feature restrictions, got %+v", admin)
	}

	standalone := loadStudentFeatureFlags(db, 0, false, nil, nil)
	if standalone.Idea || standalone.Solution || standalone.AI {
		t.Fatalf("student standalone should default to all features disabled, got %+v", standalone)
	}

	inSet := loadStudentFeatureFlags(db, u.ID, false, &ps.ID, nil)
	if !inSet.Idea || inSet.Solution || !inSet.AI {
		t.Fatalf("unexpected in-set flags: %+v", inSet)
	}
}

func openProblemsetFeatureTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.ProblemSet{},
		&models.ProblemSetMember{},
		&models.ProblemSetBan{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}
