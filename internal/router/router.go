package router

import (
	"Go_for_unity/internal/handler"
	"github.com/gin-gonic/gin"
)

func Setup(engine *gin.Engine,
	islandHandler *handler.IslandHandler,
	dataFileHandler *handler.DataFileHandler,
	exportHandler *handler.ExportHandler,
	wsHandler *handler.WebsocketHandler,
	historyTrailHandler *handler.HistoryTrailHandler,
	logHandler *handler.LogHandler) {
	// 设置静态文件服务，用于访问上传的图片
	// 前端访问 http://localhost:8080/uploads/xxx.jpg 就会映射到 ./uploads/xxx.jpg 文件
	engine.Static("/uploads", "./uploads")

	// 新增 WebSocket 路由
	engine.GET("/ws", wsHandler.ServeWS)

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

			// 导出结构化 json 接口
			// GET /api/v1/islands/:isle_id/export
			islandGroup.GET("/:isle_id/export", exportHandler.ExportIslandJSON)
		}

		// 数据相关的路由
		dataFileGroup := apiV1.Group("/data-files")
		{
			// POST /api/v1/data-files - 上传文件
			dataFileGroup.POST("", dataFileHandler.UploadDataFile)
			// GET /api/v1/data-files/isle/:isle_id - 获取某个岛屿的文件列表（分页）
			dataFileGroup.GET("/isle/:isle_id", dataFileHandler.GetDataFilesByIsle)
			// DELETE /api/v1/data-files/:id - 删除文件
			dataFileGroup.DELETE("/:id", dataFileHandler.DeleteDataFile)
			// PUT /api/v1/data-files/:id/height - 修改文件高度
			dataFileGroup.PUT("/:id/height", dataFileHandler.UpdateDataFileHeight)
		}

		// 新增历史轨迹相关路由
		trailGroup := apiV1.Group("/trails")
		{
			// POST /api/v1/trails - 上传轨迹文件
			trailGroup.POST("", historyTrailHandler.CreateTrail)
			// GET /api/v1/trails?isle_name=xxx - 获取轨迹列表
			trailGroup.GET("", historyTrailHandler.GetTrailsByIsleName)
			// GET /api/v1/trails/:id/file - 下载指定ID的轨迹文件
			trailGroup.GET("/:id/file", historyTrailHandler.GetTrailFile)
			// DELETE /api/v1/trails/:id - 删除轨迹/标注
			trailGroup.DELETE("/:id", historyTrailHandler.DeleteTrail)
		}

		// 新增日志接口
		// GET /api/v1/logs
		apiV1.GET("/logs", logHandler.GetSystemLog)
	}
}
