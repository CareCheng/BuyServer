#!/bin/bash
# ==========================================
#         User Linux 构建脚本
# ==========================================
#
# 用法:
#   ./build.sh              # 默认编译 Linux 版本（外部资源模式）
#   ./build.sh --embed      # 将前端资源嵌入到程序中（单文件模式）
#   ./build.sh --clean      # 清理构建目录
#   ./build.sh --force      # 强制重新构建前端
#
# 构建模式:
#   外部资源模式（默认）：前端资源作为独立文件，程序从 ./web/ 目录加载
#   嵌入模式（--embed）：前端资源打包进二进制文件，生成单个可执行文件
#
# 目录结构:
#   dist/
#   ├── build/              # 编译中间文件
#   │   ├── web/            # 前端 .next
#   │   └── node_modules/   # 前端依赖
#   └── linux/              # Linux 最终输出
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

for arg in "$@"; do
    case $arg in
        --embed|-e)
            EMBED_MODE=true
            ;;
        --force|-f)
            FORCE_BUILD=true
            ;;
        --clean|clean)
            CLEAN_MODE=true
            ;;
    esac
done

# 编译中间文件目录
BUILD_DIR="$SCRIPT_DIR/dist/build"
WEB_BUILD_DIR="$BUILD_DIR/web"
CACHED_NODE_MODULES="$BUILD_DIR/node_modules"
DIST_DIR="$SCRIPT_DIR/dist/linux"
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

if [ "$EMBED_MODE" = true ]; then
    print_info "构建模式: 嵌入模式（单文件）"
else
    print_info "构建模式: 外部资源模式"
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
mkdir -p "$DIST_DIR"

# ==========================================
#         前端增量构建检测
# ==========================================

# 计算前端源码哈希值
get_web_source_hash() {
    local hash_input=""
    
    # 收集 web/src 目录下所有文件内容
    if [ -d "web/src" ]; then
        hash_input=$(find web/src -type f -exec md5sum {} \; 2>/dev/null | sort | md5sum | cut -d' ' -f1)
    fi
    
    # 添加关键配置文件
    local config_files="web/package.json web/package-lock.json web/next.config.js web/next.config.mjs web/next.config.ts web/tsconfig.json web/tailwind.config.js web/tailwind.config.ts"
    local config_hash=""
    for file in $config_files; do
        if [ -f "$file" ]; then
            config_hash="$config_hash$(md5sum "$file" 2>/dev/null | cut -d' ' -f1)"
        fi
    done
    
    # 组合哈希
    echo "${hash_input}${config_hash}" | md5sum | cut -d' ' -f1
}

# 检查前端是否需要重新构建
web_needs_build() {
    local hash_file="$BUILD_DIR/.web_hash"
    local web_out_dir="web/out"
    
    # 如果输出目录不存在，需要构建
    if [ ! -d "$web_out_dir" ]; then
        return 0
    fi
    
    # 如果哈希文件不存在，需要构建
    if [ ! -f "$hash_file" ]; then
        return 0
    fi
    
    # 比较哈希值
    local current_hash=$(get_web_source_hash)
    local saved_hash=$(cat "$hash_file" 2>/dev/null)
    
    if [ "$current_hash" != "$saved_hash" ]; then
        return 0
    fi
    
    return 1
}

# 保存前端哈希值
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
    
    # 如果缓存的 node_modules 存在，复制到 web 目录
    if [ -d "$CACHED_NODE_MODULES" ] && [ ! -d "$local_node_modules" ]; then
        print_info "从缓存恢复 node_modules..."
        cp -r "$CACHED_NODE_MODULES" "$local_node_modules"
    fi
    
    # 安装依赖
    if [ ! -d "$local_node_modules" ]; then
        print_info "安装前端依赖..."
        pushd web > /dev/null
        npm install --legacy-peer-deps
        popd > /dev/null
    fi
    
    # 构建前端
    print_info "构建前端..."
    pushd web > /dev/null
    npm run build
    popd > /dev/null
    
    print_success "前端构建完成"
    
    # 保存哈希值
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
    
    # 确保 static 目录存在
    mkdir -p "$STATIC_DIR"
    
    # 清理旧的嵌入资源
    rm -rf "$EMBED_WEB_DIR"
    
    # 复制前端资源到嵌入目录
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
#         创建必要目录
# ==========================================

mkdir -p "$DIST_DIR/user_config"
mkdir -p "$DIST_DIR/Product"
mkdir -p "$DIST_DIR/backups"

# ==========================================
#         嵌入模式：准备资源
# ==========================================

if [ "$EMBED_MODE" = true ]; then
    if ! prepare_embed_resources; then
        exit 1
    fi
fi

# ==========================================
#         编译 Go
# ==========================================

print_info "========== 编译 Go 程序 =========="

BUILD_TAGS=""
if [ "$EMBED_MODE" = true ]; then
    BUILD_TAGS="-tags embed"
    print_info "编译 Go 程序（嵌入模式）..."
else
    print_info "编译 Go 程序..."
fi

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" $BUILD_TAGS -o "$DIST_DIR/UserFrontend" ./cmd/server

# 获取文件大小
FILE_SIZE=$(ls -lh "$DIST_DIR/UserFrontend" | awk '{print $5}')

if [ "$EMBED_MODE" = true ]; then
    print_success "编译完成: $DIST_DIR/UserFrontend ($FILE_SIZE) [嵌入模式]"
else
    print_success "编译完成: $DIST_DIR/UserFrontend ($FILE_SIZE) [外部资源模式]"
fi

# ==========================================
#         嵌入模式：清理临时资源
# ==========================================

if [ "$EMBED_MODE" = true ]; then
    cleanup_embed_resources
fi

# ==========================================
#         复制前端资源（非嵌入模式）
# ==========================================

if [ "$EMBED_MODE" != true ]; then
    if [ -d "web/out" ]; then
        rm -rf "$DIST_DIR/web"
        cp -r "web/out" "$DIST_DIR/web"
        print_info "已复制前端资源"
    fi
fi

# ==========================================
#         创建启动脚本
# ==========================================

if [ "$EMBED_MODE" = true ]; then
    cat > "$DIST_DIR/start.sh" << 'EOF'
#!/bin/bash
echo "========================================"
echo "  User Frontend - Starting..."
echo "  (嵌入模式 - 单文件运行)"
echo "========================================"
echo ""
echo "访问地址: http://localhost:8080/"
echo ""
chmod +x ./UserFrontend
./UserFrontend
EOF
else
    cat > "$DIST_DIR/start.sh" << 'EOF'
#!/bin/bash
echo "========================================"
echo "  User Frontend - Starting..."
echo "========================================"
echo ""
echo "访问地址: http://localhost:8080/"
echo ""
chmod +x ./UserFrontend
./UserFrontend
EOF
fi
chmod +x "$DIST_DIR/start.sh"

# ==========================================
#         构建完成
# ==========================================

echo ""
print_info "=========================================="
print_success "  构建成功!"
print_info "=========================================="
echo ""

if [ "$EMBED_MODE" = true ]; then
    print_info "构建模式: 嵌入模式（单文件）"
else
    print_info "构建模式: 外部资源模式"
fi

print_info "输出目录: $DIST_DIR"
ls -la "$DIST_DIR"
echo ""
print_info "构建目录结构:"
echo "  编译中间文件: $BUILD_DIR"
echo "    ├── web/          (前端 .next)"
echo "    └── node_modules/ (前端依赖)"
echo ""
print_info "运行命令: cd $DIST_DIR && ./start.sh"
