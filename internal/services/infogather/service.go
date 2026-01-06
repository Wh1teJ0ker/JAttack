package infogather

import (
	"JAttack/internal/db"
	"JAttack/internal/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type InfoGathering struct {
	ID        int       `json:"id"`
	Target    string    `json:"target"`
	InfoType  string    `json:"info_type"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at" ts_type:"string"`
}

type InfoService struct {
	ctx        context.Context
	dbManager  *db.Manager
	bfService  *BruteForceService
	scanCtx    context.Context
	scanCancel context.CancelFunc
	paused     atomic.Bool
	mu         sync.Mutex // 保护 scanCtx/scanCancel
	dbQueue    chan func()
}

func NewInfoService(dbManager *db.Manager, bfService *BruteForceService) *InfoService {
	s := &InfoService{
		dbManager: dbManager,
		bfService: bfService,
		dbQueue:   make(chan func(), 5000), // Buffer for DB tasks
	}
	go s.processDBQueue()
	return s
}

func (s *InfoService) processDBQueue() {
	for task := range s.dbQueue {
		task()
	}
}

func (s *InfoService) Startup(ctx context.Context) {
	s.ctx = ctx
	logger.Info("信息搜集服务已启动")
}

// StopScan 停止当前扫描任务
func (s *InfoService) StopScan() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.scanCancel != nil {
		s.scanCancel()
		logger.Info("扫描任务已停止")
	}
}

// PauseScan 暂停或恢复扫描任务
func (s *InfoService) PauseScan(pause bool) {
	s.paused.Store(pause)
	if pause {
		logger.Info("扫描任务已暂停")
	} else {
		logger.Info("扫描任务已恢复")
	}
}

// ClearData 清空数据库中的所有搜集信息
func (s *InfoService) ClearData() error {
	logger.Info("正在清除所有搜集信息")
	db := s.dbManager.GetDB()
	if db == nil {
		return errors.New("数据库未初始化")
	}

	_, err := db.Exec(`DELETE FROM info_gathering`)
	if err != nil {
		logger.Error("清除数据失败", "错误", err.Error())
		return err
	}
	logger.Info("所有搜集信息清除成功")
	return nil
}

func (s *InfoService) AddInfo(target, infoType, content string) error {
	logger.Info("正在添加搜集信息", "目标", target, "类型", infoType)

	// Push to DB queue to avoid locking
	s.dbQueue <- func() {
		// Use ExecTask to ensure serialization via dbManager
		err := s.dbManager.ExecTask(func(db *sql.DB) error {
			query := `INSERT INTO info_gathering (target, info_type, content) VALUES (?, ?, ?)`
			_, err := db.Exec(query, target, infoType, content)
			return err
		})

		if err != nil {
			logger.Error("添加信息失败", "错误", err.Error())
		}
	}

	return nil
}

func (s *InfoService) ListInfo() ([]InfoGathering, error) {
	logger.Debug("正在列出所有搜集信息")
	db := s.dbManager.GetDB()
	if db == nil {
		logger.Warn("尝试列出信息但数据库未初始化")
		return nil, nil
	}

	rows, err := db.Query(`SELECT id, target, info_type, content, created_at FROM info_gathering ORDER BY created_at DESC`)
	if err != nil {
		logger.Error("查询信息失败", "错误", err.Error())
		return nil, err
	}
	defer rows.Close()

	var results []InfoGathering
	for rows.Next() {
		var i InfoGathering
		if err := rows.Scan(&i.ID, &i.Target, &i.InfoType, &i.Content, &i.CreatedAt); err != nil {
			logger.Error("扫描信息行失败", "错误", err.Error())
			return nil, err
		}
		results = append(results, i)
	}
	logger.Debug("成功获取信息列表", "数量", len(results))
	return results, nil
}

// GetDirScanResultModel is a dummy method to force Wails to generate the DirScanResult struct binding.
// It is not meant to be called.
func (s *InfoService) GetDirScanResultModel() DirScanResult {
	return DirScanResult{}
}
