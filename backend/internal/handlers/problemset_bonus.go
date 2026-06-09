package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/models"
)

type problemSetBonusSaveReq struct {
	Date  string                    `json:"date"`
	Items []problemSetBonusSaveItem `json:"items"`
}

type problemSetBonusSaveItem struct {
	UserID uint `json:"user_id"`
	Score  int  `json:"score"`
}

type problemSetBonusRow struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Score    int    `json:"score"`
}

type problemSetBonusMatrixRow struct {
	UserID   uint           `json:"user_id"`
	Username string         `json:"username"`
	Name     string         `json:"name"`
	Scores   map[string]int `json:"scores"`
}

func normalizeProblemSetBonusDate(raw string) (string, error) {
	if raw == "" {
		raw = time.Now().Format("2006-01-02")
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02"), nil
}

func loadProblemSetBonusSwitch(db *gorm.DB, psid uint) (models.ProblemSet, error) {
	var ps models.ProblemSet
	err := db.Select("id", "visible", "enable_bonus").First(&ps, psid).Error
	return ps, err
}

func loadProblemSetBonusMembers(db *gorm.DB, psid int) ([]models.ProblemSetMember, []uint, map[uint]models.User, error) {
	var members []models.ProblemSetMember
	if err := db.Where("problem_set_id = ?", psid).Order("joined_at DESC, id DESC").Find(&members).Error; err != nil {
		return nil, nil, nil, err
	}
	if len(members) == 0 {
		return []models.ProblemSetMember{}, []uint{}, map[uint]models.User{}, nil
	}

	uids := make([]uint, 0, len(members))
	for _, m := range members {
		uids = append(uids, m.UserID)
	}
	var users []models.User
	if err := db.Where("id IN ?", uids).Find(&users).Error; err != nil {
		return nil, nil, nil, err
	}
	userByID := make(map[uint]models.User, len(users))
	for _, u := range users {
		userByID[u.ID] = u
	}
	return members, uids, userByID, nil
}

func loadProblemSetBonusRows(db *gorm.DB, psid int, scoreDate string) ([]problemSetBonusRow, error) {
	members, uids, userByID, err := loadProblemSetBonusMembers(db, psid)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return []problemSetBonusRow{}, nil
	}

	var bonusRows []models.ProblemSetDailyBonus
	if err := db.Where("problem_set_id = ? AND score_date = ? AND user_id IN ?", psid, scoreDate, uids).
		Find(&bonusRows).Error; err != nil {
		return nil, err
	}
	scoreByUser := make(map[uint]int, len(bonusRows))
	for _, row := range bonusRows {
		scoreByUser[row.UserID] = row.Score
	}

	out := make([]problemSetBonusRow, 0, len(members))
	for _, m := range members {
		u := userByID[m.UserID]
		out = append(out, problemSetBonusRow{
			UserID:   m.UserID,
			Username: u.Username,
			Name:     u.Name,
			Score:    scoreByUser[m.UserID],
		})
	}
	return out, nil
}

func loadProblemSetBonusMatrix(db *gorm.DB, psid int) ([]string, []problemSetBonusMatrixRow, error) {
	members, uids, userByID, err := loadProblemSetBonusMembers(db, psid)
	if err != nil {
		return nil, nil, err
	}
	if len(members) == 0 {
		return []string{}, []problemSetBonusMatrixRow{}, nil
	}

	var bonusRows []models.ProblemSetDailyBonus
	if err := db.Where("problem_set_id = ? AND user_id IN ?", psid, uids).
		Order("score_date ASC, user_id ASC").
		Find(&bonusRows).Error; err != nil {
		return nil, nil, err
	}

	dateSeen := make(map[string]struct{}, len(bonusRows))
	dates := make([]string, 0, len(bonusRows))
	scoreByUser := make(map[uint]map[string]int, len(members))
	for _, row := range bonusRows {
		if _, ok := dateSeen[row.ScoreDate]; !ok {
			dateSeen[row.ScoreDate] = struct{}{}
			dates = append(dates, row.ScoreDate)
		}
		if _, ok := scoreByUser[row.UserID]; !ok {
			scoreByUser[row.UserID] = map[string]int{}
		}
		scoreByUser[row.UserID][row.ScoreDate] = row.Score
	}

	out := make([]problemSetBonusMatrixRow, 0, len(members))
	for _, m := range members {
		u := userByID[m.UserID]
		scores := scoreByUser[m.UserID]
		if scores == nil {
			scores = map[string]int{}
		}
		out = append(out, problemSetBonusMatrixRow{
			UserID:   m.UserID,
			Username: u.Username,
			Name:     u.Name,
			Scores:   scores,
		})
	}
	return dates, out, nil
}

