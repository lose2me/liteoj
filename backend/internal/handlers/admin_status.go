package handlers

import (
	"context"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/liteoj/liteoj/backend/internal/services/judge"
)

// AdminStatusHandler surfaces read-only runtime & sandbox info for the admin
// dashboard. Never mutates state; safe to poll on a short interval.
type AdminStatusHandler struct {
	StartedAt   time.Time
	JudgeBase   string
	JudgeClient *judge.Client
	Queue       *judge.Queue

	// judge /version probe 有时会花 4~8 秒（WSL netsh portproxy 挂了就直接超时），
	// 这个值缓存 20 秒避免把 dashboard 的 Promise.all 拖垮。
	// 过期后下一次请求同步重取一次——业务上没有并发雪崩，admin 只有一个人在看，
	// 不必引入 singleflight。
	judgeMu   sync.Mutex
	judgeAt   time.Time
	judgeVer  string
	judgeErr  error
}

const judgeProbeTTL = 20 * time.Second

func (h *AdminStatusHandler) Status(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	ver, jerr := h.cachedJudgeVersion(c)

	qLen, qCap, workers := h.Queue.Stats()

	errStr := ""
	if jerr != nil {
		errStr = jerr.Error()
	}

	c.JSON(http.StatusOK, gin.H{
		"system": gin.H{
			"go_version":     runtime.Version(),
			"os":             runtime.GOOS,
			"arch":           runtime.GOARCH,
			"num_cpu":        runtime.NumCPU(),
			"goroutines":     runtime.NumGoroutine(),
			"alloc_mb":       m.Alloc / 1024 / 1024,
			"sys_mb":         m.Sys / 1024 / 1024,
			"uptime_seconds": int(time.Since(h.StartedAt).Seconds()),
			"started_at":     h.StartedAt.Format(time.RFC3339),
		},
		"judge": gin.H{
			"base_url":  h.JudgeBase,
			"reachable": jerr == nil,
			"version":   ver,
			"error":     errStr,
			"queue_len": qLen,
			"queue_cap": qCap,
			"workers":   workers,
		},
	})
}

// cachedJudgeVersion 返回 go-judge /version 的缓存结果。超出 TTL 时同步重取一
// 次；单次请求预算 2 秒并不再重试——失败后直接把错误缓存同样 TTL，让 UI 快速
// 拿到"offline"而不是卡 4 秒等下一个超时。
func (h *AdminStatusHandler) cachedJudgeVersion(c *gin.Context) (string, error) {
	h.judgeMu.Lock()
	defer h.judgeMu.Unlock()
	if time.Since(h.judgeAt) < judgeProbeTTL && !h.judgeAt.IsZero() {
		return h.judgeVer, h.judgeErr
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	ver, err := h.JudgeClient.Version(ctx)
	h.judgeAt = time.Now()
	h.judgeVer = ver
	h.judgeErr = err
	return ver, err
}
