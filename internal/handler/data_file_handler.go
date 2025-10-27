package handler

import (
	"Go_for_unity/internal/model"
	"Go_for_unity/internal/store"
	"archive/zip"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type DataFileHandler struct {
	dfStore *store.DataFileStore
	isStore *store.IslandStore // 需要 IslandStore 来获取岛屿信息以构建路径
}

func NewDataFileHandler(dfStore *store.DataFileStore, isStore *store.IslandStore) *DataFileHandler {
	return &DataFileHandler{dfStore: dfStore, isStore: isStore}
}

// 1. 上传文件接口
func (h *DataFileHandler) UploadDataFile(c *gin.Context) {
	// --- 1. 解析参数 ---
	isleIDStr := c.PostForm("isle_id")
	dataType := c.PostForm("data_type")
	heightStr := c.PostForm("height")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件上传失败: " + err.Error()})
		return
	}

	isleID, _ := strconv.ParseUint(isleIDStr, 10, 64)
	height, _ := strconv.ParseFloat(heightStr, 64)

	// --- 2. 获取岛屿信息，构建存储路径 ---
	island, err := h.isStore.GetByID(uint(isleID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "关联的岛屿不存在"})
		return
	}
	// 路径格式: uploads/用户名/岛屿名/文件类型/
	baseUploadDir := filepath.Join("uploads", island.BelongTo, island.IsleName, dataType)
	if err := os.MkdirAll(baseUploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文件目录失败: " + err.Error()})
		return
	}

	var finalFilePath string // 最终要存入数据库的文件路径

	// --- 3. 根据文件类型分别处理 ---
	switch dataType {
	case "shp", "tif":
		// 解压zip文件
		zipPath := filepath.Join(baseUploadDir, file.Filename)
		if err := c.SaveUploadedFile(file, zipPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存zip文件失败"})
			return
		}
		// 解压到同名文件夹
		unzipDest := strings.TrimSuffix(zipPath, filepath.Ext(zipPath))
		if err := unzip(zipPath, unzipDest); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "解压文件失败: " + err.Error()})
			return
		}
		os.Remove(zipPath) // 删除临时的zip包

		// 查找目标文件 (.shp 或 .tif)
		targetExt := "." + dataType
		foundPath, err := findFileByExt(unzipDest, targetExt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("在压缩包中未找到 %s 文件", targetExt)})
			return
		}
		finalFilePath = foundPath

	case "models", "txt", "jpg", "weather", "mapping":
		// 直接保存文件
		savePath := filepath.Join(baseUploadDir, file.Filename)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
			return
		}
		finalFilePath = savePath

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型: " + dataType})
		return
	}

	// --- 4. 创建数据库记录 ---
	// 首先获取纯数据名
	cleanDataName := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
	dataFile := model.DataFile{
		DataName: cleanDataName,
		DataType: dataType,
		DataPath: finalFilePath,
		IsleID:   uint(isleID),
		Height:   height,
	}
	if err := h.dfStore.Create(&dataFile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库记录创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件上传并处理成功", "data": dataFile})
}

// 2. 按岛屿ID分页获取文件列表
func (h *DataFileHandler) GetDataFilesByIsle(c *gin.Context) {
	isleIDStr := c.Param("isle_id")
	isleID, _ := strconv.ParseUint(isleIDStr, 10, 64)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	files, total, err := h.dfStore.GetByIsleID(uint(isleID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文件列表失败: " + err.Error()})
		return
	}

	// 将本地路径转换为可访问的URL
	for i := range files {
		files[i].DataPath = fmt.Sprintf("http://%s/%s", c.Request.Host, files[i].DataPath)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": files,
		"pagination": gin.H{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// 3. 删除文件
func (h *DataFileHandler) DeleteDataFile(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)

	// 先从数据库查找记录，获取文件路径
	file, err := h.dfStore.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件记录不存在"})
		return
	}

	// 从数据库删除记录
	if err := h.dfStore.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除数据库记录失败: " + err.Error()})
		return
	}

	// 从磁盘删除文件/文件夹
	// 如果是解压的文件，删除整个解压后的文件夹
	var pathToDel = file.DataPath
	if file.DataType == "shp" || file.DataType == "tif" {
		pathToDel = filepath.Dir(file.DataPath)
	}
	if err := os.RemoveAll(pathToDel); err != nil {
		// 即使文件删除失败，也返回成功，因为数据库记录已经删了
		c.JSON(http.StatusOK, gin.H{"message": "数据库记录删除成功，但清理磁盘文件时出错: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件删除成功"})
}

// 4. 修改文件高度
func (h *DataFileHandler) UpdateDataFileHeight(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)

	var req struct {
		Height float64 `json:"height"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	if err := h.dfStore.UpdateHeight(uint(id), req.Height); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新高度失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "高度更新成功"})
}

// --- Helper Functions ---

// unzip 解压 zip 文件
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(dest, 0755)

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// findFileByExt 在目录中查找指定后缀的文件
func findFileByExt(rootDir, ext string) (string, error) {
	var foundPath string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ext) {
			foundPath = path
			return filepath.SkipDir // 找到后停止遍历
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if foundPath == "" {
		return "", fmt.Errorf("no file with extension %s found", ext)
	}
	return foundPath, nil
}
