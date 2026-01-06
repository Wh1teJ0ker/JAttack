package infogather

import (
	"JAttack/internal/pkg/logger"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ScanConfig struct {
	Target         string `json:"target"`
	Ports          string `json:"ports"`            // 端口范围，例如 "80,443,1000-2000" 或 "common" 或 "all"
	Concurrency    int    `json:"concurrency"`      // 最大并发工作线程数
	Timeout        int    `json:"timeout"`          // 超时时间（毫秒）
	SkipAliveCheck bool   `json:"skip_alive_check"` // 跳过存活检测（ICMP）
	EnableICMP     bool   `json:"enable_icmp"`      // 启用 ICMP 存活检测
	EnablePing     bool   `json:"enable_ping"`      // ICMP 的别名
	EnableUDP      bool   `json:"enable_udp"`       // 启用 UDP 探测
}

type ScanResult struct {
	IP    string `json:"ip"`
	Alive bool   `json:"alive"`
	Ports []int  `json:"ports"`
	Time  string `json:"time"`
}

// StartScan 启动扫描任务
func (s *InfoService) StartScan(config ScanConfig) {
	s.mu.Lock()
	if s.scanCancel != nil {
		s.scanCancel() // 如果已有任务在运行，则取消
	}
	s.scanCtx, s.scanCancel = context.WithCancel(context.Background())
	s.paused.Store(false) // 重置暂停状态
	s.mu.Unlock()

	go s.runScan(config)
}

func (s *InfoService) waitIfPaused() bool {
	for s.paused.Load() {
		select {
		case <-s.scanCtx.Done():
			return true // 任务已取消
		case <-time.After(100 * time.Millisecond):
			continue
		}
	}
	select {
	case <-s.scanCtx.Done():
		return true // 任务已取消
	default:
		return false
	}
}

func (s *InfoService) runScan(config ScanConfig) {
	defer func() {
		// Give a small buffer for previous events to be processed by frontend
		time.Sleep(200 * time.Millisecond)
		s.emitLog("扫描任务完成")
		time.Sleep(100 * time.Millisecond)
		s.emitComplete()
	}()

	logger.Info("开始扫描任务", "目标", config.Target, "并发", config.Concurrency, "超时(ms)", config.Timeout)
	s.emitLog(fmt.Sprintf("开始扫描任务: %s", config.Target))

	// 如果需要，解析域名
	ips, err := parseTarget(config.Target)
	if err != nil {
		logger.Error("目标解析失败", "错误", err.Error())
		s.emitLog(fmt.Sprintf("目标解析失败: %v", err))
		return
	}

	if s.waitIfPaused() {
		return
	}

	s.emitLog(fmt.Sprintf("解析到 %d 个IP地址", len(ips)))

	var targetIPs []string

	if config.SkipAliveCheck {
		s.emitLog("跳过主机存活检测，直接进行扫描探测")
		targetIPs = ips
	} else if config.EnableICMP {
		s.emitLog("正在进行ICMP主机存活探测...")
		targetIPs = s.icmpScan(ips, config.Concurrency, config.Timeout)
		s.emitLog(fmt.Sprintf("存活主机数量: %d", len(targetIPs)))
	} else {
		// 如果禁用了 ICMP 且未显式跳过存活检测，默认假设所有 IP 都是目标
		targetIPs = ips
	}

	if s.waitIfPaused() {
		return
	}

	if len(targetIPs) == 0 {
		s.emitLog("没有发现存活主机或目标列表为空")
		return
	}

	// 扫描探测
	ports := parsePorts(config.Ports)
	if len(ports) > 0 {
		s.emitLog(fmt.Sprintf("开始TCP扫描探测，端口数量: %d", len(ports)))

		timeout := 2 * time.Second
		if config.Timeout > 0 {
			timeout = time.Duration(config.Timeout) * time.Millisecond
		}

		s.portScan(targetIPs, ports, config.Concurrency, timeout)
	}

	if s.waitIfPaused() {
		return
	}

	// UDP 扫描 (如果启用，目前仅扫描常用端口)
	if config.EnableUDP {
		s.emitLog("开始UDP服务探测...")
		udpPorts := []int{53, 123, 161} // DNS, NTP, SNMP
		s.udpScan(targetIPs, udpPorts, config.Concurrency, config.Timeout)
	}

	s.emitLog("扫描任务完成")
}

func (s *InfoService) icmpScan(ips []string, concurrency int, timeoutMs int) []string {
	var aliveIPs []string
	var mu sync.Mutex
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	timeout := time.Second
	if timeoutMs > 0 {
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}

	total := len(ips)
	for i, ip := range ips {
		if s.waitIfPaused() {
			return aliveIPs
		}

		sem <- struct{}{}
		wg.Add(1)

		// 进度更新
		if i%10 == 0 || i == total-1 {
			s.emitProgress(float64(i+1) / float64(total) * 100)
		}

		go func(ipAddr string) {
			defer wg.Done()
			defer func() { <-sem }()

			// 检查 worker 中的上下文
			select {
			case <-s.scanCtx.Done():
				return
			default:
			}

			if s.pingHost(ipAddr, timeout) {
				mu.Lock()
				aliveIPs = append(aliveIPs, ipAddr)
				mu.Unlock()
				s.emitLog(fmt.Sprintf("[ICMP] 主机存活: %s", ipAddr))

				// Save to new DB schema asynchronously
				s.dbQueue <- func() {
					_, err := s.dbManager.UpsertAsset(ipAddr, "", true)
					if err != nil {
						logger.Error("保存存活主机失败", "IP", ipAddr, "错误", err)
					}
				}

				s.saveResult(ipAddr, "ICMP", "Alive")
			}
		}(ip)
	}
	wg.Wait()
	return aliveIPs
}

func (s *InfoService) pingHost(ip string, timeout time.Duration) bool {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		return false
	}
	pinger.Count = 1
	pinger.Timeout = timeout
	pinger.SetPrivileged(false) // 尝试非特权模式 (UDP)

	err = pinger.Run()
	if err != nil {
		return false
	}
	return pinger.Statistics().PacketsRecv > 0
}

