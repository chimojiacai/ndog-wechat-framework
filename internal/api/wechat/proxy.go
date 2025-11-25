package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
)

// Proxy 微信API代理服务
type Proxy struct{}

// NewProxy 创建微信API代理实例
func NewProxy() *Proxy {
	return &Proxy{}
}

// callWechatAPI 调用微信服务API的通用函数
func (p *Proxy) callWechatAPI(r *ghttp.Request, apiType string) {
	// 从 URL 参数获取端口号
	port := r.Get("port").Int()
	if port == 0 {
		r.Response.WriteJson(map[string]interface{}{
			"code": 400,
			"msg":  "缺少 port 参数",
		})
		return
	}

	// 构建目标 URL
	url := fmt.Sprintf("http://127.0.0.1:%d/wechat/httpapi", port)

	// 获取请求体
	var requestBody map[string]interface{}
	if err := r.Parse(&requestBody); err != nil {
		r.Response.WriteJson(map[string]interface{}{
			"code": 400,
			"msg":  "请求参数解析失败",
		})
		return
	}

	// 添加 type 字段
	requestBody["type"] = apiType

	// 序列化请求体
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		r.Response.WriteJson(map[string]interface{}{
			"code": 500,
			"msg":  "请求序列化失败",
		})
		return
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		r.Response.WriteJson(map[string]interface{}{
			"code": 500,
			"msg":  "创建请求失败",
		})
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		r.Response.WriteJson(map[string]interface{}{
			"code": 500,
			"msg":  fmt.Sprintf("请求失败: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.Response.WriteJson(map[string]interface{}{
			"code": 500,
			"msg":  "读取响应失败",
		})
		return
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		r.Response.WriteJson(map[string]interface{}{
			"code": 500,
			"msg":  "响应解析失败",
		})
		return
	}

	// 返回结果
	r.Response.WriteJson(result)
}

// 基础接口 (7个)

// ChangeVersion 修改版本号
func (p *Proxy) ChangeVersion(r *ghttp.Request) {
	p.callWechatAPI(r, "changeVersion")
}

// GetLoginStatus 获取登录状态
func (p *Proxy) GetLoginStatus(r *ghttp.Request) {
	p.callWechatAPI(r, "getLoginStatus")
}

// GetLoginQrCode 获取登录二维码
func (p *Proxy) GetLoginQrCode(r *ghttp.Request) {
	p.callWechatAPI(r, "getLoginQrCode")
}

// GetSelfInfo 获取个人信息
func (p *Proxy) GetSelfInfo(r *ghttp.Request) {
	p.callWechatAPI(r, "getSelfInfo")
}

// GetAuthInfo 获取授权信息
func (p *Proxy) GetAuthInfo(r *ghttp.Request) {
	p.callWechatAPI(r, "getAuthInfo")
}

// Logout 退出登录
func (p *Proxy) Logout(r *ghttp.Request) {
	p.callWechatAPI(r, "logout")
}

// GetWechatVer 获取微信版本
func (p *Proxy) GetWechatVer(r *ghttp.Request) {
	p.callWechatAPI(r, "getWechatVer")
}

// 消息发送接口 (13个)

// SendText 发送文本消息
func (p *Proxy) SendText(r *ghttp.Request) {
	p.callWechatAPI(r, "sendText")
}

// SendImage 发送图片
func (p *Proxy) SendImage(r *ghttp.Request) {
	p.callWechatAPI(r, "sendImage")
}

// SendFile 发送文件
func (p *Proxy) SendFile(r *ghttp.Request) {
	p.callWechatAPI(r, "sendFile")
}

// SendVideo 发送视频
func (p *Proxy) SendVideo(r *ghttp.Request) {
	p.callWechatAPI(r, "sendVideo")
}

// SendEmoji 发送表情
func (p *Proxy) SendEmoji(r *ghttp.Request) {
	p.callWechatAPI(r, "sendEmoji")
}

// SendCard 发送名片
func (p *Proxy) SendCard(r *ghttp.Request) {
	p.callWechatAPI(r, "sendCard")
}

// SendLink 发送链接
func (p *Proxy) SendLink(r *ghttp.Request) {
	p.callWechatAPI(r, "sendLink")
}

// SendMiniProgram 发送小程序
func (p *Proxy) SendMiniProgram(r *ghttp.Request) {
	p.callWechatAPI(r, "sendMiniProgram")
}

// SendMusic 发送音乐
func (p *Proxy) SendMusic(r *ghttp.Request) {
	p.callWechatAPI(r, "sendMusic")
}

// SendLocation 发送位置
func (p *Proxy) SendLocation(r *ghttp.Request) {
	p.callWechatAPI(r, "sendLocation")
}

// ForwardMsg 转发消息
func (p *Proxy) ForwardMsg(r *ghttp.Request) {
	p.callWechatAPI(r, "forwardMsg")
}

// SendAtText 发送@消息
func (p *Proxy) SendAtText(r *ghttp.Request) {
	p.callWechatAPI(r, "sendAtText")
}

