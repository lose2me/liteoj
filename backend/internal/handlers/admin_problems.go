package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/models"
)

type problemUpsertReq struct {
	models.Problem
	TagIDs []uint `json:"tag_ids"`
}

func (h *AdminHandler) CreateProblem(c *gin.Context) {
	var r problemUpsertReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := r.Problem
	cfg := h.Live.Current()
	if p.TimeLimitMS == 0 {
		p.TimeLimitMS = cfg.JudgeDefaultCPU
	}
	if p.MemoryLimitMB == 0 {
		p.MemoryLimitMB = cfg.JudgeDefaultMem
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&p).Error; err != nil {
			return err
		}
		return replaceProblemTags(tx, p.ID, r.TagIDs)
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.Cache != nil {
		h.Cache.Invalidate("problems:")
	}
	h.publishProblem(p.ID)
	c.JSON(http.StatusOK, p)
}

func (h *AdminHandler) UpdateProblem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var r problemUpsertReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := r.Problem
	p.ID = uint(id)
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Problem{}).Where("id = ?", id).Updates(&p).Error; err != nil {
			return err
		}
		return replaceProblemTags(tx, uint(id), r.TagIDs)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if h.Cache != nil {
		h.Cache.Invalidate("problems:")
	}
	h.DB.Model(&models.Submission{}).Where("problem_id = ?", id).Updates(map[string]any{
		"ai_explanation":   "",
		"ai_rejected":      false,
		"ai_reject_reason": "",
	})
	h.publishProblem(uint(id))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) DeleteProblem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.Problem{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.DB.Where("problem_id = ?", id).Delete(&models.Testcase{})
	h.DB.Where("problem_id = ?", id).Delete(&models.ProblemTag{})
	h.DB.Where("problem_id = ?", id).Delete(&models.ProblemSetItem{})
	if h.Cache != nil {
		h.Cache.Invalidate("problems:")
	}
	h.publishProblem(uint(id))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) ListTestcases(c *gin.Context) {
	pid, _ := strconv.Atoi(c.Param("id"))
	var tcs []models.Testcase
	h.DB.Where("problem_id = ?", pid).Order("order_index ASC, id ASC").Find(&tcs)
	c.JSON(http.StatusOK, gin.H{"items": tcs})
}

func (h *AdminHandler) CreateTestcase(c *gin.Context) {
	pid, _ := strconv.Atoi(c.Param("id"))
	var tc models.Testcase
	if err := c.ShouldBindJSON(&tc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tc.ProblemID = uint(pid)
	if err := h.DB.Create(&tc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.publishProblem(uint(pid))
	c.JSON(http.StatusOK, tc)
}

func (h *AdminHandler) UpdateTestcase(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("tcid"))
	var tc models.Testcase
	if err := c.ShouldBindJSON(&tc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tc.ID = uint(id)
	if err := h.DB.Model(&models.Testcase{}).Where("id = ?", id).Updates(&tc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var reloaded models.Testcase
	if err := h.DB.Select("problem_id").First(&reloaded, id).Error; err == nil {
		h.publishProblem(reloaded.ProblemID)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) DeleteTestcase(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("tcid"))
	var existing models.Testcase
	h.DB.Select("problem_id").First(&existing, id)
	if err := h.DB.Delete(&models.Testcase{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing.ProblemID > 0 {
		h.publishProblem(existing.ProblemID)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func mustFindProblemSet(tx *gorm.DB, id int) (*models.ProblemSet, error) {
	var ps models.ProblemSet
	if err := tx.First(&ps, id).Error; err != nil {
		return nil, err
	}
	return &ps, nil
}

func respondNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
}
