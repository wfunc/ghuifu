# 汇付支付配置管理系统 - 部署说明

## 📋 系统概述

汇付支付配置管理系统是一个基于Web的配置管理工具，用于管理汇付支付SDK的配置信息，支持微信小程序商户配置和查询。

### 🏗️ 技术栈
- **后端**: Golang + Gin框架
- **前端**: HTML5 + JavaScript (原生)
- **SDK**: 汇付bspay-go-sdk v1.0.20
- **数据存储**: 内存存储 (配置信息)

---

## 🚀 快速部署

### 1. 下载对应平台的压缩包

根据你的服务器平台选择相应的压缩包：

- **Windows 64位**: `ghuifu-1.0.0-windows.zip`
- **Linux (自动检测架构)**: `ghuifu-1.0.0-linux.tar.gz` (支持AMD64和ARM64)
- **macOS (自动检测架构)**: `ghuifu-1.0.0-darwin.tar.gz` (支持Intel和Apple Silicon)

### 2. 解压文件

**Windows:**
```cmd
# 解压 zip 文件到目标目录
unzip ghuifu-1.0.0-windows.zip
cd ghuifu-windows
```

**Linux/macOS:**
```bash
# Linux系统解压
tar -xzf ghuifu-1.0.0-linux.tar.gz
cd ghuifu-linux

# macOS系统解压
tar -xzf ghuifu-1.0.0-darwin.tar.gz
cd ghuifu-darwin
```

### 3. 运行系统

#### 🚀 快速启动（推荐）

**Windows:**
```cmd
# 方式1: 直接运行（临时）
start.bat

# 方式2: 安装为Windows服务（推荐）
# 右键"以管理员身份运行"
install-service.bat
```

**Linux:**
```bash
# 方式1: 直接运行（临时）
./start.sh

# 方式2: 安装为systemd服务（推荐）
sudo ./install.sh
```

**macOS:**
```bash
# 方式1: 直接运行（临时）
./start.sh

# 方式2: 安装为LaunchAgent服务（推荐）
./install-service-macos.sh
```

#### 🔧 服务管理（推荐部署方式）

**Linux (systemd):**
```bash
# 安装服务
sudo ./install.sh

# 管理服务
sudo systemctl start ghuifu-payment-config     # 启动
sudo systemctl stop ghuifu-payment-config      # 停止
sudo systemctl restart ghuifu-payment-config   # 重启
sudo systemctl status ghuifu-payment-config    # 状态
sudo journalctl -u ghuifu-payment-config -f    # 查看日志

# 卸载服务
sudo ./uninstall.sh
```

**Windows (系统服务):**
```cmd
# 安装服务（需管理员权限）
install-service.bat

# 管理服务（需管理员权限）
sc start GhuifuPaymentConfig      # 启动
sc stop GhuifuPaymentConfig       # 停止  
sc query GhuifuPaymentConfig      # 状态
services.msc                      # 图形界面管理

# 卸载服务（需管理员权限）
uninstall-service.bat
```

**macOS (LaunchAgent):**
```bash
# 安装服务
./install-service-macos.sh

# 管理服务
launchctl start com.ghuifu.payment-config    # 启动
launchctl stop com.ghuifu.payment-config     # 停止
tail -f /usr/local/var/log/ghuifu.log        # 查看日志

# 卸载服务
launchctl unload ~/Library/LaunchAgents/com.ghuifu.payment-config.plist
```

### 4. 访问系统

打开浏览器访问: http://localhost:40004

---

## 📁 文件结构

解压后的目录结构：
```
ghuifu-{platform}/
├── ghuifu(.exe) 或 ghuifu-amd64/ghuifu-arm64  # 主程序可执行文件
├── static/                                    # 前端静态文件
│   ├── index.html                            # 主页面
│   └── app.js                               # JavaScript逻辑
├── config.example.json                       # 配置文件示例
├── start.sh/start.bat                       # 智能启动脚本
├── install.sh/install.bat                   # 安装脚本(可选)
├── README.md                                # 说明文档
└── DEPLOY.md                               # 本部署文档
```

**架构自动检测说明**：
- **Linux/macOS版本**: 包含`ghuifu-amd64`和`ghuifu-arm64`两个可执行文件，启动脚本会自动检测系统架构并运行对应版本
- **Windows版本**: 包含单一的`ghuifu.exe`文件（64位通用版本）

**服务管理脚本说明**：
- **Linux**: `install.sh` (systemd服务)、`uninstall.sh` (卸载服务)
- **macOS**: `install-service-macos.sh` (LaunchAgent服务)
- **Windows**: `install-service.bat` (Windows服务)、`uninstall-service.bat` (卸载服务)

---

## ⚙️ 配置说明

### 环境要求

- **端口**: 确保40004端口未被占用
- **网络**: 需要访问汇付支付API的网络连接
- **权限**: 需要创建临时配置文件的写入权限

