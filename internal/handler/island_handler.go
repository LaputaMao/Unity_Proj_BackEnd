package handler

import (
	"Go_for_unity/internal/model"
	"Go_for_unity/internal/store"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
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
	archipelagoName := c.PostForm("archipelago_name")
	country := c.PostForm("country")
	moveSpeed, _ := strconv.ParseFloat(c.DefaultPostForm("moveSpeed", "0.7"), 64)
	rotateSpeed, _ := strconv.ParseFloat(c.DefaultPostForm("rotateSpeed", "0.5"), 64)
	scaleSpeed, _ := strconv.ParseFloat(c.DefaultPostForm("scaleSpeed", "1.0"), 64)
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
		IsleName:        isleName,
		IsleDesc:        isleDesc,
		BelongTo:        belongTo,
		CenterX:         centerX,
		CenterY:         centerY,
		CameraX:         cameraX,
		CameraY:         cameraY,
		CameraZ:         cameraZ,
		IslePicPath:     picPath, // 存储相对路径
		ArchipelagoName: archipelagoName,
		Country:         country,
		MoveSpeed:       moveSpeed,
		RotateSpeed:     rotateSpeed,
		ScaleSpeed:      scaleSpeed,
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
	// --- 解析参数 ---
	// 必选参数
	owner := c.Query("belong_to")
	if owner == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "必须提供 belong_to 参数"})
		return
	}

	// 可选参数
	isleName := c.Query("isle_name")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	archipelagoName := c.Query("archipelago_name")
	country := c.Query("country")

	// --- 调用 Store 层 ---
	islands, total, err := h.store.GetByOwner(owner, isleName, archipelagoName, country, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询数据库失败: " + err.Error()})
		return
	}

	// --- 处理返回数据 ---
	// 将图片路径转换为可访问的 URL (这部分逻辑不变)
	for i := range islands {
		// 这里我们做一个健壮性检查，防止 IslePicPath 为空
		if islands[i].IslePicPath != "" {
			islands[i].IslePicPath = fmt.Sprintf("http://%s/%s", c.Request.Host, islands[i].IslePicPath)
		}
	}

	// --- 返回新的 JSON 结构 ---
	c.JSON(http.StatusOK, gin.H{
		"data": islands,
		"pagination": gin.H{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
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
// 4. 修改岛屿信息接口 (修改后，支持 multipart/form-data)
func (h *IslandHandler) UpdateIsland(c *gin.Context) {
	// 1. 获取 ID 并从数据库查找现有岛屿
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	island, err := h.store.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "岛屿不存在"})
		return
	}

	// 2. 从 form-data 中解析并更新文本/数字字段
	// 我们使用 GetPostForm 来判断字段是否存在，如果存在才更新
	// 字符串类型使用 c.PostForm()。如果字段不存在或值为空，它都会返回空字符串。
	if isleDesc := c.PostForm("isle_desc"); isleDesc != "" {
		island.IsleDesc = isleDesc
	}
	if centerXStr, ok := c.GetPostForm("center_x"); ok {
		if val, err := strconv.ParseFloat(centerXStr, 64); err == nil {
			island.CenterX = val
		}
	}
	if centerYStr, ok := c.GetPostForm("center_y"); ok {
		if val, err := strconv.ParseFloat(centerYStr, 64); err == nil {
			island.CenterY = val
		}
	}
	// ... 以此类推，更新所有可修改的字段
	if cameraXStr, ok := c.GetPostForm("camera_x"); ok {
		if val, err := strconv.ParseFloat(cameraXStr, 64); err == nil {
			island.CameraX = val
		}
	}
	if cameraYStr, ok := c.GetPostForm("camera_y"); ok {
		if val, err := strconv.ParseFloat(cameraYStr, 64); err == nil {
			island.CameraY = val
		}
	}
	if cameraZStr, ok := c.GetPostForm("camera_z"); ok {
		if val, err := strconv.ParseFloat(cameraZStr, 64); err == nil {
			island.CameraZ = val
		}
	}
	if archipelagoName := c.PostForm("archipelago_name"); archipelagoName != "" {
		island.ArchipelagoName = archipelagoName
	}
	if country := c.PostForm("country"); country != "" {
		island.Country = country
	}
	if rotateSpeedStr, ok := c.GetPostForm("rotateSpeed"); ok {
		if val, err := strconv.ParseFloat(rotateSpeedStr, 64); err == nil {
			island.RotateSpeed = val
		}
	}
	if scaleSpeedStr, ok := c.GetPostForm("scaleSpeed"); ok {
		if val, err := strconv.ParseFloat(scaleSpeedStr, 64); err == nil {
			island.ScaleSpeed = val
		}
	}

	// 3. 处理可选的图片文件上传
	newFile, err := c.FormFile("isle_pic") // 假设前端上传的文件字段名为 "isle_pic"
	if err == nil {
		// 如果 err 为 nil，说明用户上传了新图片
		log.Println("检测到新图片上传，开始处理...")

		// 3a. 删除旧图片（如果存在）
		if island.IslePicPath != "" {
			if removeErr := os.Remove(island.IslePicPath); removeErr != nil {
				// 即使删除失败，也只记录日志，不中断主流程
				log.Printf("删除旧图片失败: %s, 错误: %v", island.IslePicPath, removeErr)
			} else {
				log.Printf("成功删除旧图片: %s", island.IslePicPath)
			}
		}

		// 3b. 保存新图片
		// 使用从数据库读出的 belong_to 和 isle_name 来构建路径，确保路径正确
		uploadDir := filepath.Join("uploads", island.BelongTo, island.IsleName)
		os.MkdirAll(uploadDir, 0755) // 确保目录存在
		newPicPath := filepath.Join(uploadDir, newFile.Filename)

		if saveErr := c.SaveUploadedFile(newFile, newPicPath); saveErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存新图片失败: " + saveErr.Error()})
			return
		}
		log.Printf("成功保存新图片到: %s", newPicPath)

		// 3c. 更新数据库中的图片路径字段
		island.IslePicPath = newPicPath
	}

	// 4. 将所有更改保存到数据库
	if err := h.store.Update(island); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新数据库失败: " + err.Error()})
		return
	}

	// 5. 返回成功响应
	c.JSON(http.StatusOK, gin.H{"message": "岛屿信息更新成功", "data": island})
}
