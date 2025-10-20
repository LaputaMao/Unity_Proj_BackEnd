package handler

import (
	"Go_for_unity/internal/store"
	"Go_for_unity/internal/ws"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// --- 定义与最终 JSON 结构对应的 Go Struct ---

// LatLon 用于坐标点
type LatLon struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// LatLonHeight 用于相机位置
type LatLonHeight struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Height float64 `json:"height"`
}

// FileEntry 是各类文件列表中的通用条目
type FileEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// VectorEntry 是 shp 文件的条目，多了 Height 字段
type VectorEntry struct {
	Name   string  `json:"name"`
	Path   string  `json:"path"`
	Height float64 `json:"height"`
}

// RasterEntry 是 tif 文件的条目，多了 Height 字段
type RasterEntry struct {
	Name   string  `json:"name"`
	Path   string  `json:"path"`
	Height float64 `json:"height"`
}

// ExportedJSON 是最终生成的 JSON 的根结构
type ExportedJSON struct {
	ProjectName  string        `json:"projectName"`
	CesiumOrigin LatLon        `json:"cesiumOrigin"`
	PlayPosition LatLonHeight  `json:"playPosition"`
	Vectors      []VectorEntry `json:"vectors"`
	Rasters      []RasterEntry `json:"rasters"`
	Models       []FileEntry   `json:"models"`
	Pictures     []FileEntry   `json:"pictures"`
	Text         []FileEntry   `json:"text"`
}

// ExportHandler 负责处理导出逻辑
type ExportHandler struct {
	isStore   *store.IslandStore
	dfStore   *store.DataFileStore
	wsManager *ws.Manager
}

func NewExportHandler(isStore *store.IslandStore, dfStore *store.DataFileStore, wsManager *ws.Manager) *ExportHandler {
	return &ExportHandler{isStore: isStore, dfStore: dfStore, wsManager: wsManager}
}

// ExportIslandJSON 是导出接口的核心实现
func (h *ExportHandler) ExportIslandJSON(c *gin.Context) {
	// 1. 获取岛屿 ID
	isleIDStr := c.Param("isle_id")
	isleID, err := strconv.ParseUint(isleIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的岛屿ID"})
		return
	}

	// 2. 查询岛屿基础信息
	island, err := h.isStore.GetByID(uint(isleID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "岛屿不存在"})
		return
	}

	// 3. 查询该岛屿下的所有文件
	files, err := h.dfStore.GetAllByIsleID(uint(isleID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文件数据失败: " + err.Error()})
		return
	}

	// 4. 构建最终的 JSON 对象
	result := ExportedJSON{
		ProjectName: island.IsleName,
		CesiumOrigin: LatLon{
			Lat: island.CenterY, // 注意：lat 对应 Y
			Lon: island.CenterX, // lon 对应 X
		},
		PlayPosition: LatLonHeight{
			Lat:    island.CameraY,
			Lon:    island.CameraX,
			Height: island.CameraZ,
		},
		// 初始化空的 slice，这样即使没有数据，JSON里也会是 [] 而不是 null
		Vectors:  []VectorEntry{},
		Rasters:  []RasterEntry{},
		Models:   []FileEntry{},
		Pictures: []FileEntry{},
		Text:     []FileEntry{},
	}

	// 5. 遍历文件，分类填充到 result 中
	for _, file := range files {
		switch file.DataType {
		case "shp":
			result.Vectors = append(result.Vectors, VectorEntry{
				Name:   file.DataName,
				Path:   file.DataPath, // 直接使用数据库中的路径
				Height: file.Height,
			})
		case "tif":
			result.Rasters = append(result.Rasters, RasterEntry{
				Name: file.DataName,
				// 将路径转换为静态服务 URL
				Path:   fmt.Sprintf("http://%s/%s", c.Request.Host, file.DataPath),
				Height: file.Height,
			})
		case "models":
			result.Models = append(result.Models, FileEntry{
				Name: file.DataName,
				Path: file.DataPath, // 直接使用数据库中的路径
			})
		case "jpg":
			result.Pictures = append(result.Pictures, FileEntry{
				Name: file.DataName,
				Path: file.DataPath, // 直接使用数据库中的路径
			})
		case "txt":
			result.Text = append(result.Text, FileEntry{
				Name: file.DataName,
				Path: file.DataPath, // 直接使用数据库中的路径
			})
		}
	}

	// 6. 返回 JSON 响应
	//c.JSON(http.StatusOK, result)

	// 6. 通过 WebSocket 推送给 Unity
	h.wsManager.SendMessage(result)

	// 7. 返回 HTTP 响应给前端
	c.JSON(http.StatusOK, result)
}
