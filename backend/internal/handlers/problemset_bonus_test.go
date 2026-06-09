package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/models"
)

func TestProblemSetDailyBonusSheetCRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openProblemSetBonusTestDB(t)
	ps := models.ProblemSet{Title: "Daily Bonus", EnableBonus: true}
	if err := db.Create(&ps).Error; err != nil {
		t.Fatalf("create problemset: %v", err)
	}
	users := []models.User{
		{Username: "stu1", Name: "学生甲", PasswordHash: "x", Role: models.RoleStudent},
		{Username: "stu2", Name: "学生乙", PasswordHash: "x", Role: models.RoleStudent},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}
	members := []models.ProblemSetMember{
		{ProblemSetID: ps.ID, UserID: users[0].ID, JoinedAt: time.Now()},
		{ProblemSetID: ps.ID, UserID: users[1].ID, JoinedAt: time.Now().Add(-time.Minute)},
	}
	if err := db.Create(&members).Error; err != nil {
		t.Fatalf("create members: %v", err)
	}

	adminH := &AdminHandler{DB: db}
	date := "2026-05-06"

	getSheet := func() struct {
		Date  string `json:"date"`
		Items []struct {
			UserID uint   `json:"user_id"`
			Score  int    `json:"score"`
			Name   string `json:"name"`
		} `json:"items"`
	} {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/admin/problemsets/"+strconv.Itoa(int(ps.ID))+"/bonus?date="+date, nil)
		ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(ps.ID))}}
		adminH.ListProblemSetDailyBonus(ctx)
		if w.Code != http.StatusOK {
			t.Fatalf("list bonus sheet: code=%d body=%s", w.Code, w.Body.String())
		}
		var resp struct {
			Date  string `json:"date"`
			Items []struct {
				UserID uint   `json:"user_id"`
				Score  int    `json:"score"`
				Name   string `json:"name"`
			} `json:"items"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode bonus sheet: %v", err)
		}
		return resp
	}

	initial := getSheet()
	if initial.Date != date {
		t.Fatalf("unexpected date: %q", initial.Date)
	}
	if len(initial.Items) != 2 {
		t.Fatalf("expected 2 members, got %d", len(initial.Items))
	}
	if initial.Items[0].Score != 0 || initial.Items[1].Score != 0 {
		t.Fatalf("expected empty sheet to default to 0 scores, got %+v", initial.Items)
	}

	body := problemSetBonusSaveReq{
		Date: date,
		Items: []problemSetBonusSaveItem{
			{UserID: users[0].ID, Score: 3},
			{UserID: users[1].ID, Score: 1},
		},
	}
	buf, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/admin/problemsets/"+strconv.Itoa(int(ps.ID))+"/bonus", bytes.NewReader(buf))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(ps.ID))}}
	ctx.Set(middleware.CtxUserID, uint(999))
	adminH.SaveProblemSetDailyBonus(ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("save bonus sheet: code=%d body=%s", w.Code, w.Body.String())
	}

	saved := getSheet()
	got := map[uint]int{}
	for _, row := range saved.Items {
		got[row.UserID] = row.Score
	}
	if got[users[0].ID] != 3 || got[users[1].ID] != 1 {
		t.Fatalf("unexpected saved scores: %+v", got)
	}
}

func TestProblemsetRankingBonusIsDisplayOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openProblemSetBonusTestDB(t)
	ps := models.ProblemSet{Title: "Ranked Set", EnableBonus: true}
	if err := db.Create(&ps).Error; err != nil {
		t.Fatalf("create problemset: %v", err)
	}
	prob := models.Problem{Title: "A+B", Visible: true, TimeLimitMS: 1000, MemoryLimitMB: 128}
	if err := db.Create(&prob).Error; err != nil {
		t.Fatalf("create problem: %v", err)
	}
	if err := db.Create(&models.ProblemSetItem{ProblemSetID: ps.ID, ProblemID: prob.ID, OrderIndex: 0}).Error; err != nil {
		t.Fatalf("create item: %v", err)
	}
	users := []models.User{
		{Username: "stu1", Name: "学生甲", PasswordHash: "x", Role: models.RoleStudent},
		{Username: "stu2", Name: "学生乙", PasswordHash: "x", Role: models.RoleStudent},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}
	members := []models.ProblemSetMember{
		{ProblemSetID: ps.ID, UserID: users[0].ID, JoinedAt: time.Now()},
		{ProblemSetID: ps.ID, UserID: users[1].ID, JoinedAt: time.Now()},
	}
	if err := db.Create(&members).Error; err != nil {
		t.Fatalf("create members: %v", err)
	}
	subs := []models.Submission{
		{UserID: users[0].ID, ProblemID: prob.ID, ProblemSetID: &ps.ID, Verdict: models.VerdictWA, CreatedAt: time.Now().Add(-2 * time.Minute)},
		{UserID: users[0].ID, ProblemID: prob.ID, ProblemSetID: &ps.ID, Verdict: models.VerdictAC, CreatedAt: time.Now().Add(-time.Minute)},
		{UserID: users[1].ID, ProblemID: prob.ID, ProblemSetID: &ps.ID, Verdict: models.VerdictAC, CreatedAt: time.Now()},
	}
	if err := db.Create(&subs).Error; err != nil {
		t.Fatalf("create submissions: %v", err)
	}
	bonuses := []models.ProblemSetDailyBonus{
		{ProblemSetID: ps.ID, UserID: users[0].ID, ScoreDate: "2026-05-06", Score: 9},
		{ProblemSetID: ps.ID, UserID: users[1].ID, ScoreDate: "2026-05-06", Score: 1},
	}
	if err := db.Create(&bonuses).Error; err != nil {
		t.Fatalf("create bonuses: %v", err)
	}

	rankH := &RankingHandler{DB: db}
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/problemsets/"+strconv.Itoa(int(ps.ID))+"/ranking", nil)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(ps.ID))}}
	rankH.Problemset(ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("problemset ranking: code=%d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		BonusEnabled bool `json:"bonus_enabled"`
		Items        []struct {
			UserID     uint `json:"user_id"`
			BonusScore int  `json:"bonus_score"`
			PenaltyMin int  `json:"penalty_min"`
			ACCount    int  `json:"ac_count"`
		} `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode ranking: %v", err)
	}
	if !resp.BonusEnabled {
		t.Fatalf("expected bonus_enabled=true")
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 ranking rows, got %d", len(resp.Items))
	}
	if resp.Items[0].UserID != users[1].ID {
		t.Fatalf("expected lower-penalty user to rank first despite lower bonus, got order %+v", resp.Items)
	}
	if resp.Items[0].BonusScore != 1 || resp.Items[1].BonusScore != 9 {
		t.Fatalf("unexpected bonus scores in ranking: %+v", resp.Items)
	}
}

