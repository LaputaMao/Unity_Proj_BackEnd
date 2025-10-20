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
func (s *HistoryTrailStore) GetByIsleName(isleName string, page, pageSize int) ([]model.HistoryTrail, int64, error) {
	var trails []model.HistoryTrail
	var total int64

	// 先计算总数
	s.db.Model(&model.HistoryTrail{}).Where("isle_name = ?", isleName).Count(&total)

	// 再进行分页查询
	err := s.db.Where("isle_name = ?", isleName).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at desc"). // 按创建时间降序排序，最新的在前面
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
