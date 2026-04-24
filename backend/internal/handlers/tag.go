package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/cache"
	"github.com/liteoj/liteoj/backend/internal/models"
)

type TagHandler struct {
	DB    *gorm.DB
	Cache *cache.Cache
}

type tagDictResp struct {
	ID    uint         `json:"id"`
	Name  string       `json:"name"`
	Tags  []models.Tag `json:"tags"`
	Order int          `json:"order_index"`
}

// List returns the tag dictionary (groups with nested tags).
func (h *TagHandler) List(c *gin.Context) {
	const key = "tags:dict"
	if h.Cache != nil {
		if v, ok := h.Cache.Get(key); ok {
			c.JSON(http.StatusOK, v)
			return
		}
	}
	var groups []models.TagGroup
	h.DB.Order("order_index ASC, id ASC").Find(&groups)
	var tags []models.Tag
	h.DB.Order("order_index ASC, id ASC").Find(&tags)
	byGroup := map[uint][]models.Tag{}
	for _, t := range tags {
		byGroup[t.GroupID] = append(byGroup[t.GroupID], t)
	}
	out := make([]tagDictResp, 0, len(groups))
	for _, g := range groups {
		out = append(out, tagDictResp{ID: g.ID, Name: g.Name, Tags: byGroup[g.ID], Order: g.OrderIndex})
	}
	payload := gin.H{"groups": out}
	if h.Cache != nil {
		h.Cache.Set(key, payload, 30*time.Second)
	}
	c.JSON(http.StatusOK, payload)
}

// ---- Admin: TagGroup ----

type groupUpsertReq struct {
	Name       string `json:"name" binding:"required"`
	OrderIndex int    `json:"order_index"`
}

func (h *TagHandler) CreateGroup(c *gin.Context) {
	var r groupUpsertReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	g := &models.TagGroup{Name: r.Name, OrderIndex: r.OrderIndex}
	if err := h.DB.Create(g).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.invalidate()
	c.JSON(http.StatusOK, g)
}

func (h *TagHandler) UpdateGroup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var r groupUpsertReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Model(&models.TagGroup{}).Where("id = ?", id).Updates(map[string]any{
		"name": r.Name, "order_index": r.OrderIndex,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.invalidate()
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *TagHandler) DeleteGroup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.DB.Where(
		"tag_id IN (?)",
		h.DB.Model(&models.Tag{}).Select("id").Where("group_id = ?", id),
	).Delete(&models.ProblemTag{})
	if err := h.DB.Delete(&models.TagGroup{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.DB.Where("group_id = ?", id).Delete(&models.Tag{})
	h.invalidate()
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---- Admin: Tag ----

type tagUpsertReq struct {
	GroupID    uint   `json:"group_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	OrderIndex int    `json:"order_index"`
}

func (h *TagHandler) CreateTag(c *gin.Context) {
	var r tagUpsertReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t := &models.Tag{GroupID: r.GroupID, Name: r.Name, OrderIndex: r.OrderIndex}
	if err := h.DB.Create(t).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.invalidate()
	c.JSON(http.StatusOK, t)
}

func (h *TagHandler) UpdateTag(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var r tagUpsertReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Model(&models.Tag{}).Where("id = ?", id).Updates(map[string]any{
		"group_id": r.GroupID, "name": r.Name, "order_index": r.OrderIndex,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.invalidate()
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *TagHandler) DeleteTag(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.DB.Where("tag_id = ?", id).Delete(&models.ProblemTag{})
	if err := h.DB.Delete(&models.Tag{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.invalidate()
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *TagHandler) invalidate() {
	if h.Cache != nil {
		h.Cache.Invalidate("tags:")
		h.Cache.Invalidate("problems:") // problem list shows tag names
	}
}
