package callback

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/naidog/wechat-framework/pkg/types"
)

// SSE 客户端管理
var (
	sseClients      = make(map[chan string]bool)
	sseClientsMutex sync.RWMutex
)

// Handler 回调处理器
type Handler struct {
	pluginBroadcaster PluginBroadcaster
}

// PluginBroadcaster 插件广播接口
type PluginBroadcaster interface {
	BroadcastEventToPlugins(eventType string, eventData interface{})
}

// NewHandler 创建回调处理器实例
func NewHandler(pluginBroadcaster PluginBroadcaster) *Handler {
	return &Handler{
		pluginBroadcaster: pluginBroadcaster,
	}
}

// HandleCallback 处理微信回调事件
func (h *Handler) HandleCallback(r *ghttp.Request) {
	body := r.GetBody()
	g.Log().Debugf(r.Context(), "收到回调请求，原始数据: %s", string(body))

	var event types.CallbackEvent
	if err := r.Parse(&event); err != nil {
		g.Log().Errorf(r.Context(), "解析回调事件失败: %v, 原始数据: %s", err, string(body))
		r.Response.WriteJson(g.Map{
			"code": 400,
			"msg":  "解析失败",
		})
		return
	}

	g.Log().Infof(r.Context(), "收到事件: %s, 描述: %s", event.Type, event.Des)

	// 根据事件类型分发处理
	switch event.Type {
	case "injectSuccess":
		h.handleInjectSuccess(r.Context(), &event)
	case "loginSuccess":
		h.handleLoginSuccess(r.Context(), &event)
	case "recvMsg":
		h.handleRecvMsg(r.Context(), &event)
	case "transPay":
		h.handleTransPay(r.Context(), &event)
	case "friendReq":
		h.handleFriendReq(r.Context(), &event)
	case "groupMemberChanges":
		h.handleGroupMemberChanges(r.Context(), &event)
	case "authExpire":
		h.handleAuthExpire(r.Context(), &event)
	default:
		g.Log().Warningf(r.Context(), "未知事件类型: %s", event.Type)
	}

	// 广播事件到插件和SSE客户端
	h.broadcastEvent(event.Type, event)

	r.Response.WriteJson(g.Map{
		"code": 200,
		"msg":  "success",
	})
}

// handleInjectSuccess 处理注入成功事件
func (h *Handler) handleInjectSuccess(ctx context.Context, event *types.CallbackEvent) {
	g.Log().Infof(ctx, "注入成功 - 端口: %v, PID: %v", event.Data["port"], event.Data["pid"])
}

// handleLoginSuccess 处理登录成功事件
func (h *Handler) handleLoginSuccess(ctx context.Context, event *types.CallbackEvent) {
	wxid := gconv.String(event.Data["wxid"])
	nick := gconv.String(event.Data["nick"])
	port := event.Port
	pid := event.Pid

	g.Log().Infof(ctx, "登录成功 - 昵称: %s, wxid: %s, 端口: %d, PID: %d", nick, wxid, port, pid)

	// 更新 currentWechat.json
	h.updateCurrentWechat(ctx, event)
}

// handleRecvMsg 处理接收消息事件
func (h *Handler) handleRecvMsg(ctx context.Context, event *types.CallbackEvent) {
	msgType := gconv.Int(event.Data["msgType"])
	fromWxid := gconv.String(event.Data["fromWxid"])
	msg := gconv.String(event.Data["msg"])

	g.Log().Debugf(ctx, "收到消息 - 类型: %d, 来自: %s, 内容: %s", msgType, fromWxid, msg)
}

// handleTransPay 处理转账事件
func (h *Handler) handleTransPay(ctx context.Context, event *types.CallbackEvent) {
	fromWxid := gconv.String(event.Data["fromWxid"])
	money := gconv.String(event.Data["money"])
	memo := gconv.String(event.Data["memo"])

	g.Log().Infof(ctx, "收到转账 - 来自: %s, 金额: %s, 备注: %s", fromWxid, money, memo)
}

// handleFriendReq 处理好友请求事件
func (h *Handler) handleFriendReq(ctx context.Context, event *types.CallbackEvent) {
	wxid := gconv.String(event.Data["wxid"])
	nick := gconv.String(event.Data["nick"])
	content := gconv.String(event.Data["content"])

	g.Log().Infof(ctx, "收到好友请求 - wxid: %s, 昵称: %s, 内容: %s", wxid, nick, content)
}

// handleGroupMemberChanges 处理群成员变动事件
func (h *Handler) handleGroupMemberChanges(ctx context.Context, event *types.CallbackEvent) {
	fromWxid := gconv.String(event.Data["fromWxid"])
	eventType := gconv.String(event.Data["eventType"])

	g.Log().Infof(ctx, "群成员变动 - 群: %s, 事件: %s", fromWxid, eventType)
}

