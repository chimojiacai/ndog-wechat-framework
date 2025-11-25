package http_callback

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/naidog/wechat-framework/service/utils"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/util/gconv"
)

// 全局插件服务引用
var pluginServiceInstance interface{}

// SetPluginService 设置插件服务引用
func SetPluginService(ps interface{}) {
	pluginServiceInstance = ps
}

// SSE 客户端管理
var (
	sseClients      = make(map[chan string]bool)
	sseClientsMutex sync.RWMutex
)

// BroadcastEventToSSE 广播事件给所有 SSE 客户端
func BroadcastEventToSSE(eventType string, eventData interface{}) {
	sseClientsMutex.RLock()
	defer sseClientsMutex.RUnlock()

	// 将事件序列化为 JSON
	data, err := json.Marshal(map[string]interface{}{
		"type": eventType,
		"data": eventData,
	})
	if err != nil {
		g.Log().Errorf(nil, "SSE 事件序列化失败: %v", err)
		return
	}

	message := fmt.Sprintf("data: %s\n\n", string(data))

	// 广播给所有客户端
	for client := range sseClients {
		select {
		case client <- message:
		default:
			// 客户端阻塞，跳过
		}
	}
}

type HttpCallbackService struct{}

// 通用回调事件结构
type CallbackEvent struct {
	Type      string                 `json:"type"`      // 事件类型
	Des       string                 `json:"des"`       // 事件描述
	Data      map[string]interface{} `json:"data"`      // 事件数据
	Timestamp string                 `json:"timestamp"` // 时间戳
	Wxid      string                 `json:"wxid"`      // 微信ID
	Port      int                    `json:"port"`      // 端口
	Pid       int                    `json:"pid"`       // 进程ID
	Flag      string                 `json:"flag"`      // 标识
}

// 注入成功事件数据
type InjectSuccessData struct {
	Port string `json:"port"` // 监听端口
	Pid  string `json:"pid"`  // 进程PID
}

// 登录成功事件数据
type LoginSuccessData struct {
	Wxid      string `json:"wxid"`      // wxid
	WxNum     string `json:"wxNum"`     // 微信号
	Nick      string `json:"nick"`      // 微信昵称
	Device    string `json:"device"`    // 登录设备
	Phone     string `json:"phone"`     // 手机号
	AvatarUrl string `json:"avatarUrl"` // 头像地址
	Country   string `json:"country"`   // 国家
	Province  string `json:"province"`  // 省
	City      string `json:"city"`      // 城市
	Email     string `json:"email"`     // 邮箱
	QQ        string `json:"qq"`        // QQ
	Sign      string `json:"sign"`      // 个性签名
}

// 收到消息事件数据
type RecvMsgData struct {
	TimeStamp     string   `json:"timeStamp"`     // 13位时间戳
	FromType      int      `json:"fromType"`      // 来源类型：1私聊 2群聊 3公众号
	MsgType       int      `json:"msgType"`       // 消息类型
	MsgSource     int      `json:"msgSource"`     // 消息来源：0别人发送 1自己发送
	FromWxid      string   `json:"fromWxid"`      // 来源wxid
	FinalFromWxid string   `json:"finalFromWxid"` // 群内发言人wxid
	AtWxidList    []string `json:"atWxidList"`    // 艾特人wxid列表
	Silence       int      `json:"silence"`       // 消息免打扰：0未开启 1开启
	Membercount   int      `json:"membercount"`   // 群成员数量
	Signature     string   `json:"signature"`     // 消息签名
	Msg           string   `json:"msg"`           // 消息内容
	MsgId         string   `json:"msgId"`         // 消息ID
	SendId        string   `json:"sendId"`        // 消息发送请求ID
}

// 转账事件数据
type TransPayData struct {
	FromWxid      string `json:"fromWxid"`      // 对方wxid
	MsgSource     int    `json:"msgSource"`     // 1收到转账 2对方接收 3发出转账 4自己接收 5对方退还 6自己退还
	TransType     int    `json:"transType"`     // 1即时到账 2延时到账
	Money         string `json:"money"`         // 金额(元)
	Memo          string `json:"memo"`          // 转账备注
	Transferid    string `json:"transferid"`    // 转账ID
	Transcationid string `json:"transcationid"` // 转账ID
	Invalidtime   string `json:"invalidtime"`   // 10位时间戳
	MsgId         string `json:"msgId"`         // 消息ID
}