func (s *InfoService) portScan(ips []string, ports []int, concurrency int, timeout time.Duration) {
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	total := len(ips) * len(ports)
	count := 0

	for _, ip := range ips {
		for _, port := range ports {
			if s.waitIfPaused() {
				return
			}

			sem <- struct{}{}
			wg.Add(1)
			count++

			if count%50 == 0 || count == total {
				s.emitProgress(float64(count) / float64(total) * 100)
			}

			go func(ipAddr string, p int) {
				defer wg.Done()
				defer func() { <-sem }()

				select {
				case <-s.scanCtx.Done():
					return
				default:
				}

				if s.checkPort(ipAddr, p, timeout) {
					info := fmt.Sprintf("%d/tcp open", p)
					s.emitLog(fmt.Sprintf("[TCP] %s:%d 开放", ipAddr, p))

					// Always grab protocol info (Banner)
					// banner := GrabBanner(ipAddr, p, timeout)
					banner := ""
					serviceName := "unknown"

					// Basic service name guess based on port
					if p == 80 || p == 8080 {
						serviceName = "http"
					} else if p == 443 || p == 8443 {
						serviceName = "https"
					} else if p == 22 {
						serviceName = "ssh"
					}

					// Save Asset & Port asynchronously
					s.dbQueue <- func() {
						assetID, err := s.dbManager.UpsertAsset(ipAddr, "", true)
						if err != nil {
							logger.Error("保存主机失败", "IP", ipAddr, "错误", err)
							return
						}

						// Save Port
						portID, _ := s.dbManager.UpsertAssetPort(assetID, p, "tcp", serviceName, "", "", banner, "open")

						// If HTTP/HTTPS, create Web Service placeholder
						if serviceName == "http" || serviceName == "https" {
							protocol := "http"
							if serviceName == "https" {
								protocol = "https"
							}
							url := fmt.Sprintf("%s://%s:%d", protocol, ipAddr, p)
							s.dbManager.UpsertWebService(assetID, portID, url, "", "", banner)
						}
					}

					s.saveResult(ipAddr, "PortScan", info)
				}
			}(ip, port)
		}
	}
	wg.Wait()
}

