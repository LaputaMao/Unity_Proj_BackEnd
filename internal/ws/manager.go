package ws

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

// Manager 负责管理唯一的 WebSocket 连接
type Manager struct {
	conn *websocket.Conn
	mu   sync.RWMutex // 使用读写锁保护连接，保证并发安全
}

// NewManager 创建一个新的 Manager 实例
func NewManager() *Manager {
	return &Manager{}
}

// Register 注册并存储新的连接
func (m *Manager) Register(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conn = conn
	log.Println("WebSocket 客户端已连接")
}

// Unregister 注销连接
func (m *Manager) Unregister() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.conn != nil {
		m.conn.Close()
		m.conn = nil
		log.Println("WebSocket 客户端已断开")
	}
}

// SendMessage 向已连接的客户端发送 JSON 消息
func (m *Manager) SendMessage(message interface{}) {
	m.mu.RLock() // 使用读锁，允许多个发送操作同时进行（如果需要的话）
	defer m.mu.RUnlock()

	if m.conn == nil {
		log.Println("WebSocket 未连接，无法发送消息")
		return
	}

	// 使用 WriteJSON 可以方便地发送结构体
	err := m.conn.WriteJSON(message)
	if err != nil {
		log.Printf("通过 WebSocket 发送消息失败: %v", err)
		// 发送失败通常意味着连接已断开，可以考虑在这里触发 Unregister
	} else {
		log.Println("已通过 WebSocket成功推送消息至 Unity")
	}
}