### 配置参数说明

系统支持以下配置参数：

| 参数名 | 说明 | 是否必填 |
|--------|------|----------|
| sys_id | 系统ID | ✅ |
| product_id | 产品ID | ✅ |
| rsa_private_key | RSA私钥 | ✅ |
| environment | 环境(test/production) | ✅ |
| huifu_id | 汇付ID | ✅ (微信配置时) |
| wx_woa_app_id | 微信小程序AppID | ✅ (微信配置时) |
| wx_woa_path | 微信小程序路径 | ✅ (微信配置时) |
| fee_type | 费率类型 | ✅ (微信配置时) |

### 费率类型选项

| 代码 | 说明 |
|------|------|
| 01 | 标准费率线上（支持统一进件页面版） |
| 02 | 标准费率线下（支持统一进件页面版） |
| 03 | 非盈利行业费率 |
| 04 | 缴费行业费率 |
| 05 | 保险行业费率 |
| 06 | 行业（蓝海绿洲）活动场景费率 |
| 07 | 校园餐饮费率 |
| 08 | K12中小幼费率 |
| 09 | 非在线教培行业费率 |

---

## 🔧 使用指南

### 1. 配置系统参数
1. 访问系统主页
2. 在"配置信息"区域填入系统参数：
   - 系统ID (sys_id)
   - 产品ID (product_id) 
   - RSA私钥 (可使用"生成测试密钥"功能)
   - 选择环境 (测试/生产)
3. 点击"保存配置"

### 2. 配置微信商户
1. 在"微信商户配置"区域：
   - 选择已保存的系统配置
   - 填入汇付ID
   - 填入微信小程序AppID
   - 填入微信小程序路径
   - 选择费率类型
2. 点击"配置授权目录"

### 3. 查询微信配置
1. 选择系统配置
2. 输入汇付ID
3. 点击"查询微信配置"

### 4. 管理配置
- 在"已保存的配置"区域可以查看、选择和删除配置
- 支持删除不需要的配置

---

## 🔒 安全说明

### RSA密钥管理
- **测试环境**: 可以使用系统生成的测试密钥
- **生产环境**: 必须使用真实的RSA私钥
- **密钥格式**: 支持PKCS#1和PKCS#8格式
- **安全存储**: 密钥仅在内存中临时存储，重启后清空

### 网络安全
- 系统默认监听本地40004端口
- 生产环境建议配置反向代理和HTTPS
- 建议通过防火墙限制访问

### 数据安全
- 配置信息存储在内存中，重启后会清空
- 不会在磁盘上永久存储敏感信息
- SDK临时配置文件在使用后自动清理

---

## 🐛 故障排除

### 常见问题

**Q1: 端口8080被占用**
```bash
# Linux/macOS 查看端口占用
lsof -i:40004

# Windows 查看端口占用
netstat -ano | findstr 40004

# 解决方案：停止占用进程或修改源码中的端口号重新编译
```

**Q2: 权限问题 (Linux/macOS)**
```bash
# 给予执行权限
chmod +x ghuifu
chmod +x start.sh

# 如果需要绑定80端口，使用sudo
sudo ./ghuifu
```

**Q3: SDK初始化失败**
- 检查RSA私钥格式是否正确
- 确认sys_id和product_id是否有效
- 检查网络连接是否正常
- 查看控制台日志获取详细错误信息

**Q4: 静态文件404**
- 确保static/目录存在且包含index.html和app.js
- 检查文件权限是否正确

### 日志查看

系统运行时会在控制台输出详细日志，包括：
- API请求和响应
- SDK初始化状态
- 错误信息和调试信息

如需保存日志到文件：
```bash
# Linux/macOS
./ghuifu > ghuifu.log 2>&1

# Windows
ghuifu.exe > ghuifu.log 2>&1
```

---

## 🔄 系统升级

### 升级步骤
1. 停止当前运行的服务
2. 备份当前目录（可选）
3. 下载新版本压缩包
4. 解压并替换文件
5. 重新启动服务

### 配置迁移
- 系统配置存储在内存中，升级后需要重新配置
- 如需保留配置，建议记录相关参数

---

## 📞 技术支持

### 系统信息
- **版本**: v1.0.0
- **Go版本**: 1.19+
- **汇付SDK**: bspay-go-sdk v1.0.20

### 功能特性
- ✅ 多平台支持 (Windows/Linux/macOS)
- ✅ 实时配置管理
- ✅ 微信商户配置和查询
- ✅ RSA密钥生成和管理
- ✅ 响应式Web界面
- ✅ 错误处理和日志记录

### API文档参考
- 汇付支付开放平台: https://paas.huifu.com/open/doc
- 微信商户配置API: https://paas.huifu.com/open/doc/api/#/shgl/shjj/api_shjj_wxshpz

---

**部署完成后，即可通过浏览器访问 http://localhost:40004 开始使用汇付支付配置管理系统！** 🎉