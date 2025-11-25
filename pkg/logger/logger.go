package logger

import (
	"context"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Service 日志服务
type Service struct {
	app *application.App
	mu  sync.RWMutex
}

var (
	globalLogService *Service
	once             sync.Once
)

// NewService 创建日志服务实例
func NewService(app *application.App) *Service {
	return &Service{
		app: app,
	}
}

// SetApp 设置应用实例
func (s *Service) SetApp(app *application.App) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.app = app
}

// SendLog 发送日志到前端
// timeStamp: 时间戳
// response: 响应来源
// logType: 日志类型
// msg: 日志消息
// color: 日志颜色（可选，默认 #409EFF）
func (s *Service) SendLog(ctx context.Context, timeStamp, response, logType, msg, color string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.app == nil {
		return
	}

	// 默认颜色
	if color == "" {
		color = "#409EFF"
	}

	// 发送日志事件到前端
	s.app.Event.Emit("log:message", map[string]interface{}{
		"timeStamp": timeStamp,
		"response":  response,
		"logType":   logType,
		"msg":       msg,
		"color":     color,
	})
}

// SetGlobalLogService 设置全局日志服务
func SetGlobalLogService(service *Service) {
	once.Do(func() {
		globalLogService = service
	})
}

// GetGlobalLogService 获取全局日志服务
func GetGlobalLogService() *Service {
	return globalLogService
}
