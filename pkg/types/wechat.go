package types

// WechatAccount 微信账号信息
type WechatAccount struct {
	Wxid       string `json:"wxid"`       // 微信ID
	WxNum      string `json:"wxNum"`      // 微信号
	Nick       string `json:"nick"`       // 昵称
	AvatarUrl  string `json:"avatarUrl"`  // 头像URL
	Port       int    `json:"port"`       // 服务端口
	Pid        int    `json:"pid"`        // 进程ID
	ExpireTime string `json:"expireTime"` // 授权到期时间
	IsExpire   int    `json:"isExpire"`   // 是否已到期（1=是，0=否）
}

// WechatAccountList 微信账号列表
type WechatAccountList struct {
	List []WechatAccount `json:"list"`
}

// AuthInfo 授权信息
type AuthInfo struct {
	ExpireTime string `json:"expireTime"` // 到期时间
	IsExpire   int    `json:"isExpire"`   // 是否已到期
}

// CallbackEvent 通用回调事件结构
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

// PluginMetadata 插件元数据
type PluginMetadata struct {
	ID          string `json:"id"`          // 插件ID
	Name        string `json:"name"`        // 插件名称
	Version     string `json:"version"`     // 版本号
	Author      string `json:"author"`      // 作者
	Description string `json:"description"` // 描述
	Icon        string `json:"icon"`        // 图标路径
	Entry       string `json:"entry"`       // 入口文件
	Type        string `json:"type"`        // 插件类型
}

// PluginInfo 插件信息
type PluginInfo struct {
	Metadata PluginMetadata `json:"metadata"` // 元数据
	Path     string         `json:"path"`     // 插件路径
	Enabled  bool           `json:"enabled"`  // 是否启用
	IconURL  string         `json:"iconUrl"`  // 图标URL
	EntryURL string         `json:"entryUrl"` // 入口URL
}
