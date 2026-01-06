package vuln

import (
	"JAttack/internal/db"
	"JAttack/internal/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Vulnerability 定义漏洞信息结构体
type Vulnerability struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Product     string    `json:"product"`      // 漏洞归属/影响组件
	VulnType    string    `json:"vuln_type"`    // 漏洞类型
	Severity    string    `json:"severity"`     // 严重程度
	Description string    `json:"description"`  // 简要描述
	Details     string    `json:"details"`      // 详细信息
	Status      string    `json:"status"`       // 状态
	PocType     string    `json:"poc_type"`     // 验证类型: nuclei, python
	PocContent  string    `json:"poc_content"`  // 脚本路径或内容
	Reference   string    `json:"reference"`    // 参考链接
	CreatedAt   time.Time `json:"created_at" ts_type:"string"`
}

// VulnService 漏洞管理服务
type VulnService struct {
	ctx       context.Context
	dbManager *db.Manager
}

// NewVulnService 创建漏洞管理服务实例
func NewVulnService(dbManager *db.Manager) *VulnService {
	return &VulnService{
		dbManager: dbManager,
	}
}

// Startup 在服务启动时调用
func (s *VulnService) Startup(ctx context.Context) {
	s.ctx = ctx
	logger.Info("漏洞管理服务已启动")
	s.migrateDB()
}

// migrateDB 检查并添加缺失的数据库字段
func (s *VulnService) migrateDB() {
	database := s.dbManager.GetDB()
	if database == nil {
		return
	}

	columns := []string{"product", "vuln_type", "details", "poc_type", "poc_content", "reference"}
	for _, col := range columns {
		// SQLite 不支持 IF NOT EXISTS ADD COLUMN，所以我们尝试添加，如果失败则假设列已存在
		query := fmt.Sprintf("ALTER TABLE vulnerabilities ADD COLUMN %s TEXT", col)
		_, err := database.Exec(query)
		if err != nil {
			// 如果不是"duplicate column name"错误，记录警告
			if !strings.Contains(err.Error(), "duplicate column name") {
				logger.Warn("尝试添加数据库列失败 (可能已存在)", "列名", col, "错误", err.Error())
			}
		} else {
			logger.Info("成功添加数据库列", "列名", col)
		}
	}
}

