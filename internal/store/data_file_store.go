package store

import (
	"Go_for_unity/internal/model"
	"gorm.io/gorm"
)

type DataFileStore struct {
	db *gorm.DB
}

func NewDataFileStore(db *gorm.DB) *DataFileStore {
	return &DataFileStore{db: db}
}

// Create 创建一条文件记录
func (s *DataFileStore) Create(file *model.DataFile) error {
	return s.db.Create(file).Error
}

// GetByIsleID 分页查询某个岛屿下的所有文件
// 返回: 文件列表, 总记录数, 错误
func (s *DataFileStore) GetByIsleID(isleID uint, page, pageSize int) ([]model.DataFile, int64, error) {
	var files []model.DataFile
	var total int64

	// 计算总数
	s.db.Model(&model.DataFile{}).Where("isle_id = ?", isleID).Count(&total)

	// 分页查询
	err := s.db.Where("isle_id = ?", isleID).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&files).Error

	return files, total, err
}

// GetByID 根据 ID 查询单个文件
func (s *DataFileStore) GetByID(id uint) (*model.DataFile, error) {
	var file model.DataFile
	err := s.db.First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// UpdateHeight 更新文件的高度值
func (s *DataFileStore) UpdateHeight(id uint, height float64) error {
	return s.db.Model(&model.DataFile{}).Where("id = ?", id).Update("height", height).Error
}

// Delete 根据 ID 删除文件记录,硬删除
func (s *DataFileStore) Delete(id uint) error {
	return s.db.Unscoped().Delete(&model.DataFile{}, id).Error
}

// GetAllByIsleID 查询某个岛屿下的所有文件（不分页）
func (s *DataFileStore) GetAllByIsleID(isleID uint) ([]model.DataFile, error) {
	var files []model.DataFile
	err := s.db.Where("isle_id = ?", isleID).Find(&files).Error
	return files, err
}
