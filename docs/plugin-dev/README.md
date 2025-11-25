# 插件开发指南

本指南将帮助你快速上手奶狗微信框架的插件开发。

## 目录

- [快速开始](#快速开始)
- [插件结构](#插件结构)
- [API 参考](#api-参考)
- [事件系统](#事件系统)
- [最佳实践](#最佳实践)
- [示例插件](#示例插件)
- [常见问题](#常见问题)

## 快速开始

### 1. 创建插件目录

```
my-plugin/
├── plugin.json          # 插件元数据（必需）
├── icon.png            # 插件图标（可选）
└── frontend/           # 前端资源
    ├── index.html      # 入口页面
    ├── assets/         # 静态资源
    │   ├── css/
    │   ├── js/
    │   └── images/
    └── ...
```

### 2. 编写插件元数据

创建 `plugin.json`:

```json
{
  "id": "my-plugin",
  "name": "我的第一个插件",
  "version": "1.0.0",
  "author": "你的名字",
  "description": "这是一个示例插件",
  "icon": "icon.png",
  "entry": "frontend/index.html",
  "type": "window"
}
```

**字段说明:**

- `id`: 插件唯一标识符（必需，小写字母和连字符）
- `name`: 插件显示名称（必需）
- `version`: 版本号（必需，遵循语义化版本）
- `author`: 作者名称（必需）
- `description`: 插件描述（必需）
- `icon`: 图标文件路径（可选，相对于插件根目录）
- `entry`: 入口文件路径（必需，相对于插件根目录）
- `type`: 插件类型（目前仅支持 "window"）

### 3. 创建入口页面

创建 `frontend/index.html`:

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>我的插件</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            padding: 20px;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
        }
        button {
            padding: 10px 20px;
            margin: 5px;
            cursor: pointer;
        }
        #log {
            margin-top: 20px;
            padding: 10px;
            background: #f5f5f5;
            border-radius: 4px;
            max-height: 300px;
            overflow-y: auto;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>我的第一个插件</h1>
        <button onclick="startListening()">开始监听</button>
        <button onclick="stopListening()">停止监听</button>
        <button onclick="testSendMessage()">测试发送消息</button>
        <div id="log"></div>
    </div>

    <script src="app.js"></script>
</body>
</html>
```

### 4. 编写插件逻辑

创建 `frontend/assets/js/app.js`:

```javascript
// 配置
const API_BASE = 'http://localhost:9001/api/plugin';
const PLUGIN_ID = 'my-plugin';

let eventSource = null;
let currentAccounts = [];

// 开始监听微信事件
function startListening() {
    if (eventSource) {
        log('已经在监听中...', 'warning');
        return;
    }

    eventSource = new EventSource(`${API_BASE}/events`);
    
    eventSource.onopen = () => {
        log('连接成功，开始监听事件', 'success');
    };
    
    eventSource.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'connected') return;
        
        handleEvent(data);
    };
    
    eventSource.onerror = (error) => {
        log('连接错误: ' + error, 'error');
        stopListening();
    };
}

// 停止监听
function stopListening() {
    if (eventSource) {
        eventSource.close();
        eventSource = null;
        log('已停止监听', 'info');
    }
}

// 处理事件
function handleEvent(event) {
    log(`收到事件: ${event.type}`, 'info');
    
    switch (event.type) {
        case 'recvMsg':
            handleMessage(event.data);
            break;
        case 'loginSuccess':
            handleLogin(event.data);
            break;
        case 'friendReq':
            handleFriendRequest(event.data);
            break;
        default:
            log(`未处理的事件类型: ${event.type}`, 'warning');
    }
}

// 处理消息
function handleMessage(data) {
    const msg = data.data.msg;
    const fromWxid = data.data.fromWxid;
    log(`收到消息: ${msg} (来自: ${fromWxid})`, 'info');
    
    // 这里可以添加自动回复逻辑
    if (msg === '你好') {
        sendTextMessage(data.port, fromWxid, '你好！我是机器人');
    }
}

// 处理登录
function handleLogin(data) {
    log(`账号登录: ${data.data.nick}`, 'success');
    loadAccounts();
}

// 处理好友请求
function handleFriendRequest(data) {
    log(`收到好友请求: ${data.data.nick}`, 'info');
}

// 获取微信账号列表
async function loadAccounts() {
    try {
        const response = await fetch(`${API_BASE}/wechat`);
        const result = await response.json();
        
        if (result.code === 200) {
            currentAccounts = result.data.list || [];
            log(`当前有 ${currentAccounts.length} 个账号在线`, 'success');
        }
    } catch (error) {
        log('获取账号列表失败: ' + error, 'error');
    }
}

// 发送文本消息
async function sendTextMessage(port, wxid, message) {
    try {
        const response = await fetch(`http://localhost:9001/api/wechat/sendText?port=${port}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ wxid, msg: message })
        });
        
        const result = await response.json();
        if (result.code === 200) {
            log(`消息发送成功: ${message}`, 'success');
        } else {
            log(`消息发送失败: ${result.msg}`, 'error');
        }
    } catch (error) {
        log('发送消息失败: ' + error, 'error');
    }
}

// 测试发送消息
async function testSendMessage() {
    await loadAccounts();
    
    if (currentAccounts.length === 0) {
        log('没有在线账号', 'warning');
        return;
    }
    
    const account = currentAccounts[0];
    log(`使用账号 ${account.nick} 发送测试消息`, 'info');
    
    // 这里需要替换为实际的接收方 wxid
    // await sendTextMessage(account.port, 'wxid_xxx', '这是一条测试消息');
    log('请在代码中设置接收方 wxid', 'warning');
}

// 发送日志到主程序
async function sendLog(message, type = '信息') {
    const colors = {
        '信息': '#409EFF',
        '成功': '#67C23A',
        '警告': '#E6A23C',
        '错误': '#F56C6C'
    };
    
    try {
        await fetch(`${API_BASE}/log`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                pluginId: PLUGIN_ID,
                timeStamp: new Date().toLocaleString('zh-CN'),
                response: '我的插件',
                logType: type,
                msg: message,
                color: colors[type] || '#409EFF'
            })
        });
    } catch (error) {
        console.error('发送日志失败:', error);
    }
}

// 本地日志显示
function log(message, level = 'info') {
    const logDiv = document.getElementById('log');
    const time = new Date().toLocaleTimeString();
    const colors = {
        info: '#409EFF',
        success: '#67C23A',
        warning: '#E6A23C',
        error: '#F56C6C'
    };
    
    const entry = document.createElement('div');
    entry.style.color = colors[level] || '#333';
    entry.textContent = `[${time}] ${message}`;
    logDiv.appendChild(entry);
    logDiv.scrollTop = logDiv.scrollHeight;
    
    // 同时发送到主程序
    const typeMap = {
        info: '信息',
        success: '成功',
        warning: '警告',
        error: '错误'
    };
    sendLog(message, typeMap[level] || '信息');
}

// 页面卸载时清理
window.addEventListener('beforeunload', () => {
    stopListening();
});

// 页面加载完成后初始化
window.addEventListener('load', () => {
    log('插件已加载', 'success');
    loadAccounts();
});
```

### 5. 打包插件

将插件文件夹压缩为 `.dog` 格式：

```bash
# 使用 zip 命令
zip -r my-plugin.dog my-plugin/

# 或使用 7-Zip
7z a my-plugin.dog my-plugin/

# 或使用 PowerShell
Compress-Archive -Path my-plugin -DestinationPath my-plugin.dog
```

### 6. 安装插件

1. 打开框架主窗口
2. 进入插件管理页面
3. 点击"上传插件"按钮
4. 选择 `.dog` 文件
5. 等待上传完成
6. 刷新插件列表
7. 点击插件图标打开

## 插件结构

### 推荐的项目结构

```
my-plugin/
├── plugin.json                 # 插件元数据
├── icon.png                    # 插件图标 (128x128)
├── README.md                   # 插件说明
└── frontend/
    ├── index.html             # 入口页面
    ├── assets/
    │   ├── css/
    │   │   └── style.css      # 样式文件
    │   ├── js/
    │   │   ├── app.js         # 主逻辑
    │   │   ├── api.js         # API 封装
    │   │   └── utils.js       # 工具函数
    │   └── images/            # 图片资源
    └── lib/                   # 第三方库（如需要）
```

## API 参考

### 插件 API

#### 1. 获取配置文件

```javascript
const response = await fetch('http://localhost:9001/api/plugin/config');
const result = await response.json();
// result.data 包含配置文件内容（YAML 字符串）
```

#### 2. 获取微信账号列表

```javascript
const response = await fetch('http://localhost:9001/api/plugin/wechat');
const result = await response.json();
// result.data.list 包含账号数组
```

响应格式:
```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "wxid": "wxid_xxx",
        "wxNum": "微信号",
        "nick": "昵称",
        "avatarUrl": "头像URL",
        "port": 19088,
        "pid": 12345,
        "expireTime": "2024-12-31",
        "isExpire": 0
      }
    ]
  }
}
```

#### 3. 发送日志

```javascript
await fetch('http://localhost:9001/api/plugin/log', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        pluginId: 'my-plugin',      // 必需，与 plugin.json 中的 id 一致
        timeStamp: '2024-01-01 12:00:00',
        response: '插件名称',
        logType: '信息',
        msg: '日志内容',
        color: '#409EFF'            // 可选，默认蓝色
    })
});
```

颜色参考:
- 成功: `#67C23A`
- 信息: `#409EFF`
- 警告: `#E6A23C`
- 错误: `#F56C6C`

#### 4. 监听事件 (SSE)

```javascript
const eventSource = new EventSource('http://localhost:9001/api/plugin/events');

eventSource.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'connected') return; // 跳过连接确认
    
    console.log('事件类型:', data.type);
    console.log('事件数据:', data.data);
};

// 关闭连接
eventSource.close();
```

### 微信 API

所有微信 API 使用统一格式:

```javascript
const response = await fetch(`http://localhost:9001/api/wechat/{接口名}?port={端口号}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        // 接口参数
    })
});
```

#### 常用接口示例

**发送文本消息:**
```javascript
await fetch(`http://localhost:9001/api/wechat/sendText?port=19088`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        wxid: 'wxid_xxx',
        msg: '你好'
    })
});
```

**获取好友列表:**
```javascript
await fetch(`http://localhost:9001/api/wechat/getFriendList?port=19088`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        type: "1"  // 1=从缓存获取，2=重新遍历
    })
});
```

**获取群聊列表:**
```javascript
await fetch(`http://localhost:9001/api/wechat/getGroupList?port=19088`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        type: "2"
    })
});
```

完整 API 列表请参考 [API 文档](../api/README.md)

## 事件系统

### 事件类型

| 事件类型 | 说明 | 数据结构 |
|---------|------|---------|
| `injectSuccess` | 注入成功 | `{ port, pid }` |
| `loginSuccess` | 登录成功 | `{ wxid, nick, ... }` |
| `recvMsg` | 接收消息 | `{ fromWxid, msg, msgType, ... }` |
| `transPay` | 转账事件 | `{ fromWxid, money, memo, ... }` |
| `friendReq` | 好友请求 | `{ wxid, nick, content, ... }` |
| `groupMemberChanges` | 群成员变动 | `{ fromWxid, eventType, ... }` |
| `authExpire` | 授权到期 | `{ wxid, expireTime, ... }` |

### 事件数据格式

所有事件遵循统一格式:

```json
{
  "type": "事件类型",
  "data": {
    "type": "事件类型",
    "des": "事件描述",
    "data": { /* 具体数据 */ },
    "timestamp": "时间戳",
    "wxid": "微信ID",
    "port": 端口号,
    "pid": 进程ID,
    "flag": "标识"
  }
}
```

### 消息类型

| msgType | 说明 |
|---------|------|
| 1 | 文本消息 |
| 3 | 图片消息 |
| 34 | 语音消息 |
| 42 | 名片消息 |
| 43 | 视频消息 |
| 47 | 动态表情 |
| 48 | 地理位置 |
| 49 | 分享链接或附件 |
| 2001 | 红包消息 |
| 2002 | 小程序消息 |
| 2003 | 群邀请 |
| 10000 | 系统消息 |

## 最佳实践

### 1. 错误处理

```javascript
async function safeApiCall(apiFunc) {
    try {
        return await apiFunc();
    } catch (error) {
        console.error('API 调用失败:', error);
        log(`操作失败: ${error.message}`, 'error');
        return null;
    }
}