// 好友请求事件数据
type FriendReqData struct {
	Wxid         string `json:"wxid"`         // 微信ID
	WxNum        string `json:"wxNum"`        // 微信号
	Nick         string `json:"nick"`         // 昵称
	NickBrief    string `json:"nickBrief"`    // 昵称简拼
	NickWhole    string `json:"nickWhole"`    // 昵称全拼
	V3           string `json:"v3"`           // V3数据
	V4           string `json:"v4"`           // V4数据
	Sign         string `json:"sign"`         // 签名
	Country      string `json:"country"`      // 国家
	Province     string `json:"province"`     // 省份
	City         string `json:"city"`         // 城市
	AvatarMinUrl string `json:"avatarMinUrl"` // 头像小图
	AvatarMaxUrl string `json:"avatarMaxUrl"` // 头像大图
	Sex          string `json:"sex"`          // 性别：0未知 1男 2女
	Content      string `json:"content"`      // 附言
	Scene        string `json:"scene"`        // 来源
	ShareWxid    string `json:"shareWxid"`    // 推荐人wxid
	ShareNick    string `json:"shareNick"`    // 推荐人昵称
	GroupWxid    string `json:"groupWxid"`    // 群聊wxid
	MsgId        string `json:"msgId"`        // 消息ID
}

// 群成员变动事件数据
type GroupMemberChangesData struct {
	TimeStamp     string `json:"timeStamp"`     // 13位时间戳
	FromWxid      string `json:"fromWxid"`      // 群wxid
	FinalFromWxid string `json:"finalFromWxid"` // 变动的群成员wxid
	EventType     int    `json:"eventType"`     // 0退群 1进群
	InviterWxid   string `json:"inviterWxid"`   // 邀请人wxid(仅进群时有)
}

// 授权到期事件数据
type AuthExpireData struct {
	Wxid       string `json:"wxid"`       // wxid
	WxNum      string `json:"wxNum"`      // 微信号
	ExpireTime string `json:"expireTime"` // 到期时间
	Msg        string `json:"msg"`        // 提示消息
}

// 处理微信回调
func (s *HttpCallbackService) HandleCallback(r *ghttp.Request) {
	// 读取原始请求体用于调试
	body := r.GetBody()
	g.Log().Debugf(r.Context(), "收到回调请求，原始数据: %s", string(body))

	var event CallbackEvent

	if err := r.Parse(&event); err != nil {
		g.Log().Errorf(r.Context(), "解析回调事件失败: %v, 原始数据: %s", err, string(body))
		r.Response.WriteJson(g.Map{
			"code": 400,
			"msg":  "解析失败",
		})
		return
	}

	g.Log().Infof(r.Context(), "收到回调事件 [%s]: %v", event.Type, event)
	g.Log().Debugf(r.Context(), "事件类型: '%s', wxid: %s, port: %d, pid: %d", event.Type, event.Wxid, event.Port, event.Pid)

	// 广播事件给所有插件（Wails 窗口）
	if pluginServiceInstance != nil {
		type PluginBroadcaster interface {
			BroadcastEventToPlugins(eventType string, eventData interface{})
		}
		if ps, ok := pluginServiceInstance.(PluginBroadcaster); ok {
			ps.BroadcastEventToPlugins(event.Type, event)
		}
	}

	// 广播事件给所有 SSE 客户端（HTTP 插件）
	BroadcastEventToSSE(event.Type, event)

	// 根据事件类型处理
	switch event.Type {
	case "injectSuccess":
		s.handleInjectSuccess(r, event)
	case "loginSuccess":
		s.handleLoginSuccess(r, event)
	case "recvMsg":
		s.handleRecvMsg(r, event)
	case "transPay":
		s.handleTransPay(r, event)
	case "friendReq":
		s.handleFriendReq(r, event)
	case "groupMemberChanges":
		g.Log().Info(r.Context(), "========== 进入 groupMemberChanges 处理 ==========")
		s.handleGroupMemberChanges(r, event)
	case "authExpire":
		s.handleAuthExpire(r, event)
	default:
		g.Log().Warningf(r.Context(), "未知事件类型: '%s', 完整数据: %+v", event.Type, event)
	}

	r.Response.WriteJson(g.Map{
		"code": 200,
		"msg":  "success",
	})
}