func TestProblemSetHeadlineUsesRankingTieBreak(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openProblemSetBonusTestDB(t)
	ps := models.ProblemSet{Title: "Headline Set", Visible: true}
	if err := db.Create(&ps).Error; err != nil {
		t.Fatalf("create problemset: %v", err)
	}
	prob := models.Problem{Title: "Tie Break", Visible: true, TimeLimitMS: 1000, MemoryLimitMB: 128}
	if err := db.Create(&prob).Error; err != nil {
		t.Fatalf("create problem: %v", err)
	}
	if err := db.Create(&models.ProblemSetItem{ProblemSetID: ps.ID, ProblemID: prob.ID, OrderIndex: 0}).Error; err != nil {
		t.Fatalf("create item: %v", err)
	}
	users := []models.User{
		{Username: "stu1", Name: "学生甲", PasswordHash: "x", Role: models.RoleStudent},
		{Username: "stu2", Name: "学生乙", PasswordHash: "x", Role: models.RoleStudent},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}
	members := []models.ProblemSetMember{
		{ProblemSetID: ps.ID, UserID: users[0].ID, JoinedAt: time.Now()},
		{ProblemSetID: ps.ID, UserID: users[1].ID, JoinedAt: time.Now()},
	}
	if err := db.Create(&members).Error; err != nil {
		t.Fatalf("create members: %v", err)
	}
	subs := []models.Submission{
		{UserID: users[0].ID, ProblemID: prob.ID, ProblemSetID: &ps.ID, Verdict: models.VerdictWA, CreatedAt: time.Now().Add(-2 * time.Minute)},
		{UserID: users[0].ID, ProblemID: prob.ID, ProblemSetID: &ps.ID, Verdict: models.VerdictAC, CreatedAt: time.Now().Add(-time.Minute)},
		{UserID: users[1].ID, ProblemID: prob.ID, ProblemSetID: &ps.ID, Verdict: models.VerdictAC, CreatedAt: time.Now()},
	}
	if err := db.Create(&subs).Error; err != nil {
		t.Fatalf("create submissions: %v", err)
	}

	h := &ProblemSetHandler{DB: db}
	topName, topAC := h.problemSetLeader(ps.ID)
	if topName != users[1].Name || topAC != 1 {
		t.Fatalf("expected lower-penalty AC leader, got %q / %d", topName, topAC)
	}

	itemCount, headlineName, headlineAC := h.problemSetHeadline(ps.ID)
	if itemCount != 1 {
		t.Fatalf("expected item_count=1, got %d", itemCount)
	}
	if headlineName != users[1].Name || headlineAC != 1 {
		t.Fatalf("expected headline leader to follow ranking tie-break, got %q / %d", headlineName, headlineAC)
	}
}

