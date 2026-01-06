package settings

import (
	"JAttack/internal/db"
	"JAttack/internal/pkg/logger"
	"context"
)

type SettingsService struct {
	ctx       context.Context
	dbManager *db.Manager
}

func NewSettingsService(dbManager *db.Manager) *SettingsService {
	return &SettingsService{
		dbManager: dbManager,
	}
}

func (s *SettingsService) Startup(ctx context.Context) {
	s.ctx = ctx
	logger.Info("设置服务已启动")
}

func (s *SettingsService) InitDatabase(path string) error {
	logger.Info("正在初始化数据库", "路径", path)
	database, err := db.InitDB(path)
	if err != nil {
		logger.Error("数据库初始化失败", "错误", err.Error())
		return err
	}
	s.dbManager.SetDB(database)
	logger.Info("数据库初始化成功")
	
	// 可选: 如果需要下次自动加载，可以在此处保存路径到配置文件
	return nil
}
