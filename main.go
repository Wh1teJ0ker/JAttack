package main

import (
	"JAttack/internal/db"
	"JAttack/internal/pkg/logger"
	"JAttack/internal/services/infogather"
	"JAttack/internal/services/logs"
	"JAttack/internal/services/poc"
	"JAttack/internal/services/settings"
	"JAttack/internal/services/vuln"
	"context"
	"embed"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// //go:embed internal/db/schema.sql
// var schemaSQL string

func main() {
	// 确定运行时根目录
	// 默认为 runtime 目录，所有产生的数据都放在这里
	// 这样编译后的产物和开发环境都会保持根目录整洁

	// 如果需要支持便携式部署，可以基于可执行文件路径
	// exePath, _ := os.Executable()
	// runtimeRoot := filepath.Join(filepath.Dir(exePath), "runtime")

	// 为了简单且符合 wails dev 和 wails build 的行为
	// 我们直接使用 "runtime" 相对路径
	runtimeRoot := "runtime"

	// 确保目录结构存在
	// runtime/data/logs
	// runtime/data/jattack.db

	dataDir := filepath.Join(runtimeRoot, "data")
	logDir := filepath.Join(dataDir, "logs")

	dirs := []string{runtimeRoot, dataDir, logDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			println("无法创建目录:", dir, err.Error())
			return
		}
	}

	// 初始化日志系统
	if err := logger.InitLogger(logDir, "app"); err != nil {
		println("无法初始化日志系统:", err.Error())
		return
	}

	logger.Info("应用程序启动")

	// Initialize DB Manager
	dbManager := db.NewManager()

	// 默认初始化数据库路径为 runtime/data/jattack.db
	defaultDBPath := filepath.Join(dataDir, "jattack.db")

	// Initialize Services
	settingsService := settings.NewSettingsService(dbManager)
	bruteForceService := infogather.NewBruteForceService(dbManager)
	// fingerprintService := infogather.NewFingerprintService(dbManager)
	infoService := infogather.NewInfoService(dbManager, bruteForceService)
	vulnService := vuln.NewVulnService(dbManager)
	pocService := poc.NewPocService(dataDir)
	logService := logs.NewLogService(logDir)
	jsFinderService := infogather.NewJSFinderService(dbManager)
	assetService := infogather.NewAssetService(dbManager)

	// 尝试自动初始化数据库
	if _, err := os.Stat(defaultDBPath); err == nil || os.IsNotExist(err) {
		// 如果数据库存在或者不存在都尝试初始化（InitDB 会处理新建）
		// 注意：如果文件不存在，InitDB 会创建；如果存在，会打开
		if err := settingsService.InitDatabase(defaultDBPath); err != nil {
			logger.Error("自动初始化数据库失败", "路径", defaultDBPath, "错误", err.Error())
		} else {
			logger.Info("已自动加载数据库", "路径", defaultDBPath)
		}
	}

	// 使用选项创建应用程序
	err := wails.Run(&options.App{
		Title:  "JAttack",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			logger.Info("正在启动服务...")
			settingsService.Startup(ctx)
			bruteForceService.Startup(ctx)
			// fingerprintService.Startup(ctx)
			infoService.Startup(ctx)
			vulnService.Startup(ctx)
			pocService.Startup(ctx)
			logService.Startup(ctx)
			jsFinderService.Startup(ctx)
			assetService.Startup(ctx)
			logger.Info("服务启动完成")
		},
		Bind: []interface{}{
			settingsService,
			bruteForceService,
			// fingerprintService,
			infoService,
			vulnService,
			pocService,
			logService,
			jsFinderService,
			assetService,
		},
	})

	if err != nil {
		logger.Error("应用程序运行错误", "错误", err.Error())
	}
}