// RevokeMsg 撤回消息
func (p *Proxy) RevokeMsg(r *ghttp.Request) {
	p.callWechatAPI(r, "revokeMsg")
}

// 信息获取接口 (5个)

// GetFriendList 获取好友列表
func (p *Proxy) GetFriendList(r *ghttp.Request) {
	p.callWechatAPI(r, "getFriendList")
}

// GetGroupList 获取群聊列表
func (p *Proxy) GetGroupList(r *ghttp.Request) {
	p.callWechatAPI(r, "getGroupList")
}

// GetGroupMembers 获取群成员列表
func (p *Proxy) GetGroupMembers(r *ghttp.Request) {
	p.callWechatAPI(r, "getGroupMembers")
}

// GetContactProfile 获取联系人详细信息
func (p *Proxy) GetContactProfile(r *ghttp.Request) {
	p.callWechatAPI(r, "getContactProfile")
}

// GetDbNames 获取数据库名称
func (p *Proxy) GetDbNames(r *ghttp.Request) {
	p.callWechatAPI(r, "getDbNames")
}

// 好友管理接口 (7个)

// AddFriend 添加好友
func (p *Proxy) AddFriend(r *ghttp.Request) {
	p.callWechatAPI(r, "addFriend")
}

// AcceptFriend 同意好友请求
func (p *Proxy) AcceptFriend(r *ghttp.Request) {
	p.callWechatAPI(r, "acceptFriend")
}

// DeleteFriend 删除好友
func (p *Proxy) DeleteFriend(r *ghttp.Request) {
	p.callWechatAPI(r, "deleteFriend")
}

// SetRemark 设置备注
func (p *Proxy) SetRemark(r *ghttp.Request) {
	p.callWechatAPI(r, "setRemark")
}

// TopContact 置顶联系人
func (p *Proxy) TopContact(r *ghttp.Request) {
	p.callWechatAPI(r, "topContact")
}

// SetBlacklist 设置黑名单
func (p *Proxy) SetBlacklist(r *ghttp.Request) {
	p.callWechatAPI(r, "setBlacklist")
}

// SearchFriend 搜索好友
func (p *Proxy) SearchFriend(r *ghttp.Request) {
	p.callWechatAPI(r, "searchFriend")
}

// 群聊管理接口 (9个)

// CreateGroup 创建群聊
func (p *Proxy) CreateGroup(r *ghttp.Request) {
	p.callWechatAPI(r, "createGroup")
}

// AddGroupMember 添加群成员
func (p *Proxy) AddGroupMember(r *ghttp.Request) {
	p.callWechatAPI(r, "addGroupMember")
}

// DeleteGroupMember 删除群成员
func (p *Proxy) DeleteGroupMember(r *ghttp.Request) {
	p.callWechatAPI(r, "deleteGroupMember")
}

// QuitGroup 退出群聊
func (p *Proxy) QuitGroup(r *ghttp.Request) {
	p.callWechatAPI(r, "quitGroup")
}

// ModifyGroupName 修改群名称
func (p *Proxy) ModifyGroupName(r *ghttp.Request) {
	p.callWechatAPI(r, "modifyGroupName")
}

// ModifyGroupNotice 修改群公告
func (p *Proxy) ModifyGroupNotice(r *ghttp.Request) {
	p.callWechatAPI(r, "modifyGroupNotice")
}

// ModifyNickInGroup 修改群昵称
func (p *Proxy) ModifyNickInGroup(r *ghttp.Request) {
	p.callWechatAPI(r, "modifyNickInGroup")
}

// InviteIntoGroup 邀请进群
func (p *Proxy) InviteIntoGroup(r *ghttp.Request) {
	p.callWechatAPI(r, "inviteIntoGroup")
}

// GetGroupQrCode 获取群二维码
func (p *Proxy) GetGroupQrCode(r *ghttp.Request) {
	p.callWechatAPI(r, "getGroupQrCode")
}

// 其他接口 (6个)

// ReceiveTransfer 接收转账
func (p *Proxy) ReceiveTransfer(r *ghttp.Request) {
	p.callWechatAPI(r, "receiveTransfer")
}

// OpenBrowser 打开浏览器
func (p *Proxy) OpenBrowser(r *ghttp.Request) {
	p.callWechatAPI(r, "openBrowser")
}

// DownloadImage 下载图片
func (p *Proxy) DownloadImage(r *ghttp.Request) {
	p.callWechatAPI(r, "downloadImage")
}

// DownloadFile 下载文件
func (p *Proxy) DownloadFile(r *ghttp.Request) {
	p.callWechatAPI(r, "downloadFile")
}

// ExecuteSql 执行SQL
func (p *Proxy) ExecuteSql(r *ghttp.Request) {
	p.callWechatAPI(r, "executeSql")
}

// CallCloudFunc 调用云函数
func (p *Proxy) CallCloudFunc(r *ghttp.Request) {
	p.callWechatAPI(r, "callCloudFunc")
}
