#!/bin/bash

# 汇付支付配置管理系统打包脚本
# 支持多平台编译：Windows(exe), Linux, macOS (自动判断架构)

set -e

echo "🏗️  开始构建汇付支付配置管理系统..."

# 项目信息
APP_NAME="ghuifu"
VERSION="1.0.0"
BUILD_TIME=$(date '+%Y%m%d-%H%M%S')

# 清理构建目录
BUILD_DIR="./build"
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# 获取依赖
echo "📦 获取Go模块依赖..."
go mod tidy

# 编译目标平台 (3个平台，Linux/macOS会自动判断架构)
TARGETS=(
    "windows/amd64"
    "linux/multi"
    "darwin/multi"  
)

echo "🔨 开始多平台编译..."

for target in "${TARGETS[@]}"; do
    IFS="/" read -r GOOS GOARCH <<< "$target"
    
    if [ "$GOARCH" = "multi" ]; then
        echo "  📋 编译 $GOOS (多架构)..."
        
        # Linux和macOS支持多架构
        if [ "$GOOS" = "linux" ]; then
            ARCHS=("amd64" "arm64")
        else # darwin
            ARCHS=("amd64" "arm64")
        fi
        
        # 设置输出目录和文件名 (不包含架构)
        OUTPUT_DIR="$BUILD_DIR/${APP_NAME}-${GOOS}"
        mkdir -p "$OUTPUT_DIR"
        
        # 编译所有架构并打包到同一目录
        for arch in "${ARCHS[@]}"; do
            echo "    🔧 编译 $GOOS/$arch..."
            BINARY_NAME="${APP_NAME}-${arch}"
            
            export GOOS=$GOOS
            export GOARCH=$arch
            export CGO_ENABLED=0
            
            go build -ldflags "-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" \
                -o "$OUTPUT_DIR/$BINARY_NAME" .
        done
        
        # 创建启动脚本 (自动判断架构)
        cat > "$OUTPUT_DIR/start.sh" << 'EOF'
#!/bin/bash

# 自动检测系统架构并启动对应版本
detect_arch() {
    case $(uname -m) in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            echo "amd64"  # 默认使用 amd64
            ;;
    esac
}

ARCH=$(detect_arch)
BINARY="./ghuifu-$ARCH"

echo "🏗️  启动汇付支付配置管理系统..."
echo "📋 检测到系统架构: $ARCH"
echo "🌐 服务将运行在: http://localhost:\${PORT:-40004}"
echo "⏹️  按 Ctrl+C 停止服务"
echo ""

if [ -f "$BINARY" ]; then
    "$BINARY"
else
    echo "❌ 错误: 找不到对应架构的可执行文件 ($BINARY)"
    echo "📁 可用的文件:"
    ls -la ./ghuifu-*
    exit 1
fi
EOF
        chmod +x "$OUTPUT_DIR/start.sh"
        chmod +x "$OUTPUT_DIR/ghuifu-"*
        
        # 创建安装脚本
        cat > "$OUTPUT_DIR/install.sh" << 'EOF'
#!/bin/bash

# 汇付支付配置管理系统安装脚本
# 支持 systemd 服务管理

set -e

APP_NAME="ghuifu"
SERVICE_NAME="ghuifu-payment-config"
INSTALL_DIR="/opt/ghuifu"
SERVICE_USER="ghuifu"

echo "🏗️  开始安装汇付支付配置管理系统..."

# 检查是否以root权限运行
if [ "$EUID" -ne 0 ]; then
    echo "❌ 请以root权限运行此安装脚本"
    echo "   sudo ./install.sh"
    exit 1
fi

# 检查systemd是否可用
if ! command -v systemctl &> /dev/null; then
    echo "❌ 系统不支持 systemd，请手动运行 ./start.sh"
    exit 1
fi

# 配置服务端口
echo "⚙️  配置服务参数："
echo ""

