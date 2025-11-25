package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"golang.org/x/sys/windows"
)

type WeChatService struct{}

var (
	advapi32               = syscall.NewLazyDLL("advapi32.dll")
	procInitializeAcl      = advapi32.NewProc("InitializeAcl")
	procAddAccessDeniedAce = advapi32.NewProc("AddAccessDeniedAce")
)

type ConfigJSON struct {
	CallBackUrl      string `json:"callBackUrl"`
	Port             string `json:"port"`
	CacheData        string `json:"cacheData"`
	TimeOut          string `json:"timeOut"`
	AutoLogin        string `json:"autoLogin"`
	Ver              string `json:"ver"`
	DecryptImg       string `json:"decryptImg"`
	NoHandleMsg      string `json:"noHandleMsg"`
	Resend           string `json:"resend"`
	GroupMemberEvent string `json:"groupMemberEvent"`
	HookSilk         string `json:"hookSilk"`
	DllPath          string `json:"dllPath"`
}

// 解除微信多开限制
func (w *WeChatService) EnableMultiWeChat(ctx context.Context) error {
	mutexNames := []string{
		"_WeChat_App_Instance_Identity_Mutex_Name", // 3.x 版本
		"XWeChat_App_Instance_Identity_Mutex_Name", // 4.x 版本
	}

	for _, name := range mutexNames {
		if err := w.createAndDenyMutex(ctx, name); err != nil {
			g.Log().Warningf(ctx, "处理 Mutex %s 失败: %v", name, err)
		}
	}

	g.Log().Info(ctx, "微信多开限制已解除")
	return nil
}

func (w *WeChatService) createAndDenyMutex(ctx context.Context, mutexName string) error {
	mutexNamePtr, err := syscall.UTF16PtrFromString(mutexName)
	if err != nil {
		return fmt.Errorf("转换 Mutex 名称失败: %v", err)
	}

	// 创建 Mutex
	hMutex, err := windows.CreateMutex(nil, false, mutexNamePtr)
	if err != nil {
		return fmt.Errorf("创建 Mutex 失败: %v", err)
	}

	var sidAuthWorld windows.SidIdentifierAuthority
	sidAuthWorld.Value = [6]byte{0, 0, 0, 0, 0, 1} // SECURITY_WORLD_SID_AUTHORITY

	var pEveryoneSID *windows.SID
	err = windows.AllocateAndInitializeSid(
		&sidAuthWorld,
		1,
		windows.SECURITY_WORLD_RID,
		0, 0, 0, 0, 0, 0, 0,
		&pEveryoneSID,
	)
	if err != nil {
		return fmt.Errorf("创建 SID 失败: %v", err)
	}
	defer windows.FreeSid(pEveryoneSID)

	const aclSize = 4096
	aclBuffer := make([]byte, aclSize)
	pAcl := uintptr(unsafe.Pointer(&aclBuffer[0]))

	const ACL_REVISION = 2
	ret, _, err := procInitializeAcl.Call(pAcl, uintptr(aclSize), uintptr(ACL_REVISION))
	if ret == 0 {
		return fmt.Errorf("初始化 ACL 失败: %v", err)
	}

	const MUTEX_ALL_ACCESS = 0x1F0001
	ret, _, err = procAddAccessDeniedAce.Call(
		pAcl,
		uintptr(ACL_REVISION),
		uintptr(MUTEX_ALL_ACCESS),
		uintptr(unsafe.Pointer(pEveryoneSID)),
	)
	if ret == 0 {
		return fmt.Errorf("添加 ACE 失败: %v", err)
	}

	err = windows.SetSecurityInfo(
		hMutex,
		windows.SE_KERNEL_OBJECT,
		windows.DACL_SECURITY_INFORMATION,
		nil,
		nil,
		(*windows.ACL)(unsafe.Pointer(pAcl)),
		nil,
	)
	if err != nil {
		return fmt.Errorf("设置安全信息失败: %v", err)
	}

	g.Log().Debugf(ctx, "Mutex %s 已处理", mutexName)
	return nil
}

