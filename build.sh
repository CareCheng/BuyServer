#!/bin/bash
# ==========================================
#         User Linux 构建脚本
# ==========================================
#
# 用法:
#   ./build.sh                       # 默认编译 Linux/amd64 (外部资源模式)
#   ./build.sh --mac                 # 编译 macOS/amd64
#   ./build.sh --win                 # 编译 Windows/amd64
#   ./build.sh --all                 # 编译所有平台 (Win, Lin, Mac) 的 amd64 版本
#   ./build.sh --arm                 # 编译 arm64 架构 (配合 --all 或特定平台使用)
#   ./build.sh --all --arm --x64     # 编译所有平台的所有架构
#   ./build.sh --embed               # 嵌入模式 (单文件)
#   ./build.sh --clean               # 清理
#
# 参数:
#   --linux, --mac, --win: 指定目标操作系统
#   --all:      选中所有操作系统
#   --arm:      包含 ARM64 架构
#   --x64:      包含 AMD64 架构 (默认如果不指定 --arm)
#   --embed:    嵌入前端资源
#   --force:    强制重新构建前端
#
# ==========================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

print_info() { echo -e "${CYAN}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 解析参数
EMBED_MODE=false
FORCE_BUILD=false
CLEAN_MODE=false
BUILD_LINUX=false
BUILD_MAC=false
BUILD_WIN=false
BUILD_ALL=false
ARCH_ARM=false
ARCH_X64=false

for arg in "$@"; do
    case $arg in
        --embed|-e) EMBED_MODE=true ;;
        --force|-f) FORCE_BUILD=true ;;
        --clean|clean) CLEAN_MODE=true ;;
        --linux) BUILD_LINUX=true ;;
        --mac|--darwin|-m) BUILD_MAC=true ;;
        --win|--windows|-w) BUILD_WIN=true ;;
        --all|-a) BUILD_ALL=true ;;
        --arm) ARCH_ARM=true ;;
        --x64) ARCH_X64=true ;;
    esac
done

# 确定目标平台
TARGET_OS=""
if [ "$BUILD_ALL" = true ]; then
    TARGET_OS="linux darwin windows"
else
    [ "$BUILD_LINUX" = true ] && TARGET_OS="$TARGET_OS linux"
    [ "$BUILD_MAC" = true ] && TARGET_OS="$TARGET_OS darwin"
    [ "$BUILD_WIN" = true ] && TARGET_OS="$TARGET_OS windows"
fi

# 默认 Linux
if [ -z "$TARGET_OS" ]; then
    TARGET_OS="linux"
fi

# 确定目标架构
TARGET_ARCH=""
[ "$ARCH_ARM" = true ] && TARGET_ARCH="$TARGET_ARCH arm64"
[ "$ARCH_X64" = true ] && TARGET_ARCH="$TARGET_ARCH amd64"

# 默认 amd64
if [ -z "$TARGET_ARCH" ]; then
    TARGET_ARCH="amd64"
fi

# 编译中间文件目录
BUILD_DIR="$SCRIPT_DIR/dist/build"
WEB_BUILD_DIR="$BUILD_DIR/web"
CACHED_NODE_MODULES="$BUILD_DIR/node_modules"
DIST_ROOT="$SCRIPT_DIR/dist"
STATIC_DIR="$SCRIPT_DIR/internal/static"
EMBED_WEB_DIR="$STATIC_DIR/web"

# 清理
if [ "$CLEAN_MODE" = true ]; then
    print_info "清理构建目录..."
    rm -rf dist build
    rm -rf web/.next web/out web/node_modules
    rm -rf "$EMBED_WEB_DIR"
    print_success "清理完成"
    exit 0
fi

echo ""
print_info "=========================================="
print_info "  User Linux 构建脚本"
print_info "=========================================="
echo ""

print_info "构建目标:"
for os in $TARGET_OS; do
    for arch in $TARGET_ARCH; do
        echo "  - $os ($arch)"
    done
done
if [ "$EMBED_MODE" = true ]; then
    print_warn "  - 嵌入模式: 前端资源将打包进程序"
fi

# 检查 Go
if ! command -v go &> /dev/null; then
    print_error "未找到 Go 编译器"
    exit 1
fi
print_info "Go: $(go version)"

# 下载 Go 依赖
go mod tidy

# 检查 Node.js
SKIP_FRONTEND=""
if ! command -v npm &> /dev/null; then
    print_warn "未找到 npm，将跳过前端构建"
    SKIP_FRONTEND=true
else
    print_info "Node.js: $(node --version)"
fi

# 创建目录
mkdir -p "$BUILD_DIR"
mkdir -p "$DIST_ROOT"

# ==========================================
#         前端增量构建检测
# ==========================================

get_web_source_hash() {
    local hash_input=""
    if [ -d "web/src" ]; then
        hash_input=$(find web/src -type f -exec md5sum {} \; 2>/dev/null | sort | md5sum | cut -d' ' -f1)
    fi
    local config_files="web/package.json web/package-lock.json web/next.config.js web/next.config.mjs web/next.config.ts web/tsconfig.json web/tailwind.config.js web/tailwind.config.ts"
    local config_hash=""
    for file in $config_files; do
        if [ -f "$file" ]; then
            config_hash="$config_hash$(md5sum "$file" 2>/dev/null | cut -d' ' -f1)"
        fi
    done
    echo "${hash_input}${config_hash}" | md5sum | cut -d' ' -f1
}

web_needs_build() {
    local hash_file="$BUILD_DIR/.web_hash"
    local web_out_dir="web/out"
    if [ ! -d "$web_out_dir" ]; then return 0; fi
    if [ ! -f "$hash_file" ]; then return 0; fi
    local current_hash=$(get_web_source_hash)
    local saved_hash=$(cat "$hash_file" 2>/dev/null)
    if [ "$current_hash" != "$saved_hash" ]; then return 0; fi
    return 1
}