# 端口配置
while true; do
    echo "🌐 服务端口配置："
    echo "   默认端口: 40004"
    echo "   常用端口: 3000, 8000, 8080, 8888, 9000, 40004"
    echo ""
    read -p "请输入服务端口 (直接回车使用默认端口 40004): " SERVICE_PORT
    
    # 如果用户直接回车，使用默认端口
    if [ -z "$SERVICE_PORT" ]; then
        SERVICE_PORT=40004
        echo "✅ 使用默认端口: $SERVICE_PORT"
        break
    fi
    
    # 验证端口号是否有效
    if [[ "$SERVICE_PORT" =~ ^[0-9]+$ ]] && [ "$SERVICE_PORT" -ge 1 ] && [ "$SERVICE_PORT" -le 65535 ]; then
        # 检查端口是否被占用
        if netstat -tuln 2>/dev/null | grep -q ":$SERVICE_PORT "; then
            echo "⚠️  警告: 端口 $SERVICE_PORT 已被占用"
            read -p "是否继续使用此端口? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                echo "✅ 使用端口: $SERVICE_PORT (端口已被占用，请确保稍后处理)"
                break
            else
                continue
            fi
        else
            echo "✅ 使用端口: $SERVICE_PORT"
            break
        fi
    else
        echo "❌ 无效的端口号，请输入 1-65535 之间的数字"
    fi
done

echo ""
echo "📋 安装步骤："
echo "   1. 创建系统用户"
echo "   2. 安装文件到 $INSTALL_DIR"
echo "   3. 创建 systemd 服务 (端口: $SERVICE_PORT)"
echo "   4. 配置服务自启动"
echo ""

# 创建系统用户
if ! id "$SERVICE_USER" &>/dev/null; then
    echo "👤 创建系统用户: $SERVICE_USER"
    useradd --system --no-create-home --shell /bin/false $SERVICE_USER
else
    echo "👤 用户 $SERVICE_USER 已存在"
fi

