package handlers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/models"
)

func TestProblemSetDetailHiddenSetReturnsNotFoundForStudent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openProblemSetAccessTestDB(t)
	ps := models.ProblemSet{Title: "Hidden Set"}
	if err := db.Create(&ps).Error; err != nil {
		t.Fatalf("create problemset: %v", err)
	}
	if err := db.Model(&models.ProblemSet{}).Where("id = ?", ps.ID).Update("visible", false).Error; err != nil {
		t.Fatalf("hide problemset: %v", err)
	}

	h := &ProblemSetHandler{DB: db}
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/problemsets/"+strconv.Itoa(int(ps.ID)), nil)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(ps.ID))}}
	ctx.Set(middleware.CtxUserID, uint(101))
	ctx.Set(middleware.CtxRole, models.RoleStudent)

	h.Detail(ctx)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for hidden set detail, got code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestProblemSetJoinBannedStudentRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openProblemSetAccessTestDB(t)
	ps := models.ProblemSet{Title: "Visible Set", Visible: true}
	if err := db.Create(&ps).Error; err != nil {
		t.Fatalf("create problemset: %v", err)
	}
	user := models.User{Username: "stu1", Name: "学生甲", PasswordHash: "x", Role: models.RoleStudent}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	ban := models.ProblemSetBan{ProblemSetID: ps.ID, UserID: user.ID}
	if err := db.Create(&ban).Error; err != nil {
		t.Fatalf("create ban: %v", err)
	}

	h := &ProblemSetHandler{DB: db}
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/problemsets/"+strconv.Itoa(int(ps.ID))+"/join", nil)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(ps.ID))}}
	ctx.Set(middleware.CtxUserID, user.ID)
	ctx.Set(middleware.CtxRole, models.RoleStudent)

	h.Join(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for banned join, got code=%d body=%s", w.Code, w.Body.String())
	}
}

func openProblemSetAccessTestDB(t *testing.T) *gorm.DB {
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
		&models.ProblemSetItem{},
		&models.Problem{},
		&models.Submission{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}
