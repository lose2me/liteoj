package handlers

import (
	"encoding/json"

	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/control"
	"github.com/liteoj/liteoj/backend/internal/events"
	"github.com/liteoj/liteoj/backend/internal/models"
)

type ProblemSetHandler struct {
	DB     *gorm.DB
	Live   *control.LiveConfig
	Broker *events.Broker
}

func sanitizeProblemSetForResponse(ps models.ProblemSet) models.ProblemSet {
	ps.Password = ""
	ps.AllowedLangs = decodeAllowedLangs(ps.AllowedLangsJSON)
	return ps
}

// decodeAllowedLangs parses the persisted JSON into the transient slice. Empty
// string means "no restriction"; unreadable JSON is treated the same (safe
// degrade rather than 500 on a corrupted row).
func decodeAllowedLangs(s string) []string {
	if s == "" {
		return nil
	}
	var v []string
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil
	}
	return v
}

// encodeAllowedLangs stores the slice as JSON; nil or empty slice persists as
// "" so the "no restriction" case round-trips cleanly.
func encodeAllowedLangs(v []string) string {
	if len(v) == 0 {
		return ""
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// psListRow wraps ProblemSet with UI-facing status that the client cares about
// but we never want to persist: whether a password is set (derived, not the
// actual secret), how many problems the set contains, how many of them the
// current user has AC'd, and who the current leader is.
type psListRow struct {
	models.ProblemSet
	HasPassword bool   `json:"has_password"`
	IsMember    bool   `json:"is_member"`
	IsBanned    bool   `json:"is_banned"`
	ItemCount   int    `json:"item_count"`
	MyACCount   int    `json:"my_ac_count"`
	TopACCount  int    `json:"top_ac_count"`
	TopACName   string `json:"top_ac_name"`
}
