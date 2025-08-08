# 汇付支付配置管理系统

基于 Golang + HTML5 的汇付支付SDK配置管理工具，支持动态配置和微信商户管理。

## 🚀 快速开始

### 开发环境运行
```bash
# 安装依赖
go mod tidy

# 运行
go run *.go
```

### 生产环境部署
```bash
# 构建所有平台版本
./build.sh

# 选择对应平台的压缩包进行部署
# 详见 DEPLOY.md
```

## 📋 功能特性

- ✅ **多平台支持**: Windows, Linux (AMD64/ARM64), macOS
- ✅ **实时配置**: 动态管理汇付SDK配置参数
- ✅ **微信集成**: 支持微信小程序商户配置和查询
- ✅ **密钥管理**: RSA密钥生成和安全存储
- ✅ **响应式UI**: 现代化Web界面，支持移动端
- ✅ **错误处理**: 完善的错误处理和用户反馈
- ✅ **多环境**: 支持测试和生产环境切换

## 🏗️ 技术架构

- **后端**: Go 1.19+ + Gin框架
- **前端**: HTML5 + 原生JavaScript
- **SDK**: 汇付bspay-go-sdk v1.0.20
- **存储**: 内存存储 + 临时文件

## 📦 构建和部署

### 自动构建
```bash
# 运行构建脚本，生成所有平台的压缩包
./build.sh
```

构建完成后会在 `build/` 目录生成以下文件：
- `ghuifu-1.0.0-windows.zip` (Windows 64位通用版本)
- `ghuifu-1.0.0-linux.tar.gz` (Linux自动检测架构: AMD64/ARM64)
- `ghuifu-1.0.0-darwin.tar.gz` (macOS自动检测架构: Intel/Apple Silicon)

### 手动构建
```bash
# 构建 Linux AMD64
GOOS=linux GOARCH=amd64 go build -o ghuifu-linux-amd64

# 构建 Windows AMD64
GOOS=windows GOARCH=amd64 go build -o ghuifu-windows-amd64.exe

# 构建 Linux ARM64 (适用于树莓派等)
GOOS=linux GOARCH=arm64 go build -o ghuifu-linux-arm64
```

## 📖 使用说明

### 1. 系统配置
1. 访问 http://localhost:8080
2. 填入系统参数 (sys_id, product_id, rsa_private_key)
3. 选择环境 (测试/生产)
4. 保存配置

### 2. 微信商户配置
1. 选择已保存的系统配置
2. 填入汇付ID和微信小程序信息
3. 选择费率类型
4. 提交配置

### 3. 查询配置
- 使用"查询微信配置"功能查看当前配置状态

## 🔧 配置参数

| 参数 | 说明 | 必填 |
|------|------|------|
| sys_id | 系统ID | ✅ |
| product_id | 产品ID | ✅ |
| rsa_private_key | RSA私钥 | ✅ |
| huifu_id | 汇付ID | ✅ |
| wx_woa_app_id | 微信小程序AppID | ✅ |
| wx_woa_path | 小程序路径 | ✅ |
| fee_type | 费率类型 (01-09) | ✅ |

## 🌐 API接口

- `POST /api/config` - 保存系统配置
- `GET /api/configs` - 获取配置列表
- `DELETE /api/config/:sys_id` - 删除配置
- `POST /api/wechat-config` - 配置微信商户
- `POST /api/wechat-config-query` - 查询微信配置
- `GET /api/generate-test-key` - 生成测试密钥

## 🔒 安全特性

- RSA密钥仅在内存中临时存储
- 支持测试和生产环境隔离
- 自动清理临时配置文件
- 详细的操作日志记录

## 📋 系统要求

- Go 1.19+ (开发环境)
- 端口8080可用
- 网络连接到汇付API
- 现代浏览器支持

## 🤝 开发说明

### 项目结构
```
├── main.go              # 主程序入口
├── huifu_client.go      # 汇付客户端接口定义
├── real_client.go       # 真实SDK客户端实现
├── build.sh             # 构建脚本
├── static/              # 前端文件
│   ├── index.html      # 主页面
│   └── app.js          # JavaScript逻辑
├── DEPLOY.md           # 部署说明
└── README.md           # 项目说明
```

### 开发指南
1. 修改代码后使用 `go run *.go` 测试
2. 使用 `./build.sh` 构建生产版本
3. 前端修改在 `static/` 目录
4. API修改在 `main.go` 中对应的handler

## 📝 更新日志

### v1.0.0 (2024-01-08)
- ✅ 初始版本发布
- ✅ 支持汇付SDK配置管理
- ✅ 微信商户配置和查询功能
- ✅ 多平台打包和部署
- ✅ 完整的Web管理界面

## 📞 支持

如有问题请查看 `DEPLOY.md` 中的故障排除部分，或参考汇付支付官方文档：
- 汇付开放平台: https://paas.huifu.com/open/doc
- GitHub SDK: https://github.com/huifurepo/bspay-go-sdk

## 📄 许可证

本项目仅用于汇付支付SDK配置管理，请遵守相关API使用条款。