package handler

import (
	"Go_for_unity/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// WebsocketHandler 负责处理 WebSocket 连接请求
type WebsocketHandler struct {
	manager *ws.Manager
}

func NewWebsocketHandler(manager *ws.Manager) *WebsocketHandler {
	return &WebsocketHandler{manager: manager}
}

// upgrader 定义了 WebSocket 的一些参数，例如缓冲区大小
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 解决跨域问题：允许所有来源的连接（在生产环境中应配置得更严格）
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWS 处理 WebSocket 连接请求
func (h *WebsocketHandler) ServeWS(c *gin.Context) {
	// 将 HTTP 连接升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}

	// 注册连接
	h.manager.Register(conn)

	// 当函数退出时，自动注销并关闭连接
	defer h.manager.Unregister()

	// 启动一个循环来监听来自客户端的消息
	// 这是保持连接活动并检测断开的必要步骤
	for {
		// ReadMessage 会阻塞，直到收到消息或连接断开
		_, _, err := conn.ReadMessage()
		if err != nil {
			// 如果读取出错（例如客户端关闭了连接），就跳出循环
			log.Printf("读取 WebSocket 消息时出错: %v , 或者链接已经关闭.", err)
			break
		}
	}
}
