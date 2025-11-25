package plugin

import (
	"archive/zip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type PluginMetadata struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Entry       string `json:"entry"`
	Type        string `json:"type"`
}

type PluginInfo struct {
	Metadata PluginMetadata `json:"metadata"`
	Path     string         `json:"path"`
	Enabled  bool           `json:"enabled"`
	IconURL  string         `json:"iconUrl"`
	EntryURL string         `json:"entryUrl"`
}

type PluginService struct {
	app           *application.App
	pluginWindows map[string]*application.WebviewWindow // pluginID -> window
	windowMutex   sync.RWMutex                          // 保护 pluginWindows 的锁
	logService    interface{}                           // 日志服务引用
	pluginCache   []PluginInfo                          // 插件缓存
	cacheMutex    sync.RWMutex                          // 保护插件缓存的锁
}

// SetApp 设置应用实例
func (s *PluginService) SetApp(app *application.App) {
	s.app = app
	s.pluginWindows = make(map[string]*application.WebviewWindow)
}

// SetLogService 设置日志服务引用
func (s *PluginService) SetLogService(logService interface{}) {
	s.logService = logService
}

// RefreshPlugins 刷新插件列表（热重载）
func (s *PluginService) RefreshPlugins() ([]PluginInfo, error) {
	plugins, err := s.scanPluginsFromDisk()
	if err != nil {
		return nil, err
	}

	// 更新缓存
	s.cacheMutex.Lock()
	s.pluginCache = plugins
	s.cacheMutex.Unlock()

	g.Log().Infof(nil, "插件列表已刷新，共 %d 个插件", len(plugins))
	return plugins, nil
}

// ScanPlugins 扫描插件（优先返回缓存）
func (s *PluginService) ScanPlugins() ([]PluginInfo, error) {
	// 如果有缓存，直接返回
	s.cacheMutex.RLock()
	if len(s.pluginCache) > 0 {
		cached := s.pluginCache
		s.cacheMutex.RUnlock()
		return cached, nil
	}
	s.cacheMutex.RUnlock()

	// 首次扫描
	return s.RefreshPlugins()
}

// scanPluginsFromDisk 从磁盘扫描插件
func (s *PluginService) scanPluginsFromDisk() ([]PluginInfo, error) {
	pluginDir := "plugins"

	if !gfile.Exists(pluginDir) {
		gfile.Mkdir(pluginDir)
		return []PluginInfo{}, nil
	}

	// 先处理 .dog 压缩包
	if err := s.extractDogFiles(pluginDir); err != nil {
		g.Log().Warningf(nil, "处理 .dog 文件失败: %v", err)
	}

	var plugins []PluginInfo
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(pluginDir, entry.Name())
		metadataPath := filepath.Join(pluginPath, "plugin.json")

		if !gfile.Exists(metadataPath) {
			continue
		}

		content := gfile.GetBytes(metadataPath)
		var metadata PluginMetadata
		if err := json.Unmarshal(content, &metadata); err != nil {
			g.Log().Errorf(nil, "解析插件失败 %s: %v", entry.Name(), err)
			continue
		}

		// 生成图标和入口 URL
		iconURL := ""
		if metadata.Icon != "" {
			iconURL = fmt.Sprintf("http://localhost:9001/plugins/%s/%s", entry.Name(), metadata.Icon)
		}

		pluginInfo := PluginInfo{
			Metadata: metadata,
			Path:     pluginPath,
			Enabled:  true,
			IconURL:  iconURL,
			EntryURL: fmt.Sprintf("/plugins/%s/%s", entry.Name(), metadata.Entry),
		}

		plugins = append(plugins, pluginInfo)
	}

	return plugins, nil
}

