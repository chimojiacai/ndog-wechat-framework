package main

import (
	"embed"
	_ "embed"
	"log"
	"time"

	"github.com/naidog/wechat-framework/service/config"
	"github.com/naidog/wechat-framework/service/http_callback"
	"github.com/naidog/wechat-framework/service/plugin"
	"github.com/naidog/wechat-framework/service/utils"
	"github.com/naidog/wechat-framework/service/wechat"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

// main function serves as the application's entry point. It initializes the application, creates a window,
// and starts a goroutine that emits a time-based event every second. It subsequently runs the application and
// logs any error that might occur.
func main() {
	// 设置 GoFrame 配置文件路径
	g.Cfg().GetAdapter().(*gcfg.AdapterFile).SetPath("configs")

	// 创建微信账号服务（先创建，稍后设置 app）
	accountService := wechat.NewWechatAccountService(nil)

	// 创建日志服务（先创建，稍后设置 app）
	logService := utils.NewLogService(nil)

	// 创建插件服务（先创建，稍后设置 app）
	pluginService := &plugin.PluginService{}

	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	app := application.New(application.Options{
		Name:        "奶狗微信框架",
		Description: "奶狗微信框架",

		Services: []application.Service{
			application.NewService(&config.ConfigGetService{}),
			application.NewService(&config.ConfigSetService{}),
			application.NewService(&config.ThemeService{}),
			application.NewService(&utils.GetWechatPathService{}),
			application.NewService(&utils.NoupdateWechatService{}),
			application.NewService(&wechat.WeChatService{}),
			application.NewService(accountService),
			application.NewService(logService),
			application.NewService(pluginService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// 设置插件服务的 app 实例和日志服务
	pluginService.SetApp(app)
	pluginService.SetLogService(logService)

	// 设置插件服务到回调处理，使回调事件可以广播给插件
	http_callback.SetPluginService(pluginService)

	// 启动HTTP回调服务
	httpServer := &http_callback.HttpServerService{}
	if err := httpServer.StartServer(); err != nil {
		log.Printf("启动HTTP服务失败: %v", err)
	}

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	_ = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:               "奶狗微信框架 x64 v0.0.1",
		Width:               796,                        // 设置窗口宽度
		Height:              620,                        // 设置窗口高度
		MinWidth:            796,                        // 最小宽度（与宽度相同实现固定大小）
		MinHeight:           620,                        // 最小高度（与高度相同实现固定大小）
		MaxWidth:            796,                        // 最大宽度（与宽度相同实现固定大小）
		MaxHeight:           620,                        // 最大高度（与高度相同实现固定大小）
		DisableResize:       true,                       // 禁止调整窗口大小
		MaximiseButtonState: application.ButtonDisabled, // 禁止最大化
		DevToolsEnabled:     false,                      // 禁用开发者工具（生产环境）
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(255, 255, 255),
		Windows:          application.WindowsWindow{Theme: 0}, //这里是设框架主题，0=跟随系统，1=Dark(黑色)，2=Light(浅色)
		URL:              "/",
	})

	// 设置 app 并启动微信账号监听服务
	ctx := gctx.New()
	accountService.SetApp(app)
	go accountService.StartWatching(ctx)

	// 设置日志服务的 app 实例
	logService.SetApp(app)

	// 程序启动后发送测试日志
	go func() {
		time.Sleep(3 * time.Second) // 等待 3 秒，确保前端完全加载
		logService.SendLog(
			ctx,
			time.Now().Format("2006-01-02 15:04:05"),
			"框架",
			"公告",
			"欢迎使用奶狗微信框架！",
			"#67C23A", // 绿色
		)
		logService.SendLog(
			ctx,
			time.Now().Format("2006-01-02 15:04:05"),
			"框架",
			"公告",
			"第一次使用请先阅读框架使用文档！",
			"#67C23A", // 绿色
		)
		logService.SendLog(
			ctx,
			time.Now().Format("2006-01-02 15:04:05"),
			"框架",
			"公告",
			"仅供学习测试、请勿直接用于商业用途、违法内容、生产环境等，与作者无关！",
			"#67C23A", // 绿色
		)
		logService.SendLog(
			ctx,
			time.Now().Format("2006-01-02 15:04:05"),
			"框架",
			"全局",
			"服务已启动！",
			"#67C23A", // 绿色
		)
	}()

	// Create a goroutine that emits an event containing the current time every second.
	// The frontend can listen to this event and update the UI accordingly.
	go func() {
		// fmt.Println("窗口id：", mainWind.ID())
		// mainWind.SetAlwaysOnTop(true) // 置顶
		for {
			now := time.Now().Format(time.RFC1123)
			app.Event.Emit("time", now)
			time.Sleep(time.Second)
		}
	}()

	// Run the application. This blocks until the application has been exited.
	err := app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
