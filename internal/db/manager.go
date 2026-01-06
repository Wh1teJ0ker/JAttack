package db

import (
	"database/sql"
	"fmt"
	"sync"
)

// Manager 数据库管理器，用于管理全局数据库连接实例
type Manager struct {
	db        *sql.DB
	mu        sync.RWMutex
	writeChan chan func()
}

// NewManager 创建一个新的数据库管理器
func NewManager() *Manager {
	m := &Manager{
		writeChan: make(chan func(), 100),
	}
	go m.worker()
	return m
}

func (m *Manager) worker() {
	for task := range m.writeChan {
		task()
	}
}

// ExecTask 提交一个数据库任务到写队列并等待完成
// 这确保了所有的写操作都是串行执行的，避免 SQLite 锁竞争
func (m *Manager) ExecTask(task func(*sql.DB) error) error {
	errChan := make(chan error, 1)
	m.writeChan <- func() {
		// 注意：这里我们不需要加锁获取 DB，因为 SetDB 操作相对不频繁，
		// 而且我们在 worker 中是串行的。
		// 但为了安全起见，还是使用 GetDB()
		db := m.GetDB()
		if db == nil {
			errChan <- fmt.Errorf("database not initialized")
			return
		}
		errChan <- task(db)
	}
	return <-errChan
}

// SetDB 设置数据库连接实例
func (m *Manager) SetDB(db *sql.DB) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.db != nil {
		m.db.Close()
	}
	m.db = db
}

// GetDB 获取数据库连接实例
func (m *Manager) GetDB() *sql.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db
}