// OpenPlugin 在新窗口中打开插件
func (s *PluginService) OpenPlugin(ctx context.Context, pluginID string) error {
	if s.app == nil {
		return fmt.Errorf("应用实例未初始化")
	}

	// 获取插件信息
	plugins, err := s.ScanPlugins()
	if err != nil {
		return err
	}

	var targetPlugin *PluginInfo
	for _, plugin := range plugins {
		if plugin.Metadata.ID == pluginID {
			targetPlugin = &plugin
			break
		}
	}

	if targetPlugin == nil {
		return fmt.Errorf("插件不存在: %s", pluginID)
	}

	// 检查是否已经打开，如果已存在则先清理旧引用
	// （窗口可能已经被用户关闭，但引用还在）
	s.windowMutex.Lock()
	oldWindow, exists := s.pluginWindows[pluginID]
	if exists {
		// 尝试关闭旧窗口（如果还活着）
		if oldWindow != nil {
			oldWindow.Close()
		}
		// 清理引用
		delete(s.pluginWindows, pluginID)
		g.Log().Infof(ctx, "清理插件旧窗口引用: %s", targetPlugin.Metadata.Name)
	}
	s.windowMutex.Unlock()

	// 构建插件 URL
	pluginURL := fmt.Sprintf("http://localhost:9001%s", targetPlugin.EntryURL)

	// 创建新窗口
	window := s.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:               targetPlugin.Metadata.Name,
		Width:               520,
		Height:              380,
		URL:                 pluginURL,
		MaximiseButtonState: application.ButtonDisabled, // 禁止最大化
		DevToolsEnabled:     false,                      // 禁用开发者工具
		Mac: application.MacWindow{
			Backdrop: application.MacBackdropTranslucent,
			TitleBar: application.MacTitleBarDefault,
		},
		BackgroundColour: application.NewRGB(255, 255, 255),
		Windows:          application.WindowsWindow{Theme: 0}, //这里是设框架主题，0=跟随系统，1=Dark(黑色)，2=Light(浅色)
	})

	if window == nil {
		return fmt.Errorf("创建窗口失败")
	}

	// 保存窗口引用
	s.windowMutex.Lock()
	s.pluginWindows[pluginID] = window
	s.windowMutex.Unlock()

	// 注意：窗口关闭后需要手动清理引用，或者等待下次打开时自动检测

	g.Log().Infof(ctx, "打开插件: %s (%s)", targetPlugin.Metadata.Name, pluginURL)

	return nil
}

// ClosePlugin 关闭插件窗口并清理引用
func (s *PluginService) ClosePlugin(ctx context.Context, pluginID string) error {
	s.windowMutex.Lock()
	window, exists := s.pluginWindows[pluginID]
	if exists {
		if window != nil {
			window.Close()
		}
		delete(s.pluginWindows, pluginID)
	}
	s.windowMutex.Unlock()

	if !exists {
		return fmt.Errorf("插件未打开: %s", pluginID)
	}

	g.Log().Infof(ctx, "关闭插件: %s", pluginID)
	return nil
}

// extractDogFiles 解压 .dog 文件
func (s *PluginService) extractDogFiles(pluginDir string) error {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		// 检查是否为 .dog 文件（大小写不敏感）
		if ext != ".dog" {
			continue
		}

		dogPath := filepath.Join(pluginDir, name)
		// 插件目录名（去掉原始扩展名，兼容大小写）
		pluginName := strings.TrimSuffix(name, filepath.Ext(name))
		targetDir := filepath.Join(pluginDir, pluginName)

		// 如果目录已存在，跳过
		if gfile.Exists(targetDir) {
			g.Log().Debugf(nil, "插件目录已存在，跳过: %s", pluginName)
			continue
		}

		// 解压
		g.Log().Infof(nil, "正在解压插件: %s", entry.Name())
		if err := unzip(dogPath, targetDir); err != nil {
			g.Log().Errorf(nil, "解压失败 %s: %v", entry.Name(), err)
			continue
		}

		g.Log().Infof(nil, "插件解压成功: %s -> %s", entry.Name(), pluginName)

		// 解压成功后删除 .dog 文件
		if err := os.Remove(dogPath); err != nil {
			g.Log().Warningf(nil, "删除压缩包失败 %s: %v", entry.Name(), err)
		} else {
			g.Log().Infof(nil, "已删除压缩包: %s", entry.Name())
		}
	}

	return nil
}