// 处理注入成功事件
func (s *HttpCallbackService) handleInjectSuccess(r *ghttp.Request, event CallbackEvent) {

	port := event.Data["port"]
	pid := event.Data["pid"]

	// 获取全局日志服务
	logService := utils.GetGlobalLogService()
	if logService != nil {
		logService.SendLog(
			r.Context(),
			time.Now().Format("2006-01-02 15:04:05"),
			"框架",
			"启动",
			fmt.Sprintf("微信多开 | 端口: %v | 进程: %v", port, pid),
			"#3959CF",
		)
	}

}

// 处理登录成功事件
func (s *HttpCallbackService) handleLoginSuccess(r *ghttp.Request, event CallbackEvent) {
	data := event.Data
	// 获取外层 wxid
	wxid := event.Wxid

	// 获取全局日志服务
	logService := utils.GetGlobalLogService()
	if logService != nil {
		logService.SendLog(
			r.Context(),
			time.Now().Format("2006-01-02 15:04:05"),
			wxid,
			"登录",
			fmt.Sprintf("登录成功 | 微信号: %v", data["wxNum"]),
			"#3959CF",
		)
	}

	// 将 port 和 pid 添加到 data 中
	data["port"] = float64(event.Port)
	data["pid"] = float64(event.Pid)

	// 保存微信账号信息
	if err := s.saveWechatAccount(r.Context(), data); err != nil {
		g.Log().Warningf(r.Context(), "保存微信账号信息失败: %v", err)
	}

	time.Sleep(1 * time.Second) // 延迟 2 秒，确保账号信息已保存
	s.CheckAndUpdateAuthInfo(r.Context())

}

// 处理接收消息事件
func (s *HttpCallbackService) handleRecvMsg(r *ghttp.Request, event CallbackEvent) {
	data := event.Data

	// 获取接收消息的账号 wxid（外层）
	receiverWxid := event.Wxid

	// 使用 gconv 进行类型转换，支持所有类型
	fromType := gconv.Int(data["fromType"])
	msgType := gconv.Int(data["msgType"])

	g.Log().Debugf(r.Context(), "接收账号: %s, fromType=%d, msgType=%d", receiverWxid, fromType, msgType)

	var fromWxid, msg string
	if v, ok := data["fromWxid"].(string); ok {
		fromWxid = v
	}
	if v, ok := data["msg"].(string); ok {
		msg = v
	}

	// 判断消息来源
	var sourceDesc string
	switch fromType {
	case 1:
		sourceDesc = "私聊"
	case 2:
		sourceDesc = "群聊"
	case 3:
		sourceDesc = "公众号"
	default:
		sourceDesc = fmt.Sprintf("未知", fromType)
	}

	// 判断消息发送者 暂时用不到
	//var senderDesc string
	//if msgSource == 0 {
	//	senderDesc = "别人发送"
	//} else if msgSource == 1 {
	//	senderDesc = "自己发送"
	//} else {
	//	senderDesc = "未知来源"
	//}

	// 判断消息类型
	var msgTypeDesc string
	switch msgType {
	case 1:
		msgTypeDesc = "文本"
	case 3:
		msgTypeDesc = "图片"
	case 34:
		msgTypeDesc = "语音"
	case 42:
		msgTypeDesc = "名片"
	case 43:
		msgTypeDesc = "视频"
	case 47:
		msgTypeDesc = "动态表情"
	case 48:
		msgTypeDesc = "地理位置"
	case 49:
		msgTypeDesc = "分享链接或附件"
	case 2001:
		msgTypeDesc = "红包"
	case 2002:
		msgTypeDesc = "小程序"
	case 2003:
		msgTypeDesc = "群邀请"
	case 10000:
		msgTypeDesc = "系统消息"
	default:
		msgTypeDesc = fmt.Sprintf("未知")
	}

	g.Log().Infof(r.Context(),
		"接收账号: %s | 发送人: %v | 来源：%v | %v消息: %v",
		receiverWxid, fromWxid, sourceDesc, msgTypeDesc, msg)

	// 发送日志到前端
	logService := utils.GetGlobalLogService()
	if logService != nil {

		logService.SendLog(
			r.Context(),
			time.Now().Format("2006-01-02 15:04:05"),
			receiverWxid,
			sourceDesc,
			fmt.Sprintf("%s | 来自: %s | 消息: %s", msgTypeDesc, fromWxid, msg),
			"#3959CF",
		)
	}

	// 群聊消息特殊处理 暂时不需要
	//if fromType == 2 {
	//	if finalFromWxid, ok := data["finalFromWxid"].(string); ok {
	//		g.Log().Infof(r.Context(), "  群内发言人: %v", finalFromWxid)
	//	}
	//
	//	// 处理@消息
	//	if atList, ok := data["atWxidList"].([]interface{}); ok && len(atList) > 0 {
	//		g.Log().Infof(r.Context(), "  @了: %v", atList)
	//	}
	//}

}

