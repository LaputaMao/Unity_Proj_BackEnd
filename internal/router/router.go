package router

import (
	"Go_for_unity/internal/handler"
	"github.com/gin-gonic/gin"
)

func Setup(engine *gin.Engine, islandHandler *handler.IslandHandler) {
	// 设置静态文件服务，用于访问上传的图片
	// 前端访问 http://localhost:8080/uploads/xxx.jpg 就会映射到 ./uploads/xxx.jpg 文件
	engine.Static("/uploads", "./uploads")

	// API V1 路由组
	apiV1 := engine.Group("/api/v1")
	{
		// 岛屿相关路由
		islandGroup := apiV1.Group("/islands")
		{
			// POST /api/v1/islands - 创建岛屿
			islandGroup.POST("", islandHandler.CreateIsland)
			// GET /api/v1/islands?belong_to=xxx - 获取用户的所有岛屿
			islandGroup.GET("", islandHandler.GetIslandsByOwner)
			// DELETE /api/v1/islands/:id - 删除岛屿
			islandGroup.DELETE("/:id", islandHandler.DeleteIsland)
			// PUT /api/v1/islands/:id - 更新岛屿信息
			islandGroup.PUT("/:id", islandHandler.UpdateIsland)
		}
	}
}
