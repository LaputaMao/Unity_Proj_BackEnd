package handler

import (
	"Go_for_unity/internal/store"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// LogEntry 定义返回给前端的单条日志结构
type LogEntry struct {
	IsleName    string           `json:"isle_name"`
	IsleDesc    string           `json:"isle_desc"`
	BelongTo    string           `json:"belong_to"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	DataCounts  map[string]int64 `json:"data_counts"`  // 例如: {"shp": 5, "weather": 1}
	TrailCounts map[string]int64 `json:"trail_counts"` // 例如: {"history_trail": 2, "annotation": 3}
}

type LogHandler struct {
	isStore *store.IslandStore
	dfStore *store.DataFileStore
	htStore *store.HistoryTrailStore
}

func NewLogHandler(isStore *store.IslandStore, dfStore *store.DataFileStore, htStore *store.HistoryTrailStore) *LogHandler {
	return &LogHandler{
		isStore: isStore,
		dfStore: dfStore,
		htStore: htStore,
	}
}

// GetSystemLog 获取系统全量日志
func (h *LogHandler) GetSystemLog(c *gin.Context) {
	// 1. 获取所有岛屿
	islands, err := h.isStore.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取岛屿列表失败: " + err.Error()})
		return
	}

	// 2. 获取所有 DataFile 的统计数据
	dfCounts, err := h.dfStore.GetGlobalCounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件统计失败: " + err.Error()})
		return
	}

	// 3. 获取所有 HistoryTrail 的统计数据
	htCounts, err := h.htStore.GetGlobalCounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取轨迹统计失败: " + err.Error()})
		return
	}

	// --- 数据预处理：将统计列表转换为 Map 以便快速查找 ---

	// Map 结构: IsleID -> DataType -> Count
	dfMap := make(map[uint]map[string]int64)
	for _, item := range dfCounts {
		if _, ok := dfMap[item.IsleID]; !ok {
			dfMap[item.IsleID] = make(map[string]int64)
		}
		dfMap[item.IsleID][item.DataType] = item.Total
	}

	// Map 结构: IsleName -> Category -> Count
	htMap := make(map[string]map[string]int64)
	for _, item := range htCounts {
		if _, ok := htMap[item.IsleName]; !ok {
			htMap[item.IsleName] = make(map[string]int64)
		}
		htMap[item.IsleName][item.Category] = item.Total
	}

	// --- 4. 组装最终结果 ---
	var logs []LogEntry

	for _, isle := range islands {
		// 获取该岛屿的文件统计，如果为 nil 则初始化为空 map
		dCounts := dfMap[isle.ID]
		if dCounts == nil {
			dCounts = make(map[string]int64)
		}

		// 获取该岛屿的轨迹统计，如果为 nil 则初始化为空 map
		tCounts := htMap[isle.IsleName]
		if tCounts == nil {
			tCounts = make(map[string]int64)
		}

		entry := LogEntry{
			IsleName:    isle.IsleName,
			IsleDesc:    isle.IsleDesc,
			BelongTo:    isle.BelongTo,
			CreatedAt:   isle.CreatedAt,
			UpdatedAt:   isle.UpdatedAt,
			DataCounts:  dCounts,
			TrailCounts: tCounts,
		}
		logs = append(logs, entry)
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}
