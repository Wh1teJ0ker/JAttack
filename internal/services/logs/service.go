package logs

import (
	"JAttack/internal/pkg/logger"
	"bufio"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LogEntry 定义单条日志内容
type LogEntry struct {
	Content string `json:"content"`
}

// LogFile 定义日志文件信息
type LogFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// LogService 日志管理服务
type LogService struct {
	ctx    context.Context
	logDir string
}

// NewLogService 创建日志管理服务实例
func NewLogService(logDir string) *LogService {
	return &LogService{
		logDir: logDir,
	}
}

// Startup 在服务启动时调用
func (s *LogService) Startup(ctx context.Context) {
	s.ctx = ctx
	logger.Info("日志管理服务已启动")
}

// ListLogFiles 列出 data/logs 目录下的所有日志文件
func (s *LogService) ListLogFiles() ([]LogFile, error) {
	logger.Debug("正在列出日志文件")
	entries, err := os.ReadDir(s.logDir)
	if err != nil {
		logger.Error("读取日志目录失败", "错误", err.Error())
		return nil, err
	}

	var logs []LogFile
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
			logs = append(logs, LogFile{
				Name: entry.Name(),
				Path: filepath.Join(s.logDir, entry.Name()),
			})
		}
	}

	// 按文件名倒序排列（通常最新的日期在最后，或者名字排序）
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Name > logs[j].Name
	})

	logger.Debug("成功获取日志文件列表", "数量", len(logs))
	return logs, nil
}

// ReadLogFile 读取指定日志文件的内容
func (s *LogService) ReadLogFile(filename string) ([]string, error) {
	logger.Debug("正在读取日志文件", "文件名", filename)
	
	// 防止路径遍历攻击，只允许读取 logDir 下的文件
	targetPath := filepath.Join(s.logDir, filepath.Base(filename))
	
	file, err := os.Open(targetPath)
	if err != nil {
		logger.Error("打开日志文件失败", "错误", err.Error())
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		logger.Error("读取日志文件内容失败", "错误", err.Error())
		return nil, err
	}

	logger.Debug("成功读取日志文件", "行数", len(lines))
	return lines, nil
}
