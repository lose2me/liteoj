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

func (h *AdminHandler) CreateProblemSet(c *gin.Context) {
	var ps models.ProblemSet
	if err := c.ShouldBindJSON(&ps); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ps.AllowedLangsJSON = encodeAllowedLangs(ps.AllowedLangs)
	if err := h.DB.Create(&ps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.publishPS(ps.ID)
	c.JSON(http.StatusOK, ps)
}

func (h *AdminHandler) UpdateProblemSet(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var ps models.ProblemSet
	if err := c.ShouldBindJSON(&ps); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]any{
		"title":              ps.Title,
		"password":           ps.Password,
		"start_time":         ps.StartTime,
		"end_time":           ps.EndTime,
		"allowed_langs_json": encodeAllowedLangs(ps.AllowedLangs),
		"visible":            ps.Visible,
		"disable_idea":       ps.EnableIdea,
		"disable_solution":   ps.EnableSolution,
		"disable_ai":         ps.EnableAI,
		"enable_bonus":       ps.EnableBonus,
	}
	if err := h.DB.Model(&models.ProblemSet{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.publishPS(uint(id))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) DeleteProblemSet(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.ProblemSet{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.DB.Where("problem_set_id = ?", id).Delete(&models.ProblemSetItem{})
	h.DB.Where("problem_set_id = ?", id).Delete(&models.ProblemSetMember{})
	h.DB.Where("problem_set_id = ?", id).Delete(&models.ProblemSetBan{})
	h.DB.Where("problem_set_id = ?", id).Delete(&models.ProblemSetDailyBonus{})
	h.publishPS(uint(id))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) ListProblemSetMembers(c *gin.Context) {
	psid, _ := strconv.Atoi(c.Param("id"))
	var members []models.ProblemSetMember
	if err := h.DB.Where("problem_set_id = ?", psid).Order("joined_at DESC").Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(members) == 0 {
		c.JSON(http.StatusOK, gin.H{"items": []any{}})
		return
	}
	uids := make([]uint, 0, len(members))
	for _, m := range members {
		uids = append(uids, m.UserID)
	}
	var users []models.User
	h.DB.Where("id IN ?", uids).Find(&users)
	byID := map[uint]models.User{}
	for _, u := range users {
		byID[u.ID] = u
	}
	type row struct {
		UserID   uint      `json:"user_id"`
		Username string    `json:"username"`
		Name     string    `json:"name"`
		JoinedAt time.Time `json:"joined_at"`
	}
	out := make([]row, 0, len(members))
	for _, m := range members {
		u := byID[m.UserID]
		out = append(out, row{UserID: m.UserID, Username: u.Username, Name: u.Name, JoinedAt: m.JoinedAt})
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *AdminHandler) RemoveProblemSetMember(c *gin.Context) {
	psid, _ := strconv.Atoi(c.Param("id"))
	uid, _ := strconv.Atoi(c.Param("uid"))
	actor := middleware.CurrentUserID(c)
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("problem_set_id = ? AND user_id = ?", psid, uid).
			Delete(&models.ProblemSetMember{}).Error; err != nil {
			return err
		}
		var existing int64
		tx.Model(&models.ProblemSetBan{}).
			Where("problem_set_id = ? AND user_id = ?", psid, uid).Count(&existing)
		if existing == 0 {
			ban := models.ProblemSetBan{
				ProblemSetID: uint(psid),
				UserID:       uint(uid),
				BannedAt:     time.Now(),
				BannedBy:     actor,
			}
			if err := tx.Create(&ban).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.publishMembers(uint(psid), uint(uid))
	c.JSON(http.StatusOK, gin.H{"ok": true, "banned": true})
}

func (h *AdminHandler) ListProblemSetBans(c *gin.Context) {
	psid, _ := strconv.Atoi(c.Param("id"))
	var bans []models.ProblemSetBan
	if err := h.DB.Where("problem_set_id = ?", psid).Order("banned_at DESC").Find(&bans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(bans) == 0 {
		c.JSON(http.StatusOK, gin.H{"items": []any{}})
		return
	}
	uids := make([]uint, 0, len(bans))
	for _, b := range bans {
		uids = append(uids, b.UserID)
	}
	var users []models.User
	h.DB.Where("id IN ?", uids).Find(&users)
	byID := map[uint]models.User{}
	for _, u := range users {
		byID[u.ID] = u
	}
	type row struct {
		UserID   uint      `json:"user_id"`
		Username string    `json:"username"`
		Name     string    `json:"name"`
		BannedAt time.Time `json:"banned_at"`
	}
	out := make([]row, 0, len(bans))
	for _, b := range bans {
		u := byID[b.UserID]
		out = append(out, row{UserID: b.UserID, Username: u.Username, Name: u.Name, BannedAt: b.BannedAt})
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *AdminHandler) UnbanProblemSetMember(c *gin.Context) {
	psid, _ := strconv.Atoi(c.Param("id"))
	uid, _ := strconv.Atoi(c.Param("uid"))
	if err := h.DB.Where("problem_set_id = ? AND user_id = ?", psid, uid).Delete(&models.ProblemSetBan{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.publishMembers(uint(psid), uint(uid))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) ToggleProblemSetVisibility(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var ps models.ProblemSet
	if err := h.DB.First(&ps, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	next := !ps.Visible
	if err := h.DB.Model(&models.ProblemSet{}).Where("id = ?", id).Update("visible", next).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.publishPS(uint(id))
	c.JSON(http.StatusOK, gin.H{"ok": true, "visible": next})
}

func (h *AdminHandler) CopyProblemSet(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var src models.ProblemSet
	if err := h.DB.First(&src, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	dup := models.ProblemSet{
		Title:            i18n.ProblemSetCopyTitle(src.Title),
		AllowedLangsJSON: src.AllowedLangsJSON,
		Password:         src.Password,
		StartTime:        src.StartTime,
		EndTime:          src.EndTime,
		EnableIdea:       src.EnableIdea,
		EnableSolution:   src.EnableSolution,
		EnableAI:         src.EnableAI,
		EnableBonus:      src.EnableBonus,
		CreatedBy:        middleware.CurrentUserID(c),
	}
	if err := h.DB.Create(&dup).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var items []models.ProblemSetItem
	h.DB.Where("problem_set_id = ?", id).Order("order_index ASC, id ASC").Find(&items)
	if len(items) > 0 {
		cloned := make([]models.ProblemSetItem, len(items))
		for i, it := range items {
			cloned[i] = models.ProblemSetItem{
				ProblemSetID: dup.ID,
				ProblemID:    it.ProblemID,
				OrderIndex:   it.OrderIndex,
			}
		}
		if err := h.DB.Create(&cloned).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	h.publishPS(dup.ID)
	c.JSON(http.StatusOK, gin.H{"id": dup.ID, "title": dup.Title})
}

type setProblemsReq struct {
	ProblemIDs []uint `json:"problem_ids"`
}

func (h *AdminHandler) SetProblemSetItems(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req setProblemsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Where("problem_set_id = ?", id).Delete(&models.ProblemSetItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]models.ProblemSetItem, 0, len(req.ProblemIDs))
	for i, pid := range req.ProblemIDs {
		items = append(items, models.ProblemSetItem{
			ProblemSetID: uint(id), ProblemID: pid, OrderIndex: i,
		})
	}
	if len(items) > 0 {
		if err := h.DB.Create(&items).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	h.publishPS(uint(id))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