func (s *InfoService) udpScan(ips []string, ports []int, concurrency int, timeoutMs int) {
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	// timeout := time.Second * 2
	// if timeoutMs > 0 {
	// 	timeout = time.Duration(timeoutMs) * time.Millisecond
	// }

	for _, ip := range ips {
		for _, port := range ports {
			if s.waitIfPaused() {
				return
			}

			sem <- struct{}{}
			wg.Add(1)
			go func(ipAddr string, p int) {
				defer wg.Done()
				defer func() { <-sem }()

				select {
				case <-s.scanCtx.Done():
					return
				default:
				}

				// resp := UDPProbe(ipAddr, p, timeout)
				// if resp != "" {
				// 	s.emitLog(fmt.Sprintf("[UDP] %s:%d %s", ipAddr, p, resp))
				// 	s.saveResult(ipAddr, "UDP", fmt.Sprintf("%d/udp open: %s", p, resp))
				// }
			}(ip, port)
		}
	}
	wg.Wait()
}

func (s *InfoService) checkPort(ip string, port int, timeout time.Duration) bool {
	address := net.JoinHostPort(ip, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (s *InfoService) emitLog(message string) {
	logger.Info(message)
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "scan:log", message)
	}
}

func (s *InfoService) emitProgress(percentage float64) {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "scan:progress", percentage)
	}
}

func (s *InfoService) emitComplete() {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "scan:complete", true)
	}
}

func (s *InfoService) saveResult(target, infoType, content string) {
	// Reuse AddInfo but handle errors silently or log them
	if err := s.AddInfo(target, infoType, content); err != nil {
		logger.Error("保存扫描结果失败", "目标", target, "错误", err.Error())
	}
	// Also emit result to frontend for realtime display
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "scan:result", map[string]string{
			"target":    target,
			"info_type": infoType,
			"content":   content,
			"time":      time.Now().Format("15:04:05"),
		})
	}
}

// Helpers
func parseTarget(target string) ([]string, error) {
	var allIPs []string

	// 处理逗号分隔的多个目标
	targets := strings.Split(target, ",")
	for _, t := range targets {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}

		if strings.Contains(t, "/") {
			// CIDR 网段
			ip, ipnet, err := net.ParseCIDR(t)
			if err != nil {
				return nil, fmt.Errorf("invalid CIDR: %s, error: %v", t, err)
			}
			var ips []string
			for currentIP := ip.Mask(ipnet.Mask); ipnet.Contains(currentIP); inc(currentIP) {
				ips = append(ips, currentIP.String())
			}
			// 移除网络地址和广播地址
			if len(ips) > 2 {
				allIPs = append(allIPs, ips[1:len(ips)-1]...)
			} else {
				allIPs = append(allIPs, ips...)
			}
		} else {
			// 单个 IP 或 域名
			// 先尝试解析为 IP
			if net.ParseIP(t) != nil {
				allIPs = append(allIPs, t)
			} else {
				// 尝试解析域名
				ips, err := net.LookupIP(t)
				if err == nil {
					for _, ip := range ips {
						if ipv4 := ip.To4(); ipv4 != nil {
							allIPs = append(allIPs, ipv4.String())
						}
					}
				} else {
					// 如果解析失败，暂时记录错误或跳过？
					// 这里为了简单，如果不是IP也不是CIDR且解析失败，返回错误
					return nil, fmt.Errorf("invalid target: %s", t)
				}
			}
		}
	}

	if len(allIPs) == 0 {
		return nil, fmt.Errorf("no valid targets found")
	}

	// 去重
	uniqueIPs := make(map[string]bool)
	var result []string
	for _, ip := range allIPs {
		if !uniqueIPs[ip] {
			uniqueIPs[ip] = true
			result = append(result, ip)
		}
	}

	return result, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func parsePorts(portsStr string) []int {
	if portsStr == "all" {
		ports := make([]int, 65535)
		for i := 0; i < 65535; i++ {
			ports[i] = i + 1
		}
		return ports
	}
	if portsStr == "common" {
		// 常用端口 top list
		return []int{
			21, 22, 23, 25, 53, 80, 110, 111, 135, 139, 143, 443, 445, 993, 995,
			1433, 1521, 3306, 3389, 5432, 5900, 6379, 8080, 8443, 27017,
		}
	}

	var ports []int

	parts := strings.Split(portsStr, ",")
	for _, part := range parts {
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) == 2 {
				start, _ := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
				end, _ := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
				for i := start; i <= end; i++ {
					ports = append(ports, i)
				}
			}
		} else {
			p, _ := strconv.Atoi(strings.TrimSpace(part))
			if p > 0 {
				ports = append(ports, p)
			}
		}
	}
	return ports
}
