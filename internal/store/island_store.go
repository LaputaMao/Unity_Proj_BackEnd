package store

import (
	"Go_for_unity/internal/model"
	"gorm.io/gorm"
)

// IslandStore 定义了所有与岛屿相关的数据库操作
type IslandStore struct {
	db *gorm.DB
}

// NewIslandStore 创建一个新的 IslandStore
func NewIslandStore(db *gorm.DB) *IslandStore {
	return &IslandStore{db: db}
}

// Create 创建一个新的岛屿记录
func (s *IslandStore) Create(island *model.Island) error {
	return s.db.Create(island).Error
}

// GetByOwner 根据用户名查询所有岛屿
func (s *IslandStore) GetByOwner(owner string) ([]model.Island, error) {
	var islands []model.Island
	err := s.db.Where("belong_to = ?", owner).Find(&islands).Error
	return islands, err
}

// GetByID 根据 ID 查询单个岛屿
func (s *IslandStore) GetByID(id uint) (*model.Island, error) {
	var island model.Island
	err := s.db.First(&island, id).Error
	if err != nil {
		return nil, err
	}
	return &island, nil
}

// Update 更新一个岛屿的信息
func (s *IslandStore) Update(island *model.Island) error {
	// 使用 Save 会更新所有字段，即使是零值
	// 如果只想更新非零值字段，可以使用 Updates
	return s.db.Save(island).Error
}

// Delete 根据 ID 删除一个岛屿 (硬删除)
func (s *IslandStore) Delete(id uint) error {
	// 在 Delete 前调用 Unscoped()来实现硬删除
	return s.db.Unscoped().Delete(&model.Island{}, id).Error
}
