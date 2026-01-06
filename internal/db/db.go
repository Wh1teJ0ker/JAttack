package db

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// InitDB 初始化 SQLite 数据库
// path: 数据库文件路径
func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 执行 schema
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("执行数据库 Schema 失败: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		fmt.Printf("Warning: Failed to enable WAL mode: %v\n", err)
	}
	// Set busy timeout to 5 seconds to avoid "database is locked" errors
	if _, err := db.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
		fmt.Printf("Warning: Failed to set busy_timeout: %v\n", err)
	}

	return db, nil
}
