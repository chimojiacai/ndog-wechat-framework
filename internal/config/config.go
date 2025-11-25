package config

import (
	"fmt"

	"github.com/gogf/gf/v2/encoding/gyaml"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
)

const (
	ConfigFilePath = "configs/config.yaml"
)

// Service 配置服务
type Service struct{}

// NewService 创建配置服务实例
func NewService() *Service {
	return &Service{}
}

// GetWechatConfig 获取微信配置信息
func (s *Service) GetWechatConfig() map[string]interface{} {
	ctx := gctx.New()
	v, err := g.Cfg().Get(ctx, "wechat")
	if err != nil {
		g.Log().Errorf(ctx, "获取微信配置失败: %v", err)
		return nil
	}
	return v.Map()
}

// GetClearLog 获取日志清理阈值
func (s *Service) GetClearLog() int {
	ctx := gctx.New()
	v, err := g.Cfg().Get(ctx, "wechat.clearLog")
	if err != nil {
		g.Log().Warningf(ctx, "获取 clearLog 配置失败: %v", err)
		return 500 // 默认值
	}
	return v.Int()
}

// SetWechatConfig 设置微信配置信息
func (s *Service) SetWechatConfig(obj map[string]interface{}) (bool, error) {
	ctx := gctx.New()

	if !gfile.Exists(ConfigFilePath) {
		return false, fmt.Errorf("配置文件不存在: %s", ConfigFilePath)
	}

	content := gfile.GetBytes(ConfigFilePath)
	if len(content) == 0 {
		return false, fmt.Errorf("配置文件不可为空")
	}

	var configMap map[string]interface{}
	if err := gyaml.DecodeTo(content, &configMap); err != nil {
		return false, fmt.Errorf("解析配置文件失败: %v", err)
	}

	if configMap["wechat"] == nil {
		configMap["wechat"] = make(map[string]interface{})
	}

	wechatConfig, ok := configMap["wechat"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("配置文件格式错误")
	}

	for key, value := range obj {
		wechatConfig[key] = value
	}

	yamlBytes, err := gyaml.Encode(configMap)
	if err != nil {
		return false, fmt.Errorf("配置文件编码失败: %v", err)
	}

	if err := gfile.PutBytes(ConfigFilePath, yamlBytes); err != nil {
		return false, fmt.Errorf("写入配置文件失败: %v", err)
	}

	// 清除配置缓存
	g.Cfg().GetAdapter().(*gcfg.AdapterFile).Clear()

	g.Log().Infof(ctx, "微信配置更新成功")
	return true, nil
}

// GetTheme 获取当前主题设置
func (s *Service) GetTheme() (string, error) {
	ctx := gctx.New()
	v, err := g.Cfg().Get(ctx, "system.theme")
	if err != nil {
		return "light", nil
	}
	theme := v.String()
	if theme == "" {
		return "light", nil
	}
	return theme, nil
}

// SetTheme 设置主题
// theme: "light" (亮色主题), "dark" (黑色主题), "system" (跟随系统)
func (s *Service) SetTheme(theme string) (bool, error) {
	ctx := gctx.New()

	// 验证主题值
	if theme != "light" && theme != "dark" && theme != "system" {
		return false, fmt.Errorf("无效的主题值，仅支持: light, dark, system")
	}

	if !gfile.Exists(ConfigFilePath) {
		return false, fmt.Errorf("配置文件不存在: %s", ConfigFilePath)
	}

	content := gfile.GetBytes(ConfigFilePath)
	if len(content) == 0 {
		return false, fmt.Errorf("配置文件不可为空")
	}

	var configMap map[string]interface{}
	if err := gyaml.DecodeTo(content, &configMap); err != nil {
		return false, fmt.Errorf("解析配置文件失败: %v", err)
	}

	if configMap["system"] == nil {
		configMap["system"] = make(map[string]interface{})
	}

	systemConfig, ok := configMap["system"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("system配置格式错误")
	}

	systemConfig["theme"] = theme

	yamlBytes, err := gyaml.Encode(configMap)
	if err != nil {
		return false, fmt.Errorf("配置文件编码失败: %v", err)
	}

	if err := gfile.PutBytes(ConfigFilePath, yamlBytes); err != nil {
		return false, fmt.Errorf("写入配置文件失败: %v", err)
	}

	g.Cfg().GetAdapter().(*gcfg.AdapterFile).Clear()
	g.Log().Infof(ctx, "主题设置更新成功: %s", theme)
	return true, nil
}
