# ==========================================
#         User 统一构建脚本
# ==========================================
#
# 用法:
#   .\build.ps1              # 默认编译 Windows 版本（外部资源模式）
#   .\build.ps1 --linux      # 仅编译 Linux 版本
#   .\build.ps1 --all        # 编译所有平台版本
#   .\build.ps1 --embed      # 将前端资源嵌入到程序中（单文件模式）
#   .\build.ps1 --clean      # 清理构建目录
#   .\build.ps1 --build      # 编译后移动前端缓存到 dist/build
#
# 构建模式:
#   外部资源模式（默认）：前端资源作为独立文件，程序从 ./web/ 目录加载
#   嵌入模式（--embed）：前端资源打包进二进制文件，生成单个可执行文件
#
# 目录结构:
#   dist/
#   ├── build/              # 编译中间文件（仅 -Build 模式）
#   │   ├── web/            # 前端 .next
#   │   └── node_modules/   # 前端依赖
#   ├── windows/            # Windows 最终输出
#   └── linux/              # Linux 最终输出
#
# ==========================================

param(
    [switch]$Linux,
    [switch]$All,
    [switch]$Clean,
    [switch]$Force,
    [switch]$SkipFrontend,
    [switch]$SkipPause,
    [switch]$Build,  # 是否移动前端缓存到 dist/build
    [switch]$Embed   # 是否将前端资源嵌入到程序中
)

$ErrorActionPreference = "Stop"
$ScriptPath = $PSScriptRoot
Set-Location $ScriptPath

# ==========================================
#         颜色和日志函数
# ==========================================

function Write-Info { Write-Host "[INFO] $args" -ForegroundColor Cyan }
function Write-Success { Write-Host "[SUCCESS] $args" -ForegroundColor Green }
function Write-Warning { Write-Host "[WARNING] $args" -ForegroundColor Yellow }
function Write-Err { Write-Host "[ERROR] $args" -ForegroundColor Red }

# ==========================================
#         配置
# ==========================================

$ProjectName = "UserFrontend"
$SourceDir = $ScriptPath
$DistDir = Join-Path $ScriptPath "dist"
$BuildDir = Join-Path $DistDir "build"
$WebDir = Join-Path $SourceDir "web"
$StaticDir = Join-Path $SourceDir "internal" "static"

# 编译中间文件目录
$WebBuildDir = Join-Path $BuildDir "web"
$CachedNodeModules = Join-Path $BuildDir "node_modules"

# ==========================================
#         清理函数
# ==========================================

if ($Clean) {
    Write-Host ""
    Write-Info "========== 清理构建目录 =========="
    
    if (Test-Path $DistDir) {
        Remove-Item -Path $DistDir -Recurse -Force
        Write-Success "已清理: $DistDir"
    }
    
    # 清理嵌入式资源目录
    $EmbedWebDir = Join-Path $StaticDir "web"
    if (Test-Path $EmbedWebDir) {
        Remove-Item -Path $EmbedWebDir -Recurse -Force
        Write-Success "已清理嵌入式资源: $EmbedWebDir"
    }
    
    if ($Force) {
        # 清理旧的 build 目录
        $OldBuildDir = Join-Path $ScriptPath "build"
        if (Test-Path $OldBuildDir) {
            Remove-Item -Path $OldBuildDir -Recurse -Force
            Write-Success "已清理: $OldBuildDir"
        }
        
        # 清理前端构建缓存
        $WebNextDir = Join-Path $WebDir ".next"
        $WebOutDir = Join-Path $WebDir "out"
        $LocalNodeModules = Join-Path $WebDir "node_modules"
        if (Test-Path $WebNextDir) { Remove-Item -Path $WebNextDir -Recurse -Force }
        if (Test-Path $WebOutDir) { Remove-Item -Path $WebOutDir -Recurse -Force }
        if (Test-Path $LocalNodeModules) { 
            $item = Get-Item $LocalNodeModules -Force
            if ($item.Attributes -band [System.IO.FileAttributes]::ReparsePoint) {
                cmd /c rmdir "$LocalNodeModules" 2>$null
            } else {
                Remove-Item -Path $LocalNodeModules -Recurse -Force
            }
        }
        Write-Success "已清理前端缓存"
    }
    
    Write-Success "清理完成"
    if (-not $SkipPause) {
        Write-Host "`n按任意键退出..." -NoNewline
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    }
    exit 0
}