// 处理转账事件
func (s *HttpCallbackService) handleTransPay(r *ghttp.Request, event CallbackEvent) {
	data := event.Data
	// 获取外层 wxid
	wxid := event.Wxid

	var fromWxid, money, memo string

	// 使用 gconv 进行类型转换
	msgSource := gconv.Int(data["msgSource"])

	if v, ok := data["fromWxid"].(string); ok {
		fromWxid = v
	}
	if v, ok := data["money"].(string); ok {
		money = v
	}
	if v, ok := data["memo"].(string); ok {
		memo = v
	}

	g.Log().Debugf(r.Context(), "转账事件 msgSource=%d", msgSource)

	var sourceDesc string
	switch msgSource {
	case 1:
		sourceDesc = "收到转账"
	case 2:
		sourceDesc = "对方接收转账"
	case 3:
		sourceDesc = "发出转账"
	case 4:
		sourceDesc = "自己接收转账"
	case 5:
		sourceDesc = "对方退还"
	case 6:
		sourceDesc = "自己退还"
	}

	g.Log().Infof(r.Context(),
		"[转账事件] 账号: %s | %s | 对方: %v | 金额: %v元 | 备注: %v",
		wxid, sourceDesc, fromWxid, money, memo,
	)

	// 发送日志到前端
	logService := utils.GetGlobalLogService()
	if logService != nil {
		// 根据转账类型设置颜色
		var color string
		g.Log().Debugf(r.Context(), "前端日志 msgSource=%d", msgSource)
		switch msgSource {
		case 1, 4: // 收到转账、自己接收转账
			color = "#3959CF"

		case 2, 5, 6: // 对方接收、退还
			color = "#909399"

		case 3: // 发出转账
			color = "#FA8C16"

		default:
			color = "#3959CF"

		}

		logService.SendLog(
			r.Context(),
			time.Now().Format("2006-01-02 15:04:05"),
			wxid,
			"转账",
			fmt.Sprintf("%s | 对方: %s | 金额: %s元 | 备注: %s", sourceDesc, fromWxid, money, memo),
			color,
		)
	}

	// TODO: 处理转账逻辑，如自动接收转账
}

// 处理好友请求事件
func (s *HttpCallbackService) handleFriendReq(r *ghttp.Request, event CallbackEvent) {
	data := event.Data
	// 获取外层 wxid（接收请求的账号）
	receiverWxid := event.Wxid

	var wxid, wxNum, nick, content, scene, v3, v4 string

	if v, ok := data["wxid"].(string); ok {
		wxid = v
	}
	if v, ok := data["wxNum"].(string); ok {
		wxNum = v
	}
	if v, ok := data["nick"].(string); ok {
		nick = v
	}
	if v, ok := data["content"].(string); ok {
		content = v
	}
	if v, ok := data["scene"].(string); ok {
		scene = v
	}
	if v, ok := data["v3"].(string); ok {
		v3 = v
	}
	if v, ok := data["v4"].(string); ok {
		v4 = v
	}

	var sceneDesc string
	switch scene {
	case "1":
		sceneDesc = "QQ"
	case "3":
		sceneDesc = "微信号"
	case "6":
		sceneDesc = "单向添加"
	case "10", "13":
		sceneDesc = "通讯录"
	case "14":
		sceneDesc = "群聊"
	case "15":
		sceneDesc = "手机号"
	case "17":
		sceneDesc = "名片"
	case "30":
		sceneDesc = "扫一扫"
	default:
		sceneDesc = "未知"
	}

	g.Log().Infof(r.Context(),
		"[好友请求] 接收账号: %s | 请求人 wxid: %v | 微信号: %v | 昵称: %v | 附言: %v | 来源: %s",
		receiverWxid, wxid, wxNum, nick, content, sceneDesc,
	)
	g.Log().Debugf(r.Context(), "  V3: %v", v3)
	g.Log().Debugf(r.Context(), "  V4: %v", v4)

	// 发送日志到前端
	logService := utils.GetGlobalLogService()
	if logService != nil {
		logService.SendLog(
			r.Context(),
			time.Now().Format("2006-01-02 15:04:05"),
			receiverWxid,
			"好友请求",
			fmt.Sprintf("好友请求| 来自: %s | 昵称: %s | 附言: %s | 来源: %s", wxid, nick, content, sceneDesc),
			"#FA8C16",
		)
	}

}

