package handler

import (
	"Go_for_unity/internal/model"
	"Go_for_unity/internal/store"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type HistoryTrailHandler struct {
	store *store.HistoryTrailStore
}

func NewHistoryTrailHandler(store *store.HistoryTrailStore) *HistoryTrailHandler {
	return &HistoryTrailHandler{store: store}
}

// 1. 上传并持久化历史轨迹文件
func (h *HistoryTrailHandler) CreateTrail(c *gin.Context) {
	// 从表单获取岛屿名
	isleName := c.PostForm("isle_name")
	category := c.PostForm("category")
	if isleName == "" || category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "isle_name 和 category 不能为空"})
		return
	}

	// 从表单获取文件
	file, err := c.FormFile("file") // 假设 Unity 上传时字段名为 "file"
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件上传失败: " + err.Error()})
		return
	}

	// 存储路径可以根据 category 再做一层隔离，更清晰
	// 例如：uploads/trails/岛屿名/history_trail/
	// 或   uploads/trails/岛屿名/annotation/
	uploadDir := filepath.Join("uploads", "trails", isleName, category)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建存储目录失败: " + err.Error()})
		return
	}

	// 保存文件
	savePath := filepath.Join(uploadDir, file.Filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败: " + err.Error()})
		return
	}

	// 创建数据库记录
	trail := model.HistoryTrail{
		IsleName:  isleName,
		TrailName: file.Filename,
		TrailPath: savePath,
		Category:  category,
	}

	if err := h.store.Create(&trail); err != nil {
		// 如果数据库创建失败，最好把刚刚保存的文件也删掉，保持数据一致性
		os.Remove(savePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库记录创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件上传成功", "data": trail})
}

// 2. 根据岛屿名分页查询轨迹列表
func (h *HistoryTrailHandler) GetTrailsByIsleName(c *gin.Context) {
	isleName := c.Query("isle_name")
	category := c.Query("category")
	if isleName == "" || category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "必须提供 isle_name 和 category 参数"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	trails, total, err := h.store.GetByIsleName(isleName, category, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询轨迹列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": trails,
		"pagination": gin.H{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// 3. 根据 ID 返回轨迹文件
func (h *HistoryTrailHandler) GetTrailFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	// 从数据库查找记录
	trail, err := h.store.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "轨迹记录不存在"})
		return
	}

	// 直接将文件作为响应返回
	// Gin 会自动设置 Content-Type 为 application/json
	c.File(trail.TrailPath)
}
