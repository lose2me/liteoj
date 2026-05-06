package handlers

import (
	"bytes"
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

func TestLoginInvalidatesPreviousToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openAuthTestDB(t)
	seedAuthTestUser(t, db, "alice", "secret123")
	router := newAuthTestRouter(db)

	first := authLogin(t, router, "alice", "secret123", http.StatusOK)
	second := authLogin(t, router, "alice", "secret123", http.StatusOK)

	authGet(t, router, "/api/me", first, http.StatusUnauthorized)
	authGet(t, router, "/api/me", second, http.StatusOK)

	var u models.User
	if err := db.Select("login_version").Where("username = ?", "alice").First(&u).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if u.LoginVersion != 2 {
		t.Fatalf("expected login_version 2 after two logins, got %d", u.LoginVersion)
	}
}

func TestChangePasswordInvalidatesCurrentToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openAuthTestDB(t)
	seedAuthTestUser(t, db, "alice", "secret123")
	router := newAuthTestRouter(db)

	token := authLogin(t, router, "alice", "secret123", http.StatusOK)
	authPostJSON(t, router, "/api/me/password", token, map[string]string{
		"old_password": "secret123",
		"new_password": "newsecret456",
	}, http.StatusOK)

	authGet(t, router, "/api/me", token, http.StatusUnauthorized)
	authLogin(t, router, "alice", "secret123", http.StatusUnauthorized)
	fresh := authLogin(t, router, "alice", "newsecret456", http.StatusOK)
	authGet(t, router, "/api/me", fresh, http.StatusOK)
}

func openAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func seedAuthTestUser(t *testing.T, db *gorm.DB, username, password string) {
	t.Helper()

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if err := db.Create(&models.User{
		Username:     username,
		Name:         "Alice",
		PasswordHash: hash,
		Role:         models.RoleStudent,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func newAuthTestRouter(db *gorm.DB) *gin.Engine {
	cfg := &config.Config{
		JWTSecret:   "test-secret",
		JWTTTLHours: 1,
	}
	authH := &AuthHandler{DB: db, C: cfg}

	r := gin.New()
	r.POST("/api/auth/login", authH.Login)

	authed := r.Group("/api")
	authed.Use(middleware.Auth(cfg, db))
	authed.GET("/me", authH.Me)
	authed.POST("/me/password", authH.ChangePassword)
	return r
}

func authLogin(t *testing.T, router http.Handler, username, password string, wantStatus int) string {
	t.Helper()

	w := authPostJSON(t, router, "/api/auth/login", "", map[string]string{
		"username": username,
		"password": password,
	}, wantStatus)
	if wantStatus != http.StatusOK {
		return ""
	}
	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if resp.Token == "" {
		t.Fatalf("expected token in login response")
	}
	return resp.Token
}

func authGet(t *testing.T, router http.Handler, path, token string, wantStatus int) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != wantStatus {
		t.Fatalf("GET %s: expected %d, got %d body=%s", path, wantStatus, w.Code, w.Body.String())
	}
	return w
}

func authPostJSON(t *testing.T, router http.Handler, path, token string, body any, wantStatus int) *httptest.ResponseRecorder {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != wantStatus {
		t.Fatalf("POST %s: expected %d, got %d body=%s", path, wantStatus, w.Code, w.Body.String())
	}
	return w
}
