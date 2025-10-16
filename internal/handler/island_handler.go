package handler

import (
	"Go_for_unity/internal/model"
	"Go_for_unity/internal/store"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type IslandHandler struct {
	store *store.IslandStore
}

func NewIslandHandler(s *store.IslandStore) *IslandHandler {
	return &IslandHandler{store: s}
}

// CreateIsland 1. 创建岛屿接口
func (h *IslandHandler) CreateIsland(c *gin.Context) {
	// 解析表单数据
	isleName := c.PostForm("isle_name")
	isleDesc := c.PostForm("isle_desc")
	belongTo := c.PostForm("belong_to")

	// 转换坐标值为 float64
	centerX, _ := strconv.ParseFloat(c.PostForm("center_x"), 64)
	centerY, _ := strconv.ParseFloat(c.PostForm("center_y"), 64)
	cameraX, _ := strconv.ParseFloat(c.PostForm("camera_x"), 64)
	cameraY, _ := strconv.ParseFloat(c.PostForm("camera_y"), 64)
	cameraZ, _ := strconv.ParseFloat(c.PostForm("camera_z"), 64)

	// 处理文件上传
	file, err := c.FormFile("isle_pic")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片上传失败: " + err.Error()})
		return
	}

	// 创建存储路径: ./uploads/用户名/岛屿名/
	uploadDir := filepath.Join("uploads", belongTo, isleName)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败: " + err.Error()})
		return
	}

	// 保存文件
	picPath := filepath.Join(uploadDir, file.Filename)
	if err := c.SaveUploadedFile(file, picPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败: " + err.Error()})
		return
	}

	// 创建 Island 对象
	island := model.Island{
		IsleName:    isleName,
		IsleDesc:    isleDesc,
		BelongTo:    belongTo,
		CenterX:     centerX,
		CenterY:     centerY,
		CameraX:     cameraX,
		CameraY:     cameraY,
		CameraZ:     cameraZ,
		IslePicPath: picPath, // 存储相对路径
	}

	// 保存到数据库
	if err := h.store.Create(&island); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "岛屿创建成功", "data": island})
}

// GetIslandsByOwner 2. 获取用户的所有岛屿信息接口
func (h *IslandHandler) GetIslandsByOwner(c *gin.Context) {
	owner := c.Query("belong_to")
	if owner == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "必须提供 belong_to 参数"})
		return
	}

	islands, err := h.store.GetByOwner(owner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询数据库失败: " + err.Error()})
		return
	}

	// 这里我们不返回文件流，而是返回一个可供前端访问的 URL
	// 这样更符合 RESTful API 的设计
	for i := range islands {
		islands[i].IslePicPath = fmt.Sprintf("http://%s/%s", c.Request.Host, islands[i].IslePicPath)
	}

	c.JSON(http.StatusOK, gin.H{"data": islands})
}

// DeleteIsland 3. 删除岛屿接口
func (h *IslandHandler) DeleteIsland(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	// 删除前先获取岛屿信息，以便删除文件
	island, err := h.store.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "岛屿不存在"})
		return
	}

	// 从数据库删除
	if err := h.store.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除数据库记录失败: " + err.Error()})
		return
	}

	// 从磁盘删除整个岛屿目录
	// uploads/用户名/岛屿名
	islandDir := filepath.Dir(island.IslePicPath)
	if err := os.RemoveAll(islandDir); err != nil {
		// 即使文件删除失败，数据库记录也已经删了，这里只记录日志或返回一个警告
		c.JSON(http.StatusOK, gin.H{"message": "数据库记录删除成功，但清理文件时出错: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "岛屿删除成功"})
}

// UpdateIsland 4. 修改岛屿信息接口
func (h *IslandHandler) UpdateIsland(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	// 查找现有岛屿
	island, err := h.store.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "岛屿不存在"})
		return
	}

	// 绑定 JSON 数据到临时结构体
	var updateData struct {
		IsleDesc *string  `json:"isle_desc"`
		CenterX  *float64 `json:"center_x"`
		CenterY  *float64 `json:"center_y"`
		CameraX  *float64 `json:"camera_x"`
		CameraY  *float64 `json:"camera_y"`
		CameraZ  *float64 `json:"camera_z"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
		return
	}

	// 更新字段 (只更新传入的字段)
	if updateData.IsleDesc != nil {
		island.IsleDesc = *updateData.IsleDesc
	}
	if updateData.CenterX != nil {
		island.CenterX = *updateData.CenterX
	}
	// ... 以此类推，更新其他字段
	if updateData.CenterY != nil {
		island.CenterY = *updateData.CenterY
	}
	if updateData.CameraX != nil {
		island.CameraX = *updateData.CameraX
	}
	if updateData.CameraY != nil {
		island.CameraY = *updateData.CameraY
	}
	if updateData.CameraZ != nil {
		island.CameraZ = *updateData.CameraZ
	}

	// 保存回数据库
	if err := h.store.Update(island); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新数据库失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "岛屿信息更新成功", "data": island})
}