// 处理群成员变动事件
func (s *HttpCallbackService) handleGroupMemberChanges(r *ghttp.Request, event CallbackEvent) {

	data := event.Data
	// 获取外层 wxid
	wxid := event.Wxid

	var fromWxid, finalFromWxid, inviterWxid string

	// 使用 gconv 进行类型转换
	eventType := gconv.Int(data["eventType"])

	if v, ok := data["fromWxid"].(string); ok {
		fromWxid = v
	}
	if v, ok := data["finalFromWxid"].(string); ok {
		finalFromWxid = v
	}
	if v, ok := data["inviterWxid"].(string); ok {
		inviterWxid = v
	}

	g.Log().Debugf(r.Context(), "群成员变动: eventType=%d, fromWxid=%s, finalFromWxid=%s, inviterWxid=%s",
		eventType, fromWxid, finalFromWxid, inviterWxid)

	var eventDesc string
	if eventType == 0 {
		eventDesc = "退群"
	} else {
		eventDesc = "进群"
	}

	if inviterWxid != "" {
		g.Log().Infof(r.Context(),
			"[群成员变动] 账号: %s | %s | 群: %v | 成员: %v | 邀请人: %v",
			wxid, eventDesc, fromWxid, finalFromWxid, inviterWxid,
		)
	} else {
		g.Log().Infof(r.Context(),
			"[群成员变动] 账号: %s | %s | 群: %v | 成员: %v",
			wxid, eventDesc, fromWxid, finalFromWxid,
		)
	}

	// 发送日志到前端
	logService := utils.GetGlobalLogService()
	if logService != nil {
		var color string
		if eventType == 0 {
			color = "#909399" // 退群 - 灰色
		} else {
			color = "#52C41A" // 进群 - 绿色
		}

		var msg string
		if inviterWxid != "" {
			msg = fmt.Sprintf("%s | 群: %s | 成员: %s | 邀请人: %s", eventDesc, fromWxid, finalFromWxid, inviterWxid)
		} else {
			msg = fmt.Sprintf("%s | 群: %s | 成员: %s", eventDesc, fromWxid, finalFromWxid)
		}

		logService.SendLog(
			r.Context(),
			time.Now().Format("2006-01-02 15:04:05"),
			wxid,
			"群变动",
			msg,
			color,
		)
	}

}

// 处理授权到期事件
func (s *HttpCallbackService) handleAuthExpire(r *ghttp.Request, event CallbackEvent) {
	data := event.Data
	// 获取外层 wxid
	wxid := event.Wxid

	var expireTime, msg string

	if v, ok := data["expireTime"].(string); ok {
		expireTime = v
	}
	if v, ok := data["msg"].(string); ok {
		msg = v
	}

	g.Log().Warningf(r.Context(),
		"[授权到期] wxid: %v, 到期时间: %v, 消息: %v",
		wxid, expireTime, msg,
	)
	// 获取全局日志服务
	logService := utils.GetGlobalLogService()
	if logService != nil {
		logService.SendLog(
			r.Context(),
			time.Now().Format("2006-01-02 15:04:05"),
			wxid,
			"授权",
			fmt.Sprintf("授权已到期 | 到期时间: %s", expireTime),
			"#F5222D",
		)
	}
}

// WechatAccount 微信账号信息
type WechatAccount struct {
	Wxid       string `json:"wxid"`
	WxNum      string `json:"wxNum"`
	Nick       string `json:"nick"`
	AvatarUrl  string `json:"avatarUrl"`
	Port       int    `json:"port"`
	Pid        int    `json:"pid"`
	ExpireTime string `json:"expireTime,omitempty"` // 授权到期时间
	IsExpire   int    `json:"isExpire"`             // 是否已到期（1=是，0=否）
}

// WechatAccountList 微信账号列表
type WechatAccountList struct {
	List []WechatAccount `json:"list"`
}

var (
	wechatAccountFile = "resources/currentWechat.json"
	fileMutex         sync.Mutex
)