# ==========================================
#         确定构建目标
# ==========================================

$BuildWindows = $true
$BuildLinux = $false

if ($Linux) {
    $BuildWindows = $false
    $BuildLinux = $true
} elseif ($All) {
    $BuildWindows = $true
    $BuildLinux = $true
}

$StartTime = Get-Date

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  $ProjectName 构建脚本" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Info "构建目标:"
if ($BuildWindows) { Write-Host "  - Windows (amd64)" -ForegroundColor Gray }
if ($BuildLinux) { Write-Host "  - Linux (amd64)" -ForegroundColor Gray }
if ($Embed) { Write-Host "  - 嵌入模式: 前端资源将打包进程序" -ForegroundColor Yellow }

# ==========================================
#         检查编译器
# ==========================================

Write-Host ""
Write-Info "========== 检查编译环境 =========="

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Err "未找到 Go 编译器"
    exit 1
}
$GoVersion = go version
Write-Info "Go: $GoVersion"

if (-not $SkipFrontend) {
    if (-not (Get-Command npm -ErrorAction SilentlyContinue)) {
        Write-Warning "未找到 npm，将跳过前端构建"
        $SkipFrontend = $true
    } else {
        $NodeVersion = node --version
        Write-Info "Node.js: $NodeVersion"
    }
}

# 下载 Go 依赖
Write-Info "下载 Go 依赖..."
go mod tidy
if ($LASTEXITCODE -ne 0) { throw "go mod tidy 失败" }

# ==========================================
#         前端增量构建检测
# ==========================================

# 计算前端源码哈希值
function Get-WebSourceHash {
    $hashFiles = @()
    
    # 收集 web/src 目录下所有文件
    $srcDir = Join-Path $WebDir "src"
    if (Test-Path $srcDir) {
        $hashFiles += Get-ChildItem -Path $srcDir -Recurse -File | ForEach-Object { $_.FullName }
    }
    
    # 添加关键配置文件
    $configFiles = @("package.json", "package-lock.json", "next.config.js", "next.config.mjs", "next.config.ts", "tsconfig.json", "tailwind.config.js", "tailwind.config.ts")
    foreach ($file in $configFiles) {
        $filePath = Join-Path $WebDir $file
        if (Test-Path $filePath) {
            $hashFiles += $filePath
        }
    }
    
    if ($hashFiles.Count -eq 0) {
        return $null
    }
    
    # 计算所有文件内容的组合哈希
    $md5 = [System.Security.Cryptography.MD5]::Create()
    $allBytes = @()
    foreach ($file in ($hashFiles | Sort-Object)) {
        $content = [System.IO.File]::ReadAllBytes($file)
        $allBytes += $content
    }
    $hash = $md5.ComputeHash($allBytes)
    return [BitConverter]::ToString($hash) -replace '-', ''
}

# 检查前端是否需要重新构建
function Test-WebNeedsBuild {
    $hashFile = Join-Path $BuildDir ".web_hash"
    $webOutDir = Join-Path $WebDir "out"
    
    # 如果输出目录不存在，需要构建
    if (-not (Test-Path $webOutDir)) {
        return $true
    }
    
    # 如果哈希文件不存在，需要构建
    if (-not (Test-Path $hashFile)) {
        return $true
    }
    
    # 比较哈希值
    $currentHash = Get-WebSourceHash
    if ($null -eq $currentHash) {
        return $true
    }
    
    $savedHash = Get-Content $hashFile -Raw -ErrorAction SilentlyContinue
    if ($savedHash -ne $currentHash) {
        return $true
    }
    
    return $false
}

