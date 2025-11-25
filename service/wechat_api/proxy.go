package wechat_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
)

// WechatAPIProxyService 微信API代理服务
type WechatAPIProxyService struct{}

// callWechatAPI 调用微信服务API的通用函数
func (s *WechatAPIProxyService) callWechatAPI(r *ghttp.Request, apiType string) {
	// 从 URL 参数获取端口号
	port := r.Get("port").Int()
	if port == 0 {
		r.Response.WriteJson(map[string]interface{}{
			"code": 400,
			"msg":  "缺少 port 参数",
		})
		return
	}

	// 读取原始请求体（不解析，直接转发）
	bodyBytes := r.GetBody()

	// 解析为 map 以便构建新的请求体
	var requestData map[string]interface{}
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
			r.Response.WriteJson(map[string]interface{}{
				"code": 400,
				"msg":  "参数解析失败: " + err.Error(),
			})
			return
		}
	} else {
		requestData = make(map[string]interface{})
	}

	// 构建微信API请求体
	wechatRequest := map[string]interface{}{
		"type": apiType,
		"data": requestData,
	}

	// 构建目标URL（统一的微信API路径）
	targetURL := fmt.Sprintf("http://127.0.0.1:%d/wechat/httpapi", port)

	// 将请求数据转为JSON
	jsonData, err := json.Marshal(wechatRequest)
	if err != nil {
		r.Response.WriteJson(map[string]interface{}{
			"code": 500,
			"msg":  "JSON序列化失败: " + err.Error(),
		})
		return
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 发送请求到微信服务
	resp, err := client.Post(targetURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		r.Response.WriteJson(map[string]interface{}{
			"code": 500,
			"msg":  "请求微信服务失败: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.Response.WriteJson(map[string]interface{}{
			"code": 500,
			"msg":  "读取响应失败: " + err.Error(),
		})
		return
	}

	// 原封不动返回微信API的响应
	r.Response.Header().Set("Content-Type", "application/json")
	r.Response.Write(body)
}

// ==================== 基础接口 ====================

// EditVersion 修改微信版本号
func (s *WechatAPIProxyService) EditVersion(r *ghttp.Request) {
	s.callWechatAPI(r, "editVersion")
}

// GetLoginStatus 获取登录状态
func (s *WechatAPIProxyService) GetLoginStatus(r *ghttp.Request) {
	s.callWechatAPI(r, "getLoginStatus")
}

// GetLoginQRCode 获取登录二维码
func (s *WechatAPIProxyService) GetLoginQRCode(r *ghttp.Request) {
	s.callWechatAPI(r, "getLoginQRCode")
}

// SetDownloadImage 设置下载图片时间
func (s *WechatAPIProxyService) SetDownloadImage(r *ghttp.Request) {
	s.callWechatAPI(r, "setDownloadImage")
}

// DecryptImage 解密dat图片
func (s *WechatAPIProxyService) DecryptImage(r *ghttp.Request) {
	s.callWechatAPI(r, "decryptImage")
}

// CheckWeChat 微信状态检测
func (s *WechatAPIProxyService) CheckWeChat(r *ghttp.Request) {
	s.callWechatAPI(r, "checkWeChat")
}

// GetAuthInfo 查询授权信息
func (s *WechatAPIProxyService) GetAuthInfo(r *ghttp.Request) {
	s.callWechatAPI(r, "getAuthInfo")
}

// ==================== 消息发送接口 ====================

// SendText 发送文本消息
func (s *WechatAPIProxyService) SendText(r *ghttp.Request) {
	s.callWechatAPI(r, "sendText")
}

// SendText2 发送文本消息2
func (s *WechatAPIProxyService) SendText2(r *ghttp.Request) {
	s.callWechatAPI(r, "sendText2")
}

// SendReferText 发送引用回复文本
func (s *WechatAPIProxyService) SendReferText(r *ghttp.Request) {
	s.callWechatAPI(r, "sendReferText")
}

// SendImage 发送图片
func (s *WechatAPIProxyService) SendImage(r *ghttp.Request) {
	s.callWechatAPI(r, "sendImage")
}

// SendFile 发送文件
func (s *WechatAPIProxyService) SendFile(r *ghttp.Request) {
	s.callWechatAPI(r, "sendFile")
}

// SendGif 发送动态表情
func (s *WechatAPIProxyService) SendGif(r *ghttp.Request) {
	s.callWechatAPI(r, "sendGif")
}

// SendShareUrl 发送分享链接
func (s *WechatAPIProxyService) SendShareUrl(r *ghttp.Request) {
	s.callWechatAPI(r, "sendShareUrl")
}

// SendApplet 发送小程序
func (s *WechatAPIProxyService) SendApplet(r *ghttp.Request) {
	s.callWechatAPI(r, "sendApplet")
}

// SendMusic 发送音乐分享
func (s *WechatAPIProxyService) SendMusic(r *ghttp.Request) {
	s.callWechatAPI(r, "sendMusic")
}

// SendChatLog 发送聊天记录
func (s *WechatAPIProxyService) SendChatLog(r *ghttp.Request) {
	s.callWechatAPI(r, "sendChatLog")
}

// SendCard 发送名片消息
func (s *WechatAPIProxyService) SendCard(r *ghttp.Request) {
	s.callWechatAPI(r, "sendCard")
}

// SendXml 发送XML
func (s *WechatAPIProxyService) SendXml(r *ghttp.Request) {
	s.callWechatAPI(r, "sendXml")
}

