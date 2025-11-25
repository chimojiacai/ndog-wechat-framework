// service/http_callback/server.go
package http_callback

import (
	"context"
	"fmt"

	"github.com/naidog/wechat-framework/service/wechat_api"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
)

type HttpServerService struct {
	server *ghttp.Server
}

// StartServer 启动HTTP服务
func (s *HttpServerService) StartServer() error {
	ctx := gctx.New()

	// 读取配置
	address, err := g.Cfg().Get(ctx, "server.address")
	if err != nil || address.String() == "" {
		return fmt.Errorf("获取服务地址配置失败")
	}

	// 创建HTTP服务器
	s.server = g.Server()
	s.server.SetAddr(address.String())

	// 设置最大上传文件大小为 100MB
	s.server.SetClientMaxBodySize(100 * 1024 * 1024)

	// 配置 CORS 跨域支持
	s.server.Use(ghttp.MiddlewareCORS)

	// 注册回调路由
	callbackService := &HttpCallbackService{}
	s.server.BindHandler("/wechat/callback", callbackService.HandleCallback)

	// 注册插件 API 路由
	pluginAPIService := &PluginAPIService{}
	s.server.BindHandler("/api/plugin/config", pluginAPIService.GetConfig)
	s.server.BindHandler("/api/plugin/wechat", pluginAPIService.GetCurrentWechat)
	s.server.BindHandler("/api/plugin/log", pluginAPIService.SendLog)
	s.server.BindHandler("/api/plugin/events", pluginAPIService.EventStream)
	s.server.BindHandler("/api/plugin/upload", pluginAPIService.UploadPlugin)
	g.Log().Info(ctx, "插件 API 服务已启用（包括 SSE 事件流和文件上传）")

	// 注册插件静态文件服务
	s.server.AddStaticPath("/plugins", "plugins")
	g.Log().Info(ctx, "插件静态文件服务已启用: /plugins -> plugins/")

	// 注册微信API代理路由
	wechatAPI := &wechat_api.WechatAPIProxyService{}
	// 基础接口
	s.server.BindHandler("/api/wechat/editVersion", wechatAPI.EditVersion)
	s.server.BindHandler("/api/wechat/getLoginStatus", wechatAPI.GetLoginStatus)
	s.server.BindHandler("/api/wechat/getLoginQRCode", wechatAPI.GetLoginQRCode)
	s.server.BindHandler("/api/wechat/setDownloadImage", wechatAPI.SetDownloadImage)
	s.server.BindHandler("/api/wechat/decryptImage", wechatAPI.DecryptImage)
	s.server.BindHandler("/api/wechat/checkWeChat", wechatAPI.CheckWeChat)
	s.server.BindHandler("/api/wechat/getAuthInfo", wechatAPI.GetAuthInfo)
	// 消息发送接口
	s.server.BindHandler("/api/wechat/sendText", wechatAPI.SendText)
	s.server.BindHandler("/api/wechat/sendText2", wechatAPI.SendText2)
	s.server.BindHandler("/api/wechat/sendReferText", wechatAPI.SendReferText)
	s.server.BindHandler("/api/wechat/sendImage", wechatAPI.SendImage)
	s.server.BindHandler("/api/wechat/sendFile", wechatAPI.SendFile)
	s.server.BindHandler("/api/wechat/sendGif", wechatAPI.SendGif)
	s.server.BindHandler("/api/wechat/sendShareUrl", wechatAPI.SendShareUrl)
	s.server.BindHandler("/api/wechat/sendApplet", wechatAPI.SendApplet)
	s.server.BindHandler("/api/wechat/sendMusic", wechatAPI.SendMusic)
	s.server.BindHandler("/api/wechat/sendChatLog", wechatAPI.SendChatLog)
	s.server.BindHandler("/api/wechat/sendCard", wechatAPI.SendCard)
	s.server.BindHandler("/api/wechat/sendXml", wechatAPI.SendXml)
	s.server.BindHandler("/api/wechat/sendLocationInfo", wechatAPI.SendLocationInfo)
	// 信息获取接口
	s.server.BindHandler("/api/wechat/getSelfInfo", wechatAPI.GetSelfInfo)
	s.server.BindHandler("/api/wechat/getLabelList", wechatAPI.GetLabelList)
	s.server.BindHandler("/api/wechat/getFriendList", wechatAPI.GetFriendList)
	s.server.BindHandler("/api/wechat/getGroupList", wechatAPI.GetGroupList)
	s.server.BindHandler("/api/wechat/getPublicList", wechatAPI.GetPublicList)
	// 好友管理接口
	s.server.BindHandler("/api/wechat/agreeFriendReq", wechatAPI.AgreeFriendReq)
	s.server.BindHandler("/api/wechat/addFriendByV3", wechatAPI.AddFriendByV3)
	s.server.BindHandler("/api/wechat/addFriendByGroupWxid", wechatAPI.AddFriendByGroupWxid)
	s.server.BindHandler("/api/wechat/delFriend", wechatAPI.DelFriend)
	s.server.BindHandler("/api/wechat/editObjRemark", wechatAPI.EditObjRemark)
	s.server.BindHandler("/api/wechat/queryNewFriend", wechatAPI.QueryNewFriend)
	s.server.BindHandler("/api/wechat/queryObj", wechatAPI.QueryObj)
	// 群聊管理接口
	s.server.BindHandler("/api/wechat/quitGroup", wechatAPI.QuitGroup)
	s.server.BindHandler("/api/wechat/createGroup", wechatAPI.CreateGroup)
	s.server.BindHandler("/api/wechat/queryGroup", wechatAPI.QueryGroup)
	s.server.BindHandler("/api/wechat/addMembers", wechatAPI.AddMembers)
	s.server.BindHandler("/api/wechat/inviteMembers", wechatAPI.InviteMembers)
	s.server.BindHandler("/api/wechat/delMembers", wechatAPI.DelMembers)
	s.server.BindHandler("/api/wechat/getMemberList", wechatAPI.GetMemberList)
	s.server.BindHandler("/api/wechat/getMemberNick", wechatAPI.GetMemberNick)
	s.server.BindHandler("/api/wechat/editSelfMemberNick", wechatAPI.EditSelfMemberNick)
	// 转账与其他接口
	s.server.BindHandler("/api/wechat/confirmTrans", wechatAPI.ConfirmTrans)
	s.server.BindHandler("/api/wechat/returnTrans", wechatAPI.ReturnTrans)
	s.server.BindHandler("/api/wechat/openBrowser", wechatAPI.OpenBrowser)
	s.server.BindHandler("/api/wechat/runCloudFunction", wechatAPI.RunCloudFunction)
	s.server.BindHandler("/api/wechat/setReadStatus", wechatAPI.SetReadStatus)
	s.server.BindHandler("/api/wechat/authCami", wechatAPI.AuthCami)
	s.server.BindHandler("/api/wechat/revokeMyMsg", wechatAPI.RevokeMyMsg)
	g.Log().Info(ctx, "微信API代理服务已启用: 47个接口")

	// 启动服务
	go func() {
		g.Log().Infof(ctx, "HTTP回调服务启动在: %s", address.String())
		if err := s.server.Start(); err != nil {
			g.Log().Errorf(ctx, "HTTP服务启动失败: %v", err)
		}
	}()

	// 这里同步，不需要异步
	callbackService.CheckAndUpdateAuthInfo(ctx)

	return nil
}

// StopServer 停止HTTP服务
func (s *HttpServerService) StopServer(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown()
	}
	return nil
}
