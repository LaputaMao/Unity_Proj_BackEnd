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
// 新增参数: isleName (用于模糊搜索), page, pageSize
// 返回值: 岛屿列表, 总记录数, 错误
func (s *IslandStore) GetByOwner(owner, isleName, archipelagoName, country string, page, pageSize int) ([]model.Island, int64, error) {
	var islands []model.Island
	var total int64

	// 1. 构建基础查询，这是 GORM 中处理动态查询的最佳方式
	query := s.db.Model(&model.Island{}).Where("belong_to = ?", owner)

	// 2. 动态添加可选的查询条件
	//如果这两个参数不为空，就往 query 中追加 Where 条件。
	//GORM 会智能地将这些 Where 条件用 AND 连接起来，完美地实现了我们的组合查询需求。
	if isleName != "" {
		query = query.Where("isle_name LIKE ?", "%"+isleName+"%")
	}
	if archipelagoName != "" {
		// 精确匹配群岛名称
		query = query.Where("archipelago_name = ?", archipelagoName)
	}
	if country != "" {
		// 精确匹配国家
		query = query.Where("country = ?", country)
	}

	// 3. 先执行 Count() 获取满足条件的总记录数，用于分页
	// 注意：Count() 需要在 Limit() 和 Offset() 之前执行
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 4. 添加分页和排序条件，然后执行查询获取数据
	err = query.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at desc"). // 按创建时间降序排序
		Find(&islands).Error

	return islands, total, err
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

// GetAll 获取所有岛屿列表 (不分页，用于日志统计)
func (s *IslandStore) GetAll() ([]model.Island, error) {
	var islands []model.Island
	// 只查询需要的字段，减少传输数据量
	err := s.db.Select("id, isle_name, isle_desc, belong_to, created_at, updated_at").Find(&islands).Error
	return islands, err
}
