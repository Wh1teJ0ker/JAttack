package infogather

import (
	"JAttack/internal/config"
	"JAttack/internal/pkg/logger"
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type DirScanConfig struct {
	Target         string   `json:"target"`
	Extensions     []string `json:"extensions"`
	Threads        int      `json:"threads"`
	Timeout        int      `json:"timeout"`
	Exclude404     bool     `json:"exclude_404"`
	Redirects      bool     `json:"redirects"`
	CustomDict     string   `json:"custom_dict"`
	RecursionDepth int      `json:"recursion_depth"`
}

type DirScanResult struct {
	URL         string `json:"url"`
	Status      int    `json:"status"`
	Size        int64  `json:"size"`
	Location    string `json:"location,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Title       string `json:"title,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

// GetDictionaries 返回可用字典文件列表
func (s *InfoService) GetDictionaries() []Dictionary {
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

// SelectDictionaryFile 选择字典文件
func (s *InfoService) SelectDictionaryFile() string {
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

func (s *InfoService) StartDirScan(config DirScanConfig) {
	s.mu.Lock()
	if s.scanCancel != nil {
		s.scanCancel()
	}
	s.scanCtx, s.scanCancel = context.WithCancel(context.Background())
	s.paused.Store(false)
	s.mu.Unlock()

	// Ensure DB column exists (Quick hack for migration)
	if db := s.dbManager.GetDB(); db != nil {
		// Ignore error if column exists
		db.Exec("ALTER TABLE dir_scan_results ADD COLUMN fingerprint TEXT")
	}

	go s.runDirScan(config)
}

func (s *InfoService) runDirScan(config DirScanConfig) {
	defer func() {
		s.emitLog("目录扫描任务完成")
		runtime.EventsEmit(s.ctx, "dirScanComplete")
	}()

	logger.Info("开始目录扫描", "目标", config.Target, "并发", config.Threads, "递归深度", config.RecursionDepth)
	s.emitLog(fmt.Sprintf("开始目录扫描: %s (递归深度: %d)", config.Target, config.RecursionDepth))

	if !strings.HasPrefix(config.Target, "http://") && !strings.HasPrefix(config.Target, "https://") {
		config.Target = "http://" + config.Target
	}
	config.Target = strings.TrimRight(config.Target, "/")

	wordlistPath := config.CustomDict
	if wordlistPath == "" {
		candidates := []string{
			"internal/dict/dicc.txt",
			"data/dicc.txt",
			"dicc.txt",
		}

		for _, path := range candidates {
			if _, err := os.Stat(path); err == nil {
				wordlistPath = path
				break
			}
		}

		if wordlistPath == "" {
			s.emitLog("未找到默认字典文件")
			return
		}
	}

	lines, err := s.loadWordlist(wordlistPath)
	if err != nil {
		s.emitLog(fmt.Sprintf("加载字典失败: %v", err))
		return
	}

	s.emitLog(fmt.Sprintf("字典加载成功，共 %d 行", len(lines)))

	// Resolve WebService ID for new DB schema
	var webServiceID int64
	if s.dbManager != nil {
		u, err := url.Parse(config.Target)
		if err == nil {
			host := u.Hostname()
			portStr := u.Port()
			if portStr == "" {
				if u.Scheme == "https" {
					portStr = "443"
				} else {
					portStr = "80"
				}
			}
			port, _ := strconv.Atoi(portStr)

			ips, err := net.LookupHost(host)
			if err == nil && len(ips) > 0 {
				ip := ips[0]
				// Upsert Asset
				if assetID, err := s.dbManager.UpsertAsset(ip, "", true); err == nil {
					// Upsert Port
					if portID, err := s.dbManager.UpsertAssetPort(assetID, port, "tcp", u.Scheme, "", "", "", "open"); err == nil {
						// Upsert WebService
						webServiceID, _ = s.dbManager.UpsertWebService(assetID, portID, config.Target, "", "", "")
					}
				}
			} else {
				// Fallback if host is IP
				if net.ParseIP(host) != nil {
					if assetID, err := s.dbManager.UpsertAsset(host, "", true); err == nil {
						if portID, err := s.dbManager.UpsertAssetPort(assetID, port, "tcp", u.Scheme, "", "", "", "open"); err == nil {
							webServiceID, _ = s.dbManager.UpsertWebService(assetID, portID, config.Target, "", "", "")
						}
					}
				}
			}
		}
	}

	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Millisecond,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !config.Redirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// BFS Level Management
	currentTargets := []string{config.Target}
	visited := make(map[string]bool)
	visited[config.Target] = true
	var visitedMu sync.Mutex

	type Job struct {
		BaseURL string
		Path    string
	}

	for depth := 0; depth <= config.RecursionDepth; depth++ {
		if len(currentTargets) == 0 {
			break
		}

		if depth > 0 {
			s.emitLog(fmt.Sprintf("进入第 %d 层递归，当前层目标数: %d", depth, len(currentTargets)))
		}

		jobs := make(chan Job, config.Threads*10)
		var wg sync.WaitGroup

		var nextLevelTargets []string
		var nextLevelMu sync.Mutex

		// Start workers
		for i := 0; i < config.Threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					if s.waitIfPaused() {
						return
					}

					select {
					case <-s.scanCtx.Done():
						return
					default:
					}

					reqURL := job.BaseURL + "/" + strings.TrimLeft(job.Path, "/")

					req, err := http.NewRequestWithContext(s.scanCtx, "GET", reqURL, nil)
					if err != nil {
						continue
					}
					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

					resp, err := client.Do(req)
					if err != nil {
						continue
					}

					if config.Exclude404 && resp.StatusCode == 404 {
						resp.Body.Close()
						continue
					}

					body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
					resp.Body.Close()
					size := int64(len(body))

					contentType := resp.Header.Get("Content-Type")
					title := ""
					if strings.Contains(strings.ToLower(contentType), "text/html") {
						re := regexp.MustCompile(`(?i)<title>(.*?)</title>`)
						matches := re.FindStringSubmatch(string(body))
						if len(matches) > 1 {
							title = strings.TrimSpace(matches[1])
							if len(title) > 100 {
								title = title[:100] + "..."
							}
						}
					}

					location := ""
					if resp.StatusCode >= 300 && resp.StatusCode < 400 {
						location = resp.Header.Get("Location")
					}

					fingerprint := ""

					result := DirScanResult{
						URL:         reqURL,
						Status:      resp.StatusCode,
						Size:        size,
						Location:    location,
						Fingerprint: fingerprint,
						Title:       title,
						ContentType: contentType,
					}

					runtime.EventsEmit(s.ctx, "dirScanResult", result)

					if err := s.saveDirScanResult(config.Target, result, webServiceID); err != nil {
						logger.Error("保存目录扫描结果失败", "url", reqURL, "error", err)
					}

					// Recursion Check
					if depth < config.RecursionDepth {
						isDir := strings.HasSuffix(reqURL, "/") || (location != "" && strings.HasSuffix(location, "/"))
						if resp.StatusCode == 403 {
							isDir = true
						}

						if isDir {
							nextTarget := reqURL
							if location != "" {
								if strings.HasPrefix(location, "http") {
									nextTarget = location
								} else if strings.HasPrefix(location, "/") {
									u, _ := url.Parse(reqURL)
									nextTarget = u.Scheme + "://" + u.Host + location
								} else {
									nextTarget = strings.TrimRight(reqURL, "/") + "/" + location
								}
							}
							nextTarget = strings.TrimRight(nextTarget, "/")

							visitedMu.Lock()
							if !visited[nextTarget] {
								visited[nextTarget] = true
								visitedMu.Unlock()

								nextLevelMu.Lock()
								nextLevelTargets = append(nextLevelTargets, nextTarget)
								nextLevelMu.Unlock()
							} else {
								visitedMu.Unlock()
							}
						}
					}
				}
			}()
		}

		go func() {
			defer close(jobs)
			for _, target := range currentTargets {
				for _, line := range lines {
					select {
					case <-s.scanCtx.Done():
						return
					default:
					}

					if strings.Contains(line, "%EXT%") {
						for _, ext := range config.Extensions {
							cleanExt := strings.TrimPrefix(ext, ".")
							path := strings.ReplaceAll(line, "%EXT%", cleanExt)
							jobs <- Job{BaseURL: target, Path: path}
						}
					} else {
						jobs <- Job{BaseURL: target, Path: line}
					}
				}
			}
		}()

		wg.Wait()
		currentTargets = nextLevelTargets

		select {
		case <-s.scanCtx.Done():
			return
		default:
		}
	}
}

func (s *InfoService) loadWordlist(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" && !strings.HasPrefix(text, "#") {
			lines = append(lines, text)
		}
	}
	return lines, scanner.Err()
}

func (s *InfoService) saveDirScanResult(target string, result DirScanResult, webServiceID int64) error {
	db := s.dbManager.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Save to new IP-centric table if possible
	if webServiceID > 0 {
		// Extract path from URL
		u, err := url.Parse(result.URL)
		path := "/"
		if err == nil {
			path = u.Path
		}
		s.dbManager.AddWebDirectory(webServiceID, path, result.Status, int(result.Size), result.Title, result.ContentType, result.Location)
	}

	query := `INSERT INTO dir_scan_results (target, url, status, size, location, fingerprint) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, target, result.URL, result.Status, result.Size, result.Location, result.Fingerprint)
	return err
}
