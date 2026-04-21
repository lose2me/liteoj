package judge

import (
	"context"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/events"
	"github.com/liteoj/liteoj/backend/internal/models"
)

// Queue serializes judge work onto a fixed worker pool. With workers=1 it
// gives SQLite a single writer, avoiding "database is locked" under load.
type Queue struct {
	runner     *Runner
	db         *gorm.DB
	broker     *events.Broker
	ch         chan job
	wg         sync.WaitGroup
	workers    int
	jobTimeout time.Duration
}

type job struct {
	submissionID uint
	userID       uint
	problemID    uint
	problemSetID *uint
	lang         string
	code         string
	testcases    []models.Testcase
	cpuMS        int
	memMB        int
}

// NewQueue builds a judge queue with the given worker count, channel capacity,
// and per-job timeout. jobTimeout caps how long a single submission can stay
// in-flight from worker pick-up to final verdict; it mostly matters when
// go-judge is unreachable — without a cap, stuck TCP connects would block the
// worker indefinitely and back up the queue (queue_workers=1 + dead sandbox =
// everything stuck in PENDING). jobTimeout <= 0 falls back to 120s.
func NewQueue(db *gorm.DB, runner *Runner, broker *events.Broker, workers, cap int, jobTimeout time.Duration) *Queue {
	if workers < 1 {
		workers = 1
	}
	if cap < 1 {
		cap = 64
	}
	if jobTimeout <= 0 {
		jobTimeout = 120 * time.Second
	}
	q := &Queue{
		runner: runner, db: db, broker: broker,
		ch: make(chan job, cap), workers: workers, jobTimeout: jobTimeout,
	}
	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go q.loop()
	}
	return q
}

// Stats returns a snapshot of queue depth, capacity, and worker count for the
// admin dashboard. Safe to call from any goroutine.
func (q *Queue) Stats() (queueLen, queueCap, workers int) {
	return len(q.ch), cap(q.ch), q.workers
}

func (q *Queue) Enqueue(sub *models.Submission, tcs []models.Testcase, cpuMS, memMB int) {
	q.ch <- job{
		submissionID: sub.ID,
		userID:       sub.UserID,
		problemID:    sub.ProblemID,
		problemSetID: sub.ProblemSetID,
		lang:         sub.Language,
		code:         sub.Code,
		testcases:    tcs,
		cpuMS:        cpuMS,
		memMB:        memMB,
	}
}

func (q *Queue) loop() {
	defer q.wg.Done()
	for j := range q.ch {
		q.run(j)
	}
}

func (q *Queue) run(j job) {
	ctx, cancel := context.WithTimeout(context.Background(), q.jobTimeout)
	defer cancel()
	result, err := q.runner.Judge(ctx, RunnerInput{
		Lang: j.lang, Code: j.code, Testcases: j.testcases,
		CPULimitMS: j.cpuMS, MemLimitMB: j.memMB,
	})
	if err != nil {
		log.Printf("judge queue: submission %d error: %v", j.submissionID, err)
		q.db.Model(&models.Submission{}).Where("id = ?", j.submissionID).Updates(map[string]any{
			"verdict": models.VerdictSE,
			"message": err.Error(),
		})
		q.publishDone(j, models.VerdictSE, 0, 0)
		return
	}
	q.db.Model(&models.Submission{}).Where("id = ?", j.submissionID).Updates(map[string]any{
		"verdict":              result.Verdict,
		"message":              result.Message,
		"time_used_ms":         result.TimeMS,
		"memory_used_kb":       result.MemoryKB,
		"testcase_result_json": result.CaseResultRaw,
	})
	q.publishDone(j, result.Verdict, result.TimeMS, result.MemoryKB)
}

// publishDone broadcasts the final verdict so connected SSE clients can
// refetch (submissions list, personal page, ranking, problemset progress).
// Broker may be nil in tests; guard accordingly.
func (q *Queue) publishDone(j job, verdict string, timeMS, memKB int) {
	if q.broker == nil {
		return
	}
	q.broker.Publish(events.Event{
		Type: "submission:done",
		Data: map[string]any{
			"id":             j.submissionID,
			"user_id":        j.userID,
			"problem_id":     j.problemID,
			"problemset_id":  j.problemSetID,
			"verdict":        verdict,
			"time_used_ms":   timeMS,
			"memory_used_kb": memKB,
		},
	})
}
