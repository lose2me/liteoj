package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/models"
	"github.com/liteoj/liteoj/backend/internal/services/judge"
)

type pendingSubmissionResumer interface {
	Enqueue(sub *models.Submission, tcs []models.Testcase, cpuMS, memMB int) bool
}

type pendingResumeDeps struct {
	db    *gorm.DB
	queue pendingSubmissionResumer
	cpu   int
	mem   int
}

func newPendingResumeDeps(db *gorm.DB, q *judge.Queue, defaultCPU, defaultMem int) pendingResumeDeps {
	return pendingResumeDeps{
		db:    db,
		queue: q,
		cpu:   defaultCPU,
		mem:   defaultMem,
	}
}

func resumePendingSubmissions(deps pendingResumeDeps) (pendingCount int, resumedCount int, err error) {
	if deps.db == nil || deps.queue == nil {
		return 0, 0, errors.New("resume_pending_unavailable")
	}

	var pending []models.Submission
	if err = deps.db.
		Where("verdict = ?", models.VerdictPending).
		Order("id ASC").
		Find(&pending).Error; err != nil {
		return 0, 0, err
	}

	pendingCount = len(pending)
	for _, sub := range pending {
		var p models.Problem
		if err = deps.db.First(&p, sub.ProblemID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				err = nil
				continue
			}
			return pendingCount, resumedCount, err
		}

		var tcs []models.Testcase
		if err = deps.db.Where("problem_id = ?", sub.ProblemID).
			Order("order_index ASC, id ASC").
			Find(&tcs).Error; err != nil {
			return pendingCount, resumedCount, err
		}
		if len(tcs) == 0 {
			continue
		}

		cpu := p.TimeLimitMS
		if cpu == 0 {
			cpu = deps.cpu
		}
		mem := p.MemoryLimitMB
		if mem == 0 {
			mem = deps.mem
		}

		subCopy := sub
		if deps.queue.Enqueue(&subCopy, tcs, cpu, mem) {
			resumedCount++
		}
	}
	return pendingCount, resumedCount, nil
}

var _ pendingSubmissionResumer = (*judge.Queue)(nil)

func (h *AdminHandler) ResumePendingSubmissions(c *gin.Context) {
	if h == nil || h.DB == nil || h.Queue == nil || h.Live == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "resume_pending_unavailable"})
		return
	}
	cfg := h.Live.Current()

	pendingCount, resumedCount, err := resumePendingSubmissions(
		newPendingResumeDeps(h.DB, h.Queue, cfg.JudgeDefaultCPU, cfg.JudgeDefaultMem),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"pending_count": pendingCount,
		"resumed_count": resumedCount,
	})
}