func TestProblemSetBonusMatrixGroupsScoresByDateColumn(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openProblemSetBonusTestDB(t)
	ps := models.ProblemSet{Title: "Bonus Matrix", EnableBonus: true}
	if err := db.Create(&ps).Error; err != nil {
		t.Fatalf("create problemset: %v", err)
	}
	users := []models.User{
		{Username: "stu1", Name: "学生甲", PasswordHash: "x", Role: models.RoleStudent},
		{Username: "stu2", Name: "学生乙", PasswordHash: "x", Role: models.RoleStudent},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}
	members := []models.ProblemSetMember{
		{ProblemSetID: ps.ID, UserID: users[0].ID, JoinedAt: time.Now()},
		{ProblemSetID: ps.ID, UserID: users[1].ID, JoinedAt: time.Now()},
	}
	if err := db.Create(&members).Error; err != nil {
		t.Fatalf("create members: %v", err)
	}
	bonuses := []models.ProblemSetDailyBonus{
		{ProblemSetID: ps.ID, UserID: users[0].ID, ScoreDate: "2026-05-05", Score: 2},
		{ProblemSetID: ps.ID, UserID: users[0].ID, ScoreDate: "2026-05-06", Score: 3},
		{ProblemSetID: ps.ID, UserID: users[1].ID, ScoreDate: "2026-05-06", Score: 1},
	}
	if err := db.Create(&bonuses).Error; err != nil {
		t.Fatalf("create bonuses: %v", err)
	}

	adminH := &AdminHandler{DB: db}
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/admin/problemsets/"+strconv.Itoa(int(ps.ID))+"/bonus", nil)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(ps.ID))}}
	adminH.ListProblemSetDailyBonus(ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("list bonus matrix: code=%d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Dates []string `json:"dates"`
		Items []struct {
			UserID uint           `json:"user_id"`
			Scores map[string]int `json:"scores"`
		} `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode bonus matrix: %v", err)
	}
	if len(resp.Dates) != 2 || resp.Dates[0] != "2026-05-05" || resp.Dates[1] != "2026-05-06" {
		t.Fatalf("unexpected date columns: %+v", resp.Dates)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(resp.Items))
	}
	got := map[uint]map[string]int{}
	for _, row := range resp.Items {
		got[row.UserID] = row.Scores
	}
	if got[users[0].ID]["2026-05-05"] != 2 || got[users[0].ID]["2026-05-06"] != 3 {
		t.Fatalf("unexpected first row scores: %+v", got[users[0].ID])
	}
	if got[users[1].ID]["2026-05-06"] != 1 {
		t.Fatalf("unexpected second row scores: %+v", got[users[1].ID])
	}
}

func openProblemSetBonusTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.ProblemSet{},
		&models.ProblemSetItem{},
		&models.ProblemSetMember{},
		&models.ProblemSetBan{},
		&models.ProblemSetDailyBonus{},
		&models.Problem{},
		&models.Submission{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}
