package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"user-frontend/internal/api"
	"user-frontend/internal/cache"
	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal("获取程序路径失败:", err)
	}
	execDir := filepath.Dir(execPath)

	// 配置目录
	configDir := filepath.Join(execDir, "user_config")

	// 初始化全局配置（设置默认值）
	cfg, err := config.InitConfig(configDir)
	if err != nil {
		log.Printf("警告: 加载配置失败: %v，使用默认配置", err)
	}

	// 初始化配置数据库（SQLite，存储数据库连接配置）
	if err := model.InitConfigDB(configDir); err != nil {
		log.Fatalf("初始化配置数据库失败: %v", err)
	}
	log.Println("配置数据库初始化成功")

	// 初始化配置服务（传入配置目录路径）
	configSvc := service.InitConfigServiceWithDir(model.ConfigDB, configDir)

	// 初始化加密密钥（如果不存在则自动生成）
	if err := configSvc.InitEncryptionKey(); err != nil {
		log.Printf("警告: 初始化加密密钥失败: %v", err)
	}

	// 检查是否需要从旧的JSON配置文件迁移（旧文件在1/db_config.json）
	oldConfigPath := filepath.Join(execDir, "1", "db_config.json")
	if _, err := os.Stat(oldConfigPath); err == nil {
		log.Println("检测到旧的db_config.json配置文件，正在迁移...")
		if err := configSvc.MigrateFromJSON(oldConfigPath); err != nil {
			log.Printf("警告: 迁移旧配置失败: %v", err)
		} else {
			log.Println("配置迁移成功")
			// 迁移成功后重命名旧文件
			backupPath := oldConfigPath + ".bak"
			if err := os.Rename(oldConfigPath, backupPath); err != nil {
				log.Printf("警告: 备份旧配置文件失败: %v", err)
			} else {
				log.Printf("旧配置文件已备份为: %s", backupPath)
			}
		}
	}

	// 从SQLite加载数据库配置到全局配置
	configSvc.LoadDBConfigToGlobal()
	log.Println("已从配置数据库加载数据库连接配置")

	// 从配置数据库加载服务器端口
	if serverPort, err := configSvc.GetServerPort(); err == nil && serverPort > 0 {
		cfg.ServerConfig.Port = serverPort
		log.Printf("已从配置数据库加载服务器端口: %d", serverPort)
	}

	// 设置数据库配置服务到API层
	api.InitDBConfigService(configSvc)

	// 尝试连接主数据库
	dbCfg := &config.GlobalConfig.DBConfig
	if err := model.InitDB(dbCfg); err != nil {
		log.Printf("警告: 主数据库连接失败: %v", err)
		
		// 如果是首次启动或配置的数据库无法连接，使用本地SQLite作为默认主数据库
		defaultDBPath := filepath.Join(configDir, "user_data.db")
		log.Printf("将使用本地SQLite数据库: %s", defaultDBPath)
		
		// 创建默认的SQLite数据库配置
		defaultDBCfg := &config.DBConfig{
			Type:     "sqlite",
			Database: defaultDBPath,
		}
		
		// 尝试连接默认SQLite数据库
		if err := model.InitDB(defaultDBCfg); err != nil {
			log.Printf("错误: 无法初始化默认数据库: %v", err)
			log.Println("程序将以无数据库模式运行，仅管理页面可用")
		} else {
			log.Println("默认SQLite数据库初始化成功")
			// 更新全局配置
			config.GlobalConfig.SetDBConfig(*defaultDBCfg)
			// 保存到配置数据库，下次启动时自动使用
			if err := configSvc.SaveDBConfig(defaultDBCfg); err != nil {
				log.Printf("警告: 保存默认数据库配置失败: %v", err)
			}
			// 设置主数据库到配置服务
			configSvc.SetMainDB(model.DB)
		}
	} else {
		log.Println("主数据库连接成功")
		// 设置主数据库到配置服务
		configSvc.SetMainDB(model.DB)
	}

	// 初始化缓存系统
	initCacheSystem(configSvc)

	// 初始化服务
	api.InitServices(cfg)

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由
	r := gin.Default()

	// 设置信任代理
	r.SetTrustedProxies(nil)

	// 注册路由
	api.RegisterRoutes(r, cfg)

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.ServerConfig.Port)
	log.Printf("服务器启动在 http://localhost%s", addr)
	log.Printf("管理后台: http://localhost%s/%s", addr, cfg.ServerConfig.AdminSuffix)

	if cfg.ServerConfig.UseHTTPS && cfg.ServerConfig.CertFile != "" && cfg.ServerConfig.KeyFile != "" {
		log.Println("使用HTTPS模式")
		if err := r.RunTLS(addr, cfg.ServerConfig.CertFile, cfg.ServerConfig.KeyFile); err != nil {
			log.Fatal("启动服务器失败:", err)
		}
	} else {
		if err := r.Run(addr); err != nil {
			log.Fatal("启动服务器失败:", err)
		}
	}
}

// initCacheSystem 初始化缓存系统
func initCacheSystem(configSvc *service.ConfigService) {
	// 从配置数据库加载Redis配置
	redisConfig, err := configSvc.GetRedisConfig()
	if err != nil {
		log.Printf("警告: 加载Redis配置失败: %v, 将使用本地缓存", err)
		cache.InitCacheManager(nil)
		return
	}

	// 检查是否启用Redis
	if !redisConfig.Enabled {
		log.Println("Redis未启用，使用本地内存缓存")
		cache.InitCacheManager(nil)
		return
	}

	// 转换为缓存配置并初始化
	cacheConfig := &cache.RedisConfig{
		Enabled:          redisConfig.Enabled,
		Mode:             redisConfig.Mode,
		Host:             redisConfig.Host,
		Port:             redisConfig.Port,
		Password:         redisConfig.Password,
		Database:         redisConfig.Database,
		SentinelAddrs:    redisConfig.SentinelAddrs,
		SentinelMaster:   redisConfig.SentinelMaster,
		SentinelPassword: redisConfig.SentinelPassword,
		ClusterAddrs:     redisConfig.ClusterAddrs,
		PoolSize:         redisConfig.PoolSize,
		MinIdleConns:     redisConfig.MinIdleConns,
		MaxRetries:       redisConfig.MaxRetries,
		DialTimeout:      redisConfig.DialTimeout,
		ReadTimeout:      redisConfig.ReadTimeout,
		WriteTimeout:     redisConfig.WriteTimeout,
		KeyPrefix:        redisConfig.KeyPrefix,
		DefaultTTL:       redisConfig.DefaultTTL,
		EnableMetrics:    redisConfig.EnableMetrics,
		TLSEnabled:       redisConfig.TLSEnabled,
		TLSCert:          redisConfig.TLSCert,
		TLSKey:           redisConfig.TLSKey,
		TLSCACert:        redisConfig.TLSCACert,
	}
	cache.InitCacheManager(cacheConfig)

	// 验证连接状态
	manager := cache.GetCacheManager()
	if manager != nil {
		if err := manager.Ping(); err != nil {
			log.Printf("警告: Redis连接失败: %v, 已降级到本地缓存", err)
		} else {
			log.Printf("Redis缓存已连接 (%s模式)", redisConfig.Mode)
		}
	}
}
