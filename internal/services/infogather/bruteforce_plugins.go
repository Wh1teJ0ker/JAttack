package infogather

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hirochachacha/go-smb2"
	"github.com/jlaffaye/ftp"
	_ "github.com/lib/pq"
	"github.com/masterzen/winrm"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/mitchellh/go-vnc"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/ssh"

	"github.com/tomatome/grdp/core"
	"github.com/tomatome/grdp/glog"
	"github.com/tomatome/grdp/plugin"
	"github.com/tomatome/grdp/protocol/nla"
	"github.com/tomatome/grdp/protocol/pdu"
	"github.com/tomatome/grdp/protocol/sec"
	"github.com/tomatome/grdp/protocol/t125"
	"github.com/tomatome/grdp/protocol/tpkt"
	"github.com/tomatome/grdp/protocol/x224"
)

// BruteForcePlugin 定义爆破插件函数签名
type BruteForcePlugin func(host string, port int, user, pass string, timeout time.Duration) bool

// 插件映射表
var plugins = map[string]BruteForcePlugin{
	"ssh":        trySSH,
	"ftp":        tryFTP,
	"mysql":      tryMySQL,
	"redis":      tryRedis,
	"postgres":   tryPostgres,
	"postgresql": tryPostgres,
	"mssql":      tryMSSQL,
	"mongodb":    tryMongoDB,
	"smb":        trySMB,
	"telnet":     tryTelnet,
	"vnc":        tryVNC,
	"ldap":       tryLDAP,
	"winrm":      tryWinRM,
	"rdp":        tryRDP, // RDP 仅支持探测/基本连接
}

// trySSH 尝试 SSH 登录
func trySSH(host string, port int, user, pass string, timeout time.Duration) bool {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}
	client, err := ssh.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)), config)
	if err != nil {
		return false
	}
	client.Close()
	return true
}

// tryFTP 尝试 FTP 登录
func tryFTP(host string, port int, user, pass string, timeout time.Duration) bool {
	c, err := ftp.Dial(fmt.Sprintf("%s:%d", host, port), ftp.DialWithTimeout(timeout))
	if err != nil {
		return false
	}
	defer c.Quit()

	err = c.Login(user, pass)
	return err == nil
}

// tryMySQL 尝试 MySQL 登录
func tryMySQL(host string, port int, user, pass string, timeout time.Duration) bool {
	// timeout for connection, read, write
	timeoutSec := int(timeout.Seconds())
	if timeoutSec < 1 {
		timeoutSec = 1
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=%ds&readTimeout=%ds&writeTimeout=%ds",
		user, pass, host, port, timeoutSec, timeoutSec, timeoutSec)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return false
	}
	defer db.Close()

	// Open 不会建立连接，需要 Ping
	err = db.Ping()
	return err == nil
}

// tryPostgres 尝试 PostgreSQL 登录
func tryPostgres(host string, port int, user, pass string, timeout time.Duration) bool {
	// postgres://user:password@host:port/dbname?sslmode=disable
	// dbname 默认为 'postgres' 或 user
	timeoutSec := int(timeout.Seconds())
	if timeoutSec < 1 {
		timeoutSec = 1
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=disable&connect_timeout=%d",
		url.QueryEscape(user), url.QueryEscape(pass), host, port, timeoutSec)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return false
	}
	defer db.Close()

	err = db.Ping()
	return err == nil
}

// tryRedis 尝试 Redis 登录
func tryRedis(host string, port int, user, pass string, timeout time.Duration) bool {
	rdb := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%d", host, port),
		Password:    pass, // no password set
		DB:          0,    // use default DB
		Username:    user, // Redis 6+ ACL, usually empty for older
		DialTimeout: timeout,
	})
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	rdb.Close()
	return err == nil
}

// tryMSSQL 尝试 MSSQL 登录
func tryMSSQL(host string, port int, user, pass string, timeout time.Duration) bool {
	query := url.Values{}
	query.Add("database", "master")
	timeoutSec := int(timeout.Seconds())
	if timeoutSec < 1 {
		timeoutSec = 1
	}
	query.Add("connection timeout", strconv.Itoa(timeoutSec))

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(user, pass),
		Host:     fmt.Sprintf("%s:%d", host, port),
		RawQuery: query.Encode(),
	}

	db, err := sql.Open("sqlserver", u.String())
	if err != nil {
		return false
	}
	defer db.Close()

	err = db.Ping()
	return err == nil
}

// tryMongoDB 尝试 MongoDB 登录
func tryMongoDB(host string, port int, user, pass string, timeout time.Duration) bool {
	// mongodb://user:pass@host:port
	timeoutMs := int64(timeout.Milliseconds())
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/?connectTimeoutMS=%d",
		url.QueryEscape(user), url.QueryEscape(pass), host, port, timeoutMs)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return false
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			// ignore
		}
	}()

	// 检查连接
	err = client.Ping(ctx, nil)
	return err == nil
}