// saveWechatAccount 保存微信账号信息到 currentWechat.json
func (s *HttpCallbackService) saveWechatAccount(ctx context.Context, data map[string]interface{}) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	// 提取账号信息
	var account WechatAccount

	if v, ok := data["wxid"].(string); ok {
		account.Wxid = v
	}
	if v, ok := data["wxNum"].(string); ok {
		account.WxNum = v
	}
	if v, ok := data["nick"].(string); ok {
		account.Nick = v
	}
	if v, ok := data["avatarUrl"].(string); ok {
		account.AvatarUrl = v
	}
	if v, ok := data["port"].(float64); ok {
		account.Port = int(v)
	}
	if v, ok := data["pid"].(float64); ok {
		account.Pid = int(v)
	}

	// wxid 为必须字段
	if account.Wxid == "" {
		return fmt.Errorf("wxid 为空")
	}

	// 读取现有数据
	var accountList WechatAccountList

	// 确保 resource 目录存在
	resourceDir := filepath.Dir(wechatAccountFile)
	if !gfile.Exists(resourceDir) {
		if err := gfile.Mkdir(resourceDir); err != nil {
			return fmt.Errorf("创建 resource 目录失败: %v", err)
		}
	}

	// 如果文件存在，读取现有数据
	if gfile.Exists(wechatAccountFile) {
		fileData, err := os.ReadFile(wechatAccountFile)
		if err != nil {
			return fmt.Errorf("读取文件失败: %v", err)
		}

		if len(fileData) > 0 {
			if err := json.Unmarshal(fileData, &accountList); err != nil {
				return fmt.Errorf("解析 JSON 失败: %v", err)
			}
		}
	}

	// 检查是否已存在该 wxid
	found := false
	for i, acc := range accountList.List {
		if acc.Wxid == account.Wxid {
			// 更新现有记录
			accountList.List[i] = account
			found = true
			g.Log().Debugf(ctx, "更新微信账号: %s", account.Wxid)
			break
		}
	}

	// 如果不存在，追加新记录
	if !found {
		accountList.List = append(accountList.List, account)
		g.Log().Infof(ctx, "新增微信账号: %s", account.Wxid)
	}

	// 序列化为 JSON
	jsonData, err := json.MarshalIndent(accountList, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(wechatAccountFile, jsonData, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	g.Log().Debugf(ctx, "微信账号信息已保存到: %s", wechatAccountFile)
	return nil
}

// CheckAndUpdateAuthInfo 检查并更新所有微信账号的授权信息
func (s *HttpCallbackService) CheckAndUpdateAuthInfo(ctx context.Context) {
	g.Log().Info(ctx, "开始检查微信账号授权信息...")

	// 读取 currentWechat.json
	if !gfile.Exists(wechatAccountFile) {
		g.Log().Debug(ctx, "currentWechat.json 文件不存在，跳过检查")
		return
	}

	fileData, err := os.ReadFile(wechatAccountFile)
	if err != nil {
		g.Log().Warningf(ctx, "读取 currentWechat.json 失败: %v", err)
		return
	}

	var accountList WechatAccountList
	if len(fileData) > 0 {
		if err := json.Unmarshal(fileData, &accountList); err != nil {
			g.Log().Warningf(ctx, "解析 currentWechat.json 失败: %v", err)
			return
		}
	}

	if len(accountList.List) == 0 {
		g.Log().Debug(ctx, "没有微信账号需要检查")
		return
	}

	g.Log().Infof(ctx, "发现 %d 个微信账号，开始并发检查授权信息...", len(accountList.List))

	// 并发检查每个账号
	var wg sync.WaitGroup
	var mu sync.Mutex
	updatedAccounts := make([]WechatAccount, 0)

	for _, account := range accountList.List {
		wg.Add(1)
		go func(acc WechatAccount) {
			defer wg.Done()

			// 查询授权信息
			if authInfo, ok := s.queryAuthInfo(ctx, acc.Port); ok {
				// 更新授权信息
				acc.ExpireTime = authInfo.ExpireTime
				acc.IsExpire = authInfo.IsExpire

				mu.Lock()
				updatedAccounts = append(updatedAccounts, acc)
				mu.Unlock()

				g.Log().Infof(ctx, "✓ %s (端口:%d) 授权信息更新成功，到期时间: %s", acc.Wxid, acc.Port, authInfo.ExpireTime)
			} else {
				g.Log().Warningf(ctx, "✗ %s (端口:%d) 请求失败，已从列表中移除", acc.Wxid, acc.Port)
			}
		}(account)
	}

	wg.Wait()

	// 保存更新后的列表
	fileMutex.Lock()
	defer fileMutex.Unlock()

	accountList.List = updatedAccounts
	jsonData, err := json.MarshalIndent(accountList, "", "  ")
	if err != nil {
		g.Log().Errorf(ctx, "序列化 JSON 失败: %v", err)
		return
	}

	if err := os.WriteFile(wechatAccountFile, jsonData, 0644); err != nil {
		g.Log().Errorf(ctx, "写入文件失败: %v", err)
		return
	}

	g.Log().Infof(ctx, "授权信息检查完成，剩余 %d 个有效账号", len(updatedAccounts))
}

// AuthInfoResult 授权信息结果
type AuthInfoResult struct {
	ExpireTime string `json:"expireTime"`
	IsExpire   int    `json:"isExpire"`
}

// AuthInfoResponse 授权信息响应
type AuthInfoResponse struct {
	Code   int            `json:"code"`
	Msg    string         `json:"msg"`
	Result AuthInfoResult `json:"result"`
}

// queryAuthInfo 查询单个微信的授权信息
func (s *HttpCallbackService) queryAuthInfo(ctx context.Context, port int) (*AuthInfoResult, bool) {
	url := fmt.Sprintf("http://127.0.0.1:%d/wechat/httpapi", port)

	// 构造请求体
	requestBody := map[string]string{
		"type": "getAuthInfo",
	}
	jsonData, _ := json.Marshal(requestBody)

	// 创建 HTTP 请求
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		g.Log().Debugf(ctx, "创建请求失败 (port:%d): %v", port, err)
		return nil, false
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		g.Log().Debugf(ctx, "请求失败 (port:%d): %v", port, err)
		return nil, false
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		g.Log().Debugf(ctx, "读取响应失败 (port:%d): %v", port, err)
		return nil, false
	}

	// 解析响应
	var authResp AuthInfoResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		g.Log().Debugf(ctx, "解析响应失败 (port:%d): %v", port, err)
		return nil, false
	}

	if authResp.Code != 200 {
		g.Log().Debugf(ctx, "请求失败 (port:%d): code=%d, msg=%s", port, authResp.Code, authResp.Msg)
		return nil, false
	}

	return &authResp.Result, true
}

