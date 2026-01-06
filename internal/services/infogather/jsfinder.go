package infogather

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"JAttack/internal/db"
	"JAttack/internal/pkg/logger"
)

// Pre-compiled regex rules for sensitive information
var secretRules = []struct {
	Name  string
	Regex *regexp.Regexp
}{
	// Cloud Providers
	{"Aliyun AccessKey", regexp.MustCompile(`\bLTAI[a-zA-Z0-9]{20}\b`)},
	{"Aliyun Secret", regexp.MustCompile(`(?i)(?:aliyun|access_key|access_token).{0,20}['"]([0-9a-zA-Z]{30})['"]`)},
	{"AWS AccessKey", regexp.MustCompile(`\b((?:AKIA|ABIA|ACCA|ASIA)[0-9A-Z]{16})\b`)},
	{"AWS Secret", regexp.MustCompile(`(?i)aws.{0,20}['"]([0-9a-zA-Z\/+]{40})['"]`)},
	{"Google API Key", regexp.MustCompile(`\bAIza[0-9A-Za-z\\-_]{35}\b`)},
	{"Google OAuth", regexp.MustCompile(`[0-9]+-[0-9A-Za-z_]{32}\.apps\.googleusercontent\.com`)},
	{"Tencent SecretId", regexp.MustCompile(`\bAKID[a-zA-Z0-9]{32}\b`)},
	{"Tencent SecretKey", regexp.MustCompile(`\b(?i)SecretKey\s*[:=]\s*['"]([A-Za-z0-9]{32})['"]`)},
	{"Azure SharedKey", regexp.MustCompile(`[a-z0-9]+\.blob\.core\.windows\.net`)},
	{"Heroku API Key", regexp.MustCompile(`(?i)heroku.{0,20}[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)},

	// SaaS & Tools
	{"Slack Token", regexp.MustCompile(`xox[baprs]-[a-zA-Z0-9]{10,48}`)},
	{"Slack Webhook", regexp.MustCompile(`https://hooks\.slack\.com/services/T[0-9A-Z]{9}/B[0-9A-Z]{9}/[a-zA-Z0-9]{24}`)},
	{"GitHub Token", regexp.MustCompile(`(gh[pousr]_[a-zA-Z0-9]{36,255})`)},
	{"Stripe Key", regexp.MustCompile(`(?:r|s)k_(?:live|test)_[0-9a-zA-Z]{24}`)},
	{"PayPal Token", regexp.MustCompile(`access_token\$production\$[0-9a-z]{16}\$[0-9a-f]{32}`)},
	{"Twilio SID", regexp.MustCompile(`\bAC[a-z0-9]{32}\b`)},
	{"Mailgun API Key", regexp.MustCompile(`key-[0-9a-zA-Z]{32}`)},
	{"Telegram Bot Token", regexp.MustCompile(`[0-9]+:AA[0-9A-Za-z\-_]{33}`)},
	{"Facebook Access Token", regexp.MustCompile(`EAACEdEose0cBA[0-9A-Za-z]+`)},
	{"Square Access Token", regexp.MustCompile(`sq0atp-[0-9A-Za-z\-_]{22}`)},
	{"Square OAuth Secret", regexp.MustCompile(`sq0csp-[0-9A-Za-z\-_]{43}`)},

	// Crypto
	{"Private Key", regexp.MustCompile(`-----BEGIN ((EC|PGP|DSA|RSA|OPENSSH) )?PRIVATE KEY( BLOCK)?-----`)},

	// Generic High Entropy
	{"Generic API Key", regexp.MustCompile(`(?i)(?:api_?key|access_?token|secret|password|auth|auth_token)\s*[:=]\s*['"]([a-zA-Z0-9_\-]{16,64})['"]`)},
	{"Bearer Token", regexp.MustCompile(`(?i)Bearer\s+[a-zA-Z0-9\-\._~\+/]+=*`)},
	{"URI Credentials", regexp.MustCompile(`(?i)[a-z]+://[^/\s]+:[^/\s]+@[^/\s]+`)},
	{"JWT", regexp.MustCompile(`eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]*`)},
	{"Basic Auth", regexp.MustCompile(`(?i)Basic\s+[a-zA-Z0-9+/]+={0,2}`)},
}

type JSFinderService struct {
	ctx       context.Context
	client    *http.Client
	dbManager *db.Manager
}

type JSFinderOptions struct {
	DeepScan     bool `json:"deep_scan"`     // Recursively scan JS files
	ActiveScan   bool `json:"active_scan"`   // Active request to verify/find more
	DangerFilter bool `json:"danger_filter"` // Skip dangerous operations in active scan
	Concurrency  int  `json:"concurrency"`   // Concurrency count
	Timeout      int  `json:"timeout"`       // Timeout in milliseconds
}

type JSFindResult struct {
	URL           string   `json:"url"`
	Endpoints     []string `json:"endpoints"`
	JSFiles       []string `json:"js_files"`
	SensitiveInfo []string `json:"sensitive_info"`
	Error         string   `json:"error,omitempty"`
}

func NewJSFinderService(dbManager *db.Manager) *JSFinderService {
	return &JSFinderService{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		dbManager: dbManager,
	}
}

func (s *JSFinderService) Startup(ctx context.Context) {
	s.ctx = ctx
}

func (s *JSFinderService) FindJS(targetURL string, options JSFinderOptions) JSFindResult {
	result := JSFindResult{
		URL:           targetURL,
		Endpoints:     []string{},
		JSFiles:       []string{},
		SensitiveInfo: []string{},
	}

	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "http://" + targetURL
	}

	logger.Info("Starting JSFinder", "target", targetURL, "deep", options.DeepScan, "active", options.ActiveScan)

	// Defaults
	concurrency := 10
	if options.Concurrency > 0 {
		concurrency = options.Concurrency
	}
	timeout := 10 * time.Second
	if options.Timeout > 0 {
		timeout = time.Duration(options.Timeout) * time.Millisecond
	}

	// Thread-safe containers
	var mu sync.Mutex
	visited := make(map[string]bool)
	jsQueue := make([]string, 0)

	// Helper to add data
	addData := func(eps []string, js []string, info []string) {
		mu.Lock()
		defer mu.Unlock()
		result.Endpoints = append(result.Endpoints, eps...)
		result.JSFiles = append(result.JSFiles, js...)
		result.SensitiveInfo = append(result.SensitiveInfo, info...)

		// Add new JS to queue if not visited
		for _, j := range js {
			// Normalize JS URL
			fullJS, err := resolveURL(targetURL, j) // simplified resolution context
			if err == nil && !visited[fullJS] {
				// Only queue if we haven't seen it
				// Note: Actual queueing happens in the main loop logic,
				// but here we just collect them.
			}
			_ = fullJS
		}
	}

	// 1. Fetch Main Page
	body, err := s.fetch(targetURL, timeout)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to fetch target: %v", err)
		return result
	}
	visited[targetURL] = true

	// Analyze Main Page
	eps, jsLinks, info := s.analyzeContent(body)
	addData(eps, jsLinks, info)

	// Prepare queue for JS files
	// We use a simple map to avoid duplicates in queue
	queued := make(map[string]bool)
	queued[targetURL] = true

	for _, js := range jsLinks {
		fullJS, err := resolveURL(targetURL, js)
		if err == nil && !queued[fullJS] {
			jsQueue = append(jsQueue, fullJS)
			queued[fullJS] = true
		}
	}

	// 2. Process JS Files (BFS / Concurrent)
	// We will limit recursion depth by just processing the queue once (Level 1)
	// If DeepScan is true, we could add newly found JS from Level 1 to a Level 2 queue.
	// For now, let's just process all discovered JS files from the main page + any from them if DeepScan is true.

	// A simple worker pool
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	// If DeepScan is enabled, we might want to do 2 levels.
	// Level 1: JS found on HTML.
	// Level 2: JS found in Level 1 JS.

	levels := 1
	if options.DeepScan {
		levels = 2
	}

	currentLevelQueue := jsQueue

	for l := 0; l < levels; l++ {
		nextLevelQueue := []string{}
		var nlMu sync.Mutex

		logger.Info("Processing JS Level", "level", l+1, "count", len(currentLevelQueue))

		for _, jsURL := range currentLevelQueue {
			wg.Add(1)
			sem <- struct{}{}
			go func(urlStr string) {
				defer wg.Done()
				defer func() { <-sem }()

				// Fetch
				content, err := s.fetch(urlStr, timeout)
				if err != nil {
					return
				}

				// Analyze
				e, j, i := s.analyzeContent(content)

				// Mark info source
				for idx := range i {
					i[idx] = fmt.Sprintf("[%s] %s", urlStr, i[idx])
				}

				addData(e, j, i)

				// If DeepScan, collect JS for next level
				if options.DeepScan && l < levels-1 {
					nlMu.Lock()
					for _, newJS := range j {
						full, err := resolveURL(urlStr, newJS)
						if err == nil {
							nextLevelQueue = append(nextLevelQueue, full)
						}
					}
					nlMu.Unlock()
				}

				// If ActiveScan, try to poke endpoints?
				// API Sword does active requests.
				// We can do it in a separate phase after collection to avoid noise/dos.
			}(jsURL)
		}
		wg.Wait()

		// Filter duplicates for next level
		if options.DeepScan {
			uniqueNext := []string{}
			for _, u := range nextLevelQueue {
				if !queued[u] {
					queued[u] = true
					uniqueNext = append(uniqueNext, u)
				}
			}
			currentLevelQueue = uniqueNext
		}
	}

	// 3. Deduplicate results
	result.Endpoints = unique(result.Endpoints)
	result.JSFiles = unique(result.JSFiles)
	result.SensitiveInfo = unique(result.SensitiveInfo)

	// 4. Active Scan (Verification)
	if options.ActiveScan {
		logger.Info("Starting Active Scan on endpoints", "count", len(result.Endpoints))
		activeResults := s.performActiveScan(targetURL, result.Endpoints, options.DangerFilter, concurrency, timeout)
		// Merge active results (e.g., mark them or add to sensitive info if interesting)
		// For now, let's just add found valid endpoints to sensitive info as "Verified API"
		for _, r := range activeResults {
			result.SensitiveInfo = append(result.SensitiveInfo, "Verified API: "+r)
		}
	}

	// 5. Save Results to DB
	s.saveResults(result)

	return result
}

