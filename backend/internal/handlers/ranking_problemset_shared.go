package handlers

import (
	"time"

	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/models"
)

type problemSetRankSubmissionRow struct {
	UserID        uint
	ProblemID     uint
	Verdict       string
	CreatedAt     time.Time
	AIExplanation string
	AIRejected    bool
}

type problemSetRankPairKey struct {
	userID    uint
	problemID uint
}

type problemSetRankPairAgg struct {
	FirstAC        bool
	WABeforeAC     int
	AIUsedBeforeAC int
	LatestNonAC    string
	LatestNonACAt  time.Time
}

func newProblemSetRankRow(userID uint) *rankRow {
	return &rankRow{
		UserID:  userID,
		Results: map[uint]string{},
	}
}

func loadProblemSetMemberIDs(db *gorm.DB, psid uint) ([]uint, error) {
	var memberIDs []uint
	if err := db.Model(&models.ProblemSetMember{}).
		Where("problem_set_id = ?", psid).
		Pluck("user_id", &memberIDs).Error; err != nil {
		return nil, err
	}
	return memberIDs, nil
}

func loadProblemSetRankSubmissionRows(
	db *gorm.DB,
	psid uint,
	since time.Time,
) ([]problemSetRankSubmissionRow, error) {
	q := db.Table("submissions AS s").
		Select("s.user_id, s.problem_id, s.verdict, s.created_at, s.ai_explanation, s.ai_rejected").
		Joins("JOIN problem_set_items psi ON psi.problem_set_id = s.problem_set_id AND psi.problem_id = s.problem_id").
		Where("s.problem_set_id = ?", psid).
		Where("s.verdict IN ?", []string{
			models.VerdictAC, models.VerdictWA, models.VerdictTLE,
			models.VerdictMLE, models.VerdictOLE, models.VerdictRE,
			models.VerdictCE, models.VerdictPE,
		}).
		Where("NOT EXISTS (SELECT 1 FROM problem_set_bans b WHERE b.problem_set_id = s.problem_set_id AND b.user_id = s.user_id)").
		Order("s.user_id ASC, s.problem_id ASC, s.created_at ASC")
	if !since.IsZero() {
		q = q.Where("s.created_at >= ?", since)
	}

	var raw []problemSetRankSubmissionRow
	if err := q.Scan(&raw).Error; err != nil {
		return nil, err
	}
	return raw, nil
}

func loadProblemSetRankRows(db *gorm.DB, psid uint, since time.Time) (map[uint]*rankRow, error) {
	memberIDs, err := loadProblemSetMemberIDs(db, psid)
	if err != nil {
		return nil, err
	}
	raw, err := loadProblemSetRankSubmissionRows(db, psid, since)
	if err != nil {
		return nil, err
	}

	perUser := map[uint]*rankRow{}
	for _, uid := range memberIDs {
		perUser[uid] = newProblemSetRankRow(uid)
	}

	pairs := map[problemSetRankPairKey]*problemSetRankPairAgg{}
	for _, row := range raw {
		key := problemSetRankPairKey{userID: row.UserID, problemID: row.ProblemID}
		agg, ok := pairs[key]
		if !ok {
			agg = &problemSetRankPairAgg{}
			pairs[key] = agg
		}
		if row.Verdict == models.VerdictAC {
			if !agg.FirstAC {
				agg.FirstAC = true
			}
			continue
		}
		if !agg.FirstAC {
			agg.WABeforeAC++
			if row.AIExplanation != "" && !row.AIRejected {
				agg.AIUsedBeforeAC++
			}
		}
		if row.CreatedAt.After(agg.LatestNonACAt) {
			agg.LatestNonAC = row.Verdict
			agg.LatestNonACAt = row.CreatedAt
		}
	}

	for key, agg := range pairs {
		row, ok := perUser[key.userID]
		if !ok {
			row = newProblemSetRankRow(key.userID)
			perUser[key.userID] = row
		}
		if agg.FirstAC {
			row.ACCount++
			row.PenaltyMin += (agg.WABeforeAC + agg.AIUsedBeforeAC) * 20
			row.Results[key.problemID] = models.VerdictAC
			continue
		}
		if agg.LatestNonAC != "" {
			row.Results[key.problemID] = agg.LatestNonAC
		}
	}

	return perUser, nil
}

func applyProblemSetBonusScores(db *gorm.DB, psid uint, perUser map[uint]*rankRow) error {
	type bonusAgg struct {
		UserID uint
		Score  int
	}
	var bonusRows []bonusAgg
	if err := db.Table("problem_set_daily_bonus").
		Select("user_id, COALESCE(SUM(score), 0) AS score").
		Where("problem_set_id = ?", psid).
		Where("NOT EXISTS (SELECT 1 FROM problem_set_bans b WHERE b.problem_set_id = problem_set_daily_bonus.problem_set_id AND b.user_id = problem_set_daily_bonus.user_id)").
		Group("user_id").
		Scan(&bonusRows).Error; err != nil {
		return err
	}
	for _, row := range bonusRows {
		userRow, ok := perUser[row.UserID]
		if !ok {
			userRow = newProblemSetRankRow(row.UserID)
			perUser[row.UserID] = userRow
		}
		userRow.BonusScore = row.Score
	}
	return nil
}

func outranksProblemSetRow(a, b rankRow) bool {
	if a.ACCount != b.ACCount {
		return a.ACCount > b.ACCount
	}
	if a.PenaltyMin != b.PenaltyMin {
		return a.PenaltyMin < b.PenaltyMin
	}
	return a.UserID < b.UserID
}

func bestProblemSetRankRow(perUser map[uint]*rankRow) (*rankRow, bool) {
	var best *rankRow
	for _, row := range perUser {
		if row.ACCount == 0 {
			continue
		}
		if best == nil || outranksProblemSetRow(*row, *best) {
			best = row
		}
	}
	return best, best != nil
}
