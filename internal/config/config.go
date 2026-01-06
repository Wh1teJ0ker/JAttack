package config

import (
	"os"
	"path/filepath"
)

// 全局路径配置
var (
	ConfigDir string
	DictDir   string
)

func init() {
	// 初始化默认路径
	// 在实际应用中，可能会相对于可执行文件或用户主目录来确定
	cwd, _ := os.Getwd()
	ConfigDir = filepath.Join(cwd, "config")
	DictDir = filepath.Join(cwd, "internal", "dict")

	// 确保目录存在
	_ = os.MkdirAll(ConfigDir, 0755)
	_ = os.MkdirAll(DictDir, 0755)
}

// GetDictPath 返回字典文件的完整路径
func GetDictPath(filename string) string {
	return filepath.Join(DictDir, filename)
}
