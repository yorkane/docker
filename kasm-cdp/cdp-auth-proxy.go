// CDP Auth Proxy — 高性能 Chrome DevTools Protocol 认证代理
//
// 功能：
//   - 私有网段 (127/8, 10/8, 172.16/12, 192.168/16) 免认证直通
//   - 外部 IP 需 URL 参数 ?token=xxx 或 Authorization: Bearer xxx
//   - 未设置 CDP_TOKEN 时所有请求放行（向后兼容）
//   - 基于 Go net.TCPConn 的 splice 零拷贝转发

package main

import (
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	listenAddr   = "0.0.0.0:9222"
	upstreamAddr = "127.0.0.1:19222"
	readTimeout  = 10 * time.Second
	dialTimeout  = 5 * time.Second
)

var (
	token       string
	privateNets []*net.IPNet
)

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8", "10.0.0.0/8",
		"172.16.0.0/12", "192.168.0.0/16",
	} {
		_, n, _ := net.ParseCIDR(cidr)
		privateNets = append(privateNets, n)
	}
}

func isPrivate(ip net.IP) bool {
	for _, n := range privateNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// checkAuth 从 HTTP 请求头中校验 token
func checkAuth(header []byte) bool {
	s := string(header)
	lines := strings.SplitN(s, "\r\n", -1)
	if len(lines) == 0 {
		return false
	}

	// 解析请求行: GET /path?token=xxx HTTP/1.1
	parts := strings.SplitN(lines[0], " ", 3)
	if len(parts) >= 2 {
		if u, err := url.Parse(parts[1]); err == nil {
			if u.Query().Get("token") == token {
				return true
			}
		}
	}

	// 检查 Authorization: Bearer xxx
	for _, line := range lines[1:] {
		if len(line) > 15 && strings.EqualFold(line[:14], "authorization:") {
			val := strings.TrimSpace(line[14:])
			if len(val) > 7 && strings.EqualFold(val[:7], "bearer ") {
				if strings.TrimSpace(val[7:]) == token {
					return true
				}
			}
		}
	}
	return false
}

func handleConn(client net.Conn) {
	defer client.Close()

	host, _, _ := net.SplitHostPort(client.RemoteAddr().String())
	clientIP := net.ParseIP(host)

	// 读取初始 HTTP 请求头，设置超时
	client.SetReadDeadline(time.Now().Add(readTimeout))

	buf := make([]byte, 4096)
	n, err := client.Read(buf)
	if err != nil || n == 0 {
		return
	}

	// 清除后续数据传输的超时限制
	client.SetReadDeadline(time.Time{})

	// Token 认证：仅对非私有网段来源生效
	if token != "" && (clientIP == nil || !isPrivate(clientIP)) {
		if !checkAuth(buf[:n]) {
			client.Write([]byte(
				"HTTP/1.1 401 Unauthorized\r\n" +
					"Content-Type: text/plain\r\n" +
					"WWW-Authenticate: Bearer\r\n" +
					"Connection: close\r\n\r\n" +
					"401 Unauthorized: Invalid or missing CDP token\n"))
			log.Printf("REJECTED %s", host)
			return
		}
	}

	// 连接上游 Chrome CDP
	up, err := net.DialTimeout("tcp", upstreamAddr, dialTimeout)
	if err != nil {
		client.Write([]byte(
			"HTTP/1.1 502 Bad Gateway\r\n" +
				"Content-Type: text/plain\r\n" +
				"Connection: close\r\n\r\n" +
				"502 Chrome CDP not ready\n"))
		return
	}
	defer up.Close()

	// 转发初始请求数据
	up.Write(buf[:n])

	// 双向 splice 零拷贝代理
	done := make(chan struct{})
	go func() {
		io.Copy(up, client)
		if tc, ok := up.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
		close(done)
	}()
	io.Copy(client, up)
	<-done
}

func main() {
	token = os.Getenv("CDP_TOKEN")

	if token == "" {
		log.Println("CDP_TOKEN not set, all access allowed (no auth)")
	} else {
		log.Println("CDP Auth Proxy: token auth enabled, private nets bypass")
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening %s -> %s", listenAddr, upstreamAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConn(conn)
	}
}
