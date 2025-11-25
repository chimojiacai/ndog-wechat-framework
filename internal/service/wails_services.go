package service

import (
	"context"

	"github.com/naidog/wechat-framework/internal/config"
	"github.com/naidog/wechat-framework/internal/core/account"
	"github.com/naidog/wechat-framework/internal/core/plugin"
	"github.com/naidog/wechat-framework/internal/utils"
	"github.com/naidog/wechat-framework/pkg/logger"
	"github.com/naidog/wechat-framework/pkg/types"
)

// ConfigService Wails配置服务适配器
type ConfigService struct {
	configService *config.Service
}

// NewConfigService 创建配置服务
func NewConfigService(configService *config.Service) *ConfigService {
	return &ConfigService{configService: configService}
}

// GetWechatConfig 获取微信配置
func (s *ConfigService) GetWechatConfig() map[string]interface{} {
	return s.configService.GetWechatConfig()
}

// GetClearLog 获取日志清理阈值
func (s *ConfigService) GetClearLog() int {
	return s.configService.GetClearLog()
}

// SetWechatConfig 设置微信配置
func (s *ConfigService) SetWechatConfig(obj map[string]interface{}) (bool, error) {
	return s.configService.SetWechatConfig(obj)
}

// GetTheme 获取主题
func (s *ConfigService) GetTheme() (string, error) {
	return s.configService.GetTheme()
}

// SetTheme 设置主题
func (s *ConfigService) SetTheme(theme string) (bool, error) {
	return s.configService.SetTheme(theme)
}

// WechatPathService Wails微信路径服务适配器
type WechatPathService struct {
	pathService *utils.WechatPathService
}

// NewWechatPathService 创建微信路径服务
func NewWechatPathService(pathService *utils.WechatPathService) *WechatPathService {
	return &WechatPathService{pathService: pathService}
}

// GetWechatPaths 获取微信路径
func (s *WechatPathService) GetWechatPaths() (map[string]string, error) {
	return s.pathService.GetWechatPaths()
}

// WechatUpdateService Wails微信更新服务适配器
type WechatUpdateService struct {
	updateService *utils.WechatUpdateService
}

// NewWechatUpdateService 创建微信更新服务
func NewWechatUpdateService(updateService *utils.WechatUpdateService) *WechatUpdateService {
	return &WechatUpdateService{updateService: updateService}
}

// DisableAutoUpdate 禁用自动更新
func (s *WechatUpdateService) DisableAutoUpdate() error {
	return s.updateService.DisableAutoUpdate()
}

// AccountService Wails账号服务适配器
type AccountService struct {
	accountManager *account.Manager
}

// NewAccountService 创建账号服务
func NewAccountService(accountManager *account.Manager) *AccountService {
	return &AccountService{accountManager: accountManager}
}

// GetAccounts 获取账号列表
func (s *AccountService) GetAccounts(ctx context.Context) []types.WechatAccount {
	return s.accountManager.GetAccounts(ctx)
}

// LogService Wails日志服务适配器
type LogService struct {
	logService *logger.Service
}

// NewLogService 创建日志服务
func NewLogService(logService *logger.Service) *LogService {
	return &LogService{logService: logService}
}

// SendLog 发送日志
func (s *LogService) SendLog(ctx context.Context, timeStamp, response, logType, msg, color string) {
	s.logService.SendLog(ctx, timeStamp, response, logType, msg, color)
}

// PluginService Wails插件服务适配器
type PluginService struct {
	pluginManager *plugin.Manager
}

// NewPluginService 创建插件服务
func NewPluginService(pluginManager *plugin.Manager) *PluginService {
	return &PluginService{pluginManager: pluginManager}
}

// ScanPlugins 扫描插件
func (s *PluginService) ScanPlugins() ([]types.PluginInfo, error) {
	return s.pluginManager.ScanPlugins()
}

// RefreshPlugins 刷新插件列表
func (s *PluginService) RefreshPlugins() ([]types.PluginInfo, error) {
	return s.pluginManager.RefreshPlugins()
}

// OpenPlugin 打开插件
func (s *PluginService) OpenPlugin(ctx context.Context, pluginID string) error {
	return s.pluginManager.OpenPlugin(ctx, pluginID)
}

// ClosePlugin 关闭插件
func (s *PluginService) ClosePlugin(ctx context.Context, pluginID string) error {
	return s.pluginManager.ClosePlugin(ctx, pluginID)
}

// UninstallPlugin 卸载插件
func (s *PluginService) UninstallPlugin(ctx context.Context, pluginID string) error {
	return s.pluginManager.UninstallPlugin(ctx, pluginID)
}

// WriteFile 写入文件
func (s *PluginService) WriteFile(filePath string, base64Data string) error {
	return s.pluginManager.WriteFile(filePath, base64Data)
}