# 保存前端哈希值
function Save-WebHash {
    $hashFile = Join-Path $BuildDir ".web_hash"
    $hash = Get-WebSourceHash
    if ($null -ne $hash) {
        # 确保目录存在
        if (-not (Test-Path $BuildDir)) {
            New-Item -ItemType Directory -Path $BuildDir -Force | Out-Null
        }
        Set-Content -Path $hashFile -Value $hash -NoNewline
    }
}

# ==========================================
#         构建前端
# ==========================================

Write-Host ""
Write-Info "========== 构建 Web 前端 =========="

if ($SkipFrontend) {
    Write-Warning "跳过前端构建"
} elseif (-not (Test-Path $WebDir)) {
    Write-Warning "未找到 web 目录"
} elseif ((-not $Force) -and (-not (Test-WebNeedsBuild))) {
    Write-Info "前端未变更，跳过构建"
} else {
    $LocalNodeModules = Join-Path $WebDir "node_modules"
    
    # 确保 build 目录存在
    if (-not (Test-Path $BuildDir)) {
        New-Item -ItemType Directory -Path $BuildDir -Force | Out-Null
    }
    
    # 如果缓存存在且 web/node_modules 不存在，从缓存复制
    if ((Test-Path $CachedNodeModules) -and (-not (Test-Path $LocalNodeModules))) {
        Write-Info "从缓存恢复 node_modules..."
        Copy-Item -Path $CachedNodeModules -Destination $LocalNodeModules -Recurse
    }
    
    # 安装依赖（如果需要）
    if (-not (Test-Path $LocalNodeModules)) {
        Write-Info "安装前端依赖..."
        Push-Location $WebDir
        npm install --legacy-peer-deps
        if ($LASTEXITCODE -ne 0) {
            Write-Err "npm install 失败"
            Pop-Location
            exit 1
        }
        Pop-Location
    }
    
    # 构建前端
    Write-Info "构建前端..."
    Push-Location $WebDir
    npm run build
    if ($LASTEXITCODE -ne 0) {
        Write-Err "前端构建失败"
        Pop-Location
        exit 1
    }
    Pop-Location
    
    # 仅在 -Build 模式下移动前端构建中间文件
    if ($Build) {
        # 移动前端构建中间文件到 dist/build/web
        $WebNextDir = Join-Path $WebDir ".next"
        if (Test-Path $WebNextDir) {
            if (Test-Path $WebBuildDir) {
                Remove-Item -Path $WebBuildDir -Recurse -Force -ErrorAction SilentlyContinue
            }
            Move-Item -Path $WebNextDir -Destination $WebBuildDir -Force
            Write-Info "已移动 .next 到: $WebBuildDir"
        }
        
        # 移动 node_modules 到缓存
        if (Test-Path $LocalNodeModules) {
            if (Test-Path $CachedNodeModules) {
                Remove-Item -Path $CachedNodeModules -Recurse -Force -ErrorAction SilentlyContinue
            }
            Move-Item -Path $LocalNodeModules -Destination $CachedNodeModules -Force
            Write-Info "已移动 node_modules 到: $CachedNodeModules"
        }
    }
    
    Write-Success "前端构建完成"
    
    # 保存哈希值
    Save-WebHash
}

# ==========================================
#         准备嵌入式资源
# ==========================================

function Prepare-EmbedResources {
    Write-Host ""
    Write-Info "========== 准备嵌入式资源 =========="
    
    $WebOutDir = Join-Path $WebDir "out"
    $EmbedWebDir = Join-Path $StaticDir "web"
    
    if (-not (Test-Path $WebOutDir)) {
        Write-Err "前端构建输出目录不存在: $WebOutDir"
        Write-Err "请先构建前端（不要使用 -SkipFrontend）"
        return $false
    }
    
    # 确保 static 目录存在
    if (-not (Test-Path $StaticDir)) {
        New-Item -ItemType Directory -Path $StaticDir -Force | Out-Null
    }
    
    # 清理旧的嵌入资源
    if (Test-Path $EmbedWebDir) {
        Remove-Item -Path $EmbedWebDir -Recurse -Force
    }
    
    # 复制前端资源到嵌入目录
    Write-Info "复制前端资源到嵌入目录..."
    Copy-Item -Path $WebOutDir -Destination $EmbedWebDir -Recurse
    
    $fileCount = (Get-ChildItem -Path $EmbedWebDir -Recurse -File).Count
    $totalSize = (Get-ChildItem -Path $EmbedWebDir -Recurse -File | Measure-Object -Property Length -Sum).Sum / 1MB
    Write-Success "已准备嵌入式资源: $fileCount 个文件, $($totalSize.ToString("N2")) MB"
    
    return $true
}

