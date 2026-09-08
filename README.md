# 内网媒体联调助手

拖拽图片/视频，自动生成局域网 HTTP 链接，方便内网设备联调测试。

## 使用方式

### 安装

1. 将整个文件夹发送给需要使用的同事
2. 双击 `install.bat`（无弹窗，自动注册开机自启）
3. Chrome 打开 `chrome://extensions/` → 开启开发者模式 → 点击"加载已解压的扩展程序" → 选择本文件夹

### 使用

打开插件，拖入图片或视频，链接自动复制到剪贴板。

### 关闭服务

双击 `stop.bat`（关闭服务并取消开机自启）

### 重新启动

双击 `install.bat`

## 文件说明

| 文件 | 说明 |
|------|------|
| `install.bat` | 启动服务 + 注册开机自启 |
| `stop.bat` | 关闭服务 + 取消开机自启 |
| `server.exe` | 后端服务（Go 编译） |
| `manifest.json` | Chrome 扩展配置 |
| `popup.html` | 插件界面 |
| `popup.js` | 界面逻辑 |
| `background.js` | 后台通信 |
| `main.go` | 服务端源码 |

## 技术架构

```
Chrome 插件 (popup.js)
    ↓ base64 传输文件
Background Service Worker (background.js)
    ↓ fetch POST
Go Server (server.exe :9090)
    ↓ 保存文件 + 返回链接
http://内网IP:9090/files/xxx
```

## 开发

```bash
# 需要安装 Go: https://go.dev/dl/
go build -o server.exe main.go
```

## 端口

默认 `9090`，如需修改编辑 `main.go` 中的 `PORT` 常量。
