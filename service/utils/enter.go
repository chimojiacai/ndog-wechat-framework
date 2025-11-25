package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/naidog/wechat-framework/service/config"
)

type GetWechatPathService struct{}

type NoupdateWechatService struct{}

// 获取微信安装目录
func GetWechatInstallPath() (string, error) {
	cmd := exec.Command("cmd", "/c", "reg", "query", "HKEY_CURRENT_USER\\Software\\Tencent\\Weixin", "/v", "InstallPath")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.CombinedOutput()
	if err != nil {
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

// 获取微信缓存目录
func GetWechatCachePath() (string, error) {
	userProfile := os.Getenv("USERPROFILE")

	defaultCachePath := filepath.Join(userProfile, "Documents", "NdogCache") + string(filepath.Separator)
	// 返回默认路径（即使不存在）
	return defaultCachePath, nil
}

// 同时获取安装目录和缓存目录
func (w *GetWechatPathService) GetWechatPaths() (map[string]string, error) {
	result := make(map[string]string)

	installPath, err := GetWechatInstallPath()
	if err != nil {
		result["installationPath"] = ""
	} else {
		result["installationPath"] = installPath
	}

	cachePath, cacheErr := GetWechatCachePath()
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

func (n *NoupdateWechatService) NoupdateWechat() {
	v := config.ConfigGetService{}
	res := v.GetWechatConfig()
	if res["update"] == "1" {
		// 拒绝所有人访问
		cmd := exec.Command("icacls", "C:\\Users\\Administrator\\AppData\\Roaming\\Tencent\\xwechat\\update\\download", "/deny", "Everyone:F")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

		_, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Errorf("执行禁止自动更新失败: %v", err)
			return
		}

		fmt.Println("成功禁止微信自动更新")
		return
	}

	// 移除拒绝规则
	cmd := exec.Command("icacls", "C:\\Users\\Administrator\\AppData\\Roaming\\Tencent\\xwechat\\update\\download", "/remove:d", "Everyone")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Errorf("执行关闭微信自动更新失败: %v, 输出: %s", err, string(output))
		return
	}

	fmt.Println("成功关闭微信自动更新:", string(output))
}
