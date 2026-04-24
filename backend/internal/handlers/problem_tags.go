package handlers

import (
	"github.com/liteoj/liteoj/backend/internal/models"
	"gorm.io/gorm"
)

func normalizeTagIDs(tagIDs []uint) []uint {
	if len(tagIDs) == 0 {
		return nil
	}
	seen := make(map[uint]bool, len(tagIDs))
	out := make([]uint, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		if tagID == 0 || seen[tagID] {
			continue
		}
		seen[tagID] = true
		out = append(out, tagID)
	}
	return out
}

func replaceProblemTags(tx *gorm.DB, problemID uint, tagIDs []uint) error {
	if err := tx.Where("problem_id = ?", problemID).Delete(&models.ProblemTag{}).Error; err != nil {
		return err
	}
	tagIDs = normalizeTagIDs(tagIDs)
	if len(tagIDs) == 0 {
		return nil
	}
	rows := make([]models.ProblemTag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		rows = append(rows, models.ProblemTag{ProblemID: problemID, TagID: tagID})
	}
	return tx.Create(&rows).Error
}

func loadProblemTags(db *gorm.DB, problemID uint) ([]models.Tag, error) {
	var tags []models.Tag
	err := db.Table("tags").
		Select("tags.*").
		Joins("JOIN problem_tags pt ON pt.tag_id = tags.id").
		Joins("LEFT JOIN tag_groups tg ON tg.id = tags.group_id").
		Where("pt.problem_id = ?", problemID).
		Order("COALESCE(tg.order_index, 0) ASC, tags.order_index ASC, tags.id ASC").
		Find(&tags).Error
	return tags, err
}

func collectTagIDs(tags []models.Tag) []uint {
	if len(tags) == 0 {
		return nil
	}
	out := make([]uint, 0, len(tags))
	for _, tag := range tags {
		out = append(out, tag.ID)
	}
	return out
}