save_web_hash() {
    local hash_file="$BUILD_DIR/.web_hash"
    mkdir -p "$BUILD_DIR"
    get_web_source_hash > "$hash_file"
}

# ==========================================
#         构建前端
# ==========================================

print_info "========== 构建 Web 前端 =========="

if [ -n "$SKIP_FRONTEND" ]; then
    print_warn "跳过前端构建"
elif [ ! -d "web" ]; then
    print_warn "未找到 web 目录"
elif [ "$FORCE_BUILD" != true ] && ! web_needs_build; then
    print_info "前端未变更，跳过构建"
else
    local_node_modules="web/node_modules"
    
    if [ -d "$CACHED_NODE_MODULES" ] && [ ! -d "$local_node_modules" ]; then
        print_info "从缓存恢复 node_modules..."
        cp -r "$CACHED_NODE_MODULES" "$local_node_modules"
    fi
    
    if [ ! -d "$local_node_modules" ]; then
        print_info "安装前端依赖..."
        pushd web > /dev/null
        npm install --legacy-peer-deps
        popd > /dev/null
    fi
    
    print_info "构建前端..."
    pushd web > /dev/null
    npm run build
    popd > /dev/null
    
    print_success "前端构建完成"
    save_web_hash
fi

# ==========================================
#         准备嵌入式资源
# ==========================================

prepare_embed_resources() {
    print_info "========== 准备嵌入式资源 =========="
    local web_out_dir="web/out"
    if [ ! -d "$web_out_dir" ]; then
        print_error "前端构建输出目录不存在: $web_out_dir"
        print_error "请先构建前端"
        return 1
    fi
    mkdir -p "$STATIC_DIR"
    rm -rf "$EMBED_WEB_DIR"
    print_info "复制前端资源到嵌入目录..."
    cp -r "$web_out_dir" "$EMBED_WEB_DIR"
    local file_count=$(find "$EMBED_WEB_DIR" -type f | wc -l)
    local total_size=$(du -sh "$EMBED_WEB_DIR" | cut -f1)
    print_success "已准备嵌入式资源: $file_count 个文件, $total_size"
    return 0
}

cleanup_embed_resources() {
    if [ -d "$EMBED_WEB_DIR" ]; then
        rm -rf "$EMBED_WEB_DIR"
        print_info "已清理嵌入式资源目录"
    fi
}

# ==========================================
#         构建循环
# ==========================================

if [ "$EMBED_MODE" = true ]; then
    if ! prepare_embed_resources; then exit 1; fi
fi

START_TIME=$(date +%s)

for os in $TARGET_OS; do
    for arch in $TARGET_ARCH; do
        
        print_info "========== 构建 $os ($arch) =========="
        
        dir_name="${os}_${arch}"
        if [ "$os" = "darwin" ]; then dir_name="macos_${arch}"; fi
        
        TARGET_DIR="$DIST_ROOT/$dir_name"
        
        # 清理
        rm -rf "$TARGET_DIR"
        mkdir -p "$TARGET_DIR"
        
        # 复制资源 (非嵌入模式)
        if [ "$EMBED_MODE" != true ] && [ -d "web/out" ] && [ -z "$SKIP_FRONTEND" ]; then
            cp -r "web/out" "$TARGET_DIR/web"
            print_info "已复制前端资源"
        fi
        
        # 创建必要目录
        mkdir -p "$TARGET_DIR/user_config"
        mkdir -p "$TARGET_DIR/Product"
        mkdir -p "$TARGET_DIR/backups"
        
        # 编译
        BIN_NAME="UserFrontend"
        if [ "$os" = "windows" ]; then BIN_NAME="UserFrontend.exe"; fi
        
        BUILD_TAGS=""
        if [ "$EMBED_MODE" = true ]; then BUILD_TAGS="-tags embed"; fi
        
        print_info "编译 Go ($os/$arch)..."
        CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -ldflags="-s -w" $BUILD_TAGS -o "$TARGET_DIR/$BIN_NAME" ./cmd/server
        
        if [ $? -eq 0 ]; then
            FILE_SIZE=$(du -h "$TARGET_DIR/$BIN_NAME" | cut -f1)
            print_success "生成: $BIN_NAME ($FILE_SIZE)"
        else
            print_error "编译失败"
            exit 1
        fi
        
        # 启动脚本
        START_SCRIPT="$TARGET_DIR/start.sh"
        if [ "$os" = "windows" ]; then START_SCRIPT="$TARGET_DIR/start.bat"; fi
        
        if [ "$os" = "windows" ]; then
            cat > "$START_SCRIPT" << EOF
@echo off
title User Frontend
echo ========================================
echo   User Frontend - Starting...
echo ========================================
echo.
echo 访问地址: http://localhost:8080/
echo.
"%~dp0$BIN_NAME"
pause
EOF
        else
            cat > "$START_SCRIPT" << EOF
#!/bin/bash
echo "========================================"
echo "  User Frontend - Starting..."
echo "========================================"
echo ""
echo "访问地址: http://localhost:8080/"
echo ""
chmod +x ./$BIN_NAME
./$BIN_NAME
EOF
            chmod +x "$START_SCRIPT"
        fi
        
    done
done

if [ "$EMBED_MODE" = true ]; then
    cleanup_embed_resources
fi

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo ""
print_info "=========================================="
print_success "  构建全部完成! (耗时: ${DURATION}s)"
print_info "=========================================="
print_info "输出目录: $DIST_ROOT"
ls -F "$DIST_ROOT"
