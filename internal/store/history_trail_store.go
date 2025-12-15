package store

import (
	"Go_for_unity/internal/model"
	"gorm.io/gorm"
)

type HistoryTrailStore struct {
	db *gorm.DB
}

// TrailCountResult 用于接收聚合查询结果的临时结构
type TrailCountResult struct {
	IsleName string
	Category string
	Total    int64
}

func NewHistoryTrailStore(db *gorm.DB) *HistoryTrailStore {
	return &HistoryTrailStore{db: db}
}

// Create 创建一条新的历史轨迹记录
func (s *HistoryTrailStore) Create(trail *model.HistoryTrail) error {
	return s.db.Create(trail).Error
}

// GetByIsleName 分页查询指定岛屿的所有历史轨迹
func (s *HistoryTrailStore) GetByIsleName(isleName, category, trailName string, page, pageSize int) ([]model.HistoryTrail, int64, error) {
	var trails []model.HistoryTrail
	var total int64

	// 1. 构建基础查询
	query := s.db.Model(&model.HistoryTrail{}).Where("isle_name = ? AND category = ?", isleName, category)

	// 2. 如果 trailName 不为空，则添加模糊查询条件
	if trailName != "" {
		query = query.Where("trail_name LIKE ?", "%"+trailName+"%")
	}

	// 3. 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 4. 添加分页和排序，获取数据
	err = query.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at desc").
		Find(&trails).Error

	return trails, total, err
}

// GetByID 根据 ID 查询单条历史轨迹记录
func (s *HistoryTrailStore) GetByID(id uint) (*model.HistoryTrail, error) {
	var trail model.HistoryTrail
	err := s.db.First(&trail, id).Error
	if err != nil {
		return nil, err
	}
	return &trail, nil
}

// Delete 根据 ID 删除一条记录 (硬删除)
func (s *HistoryTrailStore) Delete(id uint) error {
	return s.db.Unscoped().Delete(&model.HistoryTrail{}, id).Error
}

// GetGlobalCounts 获取全局的轨迹统计信息
func (s *HistoryTrailStore) GetGlobalCounts() ([]TrailCountResult, error) {
	var results []TrailCountResult
	// SQL: SELECT isle_name, category, count(*) as total FROM history_trails WHERE deleted_at IS NULL GROUP BY isle_name, category
	err := s.db.Model(&model.HistoryTrail{}).
		Select("isle_name, category, count(*) as total").
		Group("isle_name, category").
		Scan(&results).Error
	return results, err
}
