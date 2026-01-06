package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	instance *slog.Logger
	once     sync.Once
)

// DailyWriter 实现了按天滚动的 io.Writer
type DailyWriter struct {
	dir      string
	prefix   string
	currFile *os.File
	currDate string
	mu       sync.Mutex
}

func NewDailyWriter(dir, prefix string) *DailyWriter {
	return &DailyWriter{
		dir:    dir,
		prefix: prefix,
	}
}

func (w *DailyWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	if w.currDate != today || w.currFile == nil {
		if err := w.rotate(today); err != nil {
			return 0, err
		}
	}

	return w.currFile.Write(p)
}

func (w *DailyWriter) rotate(date string) error {
	if w.currFile != nil {
		w.currFile.Close()
	}

	if err := os.MkdirAll(w.dir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(w.dir, fmt.Sprintf("%s-%s.log", w.prefix, date))
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	w.currFile = file
	w.currDate = date
	return nil
}

// InitLogger 初始化日志系统
// logDir: 日志目录，例如 "data/logs"
// prefix: 日志文件前缀，例如 "app"
func InitLogger(logDir, prefix string) error {
	var err error
	once.Do(func() {
		dailyWriter := NewDailyWriter(logDir, prefix)
		// 立即触发一次文件创建检查
		if _, err = dailyWriter.Write([]byte{}); err != nil {
			return // 这里的空写可能会导致问题，但主要是为了触发 rotate 检查权限
		}

		// 同时输出到控制台和文件
		multiWriter := io.MultiWriter(os.Stdout, dailyWriter)

		// 配置 Handler
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{
						Key:   "时间",
						Value: slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05")),
					}
				}
				if a.Key == slog.LevelKey {
					return slog.Attr{
						Key:   "级别",
						Value: a.Value,
					}
				}
				if a.Key == slog.MessageKey {
					return slog.Attr{
						Key:   "信息",
						Value: a.Value,
					}
				}
				return a
			},
		}

		handler := slog.NewJSONHandler(multiWriter, opts)
		instance = slog.New(handler)
	})

	return err
}

// Get 获取日志实例
func Get() *slog.Logger {
	if instance == nil {
		return slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	return instance
}

// Info 记录普通信息
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Error 记录错误信息
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// Warn 记录警告信息
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Debug 记录调试信息
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}
