package wechat

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
	"github.com/wailsapp/wails/v3/pkg/application"
)

type WechatAccountService struct {
	app               *application.App
	lastContent       string
	mu                sync.RWMutex
	wechatAccountFile string
}

// WechatAccountInfo 微信账号信息
type WechatAccountInfo struct {
	Wxid       string `json:"wxid"`
	WxNum      string `json:"wxNum"`
	Nick       string `json:"nick"`
	AvatarUrl  string `json:"avatarUrl"`
	Port       int    `json:"port"`
	Pid        int    `json:"pid"`
	ExpireTime string `json:"expireTime,omitempty"`
	IsExpire   int    `json:"isExpire"`
}

// WechatAccountListData 微信账号列表
type WechatAccountListData struct {
	List []WechatAccountInfo `json:"list"`
}

// NewWechatAccountService 创建微信账号服务
func NewWechatAccountService(app *application.App) *WechatAccountService {
	return &WechatAccountService{
		app:               app,
		wechatAccountFile: "resources/currentWechat.json",
	}
}

// SetApp 设置 app 实例
func (s *WechatAccountService) SetApp(app *application.App) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.app = app
}

// StartWatching 开始监听文件变化并推送到前端
func (s *WechatAccountService) StartWatching(ctx context.Context) {
	g.Log().Info(ctx, "开始监听微信账号文件变化...")

	// 立即发送一次初始数据
	go s.checkAndEmit(ctx)

	// 每秒检查一次文件变化
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.checkAndEmit(ctx)
			}
		}
	}()

	// 启动心跳检测，每 10 秒检查一次微信是否在线
	go s.startHeartbeat(ctx)
}

// checkAndEmit 检查文件内容是否变化，如果变化则发送事件
func (s *WechatAccountService) checkAndEmit(ctx context.Context) {
	if !gfile.Exists(s.wechatAccountFile) {
		// 文件不存在
		s.mu.RLock()
		lastContent := s.lastContent
		s.mu.RUnlock()

		// 如果之前有内容，现在文件不存在了，需要通知前端
		if lastContent != "" {
			s.mu.Lock()
			s.lastContent = ""
			s.mu.Unlock()
			g.Log().Info(ctx, "文件不存在，发送空列表")
			s.emitAccounts(ctx, []WechatAccountInfo{})
		} else if lastContent == "" {
			// 首次启动且文件不存在，也要发送
			g.Log().Info(ctx, "首次启动，文件不存在，发送空列表")
			s.emitAccounts(ctx, []WechatAccountInfo{})
		}
		return
	}

	fileData, err := os.ReadFile(s.wechatAccountFile)
	if err != nil {
		g.Log().Warningf(ctx, "读取文件失败: %v", err)
		return
	}

	content := string(fileData)

	// 检查内容是否变化
	s.mu.RLock()
	lastContent := s.lastContent
	s.mu.RUnlock()

	if content == lastContent && lastContent != "" {
		return // 内容未变化且不是首次，不发送
	}

	// 解析 JSON
	var accountList WechatAccountListData
	if len(fileData) > 0 {
		if err := json.Unmarshal(fileData, &accountList); err != nil {
			g.Log().Warningf(ctx, "解析 JSON 失败: %v", err)
			return
		}
	}

	// 更新缓存
	s.mu.Lock()
	s.lastContent = content
	s.mu.Unlock()

	g.Log().Infof(ctx, "文件内容变化，准备发送 %d 个账号", len(accountList.List))

	// 发送到前端
	s.emitAccounts(ctx, accountList.List)
}

