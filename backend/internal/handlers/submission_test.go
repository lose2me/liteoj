package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/config"
	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/models"
)

func TestSubmissionHandlerSubmitRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openSubmissionTestDB(t)
	now := time.Now()
	prob := seedSubmissionProblem(t, db)
	uid := uint(7)

	rows := []models.Submission{
		{UserID: uid, ProblemID: prob.ID, Language: "cpp", Code: "int main(){}", Verdict: models.VerdictWA, CreatedAt: now.Add(-20 * time.Second)},
		{UserID: uid, ProblemID: prob.ID, Language: "cpp", Code: "int main(){}", Verdict: models.VerdictCE, CreatedAt: now.Add(-40 * time.Second)},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed submissions: %v", err)
	}

	h := &SubmissionHandler{
		DB: db,
		C: &config.Config{
			JudgeLangs:           []string{"cpp"},
			SubmitLimitPerMinute: 2,
		},
	}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost,
		"/api/problems/"+strconv.Itoa(int(prob.ID))+"/submit",
		strings.NewReader(`{"language":"cpp","code":"int main(){return 0;}"}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(prob.ID))}}
	ctx.Set(middleware.CtxUserID, uid)
	ctx.Set(middleware.CtxRole, models.RoleStudent)

	h.Submit(ctx)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error != i18n.ErrSubmitRateLimited(2) {
		t.Fatalf("unexpected error message: %q", resp.Error)
	}

	var total int64
	if err := db.Model(&models.Submission{}).Where("user_id = ?", uid).Count(&total).Error; err != nil {
		t.Fatalf("count submissions: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 submissions after rejection, got %d", total)
	}
}

func TestEnforceSubmitRateLimitIgnoresOldSubmissions(t *testing.T) {
	db := openSubmissionTestDB(t)
	now := time.Now()
	prob := seedSubmissionProblem(t, db)
	uid := uint(9)

	rows := []models.Submission{
		{UserID: uid, ProblemID: prob.ID, Language: "cpp", Code: "old", Verdict: models.VerdictWA, CreatedAt: now.Add(-2 * time.Minute)},
		{UserID: uid, ProblemID: prob.ID, Language: "cpp", Code: "fresh", Verdict: models.VerdictWA, CreatedAt: now.Add(-10 * time.Second)},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed submissions: %v", err)
	}

	h := &SubmissionHandler{
		DB: db,
		C:  &config.Config{SubmitLimitPerMinute: 2},
	}
	if err := h.enforceSubmitRateLimit(uid, now); err != nil {
		t.Fatalf("expected old submissions to be ignored, got %v", err)
	}
}

func openSubmissionTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Problem{}, &models.Testcase{}, &models.Submission{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func seedSubmissionProblem(t *testing.T, db *gorm.DB) models.Problem {
	t.Helper()

	prob := models.Problem{
		Title:         "A+B",
		Visible:       true,
		TimeLimitMS:   1000,
		MemoryLimitMB: 128,
	}
	if err := db.Create(&prob).Error; err != nil {
		t.Fatalf("create problem: %v", err)
	}
	tc := models.Testcase{
		ProblemID:      prob.ID,
		Input:          "1 2\n",
		ExpectedOutput: "3\n",
	}
	if err := db.Create(&tc).Error; err != nil {
		t.Fatalf("create testcase: %v", err)
	}
	return prob
}