// AddVuln 添加新的漏洞记录
func (s *VulnService) AddVuln(v Vulnerability) error {
	logger.Info("正在添加漏洞", "名称", v.Name)
	db := s.dbManager.GetDB()
	if db == nil {
		return errors.New("数据库未初始化")
	}

	query := `INSERT INTO vulnerabilities (name, product, vuln_type, severity, description, details, status, poc_type, poc_content, reference) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, v.Name, v.Product, v.VulnType, v.Severity, v.Description, v.Details, v.Status, v.PocType, v.PocContent, v.Reference)
	if err != nil {
		logger.Error("添加漏洞失败", "错误", err.Error())
		return err
	}
	return nil
}

// UpdateVuln 更新漏洞记录
func (s *VulnService) UpdateVuln(v Vulnerability) error {
	logger.Info("正在更新漏洞", "ID", v.ID)
	db := s.dbManager.GetDB()
	if db == nil {
		return errors.New("数据库未初始化")
	}

	query := `UPDATE vulnerabilities SET name=?, product=?, vuln_type=?, severity=?, description=?, details=?, status=?, poc_type=?, poc_content=?, reference=? WHERE id=?`
	_, err := db.Exec(query, v.Name, v.Product, v.VulnType, v.Severity, v.Description, v.Details, v.Status, v.PocType, v.PocContent, v.Reference, v.ID)
	if err != nil {
		logger.Error("更新漏洞失败", "错误", err.Error())
		return err
	}
	return nil
}

// DeleteVuln 删除漏洞记录
func (s *VulnService) DeleteVuln(id int) error {
	logger.Info("正在删除漏洞", "ID", id)
	db := s.dbManager.GetDB()
	if db == nil {
		return errors.New("数据库未初始化")
	}

	_, err := db.Exec("DELETE FROM vulnerabilities WHERE id=?", id)
	if err != nil {
		logger.Error("删除漏洞失败", "错误", err.Error())
		return err
	}
	return nil
}

// SearchVulns 搜索漏洞
func (s *VulnService) SearchVulns(keyword string) ([]Vulnerability, error) {
	db := s.dbManager.GetDB()
	if db == nil {
		return nil, nil
	}

	var rows *sql.Rows
	var err error

	if keyword == "" {
		rows, err = db.Query(`SELECT id, name, product, vuln_type, severity, description, details, status, poc_type, poc_content, reference, created_at FROM vulnerabilities ORDER BY created_at DESC`)
	} else {
		// 优化搜索逻辑：支持多字段模糊匹配
		query := `SELECT id, name, product, vuln_type, severity, description, details, status, poc_type, poc_content, reference, created_at 
				  FROM vulnerabilities 
				  WHERE name LIKE ? OR product LIKE ? OR vuln_type LIKE ? OR description LIKE ? OR details LIKE ? 
				  ORDER BY created_at DESC`
		pattern := "%" + keyword + "%"
		rows, err = db.Query(query, pattern, pattern, pattern, pattern, pattern)
	}

	if err != nil {
		logger.Error("查询漏洞失败", "错误", err.Error())
		return nil, err
	}
	defer rows.Close()

	var results []Vulnerability
	for rows.Next() {
		var v Vulnerability
		// 处理可能为 NULL 的字段
		var product, vulnType, details, pocType, pocContent, reference sql.NullString
		
		if err := rows.Scan(&v.ID, &v.Name, &product, &vulnType, &v.Severity, &v.Description, &details, &v.Status, &pocType, &pocContent, &reference, &v.CreatedAt); err != nil {
			logger.Error("扫描漏洞行失败", "错误", err.Error())
			continue
		}
		
		v.Product = product.String
		v.VulnType = vulnType.String
		v.Details = details.String
		v.PocType = pocType.String
		v.PocContent = pocContent.String
		v.Reference = reference.String
		
		results = append(results, v)
	}
	return results, nil
}

// ListVulns 列出所有漏洞 (兼容旧接口，实际上调用 SearchVulns)
func (s *VulnService) ListVulns() ([]Vulnerability, error) {
	return s.SearchVulns("")
}

// VerifyVuln 验证漏洞
func (s *VulnService) VerifyVuln(id int, target string) (string, error) {
	logger.Info("开始验证漏洞", "ID", id, "目标", target)
	
	// 获取漏洞信息以得到 POC 路径
	db := s.dbManager.GetDB()
	if db == nil {
		return "", errors.New("数据库未初始化")
	}
	
	var v Vulnerability
	err := db.QueryRow("SELECT poc_type, poc_content FROM vulnerabilities WHERE id=?", id).Scan(&v.PocType, &v.PocContent)
	if err != nil {
		return "", fmt.Errorf("未找到指定漏洞或POC信息缺失: %v", err)
	}

	if v.PocContent == "" {
		return "", errors.New("未配置 POC 路径")
	}

	var cmd *exec.Cmd
	
	switch strings.ToLower(v.PocType) {
	case "nuclei":
		// nuclei -t <template_path> -u <target>
		logger.Info("执行 Nuclei 验证", "Template", v.PocContent)
		cmd = exec.Command("nuclei", "-t", v.PocContent, "-u", target, "-nc") // -nc: no color
	case "python":
		// python3 <script_path> <target>
		// 假设脚本接受目标 URL 作为第一个参数
		logger.Info("执行 Python 验证", "Script", v.PocContent)
		cmd = exec.Command("python3", v.PocContent, target)
	default:
		return "", fmt.Errorf("不支持的验证类型: %s", v.PocType)
	}

	// 执行命令并获取输出
	output, err := cmd.CombinedOutput()
	result := string(output)
	
	if err != nil {
		logger.Error("验证执行出错", "错误", err.Error(), "输出", result)
		return fmt.Sprintf("执行出错: %v\n\n输出:\n%s", err, result), nil // 返回错误信息但不作为 error 返回，以便前端显示输出
	}

	return result, nil
}