function Cleanup-EmbedResources {
    $EmbedWebDir = Join-Path $StaticDir "web"
    if (Test-Path $EmbedWebDir) {
        Remove-Item -Path $EmbedWebDir -Recurse -Force
        Write-Info "已清理嵌入式资源目录"
    }
}

# ==========================================
#         构建函数
# ==========================================

function Build-Platform {
    param(
        [string]$Platform,
        [string]$GOOS,
        [string]$BinaryName,
        [bool]$EmbedMode
    )
    
    Write-Host ""
    Write-Info "========== 构建 $Platform 版本 =========="
    
    $TargetDir = Join-Path $DistDir $Platform.ToLower()
    
    # 清理目标目录
    if (Test-Path $TargetDir) {
        Remove-Item -Path $TargetDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $TargetDir -Force | Out-Null
    
    # 非嵌入模式：复制前端资源
    if (-not $EmbedMode) {
        $WebOutDir = Join-Path $WebDir "out"
        if ((Test-Path $WebOutDir) -and (-not $SkipFrontend)) {
            $WebDest = Join-Path $TargetDir "web"
            Copy-Item -Path $WebOutDir -Destination $WebDest -Recurse
            Write-Info "已复制前端资源"
        }
    }
    
    # 创建必要目录
    New-Item -ItemType Directory -Path (Join-Path $TargetDir "user_config") -Force | Out-Null
    New-Item -ItemType Directory -Path (Join-Path $TargetDir "Product") -Force | Out-Null
    New-Item -ItemType Directory -Path (Join-Path $TargetDir "backups") -Force | Out-Null
    
    # 设置 Go 编译环境变量
    $env:GOOS = $GOOS
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"
    
    $OutputExe = Join-Path $TargetDir $BinaryName
    
    # 构建标签
    $buildTags = ""
    if ($EmbedMode) {
        $buildTags = "-tags embed"
        Write-Info "编译 Go 程序（嵌入模式）..."
    } else {
        Write-Info "编译 Go 程序..."
    }
    
    $buildCmd = "go build -ldflags=`"-s -w`" $buildTags -o `"$OutputExe`" ./cmd/server"
    Invoke-Expression $buildCmd
    if ($LASTEXITCODE -ne 0) {
        Write-Err "Go 编译失败"
        return $null
    }
    
    # 创建启动脚本
    if ($Platform -eq "Windows") {
        if ($EmbedMode) {
            $startScript = @"
@echo off
title User Frontend (Embedded)
echo ========================================
echo   User Frontend - Starting...
echo   (嵌入模式 - 单文件运行)
echo ========================================
echo.
echo 访问地址: http://localhost:8080/
echo.
"%~dp0$BinaryName"
pause
"@
        } else {
            $startScript = @"
@echo off
title User Frontend
echo ========================================
echo   User Frontend - Starting...
echo ========================================
echo.
echo 访问地址: http://localhost:8080/
echo.
"%~dp0$BinaryName"
pause
"@
        }
        Set-Content -Path (Join-Path $TargetDir "start.bat") -Value $startScript
    } else {
        if ($EmbedMode) {
            $startScript = @"
#!/bin/bash
echo "========================================"
echo "  User Frontend - Starting..."
echo "  (嵌入模式 - 单文件运行)"
echo "========================================"
echo ""
echo "访问地址: http://localhost:8080/"
echo ""
chmod +x ./$BinaryName
./$BinaryName
"@
        } else {
            $startScript = @"
#!/bin/bash
echo "========================================"
echo "  User Frontend - Starting..."
echo "========================================"
echo ""
echo "访问地址: http://localhost:8080/"
echo ""
chmod +x ./$BinaryName
./$BinaryName
"@
        }
        Set-Content -Path (Join-Path $TargetDir "start.sh") -Value $startScript -NoNewline
    }
    
    $OutputSize = (Get-Item $OutputExe).Length / 1MB
    $modeText = if ($EmbedMode) { "嵌入模式" } else { "外部资源模式" }
    Write-Success "已生成: $BinaryName ($($OutputSize.ToString("N2")) MB) [$modeText]"
    
    return $TargetDir
}

# ==========================================
#         执行构建
# ==========================================

try {
    $BuildResults = @()
    
    # 嵌入模式：准备资源
    if ($Embed) {
        if (-not (Prepare-EmbedResources)) {
            exit 1
        }
    }
    
    if ($BuildWindows) {
        $WinDir = Build-Platform -Platform "Windows" -GOOS "windows" -BinaryName "UserFrontend.exe" -EmbedMode $Embed
        if ($WinDir) {
            $BuildResults += @{ Platform = "Windows"; Dir = $WinDir }
        }
    }
    
    if ($BuildLinux) {
        $LinuxDir = Build-Platform -Platform "Linux" -GOOS "linux" -BinaryName "UserFrontend" -EmbedMode $Embed
        if ($LinuxDir) {
            $BuildResults += @{ Platform = "Linux"; Dir = $LinuxDir }
        }
    }
    
    # 嵌入模式：清理临时资源
    if ($Embed) {
        Cleanup-EmbedResources
    }
    
    if ($BuildResults.Count -eq 0) {
        Write-Err "没有成功构建任何平台"
        exit 1
    }
    
    $EndTime = Get-Date
    $Duration = $EndTime - $StartTime
    
    # ==========================================
    #         构建完成
    # ==========================================
    
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  构建完成!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "总耗时: $($Duration.TotalSeconds.ToString("N2")) 秒" -ForegroundColor Gray
    if ($Embed) {
        Write-Host "构建模式: 嵌入模式（单文件）" -ForegroundColor Yellow
    } else {
        Write-Host "构建模式: 外部资源模式" -ForegroundColor Gray
    }
    Write-Host ""
    
    # 显示输出目录结构
    Write-Info "输出目录结构:"
    foreach ($result in $BuildResults) {
        Write-Host ""
        Write-Host "  $($result.Platform): $($result.Dir)" -ForegroundColor Cyan
        
        $items = Get-ChildItem -Path $result.Dir
        foreach ($item in $items) {
            if ($item.PSIsContainer) {
                $subCount = (Get-ChildItem -Path $item.FullName -Recurse -File -ErrorAction SilentlyContinue).Count
                Write-Host "    ├── $($item.Name)/ ($subCount 文件)" -ForegroundColor Gray
            } else {
                $size = [math]::Round($item.Length / 1KB, 1)
                Write-Host "    ├── $($item.Name) ($size KB)" -ForegroundColor Gray
            }
        }
    }
    
    # 构建目录结构说明
    Write-Host ""
    Write-Info "构建目录结构:"
    Write-Host "  编译中间文件: $BuildDir" -ForegroundColor Gray
    Write-Host "    ├── web/          (前端 .next)" -ForegroundColor Gray
    Write-Host "    └── node_modules/ (前端依赖)" -ForegroundColor Gray
    
} catch {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "  构建失败!" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
    Write-Err "错误: $_"
    
    # 清理嵌入资源
    if ($Embed) {
        Cleanup-EmbedResources
    }
    
    exit 1
} finally {
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
    
    if (-not $SkipPause) {
        Write-Host ""
        Write-Host "按任意键退出..." -NoNewline
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    }
}
