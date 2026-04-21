// Persistent audit log of every AI call. Replaces the earlier in-memory
// registry: rows now live in the ai_tasks table so "AI 队列" reflects the
// full history of AI invocations across process restarts.
package ai

import (
	"time"

	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/events"
	"github.com/liteoj/liteoj/backend/internal/models"
)

type Queue struct {
	db     *gorm.DB
	broker *events.Broker
}

func NewQueue(db *gorm.DB, broker *events.Broker) *Queue {
	return &Queue{db: db, broker: broker}
}

// Start records a new AI call as "running" and returns the row id. The
// caller must eventually pair this with End to finalize the row.
func (q *Queue) Start(kind string, userID uint, username, subject string) uint {
	row := &models.AITask{
		Kind:      kind,
		UserID:    userID,
		Username:  username,
		Subject:   subject,
		Status:    models.AITaskStatusRunning,
		StartedAt: time.Now(),
	}
	if err := q.db.Create(row).Error; err != nil {
		return 0
	}
	q.publishChanged()
	return row.ID
}

// End marks the task as done or failed. errMsg empty ⇒ done, non-empty ⇒ failed.
// Safe to call with id=0 (no-op) when Start failed.
func (q *Queue) End(id uint, errMsg string) {
	if id == 0 {
		return
	}
	var row models.AITask
	if err := q.db.First(&row, id).Error; err != nil {
		return
	}
	now := time.Now()
	status := models.AITaskStatusDone
	if errMsg != "" {
		status = models.AITaskStatusFailed
	}
	q.db.Model(&row).Updates(map[string]any{
		"status":      status,
		"finished_at": now,
		"duration_ms": int(now.Sub(row.StartedAt).Milliseconds()),
		"error":       errMsg,
	})
	q.publishChanged()
	// ai:task:done is the targeted signal a specific client listens for to
	// pick up its own enqueued work. Keep the broadcast cheap — senders who
	// need the full payload call GET /admin/ai/tasks/:id for the Result.
	if q.broker != nil {
		q.broker.Publish(events.Event{
			Type: "ai:task:done",
			Data: map[string]any{
				"id":     row.ID,
				"kind":   row.Kind,
				"status": status,
			},
		})
	}
}

// SetPrompt stores the full prompt text sent to the model. Called even when
// the upstream call fails, so the audit log captures what was attempted.
func (q *Queue) SetPrompt(id uint, prompt string) {
	if id == 0 || prompt == "" {
		return
	}
	q.db.Model(&models.AITask{}).Where("id = ?", id).Update("prompt", prompt)
}

// SetOutput stores the raw model response body. Called on both success and
// failure paths — on failure the body often contains the diagnostic clue
// (truncated JSON, upstream error envelope, partial stream) that the
// "unexpected end of JSON input" parser error otherwise hides.
func (q *Queue) SetOutput(id uint, output string) {
	if id == 0 || output == "" {
		return
	}
	q.db.Model(&models.AITask{}).Where("id = ?", id).Update("output", output)
}

// SetResult stores the structured per-kind payload clients need after an
// async task completes (e.g. {"title":...}, the parsed TagSuggestion). Called
// only on success; on failure Result stays empty and clients show the error.
func (q *Queue) SetResult(id uint, result string) {
	if id == 0 || result == "" {
		return
	}
	q.db.Model(&models.AITask{}).Where("id = ?", id).Update("result", result)
}

func (q *Queue) publishChanged() {
	if q.broker == nil {
		return
	}
	q.broker.Publish(events.Event{Type: "ai:tasks:changed"})
}