# 创建安装目录
echo "📁 创建安装目录: $INSTALL_DIR"
mkdir -p $INSTALL_DIR
cp -r ./* $INSTALL_DIR/
chown -R $SERVICE_USER:$SERVICE_USER $INSTALL_DIR
chmod +x $INSTALL_DIR/ghuifu-*
chmod +x $INSTALL_DIR/start.sh

# 检测系统架构
detect_arch() {
    case $(uname -m) in
        x86_64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) echo "amd64" ;;
    esac
}

ARCH=$(detect_arch)
BINARY_PATH="$INSTALL_DIR/ghuifu-$ARCH"

if [ ! -f "$BINARY_PATH" ]; then
    echo "❌ 错误: 找不到对应架构的可执行文件 ($BINARY_PATH)"
    exit 1
fi

echo "🔧 创建 systemd 服务文件..."
cat > /etc/systemd/system/$SERVICE_NAME.service << SERVICEOF
[Unit]
Description=汇付支付配置管理系统
After=network.target
Wants=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$BINARY_PATH
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# 环境变量
Environment=GIN_MODE=release
Environment=PORT=$SERVICE_PORT

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$INSTALL_DIR

[Install]
WantedBy=multi-user.target
SERVICEOF

# 重新加载 systemd
echo "🔄 重新加载 systemd 配置..."
systemctl daemon-reload

# 启用服务自启动
echo "🚀 启用服务自启动..."
systemctl enable $SERVICE_NAME

echo ""
echo "✅ 安装完成！"
echo ""
echo "📋 服务管理命令："
echo "   启动服务: sudo systemctl start $SERVICE_NAME"
echo "   停止服务: sudo systemctl stop $SERVICE_NAME"
echo "   重启服务: sudo systemctl restart $SERVICE_NAME"
echo "   查看状态: sudo systemctl status $SERVICE_NAME"
echo "   查看日志: sudo journalctl -u $SERVICE_NAME -f"
echo ""
echo "🌐 服务地址: http://localhost:$SERVICE_PORT"
echo ""
echo "💡 提示："
echo "   • 服务已设置为开机自启动"
echo "   • 日志通过 journalctl 管理"
echo "   • 配置文件位于: $INSTALL_DIR"
echo ""

read -p "是否现在启动服务? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    systemctl start $SERVICE_NAME
    echo "🎉 服务已启动！"
    systemctl status $SERVICE_NAME --no-pager
else
    echo "💭 稍后可以使用以下命令启动服务:"
    echo "   sudo systemctl start $SERVICE_NAME"
fi
EOF
        chmod +x "$OUTPUT_DIR/install.sh"
        
        # 创建卸载脚本
        cat > "$OUTPUT_DIR/uninstall.sh" << 'EOF'
#!/bin/bash

# 汇付支付配置管理系统卸载脚本

set -e

SERVICE_NAME="ghuifu-payment-config"
INSTALL_DIR="/opt/ghuifu"
SERVICE_USER="ghuifu"

echo "🗑️  开始卸载汇付支付配置管理系统..."

# 检查是否以root权限运行
if [ "$EUID" -ne 0 ]; then
    echo "❌ 请以root权限运行此卸载脚本"
    echo "   sudo ./uninstall.sh"
    exit 1
fi

echo "⚠️  警告: 此操作将完全删除汇付支付配置管理系统"
echo "📁 将删除目录: $INSTALL_DIR"
echo "👤 将删除用户: $SERVICE_USER"
echo "🔧 将删除服务: $SERVICE_NAME"
echo ""

read -p "确定要继续卸载吗? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ 卸载已取消"
    exit 0
fi

# 停止并禁用服务
if systemctl is-active --quiet $SERVICE_NAME; then
    echo "🛑 停止服务..."
    systemctl stop $SERVICE_NAME
fi

if systemctl is-enabled --quiet $SERVICE_NAME; then
    echo "❌ 禁用服务自启动..."
    systemctl disable $SERVICE_NAME
fi

# 删除服务文件
if [ -f "/etc/systemd/system/$SERVICE_NAME.service" ]; then
    echo "🗑️  删除服务文件..."
    rm -f "/etc/systemd/system/$SERVICE_NAME.service"
    systemctl daemon-reload
fi

# 删除安装目录
if [ -d "$INSTALL_DIR" ]; then
    echo "📁 删除安装目录..."
    rm -rf "$INSTALL_DIR"
fi

# 删除系统用户
if id "$SERVICE_USER" &>/dev/null; then
    echo "👤 删除系统用户..."
    userdel "$SERVICE_USER" 2>/dev/null || true
fi

echo ""
echo "✅ 卸载完成！"
echo ""
echo "💡 系统已完全清理，如需重新安装请重新运行 install.sh"
EOF
        chmod +x "$OUTPUT_DIR/uninstall.sh"
        
        # macOS 特殊处理 - 使用 launchd
        if [ "$GOOS" = "darwin" ]; then
            # 创建 macOS 服务安装脚本
            cat > "$OUTPUT_DIR/install-service-macos.sh" << 'EOF'
#!/bin/bash

# macOS 系统服务安装脚本 (使用 launchd)

set -e

SERVICE_NAME="com.ghuifu.payment-config"
INSTALL_DIR="/usr/local/opt/ghuifu"
PLIST_FILE="$HOME/Library/LaunchAgents/$SERVICE_NAME.plist"

echo "🏗️  安装 macOS 系统服务..."

# 创建安装目录
echo "📁 创建安装目录: $INSTALL_DIR"
sudo mkdir -p $INSTALL_DIR
sudo cp -r ./* $INSTALL_DIR/
sudo chmod +x $INSTALL_DIR/ghuifu-*
sudo chmod +x $INSTALL_DIR/start.sh

# 检测架构
detect_arch() {
    case $(uname -m) in
        x86_64) echo "amd64" ;;
        arm64) echo "arm64" ;;
        *) echo "amd64" ;;
    esac
}

ARCH=$(detect_arch)
BINARY_PATH="$INSTALL_DIR/ghuifu-$ARCH"

# 创建 LaunchAgent plist 文件
echo "🔧 创建 LaunchAgent 配置..."
mkdir -p "$HOME/Library/LaunchAgents"

cat > "$PLIST_FILE" << PLISTEOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>$SERVICE_NAME</string>
    <key>ProgramArguments</key>
    <array>
        <string>$BINARY_PATH</string>
    </array>
    <key>WorkingDirectory</key>
    <string>$INSTALL_DIR</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/usr/local/var/log/ghuifu.log</string>
    <key>StandardErrorPath</key>
    <string>/usr/local/var/log/ghuifu.error.log</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>GIN_MODE</key>
        <string>release</string>
        <key>PORT</key>
        <string>$SERVICE_PORT</string>
    </dict>
</dict>
</plist>
PLISTEOF

# 加载服务
echo "🚀 加载服务..."
launchctl load "$PLIST_FILE"

echo ""
echo "✅ macOS 服务安装完成！"
echo ""
echo "📋 服务管理命令："
echo "   启动服务: launchctl start $SERVICE_NAME"
echo "   停止服务: launchctl stop $SERVICE_NAME"
echo "   卸载服务: launchctl unload $PLIST_FILE"
echo "   查看日志: tail -f /usr/local/var/log/ghuifu.log"
echo ""
echo "🌐 服务地址: http://localhost:$SERVICE_PORT"
EOF
            chmod +x "$OUTPUT_DIR/install-service-macos.sh"
        fi
        
    else
        # Windows 单架构处理
        echo "  📋 编译 $GOOS/$GOARCH..."
        
        OUTPUT_DIR="$BUILD_DIR/${APP_NAME}-${GOOS}"
        mkdir -p "$OUTPUT_DIR"
        
        BINARY_NAME="${APP_NAME}.exe"
        
        export GOOS=$GOOS
        export GOARCH=$GOARCH
        export CGO_ENABLED=0
        
        go build -ldflags "-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" \
            -o "$OUTPUT_DIR/$BINARY_NAME" .
        
        # Windows 启动脚本
        cat > "$OUTPUT_DIR/start.bat" << 'EOF'
@echo off
title 汇付支付配置管理系统
echo 🏗️  启动汇付支付配置管理系统...
echo 🌐 服务将运行在: http://localhost:%PORT%（默认40004）
echo ⏹️  按 Ctrl+C 停止服务
echo.
%~dp0ghuifu.exe
pause
EOF
        
        # Windows 安装脚本
        cat > "$OUTPUT_DIR/install.bat" << 'EOF'
@echo off
title 安装汇付支付配置管理系统
echo 🏗️  安装汇付支付配置管理系统...
echo.
echo 📋 安装方式：
echo    方式1: 手动运行 - 运行 start.bat
echo    方式2: Windows服务 - 运行 install-service.bat (需管理员权限)
echo.
echo 💡 推荐使用Windows服务方式，可以开机自启动
echo.
echo ✅ 文件已就绪！请选择合适的安装方式
pause
EOF
        
        # Windows 服务安装脚本
        cat > "$OUTPUT_DIR/install-service.bat" << 'EOF'
@echo off
title 安装Windows服务
echo 🏗️  安装汇付支付配置管理系统为Windows服务...

REM 检查管理员权限
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 此脚本需要管理员权限运行
    echo    请右键点击并选择"以管理员身份运行"
    pause
    exit /b 1
)

set SERVICE_NAME=GhuifuPaymentConfig
set SERVICE_DISPLAY_NAME=汇付支付配置管理系统
set INSTALL_DIR=%ProgramFiles%\Ghuifu
set BINARY_PATH=%INSTALL_DIR%\ghuifu.exe

echo 📁 创建安装目录: %INSTALL_DIR%
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

echo 📋 复制程序文件...
xcopy /E /I /Y * "%INSTALL_DIR%\"

echo 🔧 安装Windows服务...
sc create "%SERVICE_NAME%" ^
    binPath= "\"%BINARY_PATH%\"" ^
    DisplayName= "%SERVICE_DISPLAY_NAME%" ^
    start= auto ^
    depend= Tcpip

if %errorlevel% equ 0 (
    echo ✅ 服务安装成功！
    
    echo.
    echo 📋 服务管理命令：
    echo    启动服务: sc start %SERVICE_NAME%
    echo    停止服务: sc stop %SERVICE_NAME%
    echo    删除服务: sc delete %SERVICE_NAME%
    echo    查看服务: services.msc
    echo.
    echo 🌐 服务地址: http://localhost:40004
    echo.
    
    set /p choice="是否现在启动服务? (Y/N): "
    if /i "%choice%"=="Y" (
        sc start "%SERVICE_NAME%"
        echo 🎉 服务已启动！
    )
) else (
    echo ❌ 服务安装失败
    echo 💡 请检查是否已安装或使用手动方式运行 start.bat
)

echo.
pause
EOF
        
        # Windows 服务卸载脚本
        cat > "$OUTPUT_DIR/uninstall-service.bat" << 'EOF'
@echo off
title 卸载Windows服务
echo 🗑️  卸载汇付支付配置管理系统服务...

REM 检查管理员权限
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 此脚本需要管理员权限运行
    echo    请右键点击并选择"以管理员身份运行"
    pause
    exit /b 1
)

set SERVICE_NAME=GhuifuPaymentConfig
set INSTALL_DIR=%ProgramFiles%\Ghuifu

echo ⚠️  警告: 此操作将完全删除汇付支付配置管理系统
echo 📁 将删除目录: %INSTALL_DIR%
echo 🔧 将删除服务: %SERVICE_NAME%
echo.

set /p choice="确定要继续卸载吗? (Y/N): "
if /i not "%choice%"=="Y" (
    echo ❌ 卸载已取消
    pause
    exit /b 0
)

echo 🛑 停止服务...
sc stop "%SERVICE_NAME%" >nul 2>&1

echo 🗑️  删除服务...
sc delete "%SERVICE_NAME%"

echo 📁 删除安装目录...
if exist "%INSTALL_DIR%" rmdir /s /q "%INSTALL_DIR%"

echo.
echo ✅ 卸载完成！
echo.
pause
EOF
    fi
    
    # 复制静态文件和文档
    echo "    📁 复制静态文件..."
    cp -r static "$OUTPUT_DIR/"
    cp DEPLOY.md "$OUTPUT_DIR/" 2>/dev/null || true
    cp README.md "$OUTPUT_DIR/" 2>/dev/null || true
    
    # 清理 macOS 系统生成的隐藏文件
    find "$OUTPUT_DIR" -name "._*" -delete 2>/dev/null || true
    find "$OUTPUT_DIR" -name ".DS_Store" -delete 2>/dev/null || true
    
    # 创建配置示例文件
    cat > "$OUTPUT_DIR/config.example.json" << 'EOF'
{
  "sys_id": "your_system_id",
  "product_id": "your_product_id", 
  "rsa_merch_private_key": "your_rsa_private_key_content",
  "rsa_huifu_public_key": "huifu_public_key_content"
}
EOF
    
    # 打包压缩
    echo "    📦 打包 $GOOS..."
    cd "$BUILD_DIR"
    
    if [ "$GOOS" = "windows" ]; then
        # Windows 使用 zip，排除隐藏文件
        zip -r "${APP_NAME}-${VERSION}-${GOOS}.zip" "${APP_NAME}-${GOOS}/" -x "*/.*" "*/.DS_Store" > /dev/null
    else
        # Unix/Linux 使用 tar.gz，排除隐藏文件和 macOS 系统文件
        tar -czf "${APP_NAME}-${VERSION}-${GOOS}.tar.gz" \
            --exclude="._*" \
            --exclude=".DS_Store" \
            --exclude="*/._*" \
            --exclude="*/.DS_Store" \
            "${APP_NAME}-${GOOS}/"
    fi
    
    cd ..
    echo "    ✅ ${APP_NAME}-${VERSION}-${GOOS} 打包完成"
done

echo ""
echo "🎉 构建完成！输出文件："
echo "📂 构建目录: $BUILD_DIR"
echo ""

# 显示生成的文件
ls -la "$BUILD_DIR"/*.{zip,tar.gz} 2>/dev/null || echo "没有找到压缩包文件"

echo ""
echo "📋 生成的压缩包："
echo "  • Windows (64位): ${APP_NAME}-${VERSION}-windows.zip"
echo "  • Linux (自动判断架构): ${APP_NAME}-${VERSION}-linux.tar.gz"
echo "  • macOS (自动判断架构): ${APP_NAME}-${VERSION}-darwin.tar.gz"
echo ""
echo "🚀 部署方式："
echo "  1. 选择对应操作系统的压缩包"
echo "  2. 解压到目标目录"
echo "  3. 运行 start.sh/start.bat 启动服务"
echo "  4. 访问 http://localhost:40004 使用系统（默认端口）"
echo ""
echo "💡 特性说明："
echo "  • Linux/macOS版本会自动检测系统架构(AMD64/ARM64)"
echo "  • Windows版本为64位通用版本"
echo "  • 所有版本都包含完整的静态文件和文档"
echo ""
echo "✨ 构建完成时间: $(date)"