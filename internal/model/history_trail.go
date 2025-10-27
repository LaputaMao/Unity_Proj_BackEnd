package model

import "gorm.io/gorm"

// HistoryTrail 存储历史轨迹 JSON 文件的信息
type HistoryTrail struct {
	gorm.Model
	IsleName  string `gorm:"type:varchar(255);not null;index"` // 关联的岛屿名称，建立索引以加快查询
	TrailName string `gorm:"type:varchar(255);not null"`       // JSON 文件名
	TrailPath string `gorm:"type:varchar(512);not null"`       // 文件在服务器上的存储路径
	Category  string `gorm:"type:varchar(100);not null;index"` // 数据类别 (例如: 'history_trail', 'annotation')
}

func (HistoryTrail) TableName() string {
	return "history_trails"
}