// ============ 插件 HTTP API ============

type PluginAPIService struct{}

// GetConfig 获取 config.yaml
func (s *PluginAPIService) GetConfig(r *ghttp.Request) {
	configPath := "configs/config.yaml"
	if !gfile.Exists(configPath) {
		r.Response.WriteJsonExit(g.Map{
			"code": 404,
			"msg":  "配置文件不存在",
		})
		return
	}

	content := gfile.GetContents(configPath)
	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"data": content,
	})
}

// GetCurrentWechat 获取 currentWechat.json
func (s *PluginAPIService) GetCurrentWechat(r *ghttp.Request) {
	currentWechatPath := "resources/currentWechat.json"
	if !gfile.Exists(currentWechatPath) {
		r.Response.WriteJsonExit(g.Map{
			"code": 404,
			"msg":  "当前微信账号文件不存在",
		})
		return
	}

	content := gfile.GetContents(currentWechatPath)
	// 尝试解析为 JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(content), &jsonData); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "JSON 解析失败",
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"data": jsonData,
	})
}

// SendLog 插件发送日志
func (s *PluginAPIService) SendLog(r *ghttp.Request) {
	var req struct {
		PluginID  string `json:"pluginId"`
		TimeStamp string `json:"timeStamp"`
		Response  string `json:"response"`
		LogType   string `json:"logType"`
		Msg       string `json:"msg"`
		Color     string `json:"color"`
	}

	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 400,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	// 调用插件服务发送日志
	g.Log().Debugf(r.Context(), "pluginServiceInstance 是否为 nil: %v", pluginServiceInstance == nil)

	if pluginServiceInstance == nil {
		g.Log().Error(r.Context(), "插件服务实例为 nil")
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "插件服务未初始化（实例为 nil）",
		})
		return
	}

	type PluginLogSender interface {
		SendPluginLog(ctx context.Context, pluginID, timeStamp, response, logType, msg, color string) error
	}

	ps, ok := pluginServiceInstance.(PluginLogSender)
	if !ok {
		g.Log().Errorf(r.Context(), "插件服务类型断言失败，实际类型: %T", pluginServiceInstance)
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "插件服务类型不匹配",
		})
		return
	}

	if err := ps.SendPluginLog(r.Context(), req.PluginID, req.TimeStamp, req.Response, req.LogType, req.Msg, req.Color); err != nil {
		g.Log().Errorf(r.Context(), "调用 SendPluginLog 失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "发送日志失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "日志发送成功",
	})
}

