package store

import (
	"Go_for_unity/internal/model"
	"gorm.io/gorm"
)

type HistoryTrailStore struct {
	db *gorm.DB
}

func NewHistoryTrailStore(db *gorm.DB) *HistoryTrailStore {
	return &HistoryTrailStore{db: db}
}

// Create 创建一条新的历史轨迹记录
func (s *HistoryTrailStore) Create(trail *model.HistoryTrail) error {
	return s.db.Create(trail).Error
}

// GetByIsleName 分页查询指定岛屿的所有历史轨迹
func (s *HistoryTrailStore) GetByIsleName(isleName, category string, page, pageSize int) ([]model.HistoryTrail, int64, error) {
	var trails []model.HistoryTrail
	var total int64

	// 构建查询，同时根据 isle_name 和 category 筛选
	query := s.db.Model(&model.HistoryTrail{}).Where("isle_name = ? AND category = ?", isleName, category)

	// 计算总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
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
