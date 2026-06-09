package handlers

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/models"
)

type resumeEnqueueCall struct {
	submissionID uint
	cpuMS        int
	memMB        int
	tcCount      int
}

type fakePendingResumer struct {
	enqueued map[uint]struct{}
	calls    []resumeEnqueueCall
}

func (f *fakePendingResumer) Enqueue(sub *models.Submission, tcs []models.Testcase, cpuMS, memMB int) bool {
	if f.enqueued == nil {
		f.enqueued = map[uint]struct{}{}
	}
	if _, ok := f.enqueued[sub.ID]; ok {
		return false
	}
	f.enqueued[sub.ID] = struct{}{}
	f.calls = append(f.calls, resumeEnqueueCall{
		submissionID: sub.ID,
		cpuMS:        cpuMS,
		memMB:        memMB,
		tcCount:      len(tcs),
	})
	return true
}

func TestResumePendingSubmissions(t *testing.T) {
	db := openAdminSubmissionTestDB(t)
	prob1 := models.Problem{Title: "A", Visible: true, TimeLimitMS: 1500, MemoryLimitMB: 192}
	prob2 := models.Problem{Title: "B", Visible: true}
	if err := db.Create(&[]models.Problem{prob1, prob2}).Error; err != nil {
		t.Fatalf("create problems: %v", err)
	}
	var probs []models.Problem
	if err := db.Order("id ASC").Find(&probs).Error; err != nil {
		t.Fatalf("reload problems: %v", err)
	}
	prob1 = probs[0]
	prob2 = probs[1]

	tcs := []models.Testcase{
		{ProblemID: prob1.ID, Input: "1\n", ExpectedOutput: "1\n", OrderIndex: 0},
		{ProblemID: prob2.ID, Input: "2\n", ExpectedOutput: "2\n", OrderIndex: 0},
	}
	if err := db.Create(&tcs).Error; err != nil {
		t.Fatalf("create testcases: %v", err)
	}

	psid := uint(7)
	subs := []models.Submission{
		{UserID: 1, ProblemID: prob1.ID, Language: "cpp", Code: "a", Verdict: models.VerdictPending, CreatedAt: time.Now().Add(-3 * time.Minute)},
		{UserID: 2, ProblemID: prob2.ID, ProblemSetID: &psid, Language: "cpp", Code: "b", Verdict: models.VerdictPending, CreatedAt: time.Now().Add(-2 * time.Minute)},
		{UserID: 3, ProblemID: prob1.ID, Language: "cpp", Code: "c", Verdict: models.VerdictAC, CreatedAt: time.Now().Add(-time.Minute)},
	}
	if err := db.Create(&subs).Error; err != nil {
		t.Fatalf("create submissions: %v", err)
	}

	q := &fakePendingResumer{}
	deps := pendingResumeDeps{
		db:    db,
		queue: q,
		cpu:   1000,
		mem:   256,
	}

	pending, resumed, err := resumePendingSubmissions(deps)
	if err != nil {
		t.Fatalf("resume pending: %v", err)
	}
	if pending != 2 || resumed != 2 {
		t.Fatalf("expected pending=2 resumed=2, got %d %d", pending, resumed)
	}
	if len(q.calls) != 2 {
		t.Fatalf("expected 2 enqueue calls, got %d", len(q.calls))
	}
	if q.calls[0].submissionID != subs[0].ID || q.calls[0].cpuMS != 1500 || q.calls[0].memMB != 192 || q.calls[0].tcCount != 1 {
		t.Fatalf("unexpected first enqueue call: %+v", q.calls[0])
	}
	if q.calls[1].submissionID != subs[1].ID || q.calls[1].cpuMS != 1000 || q.calls[1].memMB != 256 || q.calls[1].tcCount != 1 {
		t.Fatalf("unexpected second enqueue call: %+v", q.calls[1])
	}

	pending2, resumed2, err := resumePendingSubmissions(deps)
	if err != nil {
		t.Fatalf("resume pending second call: %v", err)
	}
	if pending2 != 2 || resumed2 != 0 {
		t.Fatalf("expected second call pending=2 resumed=0, got %d %d", pending2, resumed2)
	}
}

func openAdminSubmissionTestDB(t *testing.T) *gorm.DB {
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