// emitAccounts 发送账号列表到前端
func (s *WechatAccountService) emitAccounts(ctx context.Context, accounts []WechatAccountInfo) {
	if s.app == nil {
		g.Log().Warning(ctx, "app 实例为 nil，无法发送事件")
		return
	}

	if s.app.Event == nil {
		g.Log().Warning(ctx, "app.Event 为 nil，无法发送事件")
		return
	}

	g.Log().Infof(ctx, "发送事件: wechat:accounts:update, 账号数量: %d", len(accounts))

	// 调试日志：输出数据结构
	if len(accounts) > 0 {
		jsonData, _ := json.Marshal(accounts)
		g.Log().Debugf(ctx, "发送的数据 JSON: %s", string(jsonData))
		g.Log().Debugf(ctx, "第一个账号: wxid=%s, nick=%s", accounts[0].Wxid, accounts[0].Nick)
	}

	s.app.Event.Emit("wechat:accounts:update", accounts)
	g.Log().Info(ctx, "事件发送完成")
}

// GetAccounts 获取当前账号列表（供前端主动调用）
func (s *WechatAccountService) GetAccounts(ctx context.Context) []WechatAccountInfo {
	if !gfile.Exists(s.wechatAccountFile) {
		return []WechatAccountInfo{}
	}

	fileData, err := os.ReadFile(s.wechatAccountFile)
	if err != nil {
		return []WechatAccountInfo{}
	}

	var accountList WechatAccountListData
	if len(fileData) > 0 {
		if err := json.Unmarshal(fileData, &accountList); err != nil {
			return []WechatAccountInfo{}
		}
	}

	return accountList.List
}

// startHeartbeat 启动心跳检测，定期检查微信是否在线
func (s *WechatAccountService) startHeartbeat(ctx context.Context) {
	g.Log().Info(ctx, "启动微信心跳检测服务...")

	// 等待程序启动后检查并更新所有微信账号的授权信息后再开始检测
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
			s.checkAccountsHealth(ctx)
		}
	}
}

