package account

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/naidog/wechat-framework/pkg/types"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	AccountFilePath = "resources/currentWechat.json"
)

// Manager 微信账号管理器
type Manager struct {
	app         *application.App
	lastContent string
	mu          sync.RWMutex
}

// NewManager 创建账号管理器实例
func NewManager(app *application.App) *Manager {
	return &Manager{
		app: app,
	}
}

// SetApp 设置应用实例
func (m *Manager) SetApp(app *application.App) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.app = app
}

// StartWatching 开始监听账号文件变化并推送到前端
func (m *Manager) StartWatching(ctx context.Context) {
	g.Log().Info(ctx, "开始监听微信账号文件变化...")

	// 立即发送一次初始数据
	go m.checkAndEmit(ctx)

	// 每秒检查一次文件变化
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.checkAndEmit(ctx)
			}
		}
	}()

	// 启动心跳检测，每 2 秒检查一次微信是否在线
	go m.startHeartbeat(ctx)
}

// checkAndEmit 检查文件内容是否变化，如果变化则发送事件
func (m *Manager) checkAndEmit(ctx context.Context) {
	if !gfile.Exists(AccountFilePath) {
		m.mu.RLock()
		lastContent := m.lastContent
		m.mu.RUnlock()

		if lastContent != "" {
			m.mu.Lock()
			m.lastContent = ""
			m.mu.Unlock()
			m.emitAccounts(ctx, []types.WechatAccount{})
		} else if lastContent == "" {
			m.emitAccounts(ctx, []types.WechatAccount{})
		}
		return
	}

	fileData, err := os.ReadFile(AccountFilePath)
	if err != nil {
		g.Log().Warningf(ctx, "读取账号文件失败: %v", err)
		return
	}

	content := string(fileData)

	m.mu.RLock()
	lastContent := m.lastContent
	m.mu.RUnlock()

	if content == lastContent && lastContent != "" {
		return
	}

	var accountList types.WechatAccountList
	if len(fileData) > 0 {
		if err := json.Unmarshal(fileData, &accountList); err != nil {
			g.Log().Warningf(ctx, "解析账号JSON失败: %v", err)
			return
		}
	}

	m.mu.Lock()
	m.lastContent = content
	m.mu.Unlock()

	g.Log().Infof(ctx, "账号文件内容变化，准备发送 %d 个账号", len(accountList.List))
	m.emitAccounts(ctx, accountList.List)
}

// emitAccounts 发送账号列表到前端
func (m *Manager) emitAccounts(ctx context.Context, accounts []types.WechatAccount) {
	if m.app == nil || m.app.Event == nil {
		g.Log().Warning(ctx, "应用实例未初始化，无法发送事件")
		return
	}

	g.Log().Infof(ctx, "发送事件: wechat:accounts:update, 账号数量: %d", len(accounts))
	m.app.Event.Emit("wechat:accounts:update", accounts)
}

// GetAccounts 获取当前账号列表
func (m *Manager) GetAccounts(ctx context.Context) []types.WechatAccount {
	if !gfile.Exists(AccountFilePath) {
		return []types.WechatAccount{}
	}

	fileData, err := os.ReadFile(AccountFilePath)
	if err != nil {
		return []types.WechatAccount{}
	}

	var accountList types.WechatAccountList
	if len(fileData) > 0 {
		if err := json.Unmarshal(fileData, &accountList); err != nil {
			return []types.WechatAccount{}
		}
	}

	return accountList.List
}

// startHeartbeat 启动心跳检测
func (m *Manager) startHeartbeat(ctx context.Context) {
	g.Log().Info(ctx, "启动微信心跳检测服务...")

	// 等待 15 秒，给授权检查留出时间
	time.Sleep(15 * time.Second)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			g.Log().Info(ctx, "心跳检测服务已停止")
			return
		case <-ticker.C:
			m.checkAccountsHealth(ctx)
		}
	}
}

// checkAccountsHealth 检查所有账号的健康状态
func (m *Manager) checkAccountsHealth(ctx context.Context) {
	if !gfile.Exists(AccountFilePath) {
		return
	}

	fileData, err := os.ReadFile(AccountFilePath)
	if err != nil {
		return
	}

	var accountList types.WechatAccountList
	if len(fileData) > 0 {
		if err := json.Unmarshal(fileData, &accountList); err != nil {
			return
		}
	}

	if len(accountList.List) == 0 {
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	aliveAccounts := make([]types.WechatAccount, 0)
	removedCount := 0
	updatedCount := 0

	for i := range accountList.List {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			acc := accountList.List[index]

			// 查询授权信息
			authInfo := m.getAuthInfo(ctx, acc.Port)
			if authInfo == nil {
				mu.Lock()
				removedCount++
				mu.Unlock()
				g.Log().Warningf(ctx, "检测到微信已退出: %s (端口:%d)", acc.Wxid, acc.Port)
				return
			}

			oldExpire := acc.IsExpire
			oldExpireTime := acc.ExpireTime

			acc.ExpireTime = authInfo.ExpireTime
			acc.IsExpire = authInfo.IsExpire

			if oldExpire != acc.IsExpire || oldExpireTime != acc.ExpireTime {
				mu.Lock()
				updatedCount++
				mu.Unlock()
			}

			mu.Lock()
			aliveAccounts = append(aliveAccounts, acc)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	if removedCount > 0 || updatedCount > 0 {
		if removedCount > 0 {
			g.Log().Infof(ctx, "移除 %d 个已退出的微信账号，剩余 %d 个", removedCount, len(aliveAccounts))
		}
		if updatedCount > 0 {
			g.Log().Infof(ctx, "更新 %d 个账号的授权信息", updatedCount)
		}

		m.mu.Lock()
		accountList.List = aliveAccounts
		jsonData, err := json.MarshalIndent(accountList, "", "  ")
		if err == nil {
			os.WriteFile(AccountFilePath, jsonData, 0644)
		}
		m.mu.Unlock()
	}
}

// getAuthInfo 获取授权信息
func (m *Manager) getAuthInfo(ctx context.Context, port int) *types.AuthInfo {
	url := fmt.Sprintf("http://127.0.0.1:%d/wechat/httpapi", port)

	requestBody := map[string]string{
		"type": "getAuthInfo",
	}
	jsonData, _ := json.Marshal(requestBody)

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var result struct {
		Code   int `json:"code"`
		Result struct {
			ExpireTime string `json:"expireTime"`
			IsExpire   int    `json:"isExpire"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	if result.Code != 200 {
		return nil
	}

	return &types.AuthInfo{
		ExpireTime: result.Result.ExpireTime,
		IsExpire:   result.Result.IsExpire,
	}
}