// handleAuthExpire 处理授权到期事件
func (h *Handler) handleAuthExpire(ctx context.Context, event *types.CallbackEvent) {
	wxid := event.Wxid
	expireTime := gconv.String(event.Data["expireTime"])

	g.Log().Warningf(ctx, "授权到期 - wxid: %s, 到期时间: %s", wxid, expireTime)
}

// updateCurrentWechat 更新当前微信账号信息
func (h *Handler) updateCurrentWechat(ctx context.Context, event *types.CallbackEvent) {
	accountFilePath := "resources/currentWechat.json"

	// 确保目录存在
	dir := filepath.Dir(accountFilePath)
	if !gfile.Exists(dir) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			g.Log().Errorf(ctx, "创建目录失败: %v", err)
			return
		}
	}

	// 读取现有账号列表
	var accountList types.WechatAccountList
	if gfile.Exists(accountFilePath) {
		content := gfile.GetBytes(accountFilePath)
		if len(content) > 0 {
			if err := json.Unmarshal(content, &accountList); err != nil {
				g.Log().Errorf(ctx, "解析账号文件失败: %v", err)
				accountList = types.WechatAccountList{List: []types.WechatAccount{}}
			}
		}
	}

	// 创建新账号
	newAccount := types.WechatAccount{
		Wxid:       gconv.String(event.Data["wxid"]),
		WxNum:      gconv.String(event.Data["wxNum"]),
		Nick:       gconv.String(event.Data["nick"]),
		AvatarUrl:  gconv.String(event.Data["avatarUrl"]),
		Port:       event.Port,
		Pid:        event.Pid,
		ExpireTime: "",
		IsExpire:   0,
	}

	// 检查是否已存在
	found := false
	for i, acc := range accountList.List {
		if acc.Wxid == newAccount.Wxid {
			accountList.List[i] = newAccount
			found = true
			break
		}
	}

	if !found {
		accountList.List = append(accountList.List, newAccount)
	}

	// 写入文件
	jsonData, err := json.MarshalIndent(accountList, "", "  ")
	if err != nil {
		g.Log().Errorf(ctx, "序列化账号数据失败: %v", err)
		return
	}

	if err := os.WriteFile(accountFilePath, jsonData, 0644); err != nil {
		g.Log().Errorf(ctx, "写入账号文件失败: %v", err)
		return
	}

	g.Log().Infof(ctx, "账号信息已更新: %s", newAccount.Nick)
}

// broadcastEvent 广播事件到插件和SSE客户端
func (h *Handler) broadcastEvent(eventType string, eventData interface{}) {
	// 广播到插件
	if h.pluginBroadcaster != nil {
		h.pluginBroadcaster.BroadcastEventToPlugins(eventType, eventData)
	}

	// 广播到SSE客户端
	BroadcastEventToSSE(eventType, eventData)
}

// BroadcastEventToSSE 广播事件给所有 SSE 客户端
func BroadcastEventToSSE(eventType string, eventData interface{}) {
	sseClientsMutex.RLock()
	defer sseClientsMutex.RUnlock()

	data, err := json.Marshal(map[string]interface{}{
		"type": eventType,
		"data": eventData,
	})
	if err != nil {
		g.Log().Errorf(nil, "SSE 事件序列化失败: %v", err)
		return
	}

	message := fmt.Sprintf("data: %s\n\n", string(data))

	for client := range sseClients {
		select {
		case client <- message:
		default:
			// 客户端阻塞，跳过
		}
	}
}

// AddSSEClient 添加 SSE 客户端
func AddSSEClient(client chan string) {
	sseClientsMutex.Lock()
	defer sseClientsMutex.Unlock()
	sseClients[client] = true
}

// RemoveSSEClient 移除 SSE 客户端
func RemoveSSEClient(client chan string) {
	sseClientsMutex.Lock()
	defer sseClientsMutex.Unlock()
	delete(sseClients, client)
	close(client)
}

// HandleSSEEvents 处理 SSE 事件流
func HandleSSEEvents(r *ghttp.Request) {
	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	r.Response.Header().Set("Access-Control-Allow-Origin", "*")

	client := make(chan string, 10)
	AddSSEClient(client)
	defer RemoveSSEClient(client)

	// 发送连接成功消息
	r.Response.Write([]byte("data: " + `{"type":"connected","data":{"message":"连接成功"}}` + "\n\n"))
	r.Response.Flush()

	ctx := r.Context()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case message := <-client:
			r.Response.Write([]byte(message))
			r.Response.Flush()
		case <-ticker.C:
			// 发送心跳
			r.Response.Write([]byte(": heartbeat\n\n"))
			r.Response.Flush()
		}
	}
}
