package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/auth"
	"github.com/liteoj/liteoj/backend/internal/config"
	"github.com/liteoj/liteoj/backend/internal/middleware"
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

func TestResetDataInvalidatesAllSessions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.TagGroup{},
		&models.Tag{},
		&models.Problem{},
		&models.ProblemTag{},
		&models.Testcase{},
		&models.ProblemSet{},
		&models.ProblemSetItem{},
		&models.ProblemSetMember{},
		&models.ProblemSetBan{},
		&models.Submission{},
		&models.AITask{},
		&models.HomePage{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cfg := &config.Config{
		DBDriver:                     "sqlite",
		JWTSecret:                    "test-secret",
		JWTTTLHours:                  1,
		AdminInitUsername:            "admin",
		AdminInitPassword:            "admin123",
		AdminInitName:                "Admin",
		AdminDangerSecondaryPassword: "danger-pass",
		UploadDir:                    t.TempDir() + "\\uploads",
	}
	adminHash, err := auth.HashPassword("admin123")
	if err != nil {
		t.Fatalf("hash admin password: %v", err)
	}
	userHash, err := auth.HashPassword("user12345")
	if err != nil {
		t.Fatalf("hash user password: %v", err)
	}
	admin := models.User{
		Username:     "admin",
		Name:         "Admin",
		PasswordHash: adminHash,
		Role:         models.RoleAdmin,
		LoginVersion: 3,
	}
	user := models.User{
		Username:     "alice",
		Name:         "Alice",
		PasswordHash: userHash,
		Role:         models.RoleStudent,
		LoginVersion: 2,
	}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	adminToken, err := auth.Issue(cfg.JWTSecret, cfg.JWTTTL(), &admin)
	if err != nil {
		t.Fatalf("issue admin token: %v", err)
	}
	userToken, err := auth.Issue(cfg.JWTSecret, cfg.JWTTTL(), &user)
	if err != nil {
		t.Fatalf("issue user token: %v", err)
	}

	authH := &AuthHandler{DB: db, C: cfg}
	adminH := &AdminHandler{DB: db, C: cfg}
	router := gin.New()
	api := router.Group("/api")
	api.POST("/auth/login", authH.Login)
	authed := api.Group("")
	authed.Use(middleware.Auth(cfg, db))
	authed.GET("/me", authH.Me)
	adminGroup := authed.Group("/admin")
	adminGroup.Use(middleware.AdminOnly())
	adminGroup.POST("/reset-data", adminH.ResetData)

	w := authPostJSON(t, router, "/api/admin/reset-data", adminToken, map[string]string{
		"secondary_password": "danger-pass",
	}, http.StatusOK)
	var resetResp struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resetResp); err != nil {
		t.Fatalf("decode reset response: %v", err)
	}
	if !resetResp.OK {
		t.Fatalf("expected reset response ok=true")
	}

	authGet(t, router, "/api/me", adminToken, http.StatusUnauthorized)
	authGet(t, router, "/api/me", userToken, http.StatusUnauthorized)

	newAdminToken := authLogin(t, router, "admin", "admin123", http.StatusOK)
	authGet(t, router, "/api/me", newAdminToken, http.StatusOK)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+newAdminToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected new admin token to work, got %d body=%s", rec.Code, rec.Body.String())
	}
}
