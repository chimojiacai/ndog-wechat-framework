package server

import (
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/naidog/wechat-framework/internal/api/plugin"
	"github.com/naidog/wechat-framework/internal/api/wechat"
	"github.com/naidog/wechat-framework/internal/core/callback"
)

// HTTPServer HTTP服务器
type HTTPServer struct {
	server            *ghttp.Server
	callbackHandler   *callback.Handler
	wechatProxy       *wechat.Proxy
	pluginAPI         *plugin.API
	callbackURLSuffix string
}

// NewHTTPServer 创建HTTP服务器实例
func NewHTTPServer(
	callbackHandler *callback.Handler,
	wechatProxy *wechat.Proxy,
	pluginAPI *plugin.API,
) *HTTPServer {
	return &HTTPServer{
		callbackHandler: callbackHandler,
		wechatProxy:     wechatProxy,
		pluginAPI:       pluginAPI,
	}
}

// Start 启动HTTP服务
func (s *HTTPServer) Start() error {
	ctx := gctx.New()

	// 读取配置
	address, err := g.Cfg().Get(ctx, "server.address")
	if err != nil || address.String() == "" {
		return fmt.Errorf("获取服务地址配置失败")
	}

	callbackURL, err := g.Cfg().Get(ctx, "server.callBackUrl")
	if err == nil && callbackURL.String() != "" {
		s.callbackURLSuffix = callbackURL.String()
	} else {
		s.callbackURLSuffix = "wechat/callback"
	}

	// 创建HTTP服务器
	s.server = g.Server()
	s.server.SetAddr(address.String())

	// 配置CORS
	s.server.Use(ghttp.MiddlewareCORS)

	// 注册路由
	s.registerRoutes()

	g.Log().Infof(ctx, "HTTP服务器启动在: %s", address.String())
	g.Log().Infof(ctx, "回调地址: http://localhost%s/%s", address.String(), s.callbackURLSuffix)

	// 启动服务器（非阻塞）
	go func() {
		if err := s.server.Start(); err != nil {
			g.Log().Errorf(ctx, "HTTP服务器启动失败: %v", err)
		}
	}()

	return nil
}

// Stop 停止HTTP服务
func (s *HTTPServer) Stop() error {
	if s.server != nil {
		return s.server.Shutdown()
	}
	return nil
}

