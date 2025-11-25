package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/naidog/wechat-framework/internal/config"
)

// WechatPathService 微信路径服务
type WechatPathService struct{}

// NewWechatPathService 创建微信路径服务实例
func NewWechatPathService() *WechatPathService {
	return &WechatPathService{}
}

// GetInstallPath 获取微信安装目录
func (s *WechatPathService) GetInstallPath() (string, error) {
	cmd := exec.Command("cmd", "/c", "reg", "query", "HKEY_CURRENT_USER\\Software\\Tencent\\Weixin", "/v", "InstallPath")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// 尝试常见路径
		commonPaths := []string{
			"C:\\Program Files (x86)\\Tencent\\WeChat\\WeChat.exe",
			"C:\\Program Files\\Tencent\\WeChat\\WeChat.exe",
			"D:\\Program Files (x86)\\Tencent\\WeChat\\WeChat.exe",
			"D:\\Program Files\\Tencent\\WeChat\\WeChat.exe",
		}

		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}

		return "", fmt.Errorf("未找到微信安装目录: %v", err)
	}

	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	for _, line := range lines {
		if strings.Contains(line, "InstallPath") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				installPath := parts[len(parts)-1]
				fullPath := filepath.Join(installPath, "WeChat.exe")

				if _, err := os.Stat(fullPath); err == nil {
					return fullPath, nil
				}

				return installPath, nil
			}
		}
	}

	return "", fmt.Errorf("未能从注册表解析微信安装路径")
}

// GetCachePath 获取微信缓存目录
func (s *WechatPathService) GetCachePath() (string, error) {
	userProfile := os.Getenv("USERPROFILE")
	defaultCachePath := filepath.Join(userProfile, "Documents", "NdogCache") + string(filepath.Separator)
	return defaultCachePath, nil
}

// GetWechatPaths 同时获取安装目录和缓存目录
func (s *WechatPathService) GetWechatPaths() (map[string]string, error) {
	result := make(map[string]string)

	installPath, err := s.GetInstallPath()
	if err != nil {
		result["installationPath"] = ""
	} else {
		result["installationPath"] = installPath
	}

	cachePath, cacheErr := s.GetCachePath()
	if cacheErr != nil {
		result["cachePath"] = ""
	} else {
		result["cachePath"] = cachePath
	}

	if err != nil && cacheErr != nil {
		return result, fmt.Errorf("获取微信路径失败")
	}

	return result, nil
}

// WechatUpdateService 微信更新控制服务
type WechatUpdateService struct {
	configService *config.Service
}

// NewWechatUpdateService 创建微信更新控制服务实例
func NewWechatUpdateService(configService *config.Service) *WechatUpdateService {
	return &WechatUpdateService{
		configService: configService,
	}
}

// DisableAutoUpdate 禁止微信自动更新
func (s *WechatUpdateService) DisableAutoUpdate() error {
	wechatConfig := s.configService.GetWechatConfig()
	if wechatConfig["update"] != "1" {
		return s.enableAutoUpdate()
	}

	// 拒绝所有人访问
	cmd := exec.Command("icacls", "C:\\Users\\Administrator\\AppData\\Roaming\\Tencent\\xwechat\\update\\download", "/deny", "Everyone:F")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("执行禁止自动更新失败: %v", err)
	}

	return nil
}

// enableAutoUpdate 启用微信自动更新
func (s *WechatUpdateService) enableAutoUpdate() error {
	// 移除拒绝规则
	cmd := exec.Command("icacls", "C:\\Users\\Administrator\\AppData\\Roaming\\Tencent\\xwechat\\update\\download", "/remove:d", "Everyone")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("执行关闭微信自动更新失败: %v", err)
	}

	return nil
}