// trySMB 尝试 SMB 登录
func trySMB(host string, port int, user, pass string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     user,
			Password: pass,
			Domain:   "", // workgroup?
		},
	}

	s, err := d.Dial(conn)
	if err != nil {
		return false
	}
	s.Logoff()
	return true
}

// tryTelnet 尝试 Telnet 登录 (简单模拟)
func tryTelnet(host string, port int, user, pass string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()

	// 设置读写超时
	conn.SetDeadline(time.Now().Add(timeout))

	buf := make([]byte, 4096)

	// 读取 Banner / Login 提示
	n, err := conn.Read(buf)
	if err != nil {
		return false
	}
	output := string(buf[:n])

	// 尝试发送用户名
	if strings.Contains(strings.ToLower(output), "login") || strings.Contains(strings.ToLower(output), "user") {
		conn.Write([]byte(user + "\r\n"))
	} else {
		conn.Write([]byte(user + "\r\n"))
	}

	// 等待密码提示
	time.Sleep(500 * time.Millisecond) // 稍作等待
	n, err = conn.Read(buf)
	if err != nil {
		return false
	}
	output = string(buf[:n])

	if strings.Contains(strings.ToLower(output), "pass") {
		conn.Write([]byte(pass + "\r\n"))
	} else {
		conn.Write([]byte(pass + "\r\n"))
	}

	// 等待结果
	time.Sleep(500 * time.Millisecond)
	n, err = conn.Read(buf)
	if err != nil {
		return false
	}
	output = string(buf[:n])

	// 检查失败标识
	lowerOut := strings.ToLower(output)
	if strings.Contains(lowerOut, "incorrect") || strings.Contains(lowerOut, "fail") || strings.Contains(lowerOut, "denied") {
		return false
	}

	// 检查成功标识
	if strings.Contains(output, "$") || strings.Contains(output, "#") || strings.Contains(output, ">") {
		return true
	}

	return false
}

// tryVNC 尝试 VNC 登录
func tryVNC(host string, port int, user, pass string, timeout time.Duration) bool {
	nc, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	defer nc.Close()

	// VNC 认证配置
	cfg := &vnc.ClientConfig{
		Auth: []vnc.ClientAuth{
			&vnc.PasswordAuth{Password: pass},
		},
	}

	c, err := vnc.Client(nc, cfg)
	if err != nil {
		return false
	}
	c.Close()
	return true
}

// tryLDAP 尝试 LDAP 登录
func tryLDAP(host string, port int, user, pass string, timeout time.Duration) bool {
	l, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%d", host, port))
	if err != nil {
		return false
	}
	defer l.Close()

	// Set timeout
	l.SetTimeout(timeout)

	// Simple Bind
	err = l.Bind(user, pass)
	return err == nil
}

// tryWinRM 尝试 WinRM 登录
func tryWinRM(host string, port int, user, pass string, timeout time.Duration) bool {
	endpoint := winrm.NewEndpoint(host, port, false, false, nil, nil, nil, timeout)
	client, err := winrm.NewClient(endpoint, user, pass)
	if err != nil {
		return false
	}
	// Create shell to verify
	_, err = client.CreateShell()
	return err == nil
}

// tryRDP 尝试 RDP 登录
func tryRDP(host string, port int, user, pass string, timeout time.Duration) bool {
	glog.SetLevel(glog.NONE) // Disable logs

	target := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return false
	}
	// Do not defer conn.Close() here as it is wrapped in SocketLayer
	// But if we return early on error, we must close it

	// Use empty domain if not specified (common for local accounts)
	domain := ""

	// Setup RDP protocol stack
	g_tpkt := tpkt.New(core.NewSocketLayer(conn), nla.NewNTLMv2(domain, user, pass))
	g_x224 := x224.New(g_tpkt)
	g_mcs := t125.NewMCSClient(g_x224)
	g_sec := sec.NewClient(g_mcs)
	g_pdu := pdu.NewClient(g_sec)

	// Register channels (required for protocol negotiation)
	g_channels := plugin.NewChannels(g_sec)

	// Set client core data (resolution etc)
	g_mcs.SetClientCoreData(800, 600)

	g_sec.SetUser(user)
	g_sec.SetPwd(pass)
	g_sec.SetDomain(domain)

	// Wire up listeners
	g_tpkt.SetFastPathListener(g_sec)
	g_sec.SetFastPathListener(g_pdu)
	g_sec.SetChannelSender(g_mcs)
	g_channels.SetChannelSender(g_sec)

	// Attempt connection (which includes NLA auth)
	err = g_x224.Connect()

	// Close connection
	conn.Close()

	if err != nil {
		return false
	}

	return true
}
