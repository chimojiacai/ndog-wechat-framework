package utils

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// 全局日志服务实例
var globalLogService *LogService

// LogService 日志服务
type LogService struct {
	app *application.App
}

// LogData 日志数据结构
type LogData struct {
	TimeStamp string `json:"timeStamp"` // 时间戳
	Response  string `json:"response"`  // 响应内容
	Type      string `json:"type"`      // 日志类型
	Msg       string `json:"msg"`       // 日志消息
	Color     string `json:"color"`     // 显示颜色
}

// NewLogService 创建日志服务
func NewLogService(app *application.App) *LogService {
	return &LogService{
		app: app,
	}
}

// SetApp 设置 app 实例
func (s *LogService) SetApp(app *application.App) {
	s.app = app
	// 同时设置全局实例
	globalLogService = s
}

// GetGlobalLogService 获取全局日志服务实例
func GetGlobalLogService() *LogService {
	return globalLogService
}

// SendLog 发送日志到前端
// timeStamp: 时间戳，如 "2025-11-23 18:00:00"
// response: 响应内容
// logType: 日志类型（自定义字符串，如："群聊"、"私聊"、"注入成功"等）
// msg: 日志消息
// color: 显示颜色（如：#67C23A 绿色, #E6A23C 橙色, #F56C6C 红色, #409EFF 蓝色）
func (s *LogService) SendLog(ctx context.Context, timeStamp, response, logType, msg, color string) {
	if s.app == nil {
		g.Log().Warning(ctx, "LogService: app 实例为 nil，无法发送日志事件")
		return
	}

	if s.app.Event == nil {
		g.Log().Warning(ctx, "LogService: app.Event 为 nil，无法发送日志事件")
		return
	}

	logData := LogData{
		TimeStamp: timeStamp,
		Response:  response,
		Type:      logType,
		Msg:       msg,
		Color:     color,
	}

	// 发送事件到前端
	s.app.Event.Emit("system:log", logData)

	g.Log().Debugf(ctx, "发送日志事件: type=%s, msg=%s", logType, msg)
}