func (s *JSFinderService) saveResults(result JSFindResult) {
	if s.dbManager == nil {
		return
	}

	// Parse Target URL
	u, err := url.Parse(result.URL)
	if err != nil {
		logger.Error("Failed to parse target URL for saving", "url", result.URL, "error", err)
		return
	}

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

	// Resolve IP
	ips, err := net.LookupHost(host)
	if err != nil || len(ips) == 0 {
		// Fallback: if host is IP, use it
		if net.ParseIP(host) != nil {
			ips = []string{host}
		} else {
			logger.Error("Failed to resolve host for saving", "host", host, "error", err)
			return
		}
	}
	ip := ips[0] // Use first IP

	// Upsert Asset
	assetID, err := s.dbManager.UpsertAsset(ip, "", true)
	if err != nil {
		logger.Error("Failed to upsert asset", "ip", ip, "error", err)
		return
	}

	// Upsert Port
	portID, err := s.dbManager.UpsertAssetPort(assetID, port, "tcp", u.Scheme, "", "", "", "open")
	if err != nil {
		logger.Error("Failed to upsert port", "port", port, "error", err)
		return
	}

	// Upsert WebService
	webServiceID, err := s.dbManager.UpsertWebService(assetID, portID, result.URL, "", "", "")
	if err != nil {
		logger.Error("Failed to upsert web service", "url", result.URL, "error", err)
		return
	}

	// Save JS Files
	for _, js := range result.JSFiles {
		fullJS, _ := resolveURL(result.URL, js)
		s.dbManager.AddWebJSFile(webServiceID, js, fullJS)
	}

	// Save Sensitive Info
	for _, info := range result.SensitiveInfo {
		// info format: "[source_url] Type: Value"
		source := result.URL
		content := info
		infoType := "Sensitive Info"

		if strings.HasPrefix(info, "[") {
			end := strings.Index(info, "]")
			if end > 1 {
				source = info[1:end]
				content = strings.TrimSpace(info[end+1:])
			}
		}

		parts := strings.SplitN(content, ":", 2)
		if len(parts) == 2 {
			infoType = strings.TrimSpace(parts[0])
			content = strings.TrimSpace(parts[1])
		}

		s.dbManager.AddSensitiveResult(webServiceID, source, infoType, content, "")
	}
}

