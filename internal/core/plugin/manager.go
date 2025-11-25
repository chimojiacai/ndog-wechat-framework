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
	"github.com/naidog/wechat-framework/pkg/logger"
	"github.com/naidog/wechat-framework/pkg/types"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	PluginDir = "plugins"
)

// Manager 插件管理器
type Manager struct {
	app           *application.App
	pluginWindows map[string]*application.WebviewWindow // pluginID -> window
	windowMutex   sync.RWMutex                          // 保护 pluginWindows 的锁
	logService    *logger.Service                       // 日志服务引用
	pluginCache   []types.PluginInfo                    // 插件缓存
	cacheMutex    sync.RWMutex                          // 保护插件缓存的锁
}

// NewManager 创建插件管理器实例
func NewManager(app *application.App, logService *logger.Service) *Manager {
	return &Manager{
		app:           app,
		pluginWindows: make(map[string]*application.WebviewWindow),
		logService:    logService,
		pluginCache:   make([]types.PluginInfo, 0),
	}
}

// SetApp 设置应用实例
func (m *Manager) SetApp(app *application.App) {
	m.app = app
	if m.pluginWindows == nil {
		m.pluginWindows = make(map[string]*application.WebviewWindow)
	}
}

// SetLogService 设置日志服务引用
func (m *Manager) SetLogService(logService *logger.Service) {
	m.logService = logService
}

// RefreshPlugins 刷新插件列表（热重载）
func (m *Manager) RefreshPlugins() ([]types.PluginInfo, error) {
	plugins, err := m.scanPluginsFromDisk()
	if err != nil {
		return nil, err
	}

	// 更新缓存
	m.cacheMutex.Lock()
	m.pluginCache = plugins
	m.cacheMutex.Unlock()

	g.Log().Infof(nil, "插件列表已刷新，共 %d 个插件", len(plugins))
	return plugins, nil
}

// ScanPlugins 扫描插件（优先返回缓存）
func (m *Manager) ScanPlugins() ([]types.PluginInfo, error) {
	// 如果有缓存，直接返回
	m.cacheMutex.RLock()
	if len(m.pluginCache) > 0 {
		cached := m.pluginCache
		m.cacheMutex.RUnlock()
		return cached, nil
	}
	m.cacheMutex.RUnlock()

	// 首次扫描
	return m.RefreshPlugins()
}