// 使用
const accounts = await safeApiCall(() => loadAccounts());
```

### 2. 资源清理

```javascript
// 确保在页面卸载时清理资源
window.addEventListener('beforeunload', () => {
    if (eventSource) {
        eventSource.close();
    }
    // 清理其他资源...
});
```

### 3. 配置管理

```javascript
// 使用 localStorage 保存插件配置
const config = {
    autoReply: true,
    keywords: ['你好', 'hello']
};

// 保存
localStorage.setItem('plugin-config', JSON.stringify(config));

// 读取
const savedConfig = JSON.parse(localStorage.getItem('plugin-config') || '{}');
```

### 4. 日志分级

```javascript
const Logger = {
    debug: (msg) => console.log(`[DEBUG] ${msg}`),
    info: (msg) => log(msg, 'info'),
    warn: (msg) => log(msg, 'warning'),
    error: (msg) => log(msg, 'error')
};
```

### 5. API 封装

```javascript
class WeChatAPI {
    constructor(baseUrl = 'http://localhost:9001') {
        this.baseUrl = baseUrl;
    }
    
    async sendText(port, wxid, msg) {
        const response = await fetch(`${this.baseUrl}/api/wechat/sendText?port=${port}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ wxid, msg })
        });
        return await response.json();
    }
    
    async getFriendList(port) {
        const response = await fetch(`${this.baseUrl}/api/wechat/getFriendList?port=${port}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ type: "1" })
        });
        return await response.json();
    }
}

