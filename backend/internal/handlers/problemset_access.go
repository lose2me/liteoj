package handlers

import (
	"errors"

	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/models"
)

// problemSetAccess snapshots the current viewer's relationship to one
// problem-set context. It centralizes the visibility / membership / ban /
// problem-in-set checks that are reused across problem detail, submissions,
// AI actions, and bonus-sheet reads.
type problemSetAccess struct {
	ProblemSet     models.ProblemSet
	Found          bool
	IsAdmin        bool
	IsMember       bool
	IsBanned       bool
	ProblemLinked  bool
}

func loadProblemSetAccess(
	db *gorm.DB,
	problemSetID uint,
	userID uint,
	isAdmin bool,
	problemID *uint,
) (problemSetAccess, error) {
	state := problemSetAccess{
		IsAdmin:       isAdmin,
		ProblemLinked: true,
	}
	if db == nil || problemSetID == 0 {
		return state, nil
	}
	if err := db.First(&state.ProblemSet, problemSetID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return state, nil
		}
		return state, err
	}
	state.Found = true

	if isAdmin {
		state.IsMember = true
	} else if userID > 0 {
		var memberCount int64
		if err := db.Model(&models.ProblemSetMember{}).
			Where("problem_set_id = ? AND user_id = ?", problemSetID, userID).
			Count(&memberCount).Error; err != nil {
			return state, err
		}
		state.IsMember = memberCount > 0

		var banCount int64
		if err := db.Model(&models.ProblemSetBan{}).
			Where("problem_set_id = ? AND user_id = ?", problemSetID, userID).
			Count(&banCount).Error; err != nil {
			return state, err
		}
		state.IsBanned = banCount > 0
	}

	if problemID != nil {
		var itemCount int64
		if err := db.Model(&models.ProblemSetItem{}).
			Where("problem_set_id = ? AND problem_id = ?", problemSetID, *problemID).
			Count(&itemCount).Error; err != nil {
			return state, err
		}
		state.ProblemLinked = itemCount > 0
	}

	return state, nil
}

func (a problemSetAccess) HasPassword() bool {
	return a.ProblemSet.Password != ""
}

func (a problemSetAccess) VisibleToStudent() bool {
	return a.Found && a.ProblemSet.Visible
}

func (a problemSetAccess) ContextApplies() bool {
	if !a.Found || !a.ProblemLinked {
		return false
	}
	if a.IsAdmin {
		return true
	}
	return a.ProblemSet.Visible && a.IsMember && !a.IsBanned
}

func (a problemSetAccess) StudentLockReason() string {
	if a.IsBanned {
		return "banned"
	}
	return "not_member"
}

func (a problemSetAccess) BonusAllowed() bool {
	if !a.Found || !a.ProblemSet.EnableBonus {
		return false
	}
	if a.IsAdmin {
		return true
	}
	return a.ContextApplies()
}