// registerRoutes 注册路由
func (s *HTTPServer) registerRoutes() {
	// 微信回调路由
	s.server.BindHandler("/"+s.callbackURLSuffix, s.callbackHandler.HandleCallback)

	// 插件API路由组
	pluginGroup := s.server.Group("/api/plugin")
	{
		pluginGroup.GET("/config", s.pluginAPI.GetConfig)
		pluginGroup.GET("/wechat", s.pluginAPI.GetWechat)
		pluginGroup.POST("/log", s.pluginAPI.SendLog)
		pluginGroup.POST("/upload", s.pluginAPI.UploadFile)
		pluginGroup.GET("/events", callback.HandleSSEEvents)
	}

	// 插件静态文件服务
	s.server.AddStaticPath("/plugins", "plugins")

	// 微信API代理路由组
	wechatGroup := s.server.Group("/api/wechat")
	{
		// 基础接口
		wechatGroup.POST("/changeVersion", s.wechatProxy.ChangeVersion)
		wechatGroup.POST("/getLoginStatus", s.wechatProxy.GetLoginStatus)
		wechatGroup.POST("/getLoginQrCode", s.wechatProxy.GetLoginQrCode)
		wechatGroup.POST("/getSelfInfo", s.wechatProxy.GetSelfInfo)
		wechatGroup.POST("/getAuthInfo", s.wechatProxy.GetAuthInfo)
		wechatGroup.POST("/logout", s.wechatProxy.Logout)
		wechatGroup.POST("/getWechatVer", s.wechatProxy.GetWechatVer)

		// 消息发送接口
		wechatGroup.POST("/sendText", s.wechatProxy.SendText)
		wechatGroup.POST("/sendImage", s.wechatProxy.SendImage)
		wechatGroup.POST("/sendFile", s.wechatProxy.SendFile)
		wechatGroup.POST("/sendVideo", s.wechatProxy.SendVideo)
		wechatGroup.POST("/sendEmoji", s.wechatProxy.SendEmoji)
		wechatGroup.POST("/sendCard", s.wechatProxy.SendCard)
		wechatGroup.POST("/sendLink", s.wechatProxy.SendLink)
		wechatGroup.POST("/sendMiniProgram", s.wechatProxy.SendMiniProgram)
		wechatGroup.POST("/sendMusic", s.wechatProxy.SendMusic)
		wechatGroup.POST("/sendLocation", s.wechatProxy.SendLocation)
		wechatGroup.POST("/forwardMsg", s.wechatProxy.ForwardMsg)
		wechatGroup.POST("/sendAtText", s.wechatProxy.SendAtText)
		wechatGroup.POST("/revokeMsg", s.wechatProxy.RevokeMsg)

		// 信息获取接口
		wechatGroup.POST("/getFriendList", s.wechatProxy.GetFriendList)
		wechatGroup.POST("/getGroupList", s.wechatProxy.GetGroupList)
		wechatGroup.POST("/getGroupMembers", s.wechatProxy.GetGroupMembers)
		wechatGroup.POST("/getContactProfile", s.wechatProxy.GetContactProfile)
		wechatGroup.POST("/getDbNames", s.wechatProxy.GetDbNames)

		// 好友管理接口
		wechatGroup.POST("/addFriend", s.wechatProxy.AddFriend)
		wechatGroup.POST("/acceptFriend", s.wechatProxy.AcceptFriend)
		wechatGroup.POST("/deleteFriend", s.wechatProxy.DeleteFriend)
		wechatGroup.POST("/setRemark", s.wechatProxy.SetRemark)
		wechatGroup.POST("/topContact", s.wechatProxy.TopContact)
		wechatGroup.POST("/setBlacklist", s.wechatProxy.SetBlacklist)
		wechatGroup.POST("/searchFriend", s.wechatProxy.SearchFriend)

		// 群聊管理接口
		wechatGroup.POST("/createGroup", s.wechatProxy.CreateGroup)
		wechatGroup.POST("/addGroupMember", s.wechatProxy.AddGroupMember)
		wechatGroup.POST("/deleteGroupMember", s.wechatProxy.DeleteGroupMember)
		wechatGroup.POST("/quitGroup", s.wechatProxy.QuitGroup)
		wechatGroup.POST("/modifyGroupName", s.wechatProxy.ModifyGroupName)
		wechatGroup.POST("/modifyGroupNotice", s.wechatProxy.ModifyGroupNotice)
		wechatGroup.POST("/modifyNickInGroup", s.wechatProxy.ModifyNickInGroup)
		wechatGroup.POST("/inviteIntoGroup", s.wechatProxy.InviteIntoGroup)
		wechatGroup.POST("/getGroupQrCode", s.wechatProxy.GetGroupQrCode)

		// 其他接口
		wechatGroup.POST("/receiveTransfer", s.wechatProxy.ReceiveTransfer)
		wechatGroup.POST("/openBrowser", s.wechatProxy.OpenBrowser)
		wechatGroup.POST("/downloadImage", s.wechatProxy.DownloadImage)
		wechatGroup.POST("/downloadFile", s.wechatProxy.DownloadFile)
		wechatGroup.POST("/executeSql", s.wechatProxy.ExecuteSql)
		wechatGroup.POST("/callCloudFunc", s.wechatProxy.CallCloudFunc)
	}
}

// GetCallbackURL 获取回调URL
func (s *HTTPServer) GetCallbackURL() string {
	ctx := gctx.New()
	address, _ := g.Cfg().Get(ctx, "server.address")
	return fmt.Sprintf("http://localhost%s/%s", address.String(), s.callbackURLSuffix)
}