// SendLocationInfo 发送位置信息
func (s *WechatAPIProxyService) SendLocationInfo(r *ghttp.Request) {
	s.callWechatAPI(r, "sendLocationInfo")
}

// ==================== 信息获取接口 ====================

// GetSelfInfo 获取个人信息
func (s *WechatAPIProxyService) GetSelfInfo(r *ghttp.Request) {
	s.callWechatAPI(r, "getSelfInfo")
}

// GetLabelList 获取标签列表
func (s *WechatAPIProxyService) GetLabelList(r *ghttp.Request) {
	s.callWechatAPI(r, "getLabelList")
}

// GetFriendList 获取好友列表
func (s *WechatAPIProxyService) GetFriendList(r *ghttp.Request) {
	s.callWechatAPI(r, "getFriendList")
}

// GetGroupList 获取群聊列表
func (s *WechatAPIProxyService) GetGroupList(r *ghttp.Request) {
	s.callWechatAPI(r, "getGroupList")
}

// GetPublicList 获取公众号列表
func (s *WechatAPIProxyService) GetPublicList(r *ghttp.Request) {
	s.callWechatAPI(r, "getPublicList")
}

// ==================== 好友管理接口 ====================

// AgreeFriendReq 同意好友请求
func (s *WechatAPIProxyService) AgreeFriendReq(r *ghttp.Request) {
	s.callWechatAPI(r, "agreeFriendReq")
}

// AddFriendByV3 添加好友_通过v3
func (s *WechatAPIProxyService) AddFriendByV3(r *ghttp.Request) {
	s.callWechatAPI(r, "addFriendByV3")
}

// AddFriendByGroupWxid 添加好友_通过群wxid
func (s *WechatAPIProxyService) AddFriendByGroupWxid(r *ghttp.Request) {
	s.callWechatAPI(r, "addFriendByGroupWxid")
}

// DelFriend 删除好友
func (s *WechatAPIProxyService) DelFriend(r *ghttp.Request) {
	s.callWechatAPI(r, "delFriend")
}

// EditObjRemark 修改对象备注
func (s *WechatAPIProxyService) EditObjRemark(r *ghttp.Request) {
	s.callWechatAPI(r, "editObjRemark")
}

// QueryNewFriend 查询陈生人信息
func (s *WechatAPIProxyService) QueryNewFriend(r *ghttp.Request) {
	s.callWechatAPI(r, "queryNewFriend")
}

// QueryObj 查询对象信息
func (s *WechatAPIProxyService) QueryObj(r *ghttp.Request) {
	s.callWechatAPI(r, "queryObj")
}

// ==================== 群聊管理接口 ====================

// QuitGroup 退出群聊
func (s *WechatAPIProxyService) QuitGroup(r *ghttp.Request) {
	s.callWechatAPI(r, "quitGroup")
}

// CreateGroup 创建群聊
func (s *WechatAPIProxyService) CreateGroup(r *ghttp.Request) {
	s.callWechatAPI(r, "createGroup")
}

// QueryGroup 查询群聊信息
func (s *WechatAPIProxyService) QueryGroup(r *ghttp.Request) {
	s.callWechatAPI(r, "queryGroup")
}

// AddMembers 添加群成员
func (s *WechatAPIProxyService) AddMembers(r *ghttp.Request) {
	s.callWechatAPI(r, "addMembers")
}

// InviteMembers 邀请群成员
func (s *WechatAPIProxyService) InviteMembers(r *ghttp.Request) {
	s.callWechatAPI(r, "inviteMembers")
}

// DelMembers 移除群成员
func (s *WechatAPIProxyService) DelMembers(r *ghttp.Request) {
	s.callWechatAPI(r, "delMembers")
}

// GetMemberList 获取群成员列表
func (s *WechatAPIProxyService) GetMemberList(r *ghttp.Request) {
	s.callWechatAPI(r, "getMemberList")
}

// GetMemberNick 获取群成员昵称
func (s *WechatAPIProxyService) GetMemberNick(r *ghttp.Request) {
	s.callWechatAPI(r, "getMemberNick")
}

// EditSelfMemberNick 修改自己群昵称
func (s *WechatAPIProxyService) EditSelfMemberNick(r *ghttp.Request) {
	s.callWechatAPI(r, "editSelfMemberNick")
}

// ==================== 转账与其他接口 ====================

// ConfirmTrans 确认收款
func (s *WechatAPIProxyService) ConfirmTrans(r *ghttp.Request) {
	s.callWechatAPI(r, "confirmTrans")
}

// ReturnTrans 退还收款
func (s *WechatAPIProxyService) ReturnTrans(r *ghttp.Request) {
	s.callWechatAPI(r, "returnTrans")
}

// OpenBrowser 打开浏览器
func (s *WechatAPIProxyService) OpenBrowser(r *ghttp.Request) {
	s.callWechatAPI(r, "openBrowser")
}

// RunCloudFunction 执行云函数
func (s *WechatAPIProxyService) RunCloudFunction(r *ghttp.Request) {
	s.callWechatAPI(r, "runCloudFunction")
}

// SetReadStatus 标记已读未读
func (s *WechatAPIProxyService) SetReadStatus(r *ghttp.Request) {
	s.callWechatAPI(r, "setReadStatus")
}

// AuthCami 使用授权卡密
func (s *WechatAPIProxyService) AuthCami(r *ghttp.Request) {
	s.callWechatAPI(r, "authCami")
}

// RevokeMyMsg 撤回消息
func (s *WechatAPIProxyService) RevokeMyMsg(r *ghttp.Request) {
	s.callWechatAPI(r, "revokeMyMsg")
}