// scanPluginsFromDisk 从磁盘扫描插件
func (m *Manager) scanPluginsFromDisk() ([]types.PluginInfo, error) {
	if !gfile.Exists(PluginDir) {
		gfile.Mkdir(PluginDir)
		return []types.PluginInfo{}, nil
	}

	// 先处理 .dog 压缩包
	if err := m.extractDogFiles(PluginDir); err != nil {
		g.Log().Warningf(nil, "处理 .dog 文件失败: %v", err)
	}

	var plugins []types.PluginInfo
	entries, err := os.ReadDir(PluginDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(PluginDir, entry.Name())
		metadataPath := filepath.Join(pluginPath, "plugin.json")

		if !gfile.Exists(metadataPath) {
			continue
		}

		content := gfile.GetBytes(metadataPath)
		var metadata types.PluginMetadata
		if err := json.Unmarshal(content, &metadata); err != nil {
			g.Log().Errorf(nil, "解析插件失败 %s: %v", entry.Name(), err)
			continue
		}

		// 生成图标和入口 URL
		iconURL := ""
		if metadata.Icon != "" {
			iconURL = fmt.Sprintf("http://localhost:9001/plugins/%s/%s", entry.Name(), metadata.Icon)
		}

		pluginInfo := types.PluginInfo{
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
func (m *Manager) OpenPlugin(ctx context.Context, pluginID string) error {
	if m.app == nil {
		return fmt.Errorf("应用实例未初始化")
	}

	// 获取插件信息
	plugins, err := m.ScanPlugins()
	if err != nil {
		return err
	}

	var targetPlugin *types.PluginInfo
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
	m.windowMutex.Lock()
	oldWindow, exists := m.pluginWindows[pluginID]
	if exists {
		if oldWindow != nil {
			oldWindow.Close()
		}
		delete(m.pluginWindows, pluginID)
		g.Log().Infof(ctx, "清理插件旧窗口引用: %s", targetPlugin.Metadata.Name)
	}
	m.windowMutex.Unlock()

	// 构建插件 URL
	pluginURL := fmt.Sprintf("http://localhost:9001%s", targetPlugin.EntryURL)

	// 创建新窗口
	window := m.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:               targetPlugin.Metadata.Name,
		Width:               520,
		Height:              380,
		URL:                 pluginURL,
		MaximiseButtonState: application.ButtonDisabled,
		DevToolsEnabled:     false,
		Mac: application.MacWindow{
			Backdrop: application.MacBackdropTranslucent,
			TitleBar: application.MacTitleBarDefault,
		},
		BackgroundColour: application.NewRGB(255, 255, 255),
		Windows:          application.WindowsWindow{Theme: 0},
	})

	if window == nil {
		return fmt.Errorf("创建窗口失败")
	}

	// 保存窗口引用
	m.windowMutex.Lock()
	m.pluginWindows[pluginID] = window
	m.windowMutex.Unlock()

	g.Log().Infof(ctx, "打开插件: %s (%s)", targetPlugin.Metadata.Name, pluginURL)
	return nil
}

// ClosePlugin 关闭插件窗口并清理引用
func (m *Manager) ClosePlugin(ctx context.Context, pluginID string) error {
	m.windowMutex.Lock()
	window, exists := m.pluginWindows[pluginID]
	if exists {
		if window != nil {
			window.Close()
		}
		delete(m.pluginWindows, pluginID)
	}
	m.windowMutex.Unlock()

	if !exists {
		return fmt.Errorf("插件未打开: %s", pluginID)
	}

	g.Log().Infof(ctx, "关闭插件: %s", pluginID)
	return nil
}

// extractDogFiles 解压 .dog 文件
func (m *Manager) extractDogFiles(pluginDir string) error {
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
		if ext != ".dog" {
			continue
		}

		dogPath := filepath.Join(pluginDir, name)
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
func (m *Manager) GetConfigYaml() (string, error) {
	configPath := "configs/config.yaml"
	if !gfile.Exists(configPath) {
		return "", fmt.Errorf("配置文件不存在")
	}
	content := gfile.GetContents(configPath)
	return content, nil
}

// GetCurrentWechat 获取 currentWechat.json 文件内容
func (m *Manager) GetCurrentWechat() (string, error) {
	currentWechatPath := "resources/currentWechat.json"
	if !gfile.Exists(currentWechatPath) {
		return "", fmt.Errorf("当前微信账号文件不存在")
	}
	content := gfile.GetContents(currentWechatPath)
	return content, nil
}

// SendPluginLog 插件发送日志到主程序
func (m *Manager) SendPluginLog(ctx context.Context, pluginID, timeStamp, response, logType, msg, color string) error {
	if m.logService == nil {
		return fmt.Errorf("日志服务未初始化")
	}

	m.logService.SendLog(ctx, timeStamp, response, logType, msg, color)
	return nil
}

// BroadcastEventToPlugins 向所有打开的插件广播事件
func (m *Manager) BroadcastEventToPlugins(eventType string, eventData interface{}) {
	m.windowMutex.RLock()
	defer m.windowMutex.RUnlock()

	for pluginID, window := range m.pluginWindows {
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
func (m *Manager) WriteFile(filePath string, base64Data string) error {
	// 限制只能写入 plugins 目录
	if !filepath.HasPrefix(filepath.Clean(filePath), PluginDir) {
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
func (m *Manager) UninstallPlugin(ctx context.Context, pluginID string) error {
	// 先关闭插件窗口（如果打开着）
	m.windowMutex.Lock()
	window, exists := m.pluginWindows[pluginID]
	if exists {
		if window != nil {
			window.Close()
		}
		delete(m.pluginWindows, pluginID)
	}
	m.windowMutex.Unlock()

	// 删除插件目录
	pluginDir := filepath.Join(PluginDir, pluginID)
	if !gfile.Exists(pluginDir) {
		return fmt.Errorf("插件不存在: %s", pluginID)
	}

	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("删除插件失败: %v", err)
	}

	// 清理缓存
	m.cacheMutex.Lock()
	newCache := []types.PluginInfo{}
	for _, p := range m.pluginCache {
		if p.Metadata.ID != pluginID {
			newCache = append(newCache, p)
		}
	}
	m.pluginCache = newCache
	m.cacheMutex.Unlock()

	g.Log().Infof(ctx, "插件已卸载: %s", pluginID)
	return nil
}