func (s *JSFinderService) fetch(urlStr string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Limit read to avoid huge files
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // 5MB limit
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

func (s *JSFinderService) analyzeContent(content string) ([]string, []string, []string) {
	var endpoints []string
	var jsFiles []string
	var sensitiveInfo []string

	// 1. Extract Paths / Endpoints
	// Matches: "/api/user", "v1/config", etc.
	// Quote-bounded paths
	pathRegex := regexp.MustCompile(`(?:"|')(((?:/|\.\./)[a-zA-Z0-9_./-]+)|(https?://[a-zA-Z0-9_./-]+))(?:"|')`)
	pathMatches := pathRegex.FindAllStringSubmatch(content, -1)

	for _, m := range pathMatches {
		if len(m) > 1 {
			raw := m[1]
			// Filter
			if isValidEndpoint(raw) {
				if strings.HasSuffix(raw, ".js") {
					jsFiles = append(jsFiles, raw)
				} else {
					endpoints = append(endpoints, raw)
				}
			}
		}
	}

	// 2. Sensitive Info - Rule Based
	for _, rule := range secretRules {
		// If regex has submatches, we prefer the first group (actual value without quotes/keys)
		if rule.Regex.NumSubexp() > 0 {
			matches := rule.Regex.FindAllStringSubmatch(content, -1)
			for _, m := range matches {
				if len(m) > 1 {
					val := m[1]
					if len(val) > 100 {
						val = val[:100] + "..."
					}
					sensitiveInfo = append(sensitiveInfo, fmt.Sprintf("%s: %s", rule.Name, val))
				}
			}
		} else {
			matches := rule.Regex.FindAllString(content, -1)
			for _, m := range matches {
				val := m
				if len(val) > 100 {
					val = val[:100] + "..."
				}
				sensitiveInfo = append(sensitiveInfo, fmt.Sprintf("%s: %s", rule.Name, val))
			}
		}
	}

	// 3. PII & Other Info
	// IP
	ipRegex := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	for _, ip := range ipRegex.FindAllString(content, -1) {
		if !strings.HasPrefix(ip, "0.") && !strings.HasPrefix(ip, "127.") && !strings.HasPrefix(ip, "255.") {
			sensitiveInfo = append(sensitiveInfo, "IP: "+ip)
		}
	}

	// Phone & Email
	phoneRegex := regexp.MustCompile(`\b1[3-9]\d{9}\b`)
	for _, p := range phoneRegex.FindAllString(content, -1) {
		sensitiveInfo = append(sensitiveInfo, "Phone: "+p)
	}

	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	for _, e := range emailRegex.FindAllString(content, -1) {
		if !strings.HasSuffix(e, ".png") && !strings.HasSuffix(e, ".jpg") && !strings.HasSuffix(e, ".svg") {
			sensitiveInfo = append(sensitiveInfo, "Email: "+e)
		}
	}

	return endpoints, jsFiles, sensitiveInfo
}

func (s *JSFinderService) performActiveScan(baseURL string, endpoints []string, dangerFilter bool, concurrency int, timeout time.Duration) []string {
	var valid []string
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex

	base, _ := url.Parse(baseURL)

	for _, ep := range endpoints {
		// Filter dangerous
		if dangerFilter {
			lower := strings.ToLower(ep)
			if strings.Contains(lower, "del") || strings.Contains(lower, "remove") ||
				strings.Contains(lower, "logout") || strings.Contains(lower, "exit") {
				continue
			}
		}

		// Construct URL
		target := ep
		if !strings.HasPrefix(ep, "http") {
			u, err := url.Parse(ep)
			if err != nil {
				continue
			}
			target = base.ResolveReference(u).String()
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(t string) {
			defer wg.Done()
			defer func() { <-sem }()

			// Try HEAD first
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, "HEAD", t, nil)
			req.Header.Set("User-Agent", "Mozilla/5.0")
			resp, err := s.client.Do(req)

			// If Method Not Allowed, try GET
			if err == nil && resp.StatusCode == 405 {
				ctx2, cancel2 := context.WithTimeout(context.Background(), timeout)
				defer cancel2()
				req, _ = http.NewRequestWithContext(ctx2, "GET", t, nil)
				resp, err = s.client.Do(req)
			}

			if err == nil {
				defer resp.Body.Close()
				// 200 OK, 401 Unauthorized, 403 Forbidden are interesting
				if resp.StatusCode == 200 || resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 500 {
					mu.Lock()
					valid = append(valid, fmt.Sprintf("[%d] %s", resp.StatusCode, t))
					mu.Unlock()
				}
			}
		}(target)
	}
	wg.Wait()
	return valid
}

func isValidEndpoint(s string) bool {
	// Filter common static assets
	exts := []string{".png", ".jpg", ".jpeg", ".gif", ".svg", ".css", ".ico", ".woff", ".woff2", ".ttf"}
	for _, e := range exts {
		if strings.HasSuffix(strings.ToLower(s), e) {
			return false
		}
	}
	// Filter overly short or long
	if len(s) < 2 || len(s) > 200 {
		return false
	}
	// Filter standard libs often found in requires
	if !strings.Contains(s, "/") && !strings.HasSuffix(s, ".js") {
		return false // Likely just a variable name or package name
	}
	return true
}

func resolveURL(base, ref string) (string, error) {
	u, err := url.Parse(ref)
	if err != nil {
		return "", err
	}
	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	return b.ResolveReference(u).String(), nil
}

func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