func (w *WeChatService) RunWechat() (bool, error) {
	ctx := gctx.New()

	if err := w.EnableMultiWeChat(ctx); err != nil {
		g.Log().Warningf(ctx, "解除多开限制失败: %v", err)

	}

	installPath, err := g.Cfg().Get(ctx, "wechat.installationPath")
	if err != nil || installPath.String() == "" {
		return false, fmt.Errorf("获取微信安装路径失败")
	}

	cachePath, _ := g.Cfg().Get(ctx, "wechat.cachePath")
	timeOut, _ := g.Cfg().Get(ctx, "wechat.timeOut")
	decodePict, _ := g.Cfg().Get(ctx, "wechat.decodePict")
	ignoreMsg, _ := g.Cfg().Get(ctx, "wechat.ignoreMsg")
	resend, _ := g.Cfg().Get(ctx, "wechat.resend")
	groupMemberEvent, _ := g.Cfg().Get(ctx, "wechat.groupMemberEvent")
	hookSilk, _ := g.Cfg().Get(ctx, "wechat.hookSilk")

	resourceDir := "resources"
	if !gfile.Exists(resourceDir) {
		return false, fmt.Errorf("resources 目录不存在")
	}

	versionDllSrc := filepath.Join(resourceDir, "version.dll")
	versionDllDst := filepath.Join(installPath.String(), "version.dll")

	if !gfile.Exists(versionDllSrc) {
		return false, fmt.Errorf("version.dll 文件不存在: %s", versionDllSrc)
	}

	if !gfile.Exists(versionDllDst) {
		err = gfile.CopyFile(versionDllSrc, versionDllDst)
		if err != nil {
			return false, fmt.Errorf("复制 version.dll 失败: %v", err)
		}
		g.Log().Info(ctx, "version.dll 已复制到:", versionDllDst)
	} else {
		g.Log().Info(ctx, "version.dll 已存在，跳过复制:", versionDllDst)
	}

	randomPort := rand.Intn(56536) + 9000

	dllRelPath := filepath.Join(resourceDir, "4.1.2.17.dll")
	dllAbsPath, err := filepath.Abs(dllRelPath)
	if err != nil {
		return false, fmt.Errorf("获取 dll 绝对路径失败: %v", err)
	}

	if !gfile.Exists(dllAbsPath) {
		return false, fmt.Errorf("4.1.2.17.dll 文件不存在: %s", dllAbsPath)
	}

	config := ConfigJSON{
		CallBackUrl:      "http://127.0.0.1:9001/wechat/callback",
		Port:             fmt.Sprintf("%d", randomPort),
		CacheData:        cachePath.String(),
		TimeOut:          timeOut.String(),
		AutoLogin:        "0",
		Ver:              "",
		DecryptImg:       decodePict.String(),
		NoHandleMsg:      ignoreMsg.String(),
		Resend:           resend.String(),
		GroupMemberEvent: groupMemberEvent.String(),
		HookSilk:         hookSilk.String(),
		DllPath:          dllAbsPath,
	}

	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return false, fmt.Errorf("生成 JSON 失败: %v", err)
	}

	configJsonDst := filepath.Join(installPath.String(), "config.json")
	err = os.WriteFile(configJsonDst, jsonData, 0644)
	if err != nil {
		return false, fmt.Errorf("写入 config.json 失败: %v", err)
	}

	g.Log().Info(ctx, "成功配置微信")
	g.Log().Infof(ctx, "- config.json 已更新到: %s", configJsonDst)
	g.Log().Infof(ctx, "- 随机端口: %d", randomPort)

	wechatExe := filepath.Join(installPath.String(), "Weixin.exe")
	if !gfile.Exists(wechatExe) {
		return false, fmt.Errorf("微信程序不存在: %s", wechatExe)
	}

	cmd := exec.Command(wechatExe)
	cmd.Dir = installPath.String() // 设置工作目录为微信安装目录
	// 注意：不要设置 HideWindow，否则微信窗口会被隐藏

	err = cmd.Start()
	if err != nil {
		return false, fmt.Errorf("启动微信失败: %v", err)
	}

	pid := cmd.Process.Pid
	g.Log().Infof(ctx, "- 微信进程已启动，PID: %d", pid)

	// 等待2秒后检查进程是否还存活
	time.Sleep(2 * time.Second)

	// 使用 tasklist 命令检查进程是否还在运行
	checkCmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH")
	checkCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := checkCmd.Output()
	if err != nil || len(output) == 0 || !strings.Contains(string(output), "Weixin.exe") {
		return false, fmt.Errorf("微信进程 PID %d 已退出，可能是配置错误、权限问题或DLL文件损坏", pid)
	}

	g.Log().Infof(ctx, "- 微信进程运行正常，PID: %d", pid)

	return true, nil
}
