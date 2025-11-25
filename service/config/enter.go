package config

import (
	"fmt"

	"github.com/gogf/gf/v2/encoding/gyaml"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
)

type ConfigGetService struct{}
type ConfigSetService struct{}
type ThemeService struct{}

// 获取配置文件微信配置信息
func (cf *ConfigGetService) GetWechatConfig() map[string]interface{} {
	var ctx = gctx.New()
	v, err := g.Cfg().Get(ctx, "wechat")
	if err != nil {
		fmt.Println("获取配置失败:", err)
		return nil
	}

	return v.Map()
}

// 获取日志清理阈值
func (cf *ConfigGetService) GetClearLog() int {
	var ctx = gctx.New()
	v, err := g.Cfg().Get(ctx, "wechat.clearLog")
	if err != nil {
		fmt.Println("获取 clearLog 配置失败:", err)
		return 500 // 默认值
	}
	return v.Int()
}

// 设置配置文件微信配置信息
func (cs *ConfigSetService) SetWechatConfig(obj map[string]interface{}) (bool, error) {
	var ctx = gctx.New()

	configPath := "configs/config.yaml"
	if !gfile.Exists(configPath) {
		return false, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	content := gfile.GetBytes(configPath)
	if len(content) == 0 {
		return false, fmt.Errorf("配置文件不可为空！")
	}

	var configMap map[string]interface{}
	err := gyaml.DecodeTo(content, &configMap)
	if err != nil {
		return false, fmt.Errorf("解析配置文件失败: %v", err)
	}

	if configMap["wechat"] == nil {
		configMap["wechat"] = make(map[string]interface{})
	}

	wechatConfig, ok := configMap["wechat"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("配置文件格式错误！")
	}

	for key, value := range obj {
		wechatConfig[key] = value
	}

	yamlBytes, err := gyaml.Encode(configMap)
	if err != nil {
		return false, fmt.Errorf("配置文件编码失败: %v", err)
	}

	err = gfile.PutBytes(configPath, yamlBytes)
	if err != nil {
		return false, fmt.Errorf("写入配置文件失败: %v", err)
	}

	g.Cfg().GetAdapter().(*gcfg.AdapterFile).Clear()

	fmt.Println(ctx, "微信配置更新成功~")

	return true, nil
}

// GetTheme 获取当前主题设置
func (ts *ThemeService) GetTheme() (string, error) {
	var ctx = gctx.New()
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

// 设置主题
// theme: "light" (亮色主题), "dark" (黑色主题), "system" (跟随系统)
func (ts *ThemeService) SetTheme(theme string) (bool, error) {
	var ctx = gctx.New()

	// 验证主题值
	if theme != "light" && theme != "dark" && theme != "system" {
		return false, fmt.Errorf("无效的主题值，仅支持: light, dark, system")
	}

	// 获取配置文件路径
	configPath := "configs/config.yaml"
	if !gfile.Exists(configPath) {
		return false, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	content := gfile.GetBytes(configPath)
	if len(content) == 0 {
		return false, fmt.Errorf("配置文件不可为空！")
	}

	var configMap map[string]interface{}
	err := gyaml.DecodeTo(content, &configMap)
	if err != nil {
		return false, fmt.Errorf("解析配置文件失败: %v", err)
	}

	if configMap["system"] == nil {
		configMap["system"] = make(map[string]interface{})
	}

	systemConfig, ok := configMap["system"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("system配置格式错误！")
	}

	systemConfig["theme"] = theme

	yamlBytes, err := gyaml.Encode(configMap)
	if err != nil {
		return false, fmt.Errorf("配置文件编码失败: %v", err)
	}

	err = gfile.PutBytes(configPath, yamlBytes)
	if err != nil {
		return false, fmt.Errorf("写入配置文件失败: %v", err)
	}

	g.Cfg().GetAdapter().(*gcfg.AdapterFile).Clear()
	fmt.Println(ctx, "主题设置更新成功: ", theme)
	return true, nil
}
