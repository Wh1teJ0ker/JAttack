package infogather

import (
	"JAttack/internal/config"
	"JAttack/internal/pkg/logger"
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func (s *BruteForceService) runAttack(ctx context.Context, cfg BruteForceConfig) {
	defer func() {
		s.emitLog("爆破任务结束")
	}()

	// 1. 设置超时
	timeout := 3 * time.Second
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout) * time.Millisecond
	}

	// 读取字典
	userDictPath := cfg.UserDict
	if !filepath.IsAbs(userDictPath) {
		userDictPath = config.GetDictPath(userDictPath)
	}
	users, err := readLines(userDictPath)
	if err != nil {
		s.emitLog(fmt.Sprintf("读取用户字典失败: %v", err))
		return
	}

	passDictPath := cfg.PassDict
	if !filepath.IsAbs(passDictPath) {
		passDictPath = config.GetDictPath(passDictPath)
	}
	passwords, err := readLines(passDictPath)
	if err != nil {
		s.emitLog(fmt.Sprintf("读取密码字典失败: %v", err))
		return
	}

	s.emitLog(fmt.Sprintf("加载了 %d 个用户和 %d 个密码", len(users), len(passwords)))

	var wg sync.WaitGroup
	// 默认线程数
	if cfg.Threads <= 0 {
		cfg.Threads = 10
	}
	sem := make(chan struct{}, cfg.Threads)

	var targets []BruteForceTarget
	if len(cfg.Targets) > 0 {
		targets = cfg.Targets
	} else {
		targets = s.GetTargets()
	}
	
	// 创建选中协议的 map 以便快速查找
	selectedProtocols := make(map[string]bool)
	if len(cfg.Protocols) > 0 {
		for _, p := range cfg.Protocols {
			selectedProtocols[strings.ToLower(p)] = true
		}
	}

	for _, t := range targets {
		serviceName := strings.ToLower(t.Service)

		// 检查是否在选中列表中 (如果列表为空，则默认全部)
		if len(selectedProtocols) > 0 && !selectedProtocols[serviceName] {
			continue
		}

		// 检查是否支持
		if !isSupported(serviceName) {
			s.emitLog(fmt.Sprintf("跳过不支持的服务: %s (%s:%d)", t.Service, t.IP, t.Port))
			continue
		}

		s.emitLog(fmt.Sprintf("正在爆破目标: %s:%d (%s)", t.IP, t.Port, t.Service))

		foundForTarget := false

		for _, u := range users {
			if foundForTarget { break }
			
			for _, p := range passwords {
				if foundForTarget { break }

				select {
				case <-ctx.Done():
					return
				default:
				}
				
				wg.Add(1)
				sem <- struct{}{}
				
				go func(target BruteForceTarget, user, pass string) {
					defer wg.Done()
					defer func() { <-sem }()
					
					// 快速跳过
					if foundForTarget { return }

					if s.tryLogin(target, user, pass, timeout) {
						s.emitLog(fmt.Sprintf("[SUCCESS] 发现弱口令! %s:%d (%s) -> %s / %s", target.IP, target.Port, target.Service, user, pass))
						foundForTarget = true
						
						// 保存到数据库
						if err := s.saveWeakPassword(target, user, pass); err != nil {
							logger.Error("保存弱口令失败", "error", err, "target", target.IP)
						}
					}
				}(t, u, p)
			}
		}
	}
	wg.Wait()
}

// saveWeakPassword 保存弱口令到数据库
func (s *BruteForceService) saveWeakPassword(target BruteForceTarget, user, pass string) error {
	// 1. Ensure Asset exists
	assetID, err := s.dbManager.UpsertAsset(target.IP, "", true)
	if err != nil {
		return err
	}

	// 2. Ensure Port exists
	protocol := target.Protocol
	if protocol == "" {
		protocol = "tcp"
	}
	portID, err := s.dbManager.UpsertAssetPort(assetID, target.Port, protocol, target.Service, "", "", "", "open")
	if err != nil {
		return err
	}

	// 3. Save Auth Result
	if err := s.dbManager.AddAuthResult(assetID, portID, target.Service, user, pass, true); err != nil {
		return err
	}

	// Legacy support (optional, can be removed if not needed)
	db := s.dbManager.GetDB()
	if db != nil {
		query := `INSERT INTO weak_passwords (target, port, service, username, password) VALUES (?, ?, ?, ?, ?)`
		db.Exec(query, target.IP, target.Port, target.Service, user, pass)
	}
	return nil
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			lines = append(lines, text)
		}
	}
	return lines, scanner.Err()
}

func isSupported(service string) bool {
	_, ok := plugins[strings.ToLower(service)]
	return ok
}

func (s *BruteForceService) tryLogin(t BruteForceTarget, user, pass string, timeout time.Duration) bool {
	service := strings.ToLower(t.Service)
	if plugin, ok := plugins[service]; ok {
		return plugin(t.IP, t.Port, user, pass, timeout)
	}
	return false
}
