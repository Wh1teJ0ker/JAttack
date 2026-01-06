package infogather

import (
	"JAttack/internal/config"
	"JAttack/internal/db"
	"JAttack/internal/pkg/logger"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type BruteForceTarget struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"` // tcp/udp
	Service  string `json:"service"`  // ssh, ftp, mysql, etc.
}

type BruteForceService struct {
	ctx       context.Context
	targets   []BruteForceTarget
	mu        sync.RWMutex
	dbManager *db.Manager

	// For cancellation
	attackCtx    context.Context
	attackCancel context.CancelFunc
	running      bool
}

func NewBruteForceService(dbManager *db.Manager) *BruteForceService {
	return &BruteForceService{
		targets:   make([]BruteForceTarget, 0),
		dbManager: dbManager,
	}
}

func (s *BruteForceService) Startup(ctx context.Context) {
	s.ctx = ctx
	logger.Info("爆破服务已启动")
	// Ensure dict directory exists
	if err := os.MkdirAll(config.DictDir, 0755); err != nil {
		logger.Error("无法创建字典目录", "error", err)
	}
}

// AddTarget adds a new target to the brute force list
func (s *BruteForceService) AddTarget(ip string, port int, protocol, service string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicates
	for _, t := range s.targets {
		if t.IP == ip && t.Port == port && t.Service == service {
			return
		}
	}

	target := BruteForceTarget{
		IP:       ip,
		Port:     port,
		Protocol: protocol,
		Service:  service,
	}
	s.targets = append(s.targets, target)
	logger.Info(fmt.Sprintf("自动添加爆破目标: %s:%d (%s)", ip, port, service))
}

// GetTargets returns the current list of targets
func (s *BruteForceService) GetTargets() []BruteForceTarget {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.targets
}

// ClearTargets clears all targets
func (s *BruteForceService) ClearTargets() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.targets = make([]BruteForceTarget, 0)
}

type Dictionary struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type BruteForceConfig struct {
	UserDict  string             `json:"user_dict"`
	PassDict  string             `json:"pass_dict"`
	Threads   int                `json:"threads"`
	Timeout   int                `json:"timeout"`   // 超时时间(ms)
	Protocols []string           `json:"protocols"` // 选中的协议列表
	Targets   []BruteForceTarget `json:"targets"`   // 选中的目标列表
}

// GetDictionaries 返回可用字典文件列表
func (s *BruteForceService) GetDictionaries() []Dictionary {
	files, err := os.ReadDir(config.DictDir)
	if err != nil {
		logger.Error("读取字典目录失败", "error", err)
		return []Dictionary{}
	}

	var dicts []Dictionary
	for _, f := range files {
		if !f.IsDir() {
			info, err := f.Info()
			if err != nil {
				continue
			}
			dicts = append(dicts, Dictionary{
				Name: f.Name(),
				Path: filepath.Join(config.DictDir, f.Name()),
				Size: info.Size(),
			})
		}
	}
	return dicts
}

// StartAttack 开始爆破任务
func (s *BruteForceService) StartAttack(config BruteForceConfig) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		s.emitLog("任务已在运行中")
		return
	}
	s.running = true
	s.attackCtx, s.attackCancel = context.WithCancel(context.Background())
	s.mu.Unlock()

	logger.Info("开始爆破任务", "config", config)
	s.emitLog("开始爆破任务...")

	go func() {
		defer func() {
			s.mu.Lock()
			s.running = false
			if s.attackCancel != nil {
				s.attackCancel() // 确保清理
			}
			s.mu.Unlock()
			runtime.EventsEmit(s.ctx, "bruteforce:finished", true)
		}()
		s.runAttack(s.attackCtx, config)
	}()
}

// StopAttack 停止爆破任务
func (s *BruteForceService) StopAttack() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.attackCancel != nil {
		s.attackCancel()
		s.emitLog("正在停止爆破任务...")
	}
}

// SelectDictionaryFile opens a native file dialog to select a dictionary file
func (s *BruteForceService) SelectDictionaryFile() string {
	selection, err := runtime.OpenFileDialog(s.ctx, runtime.OpenDialogOptions{
		Title: "选择字典文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Text Files (*.txt)",
				Pattern:     "*.txt",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		logger.Error("Failed to open file dialog", "error", err)
		return ""
	}
	return selection
}

func (s *BruteForceService) emitLog(msg string) {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "bruteforce:log", msg)
	}
}
