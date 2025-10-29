package model

import "gorm.io/gorm"

// TableName 为了让 GORM 创建的表名是 isles 而不是 islands，我们添加一个 TableName 方法
func (Island) TableName() string {
	return "isles"
}

type Island struct {
	//ID          uint    `gorm:"primarykey"`
	IsleName        string  `gorm:"type:varchar(255);not null;unique"` // 岛屿名，设为唯一
	IsleDesc        string  `gorm:"type:text"`                         // 岛屿描述
	BelongTo        string  `gorm:"type:varchar(255);not null;index"`  // 所属用户, 建立索引方便查询
	CenterX         float64 // 中心点坐标X
	CenterY         float64 // 中心点坐标Y
	CameraX         float64 // 默认相机位置X
	CameraY         float64 // 默认相机位置Y
	CameraZ         float64 // 默认相机位置Z
	IslePicPath     string  `gorm:"type:varchar(512)"`       // 岛屿图片存储路径
	ArchipelagoName string  `gorm:"type:varchar(255);index"` // 群岛名称，加上索引方便查询
	Country         string  `gorm:"type:varchar(255);index"` // 所属国家，也加上索引
	MoveSpeed       float64 `gorm:"default:0.7"`             // 相机移动速度为新字段设置了 default 值。这样，即使在创建时没有提供这些参数，数据库中也会有合理的默认值。
	RotateSpeed     float64 `gorm:"default:0.5"`             // 相机旋转速度
	ScaleSpeed      float64 `gorm:"default:1.0"`             // 相机缩放速度
	// gorm.Model 会自动添加 ID, CreatedAt, UpdatedAt, DeletedAt 字段
	// 这里我们手动定义，可以更灵活。如果想用 gorm.Model，可以去掉上面的 ID。
	gorm.Model
}