// ListDailyBonus returns either one day's score sheet (?date=YYYY-MM-DD) or
// the whole bonus matrix (default) to current members/admin on the detail page.
func (h *ProblemSetHandler) ListDailyBonus(c *gin.Context) {
	psid, _ := strconv.Atoi(c.Param("id"))
	uid := middleware.CurrentUserID(c)
	isAdmin := middleware.CurrentRole(c) == models.RoleAdmin
	access, err := loadProblemSetAccess(h.DB, uint(psid), uid, isAdmin, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !access.Found || (!isAdmin && !access.VisibleToStudent()) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	if !access.BonusAllowed() {
		if !access.ProblemSet.EnableBonus {
			c.JSON(http.StatusForbidden, gin.H{"error": errBonusDisabled})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": access.StudentLockReason()})
		return
	}

	if c.Query("date") == "" {
		dates, rows, err := loadProblemSetBonusMatrix(h.DB, psid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"dates": dates, "items": rows})
		return
	}

	scoreDate, err := normalizeProblemSetBonusDate(c.Query("date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_date"})
		return
	}
	rows, err := loadProblemSetBonusRows(h.DB, psid, scoreDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"date": scoreDate, "items": rows})
}

// ListProblemSetDailyBonus returns either one day's editable score sheet
// (?date=YYYY-MM-DD) or the whole editable bonus matrix (default).
func (h *AdminHandler) ListProblemSetDailyBonus(c *gin.Context) {
	psid, _ := strconv.Atoi(c.Param("id"))
	ps, err := loadProblemSetBonusSwitch(h.DB, uint(psid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	if !ps.EnableBonus {
		c.JSON(http.StatusForbidden, gin.H{"error": errBonusDisabled})
		return
	}

	if c.Query("date") == "" {
		dates, rows, err := loadProblemSetBonusMatrix(h.DB, psid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"dates": dates, "items": rows})
		return
	}

	scoreDate, err := normalizeProblemSetBonusDate(c.Query("date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_date"})
		return
	}
	rows, err := loadProblemSetBonusRows(h.DB, psid, scoreDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"date": scoreDate, "items": rows})
}

// SaveProblemSetDailyBonus replaces one day's score sheet for a problem set.
// The payload is treated as the full list for that date: omitted users become 0.
func (h *AdminHandler) SaveProblemSetDailyBonus(c *gin.Context) {
	psid, _ := strconv.Atoi(c.Param("id"))
	ps, err := loadProblemSetBonusSwitch(h.DB, uint(psid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	if !ps.EnableBonus {
		c.JSON(http.StatusForbidden, gin.H{"error": errBonusDisabled})
		return
	}

	var req problemSetBonusSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	scoreDate, err := normalizeProblemSetBonusDate(req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_date"})
		return
	}

	var members []models.ProblemSetMember
	if err := h.DB.Where("problem_set_id = ?", psid).Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	memberSet := make(map[uint]struct{}, len(members))
	for _, m := range members {
		memberSet[m.UserID] = struct{}{}
	}

	seen := map[uint]struct{}{}
	rows := make([]models.ProblemSetDailyBonus, 0, len(req.Items))
	actor := middleware.CurrentUserID(c)
	for _, item := range req.Items {
		if item.Score < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "score_must_be_non_negative"})
			return
		}
		if _, ok := memberSet[item.UserID]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_not_in_problemset"})
			return
		}
		if _, dup := seen[item.UserID]; dup {
			c.JSON(http.StatusBadRequest, gin.H{"error": "duplicate_user"})
			return
		}
		seen[item.UserID] = struct{}{}
		if item.Score == 0 {
			continue
		}
		rows = append(rows, models.ProblemSetDailyBonus{
			ProblemSetID: uint(psid),
			UserID:       item.UserID,
			ScoreDate:    scoreDate,
			Score:        item.Score,
			CreatedBy:    actor,
			UpdatedBy:    actor,
		})
	}

	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("problem_set_id = ? AND score_date = ?", psid, scoreDate).
			Delete(&models.ProblemSetDailyBonus{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.publishPS(uint(psid))
	c.JSON(http.StatusOK, gin.H{"ok": true, "date": scoreDate})
}
