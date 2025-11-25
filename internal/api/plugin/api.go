package plugin

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// PluginManager 插件管理器接口
type PluginManager interface {
	GetConfigYaml() (string, error)
	GetCurrentWechat() (string, error)
	SendPluginLog(ctx context.Context, pluginID, timeStamp, response, logType, msg, color string) error
	WriteFile(filePath string, base64Data string) error
}

// API 插件API服务
type API struct {
	pluginManager PluginManager
}

// NewAPI 创建插件API实例
func NewAPI(pluginManager PluginManager) *API {
	return &API{
		pluginManager: pluginManager,
	}
}

// GetConfig 获取配置文件
func (a *API) GetConfig(r *ghttp.Request) {
	content, err := a.pluginManager.GetConfigYaml()
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code": 200,
		"data": content,
	})
}

// GetWechat 获取当前微信账号
func (a *API) GetWechat(r *ghttp.Request) {
	content, err := a.pluginManager.GetCurrentWechat()
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code": 500,
			"msg":  err.Error(),
			"data": g.Map{"list": []interface{}{}},
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code": 200,
		"data": content,
	})
}

// SendLog 发送日志
func (a *API) SendLog(r *ghttp.Request) {
	var req struct {
		PluginID  string `json:"pluginId"`
		TimeStamp string `json:"timeStamp"`
		Response  string `json:"response"`
		LogType   string `json:"logType"`
		Msg       string `json:"msg"`
		Color     string `json:"color"`
	}

	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"code": 400,
			"msg":  "参数解析失败",
		})
		return
	}

	if err := a.pluginManager.SendPluginLog(
		r.Context(),
		req.PluginID,
		req.TimeStamp,
		req.Response,
		req.LogType,
		req.Msg,
		req.Color,
	); err != nil {
		r.Response.WriteJson(g.Map{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code": 200,
		"msg":  "success",
	})
}

// UploadFile 上传文件
func (a *API) UploadFile(r *ghttp.Request) {
	var req struct {
		FilePath string `json:"filePath"`
		FileData string `json:"fileData"` // Base64 编码
	}

	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"code": 400,
			"msg":  "参数解析失败",
		})
		return
	}

	if err := a.pluginManager.WriteFile(req.FilePath, req.FileData); err != nil {
		r.Response.WriteJson(g.Map{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code": 200,
		"msg":  "上传成功",
	})
}