const api = new WeChatAPI();
```

## 示例插件

### 自动回复插件

完整的自动回复插件示例，支持关键词匹配和自定义回复。

[查看完整代码](./examples/auto-reply/)

### 消息统计插件

统计消息数量、类型分布等信息。

[查看完整代码](./examples/message-stats/)

### 好友管理插件

批量添加好友、自动通过好友请求等。

[查看完整代码](./examples/friend-manager/)

## 常见问题

### Q: 插件无法加载？

A: 检查以下几点:
1. `plugin.json` 格式是否正确
2. 插件 ID 是否唯一
3. 入口文件路径是否正确
4. 查看控制台错误信息

### Q: 事件接收不到？

A: 确认:
1. SSE 连接是否成功建立
2. 微信账号是否已登录
3. 检查浏览器控制台是否有错误

### Q: API 调用失败？

A: 检查:
1. 端口号是否正确
2. 请求参数是否完整
3. 网络连接是否正常
4. 查看响应错误信息

### Q: 如何调试插件？

A: 
1. 在插件窗口右键选择"检查元素"
2. 使用浏览器开发者工具
3. 查看 Console 输出
4. 使用 Network 面板查看请求

### Q: 插件窗口大小如何调整？

A: 目前插件窗口大小固定为 520x380，暂不支持调整。

## 进阶主题

### 使用现代前端框架

你可以使用 React、Vue 等现代框架开发插件:

```bash
# 创建 React 项目
npx create-react-app my-plugin-frontend
cd my-plugin-frontend

# 构建
npm run build

# 将 build 目录复制到插件的 frontend 目录
```

### 使用 TypeScript

```typescript
interface WeChatAccount {
    wxid: string;
    nick: string;
    port: number;
    // ...
}

async function loadAccounts(): Promise<WeChatAccount[]> {
    const response = await fetch('http://localhost:9001/api/plugin/wechat');
    const result = await response.json();
    return result.data.list || [];
}
```

### 性能优化

1. 使用防抖和节流
2. 缓存 API 响应
3. 延迟加载资源
4. 优化事件处理

## 资源链接

- [API 完整文档](../api/README.md)
- [示例插件仓库](https://github.com/naidog/wechat-framework-plugins)
- [问题反馈](https://github.com/naidog/wechat-framework/issues)
- [讨论区](https://github.com/naidog/wechat-framework/discussions)

---

如有问题或建议，欢迎提交 Issue 或 PR！
