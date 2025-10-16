package model

import "gorm.io/gorm"

// DataFile 存储所有类型的文件信息
type DataFile struct {
	gorm.Model         // 自动包含 ID, CreatedAt, UpdatedAt, DeletedAt
	DataName   string  `gorm:"type:varchar(255);not null"`      // 文件原始名称
	DataType   string  `gorm:"type:varchar(50);not null;index"` // 文件类型 (shp, tif, models, txt, jpg)
	DataPath   string  `gorm:"type:varchar(512);not null"`      // 文件在服务器上的存储路径
	IsleID     uint    `gorm:"not null;index"`                  // 外键，关联到 isles 表的 ID
	Height     float64 // 高度值，仅用于 shp 和 tif

	Island Island `gorm:"foreignKey:IsleID" json:"-"` // 定义与 Island 的关联,用json:"-"告诉 JSON 序列化器忽略这个字段。
}

func (DataFile) TableName() string {
	return "data_files"
}
