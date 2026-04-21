package handlers

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/liteoj/liteoj/backend/internal/events"
)

// EventsHandler exposes the server-sent events endpoint. The stream is
// unauthenticated: events are broadcast globally and carry only information
// already visible on public views (submissions list, ranking, problemset
// progress). Adding auth here would require supporting token-in-query-string,
// which is awkward with the native EventSource API.
type EventsHandler struct {
	Broker *events.Broker
}

func (h *EventsHandler) Stream(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	// Disable proxy buffering (nginx) so events flush immediately.
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	ch := h.Broker.Subscribe()
	defer h.Broker.Unsubscribe(ch)

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	ctx := c.Request.Context()
	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case ev, ok := <-ch:
			if !ok {
				return false
			}
			c.SSEvent(ev.Type, ev.Data)
			return true
		case <-ticker.C:
			// Comment line keeps the connection alive through any idle proxy.
			_, _ = io.WriteString(w, ":ping\n\n")
			return true
		}
	})
}
