package handlers

import (
	"gorm.io/gorm"
)

type studentFeatureFlags struct {
	Idea     bool
	Solution bool
	AI       bool
}

func loadStudentFeatureFlags(
	db *gorm.DB,
	userID uint,
	isAdmin bool,
	problemSetID *uint,
	problemID *uint,
) studentFeatureFlags {
	if isAdmin {
		return studentFeatureFlags{Idea: true, Solution: true, AI: true}
	}
	if db == nil || problemSetID == nil || *problemSetID == 0 {
		return studentFeatureFlags{}
	}
	access, err := loadProblemSetAccess(db, *problemSetID, userID, false, problemID)
	if err != nil || !access.ContextApplies() {
		return studentFeatureFlags{}
	}
	return studentFeatureFlags{
		Idea:     access.ProblemSet.EnableIdea,
		Solution: access.ProblemSet.EnableSolution,
		AI:       access.ProblemSet.EnableAI,
	}
}
