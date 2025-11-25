package main

import (
	"context"
	// "embed"  // 暂时不使用embed
	// _ "embed"
	"log"
	"time"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/naidog/wechat-framework/internal/api/plugin"
	"github.com/naidog/wechat-framework/internal/api/wechat"
	"github.com/naidog/wechat-framework/internal/config"
	"github.com/naidog/wechat-framework/internal/core/account"
	"github.com/naidog/wechat-framework/internal/core/callback"
	pluginCore "github.com/naidog/wechat-framework/internal/core/plugin"
	"github.com/naidog/wechat-framework/internal/server"
	"github.com/naidog/wechat-framework/internal/service"
	"github.com/naidog/wechat-framework/internal/utils"
	"github.com/naidog/wechat-framework/pkg/logger"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// 前端资源embed（需要先构建前端）
// 开发阶段暂时注释，等前端构建完成后再启用
// //go:embed all:../../frontend/dist
// var assets embed.FS

// //go:embed ../../build/appicon.png
// var appIcon []byte

func main() {
	ctx := gctx.New()

	// 创建核心服务
	configService := config.NewService()
	wechatPathService := utils.NewWechatPathService()
	wechatUpdateService := utils.NewWechatUpdateService(configService)

	// 创建Wails应用
	app := application.New(application.Options{
		Name:        "奶狗微信框架",
		Description: "奶狗微信框架",
		Services:    []application.Service{},
		// 开发阶段暂不使用embed的前端资源
		// 前端需要单独运行或构建后再启用
		// Assets: application.AssetOptions{
		// 	Handler: application.AssetFileServerFS(assets),
		// },
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// 创建日志服务
	logService := logger.NewService(app)
	logger.SetGlobalLogService(logService)

	// 创建账号管理器
	accountManager := account.NewManager(app)

	// 创建插件管理器
	pluginManager := pluginCore.NewManager(app, logService)

	// 创建回调处理器
	callbackHandler := callback.NewHandler(pluginManager)

	// 创建API服务
	wechatProxy := wechat.NewProxy()
	pluginAPI := plugin.NewAPI(pluginManager)

	// 创建HTTP服务器
	httpServer := server.NewHTTPServer(callbackHandler, wechatProxy, pluginAPI)

	// 创建Wails服务适配器
	wailsConfigService := service.NewConfigService(configService)
	wailsPathService := service.NewWechatPathService(wechatPathService)
	wailsUpdateService := service.NewWechatUpdateService(wechatUpdateService)
	wailsAccountService := service.NewAccountService(accountManager)
	wailsLogService := service.NewLogService(logService)
	wailsPluginService := service.NewPluginService(pluginManager)

	// 注册Wails服务
	app.RegisterService(application.NewService(wailsConfigService))
	app.RegisterService(application.NewService(wailsPathService))
	app.RegisterService(application.NewService(wailsUpdateService))
	app.RegisterService(application.NewService(wailsAccountService))
	app.RegisterService(application.NewService(wailsLogService))
	app.RegisterService(application.NewService(wailsPluginService))

	// 创建主窗口
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "奶狗微信框架",
		Width:  1200,
		Height: 800,
		Mac: application.MacWindow{
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
			InvisibleTitleBarHeight: 50,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	// 启动HTTP服务器
	if err := httpServer.Start(); err != nil {
		log.Fatalf("HTTP服务器启动失败: %v", err)
	}

	// 启动账号监听
	go accountManager.StartWatching(ctx)

	// 启动定时任务
	go startTimedTasks(app, ctx)

	// 启动日志发送
	go startLogSender(app, logService, ctx)

	// 运行应用
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// startTimedTasks 启动定时任务
func startTimedTasks(app *application.App, ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			// 每秒发送时间事件到前端
			app.Event.Emit("time:update", map[string]interface{}{
				"time": t.Format("2006-01-02 15:04:05"),
			})
		}
	}
}

// startLogSender 启动日志发送
func startLogSender(app *application.App, logService *logger.Service, ctx context.Context) {
	time.Sleep(2 * time.Second) // 等待前端准备好

	// 发送启动日志
	logService.SendLog(
		ctx,
		time.Now().Format("2006-01-02 15:04:05"),
		"框架",
		"成功",
		"奶狗微信框架已启动",
		"#67C23A",
	)

	logService.SendLog(
		ctx,
		time.Now().Format("2006-01-02 15:04:05"),
		"框架",
		"信息",
		"HTTP服务已启动在 :9001",
		"#409EFF",
	)
}