// unzip 解压 zip 文件
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// 创建目标目录
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for _, f := range r.File {
		filePath := filepath.Join(dest, f.Name)

		// 防止目录遍历攻击
		if !strings.HasPrefix(filepath.Clean(filePath), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("非法文件路径: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, f.Mode())
			continue
		}

		// 创建文件的父目录
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		// 创建文件
		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		// 打开 zip 中的文件
		srcFile, err := f.Open()
		if err != nil {
			dstFile.Close()
			return err
		}

		// 复制内容
		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// GetConfigYaml 获取 config.yaml 文件内容
func (s *PluginService) GetConfigYaml() (string, error) {
	configPath := "configs/config.yaml"
	if !gfile.Exists(configPath) {
		return "", fmt.Errorf("配置文件不存在")
	}
	content := gfile.GetContents(configPath)
	return content, nil
}

// GetCurrentWechat 获取 currentWechat.json 文件内容
func (s *PluginService) GetCurrentWechat() (string, error) {
	currentWechatPath := "resources/currentWechat.json"
	if !gfile.Exists(currentWechatPath) {
		return "", fmt.Errorf("当前微信账号文件不存在")
	}
	content := gfile.GetContents(currentWechatPath)
	return content, nil
}

// SendPluginLog 插件发送日志到主程序
func (s *PluginService) SendPluginLog(ctx context.Context, pluginID, timeStamp, response, logType, msg, color string) error {
	if s.logService == nil {
		return fmt.Errorf("日志服务未初始化")
	}

	// 使用类型断言调用日志服务的 SendLog 方法
	type LogService interface {
		SendLog(ctx context.Context, timeStamp, response, logType, msg, color string)
	}

	if logSvc, ok := s.logService.(LogService); ok {
		logSvc.SendLog(ctx, timeStamp, response, logType, msg, color)
		return nil
	}

	return fmt.Errorf("日志服务类型不匹配")
}

// BroadcastEventToPlugins 向所有打开的插件广播事件
func (s *PluginService) BroadcastEventToPlugins(eventType string, eventData interface{}) {
	s.windowMutex.RLock()
	defer s.windowMutex.RUnlock()

	for pluginID, window := range s.pluginWindows {
		if window != nil {
			// 向插件窗口发送事件
			window.EmitEvent("wechat:event", map[string]interface{}{
				"type": eventType,
				"data": eventData,
			})
			g.Log().Debugf(nil, "向插件 %s 广播事件: %s", pluginID, eventType)
		}
	}
}

// WriteFile 写入文件（用于前端上传）
// 接收 Base64 编码的字符串，解决 Wails 二进制数据传输问题
func (s *PluginService) WriteFile(filePath string, base64Data string) error {
	// 限制只能写入 plugins 目录
	if !filepath.HasPrefix(filepath.Clean(filePath), "plugins") {
		return fmt.Errorf("只能上传到 plugins 目录")
	}

	// 解码 Base64
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("Base64 解码失败: %v", err)
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	g.Log().Infof(nil, "文件上传成功: %s (%d bytes)", filePath, len(data))
	return nil
}

// UninstallPlugin 卸载插件
func (s *PluginService) UninstallPlugin(ctx context.Context, pluginID string) error {
	// 先关闭插件窗口（如果打开着）
	s.windowMutex.Lock()
	window, exists := s.pluginWindows[pluginID]
	if exists {
		if window != nil {
			window.Close()
		}
		delete(s.pluginWindows, pluginID)
	}
	s.windowMutex.Unlock()

	// 删除插件目录
	pluginDir := filepath.Join("plugins", pluginID)
	if !gfile.Exists(pluginDir) {
		return fmt.Errorf("插件不存在: %s", pluginID)
	}

	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("删除插件失败: %v", err)
	}

	// 清理缓存
	s.cacheMutex.Lock()
	newCache := []PluginInfo{}
	for _, p := range s.pluginCache {
		if p.Metadata.ID != pluginID {
			newCache = append(newCache, p)
		}
	}
	s.pluginCache = newCache
	s.cacheMutex.Unlock()

	g.Log().Infof(ctx, "插件已卸载: %s", pluginID)
	return nil
}