// UploadPlugin 上传插件文件
func (s *PluginAPIService) UploadPlugin(r *ghttp.Request) {
	g.Log().Info(r.Context(), "开始处理插件上传请求")

	// 获取上传的文件
	file := r.GetUploadFile("file")
	if file == nil {
		g.Log().Error(r.Context(), "未找到上传文件")
		r.Response.WriteJsonExit(g.Map{
			"code": 400,
			"msg":  "未找到上传文件",
		})
		return
	}

	g.Log().Infof(r.Context(), "收到上传文件: %s, 大小: %d bytes (%.2f MB)",
		file.Filename, file.Size, float64(file.Size)/1024/1024)

	// 检查文件后缀
	if filepath.Ext(file.Filename) != ".dog" {
		g.Log().Warningf(r.Context(), "文件格式错误: %s", file.Filename)
		r.Response.WriteJsonExit(g.Map{
			"code": 400,
			"msg":  "只支持 .dog 格式的插件文件",
		})
		return
	}

	// 确保 plugins 目录存在
	pluginsDir := "plugins"
	if !gfile.Exists(pluginsDir) {
		if err := gfile.Mkdir(pluginsDir); err != nil {
			g.Log().Errorf(r.Context(), "创建 plugins 目录失败: %v", err)
			r.Response.WriteJsonExit(g.Map{
				"code": 500,
				"msg":  "创建目录失败",
			})
			return
		}
	}

	// 保存到 plugins 目录（精确文件路径）
	pluginPath := filepath.Join(pluginsDir, file.Filename)
	g.Log().Infof(r.Context(), "开始保存文件到: %s", pluginPath)

	src, err := file.Open()
	if err != nil {
		g.Log().Errorf(r.Context(), "打开上传文件失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "读取上传文件失败",
		})
		return
	}
	defer src.Close()

	dst, err := os.Create(pluginPath)
	if err != nil {
		g.Log().Errorf(r.Context(), "创建目标文件失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "保存文件失败: " + err.Error(),
		})
		return
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		g.Log().Errorf(r.Context(), "写入插件文件失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "保存文件失败: " + err.Error(),
		})
		return
	}
	dst.Close()

	g.Log().Infof(r.Context(), "插件文件上传成功: %s (%.2f MB)",
		file.Filename, float64(file.Size)/1024/1024)

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "上传成功",
		"data": g.Map{
			"filename": file.Filename,
			"size":     file.Size,
		},
	})
}

// EventStream SSE 事件流接口
func (s *PluginAPIService) EventStream(r *ghttp.Request) {
	// 设置 SSE 响应头
	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	r.Response.Header().Set("Access-Control-Allow-Origin", "*")

	// 创建客户端通道
	clientChan := make(chan string, 10)

	// 注册客户端
	sseClientsMutex.Lock()
	sseClients[clientChan] = true
	sseClientsMutex.Unlock()

	g.Log().Infof(r.Context(), "SSE 客户端已连接，当前总数: %d", len(sseClients))

	// 发送连接成功消息
	r.Response.Write("data: {\"type\":\"connected\",\"msg\":\"连接成功\"}\n\n")
	r.Response.Flush()

	// 监听客户端断开
	notify := r.Context().Done()

	// 保持连接
	for {
		select {
		case <-notify:
			// 客户端断开
			sseClientsMutex.Lock()
			delete(sseClients, clientChan)
			sseClientsMutex.Unlock()
			close(clientChan)
			g.Log().Infof(r.Context(), "SSE 客户端已断开，当前总数: %d", len(sseClients))
			return

		case msg := <-clientChan:
			// 发送消息给客户端
			r.Response.Write(msg)
			r.Response.Flush()

		case <-time.After(30 * time.Second):
			// 心跳，保持连接
			r.Response.Write(": heartbeat\n\n")
			r.Response.Flush()
		}
	}
}