// checkAccountsHealth 检查所有账号的健康状态
func (s *WechatAccountService) checkAccountsHealth(ctx context.Context) {
	if !gfile.Exists(s.wechatAccountFile) {
		return
	}

	fileData, err := os.ReadFile(s.wechatAccountFile)
	if err != nil {
		return
	}

	var accountList WechatAccountListData
	if len(fileData) > 0 {
		if err := json.Unmarshal(fileData, &accountList); err != nil {
			return
		}
	}

	if len(accountList.List) == 0 {
		return
	}

	// 并发检查每个账号
	var wg sync.WaitGroup
	var mu sync.Mutex
	aliveAccounts := make([]WechatAccountInfo, 0)
	removedCount := 0
	updatedCount := 0

	for i := range accountList.List {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			acc := accountList.List[index]

			// 直接查询授权信息（同时检查微信是否在线）
			authInfo := s.getAuthInfo(ctx, acc.Port)
			if authInfo == nil {
				// 获取失败，说明微信已退出
				mu.Lock()
				removedCount++
				mu.Unlock()
				g.Log().Warningf(ctx, "检测到微信已退出: %s (端口:%d)", acc.Wxid, acc.Port)
				return
			}

			// 微信在线，更新授权信息
			// 保存旧值
			oldExpire := acc.IsExpire
			oldExpireTime := acc.ExpireTime

			//g.Log().Debugf(ctx, "账号 %s - 旧值: isExpire=%d, expireTime=%s", acc.Wxid, oldExpire, oldExpireTime)
			//g.Log().Debugf(ctx, "账号 %s - 新值: isExpire=%d, expireTime=%s", acc.Wxid, authInfo.IsExpire, authInfo.ExpireTime)

			// 更新授权信息
			acc.ExpireTime = authInfo.ExpireTime
			acc.IsExpire = authInfo.IsExpire

			// 检查是否有变化
			if oldExpire != acc.IsExpire || oldExpireTime != acc.ExpireTime {
				mu.Lock()
				updatedCount++
				mu.Unlock()

				// 记录详细日志
				if oldExpire != acc.IsExpire {
					if acc.IsExpire == 1 {
						//	g.Log().Warningf(ctx, "检测到授权状态变化: %s (端口:%d) %d -> %d, 到期时间:%s", acc.Wxid, acc.Port, oldExpire, acc.IsExpire, acc.ExpireTime)
					} else {
						// g.Log().Infof(ctx, "检测到授权状态变化: %s (端口:%d) %d -> %d, 到期时间:%s", acc.Wxid, acc.Port, oldExpire, acc.IsExpire, acc.ExpireTime)
					}
				} else if oldExpireTime != acc.ExpireTime {
					// g.Log().Infof(ctx, "检测到授权时间变化: %s (端口:%d) %s -> %s", acc.Wxid, acc.Port, oldExpireTime, acc.ExpireTime)
				}
			}

			mu.Lock()
			aliveAccounts = append(aliveAccounts, acc)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// 如果有账号被移除或授权信息更新，更新文件
	if removedCount > 0 || updatedCount > 0 {
		if removedCount > 0 {
			g.Log().Infof(ctx, "移除 %d 个已退出的微信账号，剩余 %d 个", removedCount, len(aliveAccounts))
		}
		if updatedCount > 0 {
			g.Log().Infof(ctx, "更新 %d 个账号的授权信息", updatedCount)
		}

		s.mu.Lock()
		accountList.List = aliveAccounts
		jsonData, err := json.MarshalIndent(accountList, "", "  ")
		if err == nil {
			os.WriteFile(s.wechatAccountFile, jsonData, 0644)
		}
		s.mu.Unlock()
	}
}

// isWechatAlive 检查微信是否在线（通过 HTTP API）
func (s *WechatAccountService) isWechatAlive(ctx context.Context, port int) bool {
	url := fmt.Sprintf("http://127.0.0.1:%d/wechat/httpapi", port)

	// 构造简单的请求体
	requestBody := map[string]string{
		"type": "getAuthInfo",
	}
	jsonData, _ := json.Marshal(requestBody)

	// 创建 HTTP 请求，超时时间设置为 3 秒
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		g.Log().Debugf(ctx, "isWechatAlive - 创建请求失败 (port:%d): %v", port, err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		g.Log().Debugf(ctx, "isWechatAlive - 请求失败 (port:%d): %v", port, err)
		return false
	}
	defer resp.Body.Close()

	// 只要能连接上就认为在线，不关心返回内容
	isAlive := resp.StatusCode == 200
	g.Log().Debugf(ctx, "isWechatAlive (port:%d): %v (status:%d)", port, isAlive, resp.StatusCode)
	return isAlive
}

// AuthInfo 授权信息
type AuthInfo struct {
	ExpireTime string
	IsExpire   int
}

// getAuthInfo 获取授权信息
func (s *WechatAccountService) getAuthInfo(ctx context.Context, port int) *AuthInfo {
	url := fmt.Sprintf("http://127.0.0.1:%d/wechat/httpapi", port)

	// 构造请求体
	requestBody := map[string]string{
		"type": "getAuthInfo",
	}
	jsonData, _ := json.Marshal(requestBody)

	// 创建 HTTP 请求
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		g.Log().Debugf(ctx, "读取响应体失败 (port:%d): %v", port, err)
		return nil
	}

	// g.Log().Debugf(ctx, "响应体内容 (port:%d): %s", port, string(body))

	// 解析响应
	var result struct {
		Code   int `json:"code"`
		Result struct {
			ExpireTime string `json:"expireTime"`
			IsExpire   int    `json:"isExpire"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		g.Log().Debugf(ctx, "解析响应失败 (port:%d): %v", port, err)
		return nil
	}

	// g.Log().Debugf(ctx, "获取授权信息 (port:%d): code=%d, expireTime=%s, isExpire=%d", port, result.Code, result.Result.ExpireTime, result.Result.IsExpire)

	if result.Code != 200 {
		g.Log().Debugf(ctx, "返回码错误 (port:%d): code=%d", port, result.Code)
		return nil
	}

	return &AuthInfo{
		ExpireTime: result.Result.ExpireTime,
		IsExpire:   result.Result.IsExpire,
	}
}
