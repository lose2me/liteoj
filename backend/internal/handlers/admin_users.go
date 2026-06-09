package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/auth"
	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/models"
)

type userUpsertReq struct {
	Username string      `json:"username"`
	Name     string      `json:"name"`
	Password string      `json:"password"`
	Role     models.Role `json:"role"`
}

type userListRow struct {
	models.User
	DistinctAC  int     `json:"distinct_ac"`
	DistinctTry int     `json:"distinct_tried"`
	TotalSubs   int     `json:"total_submissions"`
	ACRate      float64 `json:"ac_rate"`
	AK          int     `json:"ak"`
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	var users []models.User
	q := h.DB.Model(&models.User{})
	if r := c.Query("role"); r != "" {
		q = q.Where("role = ?", r)
	}
	if kw := c.Query("q"); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("name LIKE ? OR username LIKE ?", like, like)
	}
	q.Order("id ASC").Find(&users)

	type agg struct {
		UserID uint
		Total  int
		ACSub  int
		Tried  int
		ACProb int
	}
	var aggs []agg
	h.DB.Raw(`
		SELECT user_id,
		       COUNT(*)                                                   AS total,
		       SUM(CASE WHEN verdict = 'AC' THEN 1 ELSE 0 END)             AS ac_sub,
		       COUNT(DISTINCT problem_id)                                  AS tried,
		       COUNT(DISTINCT CASE WHEN verdict = 'AC' THEN problem_id END) AS ac_prob
		  FROM submissions
		 GROUP BY user_id`).Scan(&aggs)
	aggByUser := make(map[uint]agg, len(aggs))
	for _, a := range aggs {
		aggByUser[a.UserID] = a
	}

	akPerUser := computeAKPerUser(h.DB, time.Time{})

	out := make([]userListRow, len(users))
	for i, u := range users {
		a := aggByUser[u.ID]
		rate := 0.0
		if a.Total > 0 {
			rate = float64(a.ACSub) / float64(a.Total)
		}
		out[i] = userListRow{
			User:        u,
			DistinctAC:  a.ACProb,
			DistinctTry: a.Tried,
			TotalSubs:   a.Total,
			ACRate:      rate,
			AK:          akPerUser[u.ID],
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *AdminHandler) UserProfile(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var u models.User
	if err := h.DB.First(&u, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrUserNotFound})
		return
	}
	uid := u.ID
	type row struct {
		Verdict string
		Count   int
	}
	var rows []row
	h.DB.Raw(`SELECT verdict, COUNT(*) AS count FROM submissions WHERE user_id = ? GROUP BY verdict`, uid).Scan(&rows)
	dist := map[string]int{}
	var total int
	for _, r := range rows {
		dist[r.Verdict] = r.Count
		total += r.Count
	}
	var distinctAC int64
	h.DB.Table("submissions").Where("user_id = ? AND verdict = ?", uid, models.VerdictAC).
		Distinct("problem_id").Count(&distinctAC)
	var distinctTried int64
	h.DB.Table("submissions").Where("user_id = ?", uid).
		Distinct("problem_id").Count(&distinctTried)
	acRate := 0.0
	if total > 0 {
		acRate = float64(dist[models.VerdictAC]) / float64(total)
	}
	ak := computeAKPerUser(h.DB, time.Time{})[uid]
	c.JSON(http.StatusOK, gin.H{
		"user":              u,
		"total_submissions": total,
		"distinct_ac":       distinctAC,
		"distinct_tried":    distinctTried,
		"ac_rate":           acRate,
		"distribution":      dist,
		"ak":                ak,
	})
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req userUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrBadRequest})
		return
	}
	if req.Role == "" {
		req.Role = models.RoleStudent
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	u := &models.User{Username: req.Username, Name: req.Name, PasswordHash: hash, Role: req.Role}
	if err := h.DB.Create(u).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req userUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrBadRequest})
		return
	}
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Password != "" {
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		updates["password_hash"] = hash
		updates["login_version"] = gorm.Expr(loginVersionBumpExpr)
	}
	if err := h.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
