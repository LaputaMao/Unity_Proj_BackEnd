package main

import (
	"Go_for_unity/internal/handler"
	"Go_for_unity/internal/model"
	"Go_for_unity/internal/router"
	"Go_for_unity/internal/store"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

func main() {
	// 1. 加载配置
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %s", err)
	}

	// 2. 初始化数据库连接
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetInt("mysql.port"),
		viper.GetString("mysql.dbname"), // 你的数据库名 unity_li
		viper.GetString("mysql.charset"),
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %s", err)
	}

	// 3. 自动迁移 (创建/更新表结构)
	err = db.AutoMigrate(&model.Island{}, &model.DataFile{})
	if err != nil {
		log.Fatalf("数据库迁移失败: %s", err)
	}
	log.Println("数据库迁移成功！")

	// 4. 依赖注入：创建 store 和 handler
	islandStore := store.NewIslandStore(db)
	islandHandler := handler.NewIslandHandler(islandStore)
	dataFileStore := store.NewDataFileStore(db)
	dataFileHandler := handler.NewDataFileHandler(dataFileStore, islandStore) // 注意这里需要传入两个 store
	exportHandler := handler.NewExportHandler(islandStore, dataFileStore)
	// 5. 初始化 Gin 引擎
	r := gin.Default()
	// 增加 Body 大小限制，防止上传大文件时出错
	r.MaxMultipartMemory = 2 << 30 // 2 GB

	// 6. 设置路由
	router.Setup(r, islandHandler, dataFileHandler, exportHandler)

	// 7. 启动服务器
	// All the Go project developed by LaputaMao will listen on port 9090 , just because 9090 like 'gogo' hhh.
	log.Println("服务器启动，监听端口 :9090")
	if err := r.Run(":9090"); err != nil {
		log.Fatalf("服务器启动失败: %s", err)
	}
}
