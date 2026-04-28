// CDP TLS Proxy — Chrome DevTools Protocol 认证代理（TLS + 路径 Token）
//
// 功能：
//   - TLS 加密（自动生成自签名证书或使用指定证书）
//   - Token 嵌入 URL 路径前缀 /<token>/... 作为认证入口
//   - 直接兼容 Playwright connectOverCDP('https://host:port/<token>/')
//   - HTTP 反向代理 + WebSocket 劫持代理
//   - 自动重写 /json/* 中的 webSocketDebuggerUrl
//   - 未设置 CDP_TOKEN 时所有请求放行（向后兼容）
//
// 环境变量：
//   CDP_TOKEN      — Token，嵌入 URL 路径。空则无需 Token
//   CDP_TLS_CERT   — TLS 证书路径（默认 /config/ssl/cdp-cert.pem）
//   CDP_TLS_KEY    — TLS 密钥路径（默认 /config/ssl/cdp-key.pem）

package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	listenAddr   = "0.0.0.0:9222"
	upstreamAddr = "127.0.0.1:19222"
)

var (
	token        string
	certFile     string
	keyFile      string
	disableHTTPS bool
)

type ctxKey string

const ctxHostKey ctxKey = "origHost"

// ───────────────────── TLS 证书 ─────────────────────

func ensureTLSCerts() error {
	if _, err := os.Stat(certFile); err == nil {
		log.Println("📄 Using existing TLS cert:", certFile)
		return nil
	}

	log.Println("🔐 Generating self-signed TLS certificate...")

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "CDP Proxy"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("create cert: %w", err)
	}

	os.MkdirAll(filepath.Dir(certFile), 0755)

	cf, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("write cert: %w", err)
	}
	pem.Encode(cf, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	cf.Close()

	kb, _ := x509.MarshalECPrivateKey(key)
	kf, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("write key: %w", err)
	}
	pem.Encode(kf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: kb})
	kf.Close()

	log.Printf("✅ TLS cert generated: %s", certFile)
	return nil
}

// ───────────────────── 工具函数 ─────────────────────

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func isWSUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

// stripPrefix 从路径中移除 /<token> 前缀
func stripPrefix(path string) string {
	if token == "" {
		return path
	}
	p := strings.TrimPrefix(path, "/"+token)
	if p == "" || p[0] != '/' {
		return "/" + p
	}
	return p
}

// ───────────────────── WebSocket 代理 ─────────────────────

func proxyWebSocket(w http.ResponseWriter, r *http.Request) {
	targetPath := stripPrefix(r.URL.Path)
	if r.URL.RawQuery != "" {
		targetPath += "?" + r.URL.RawQuery
	}

	// 连接上游 Chrome
	upConn, err := net.DialTimeout("tcp", upstreamAddr, 5*time.Second)
	if err != nil {
		http.Error(w, "502 Chrome CDP not ready", http.StatusBadGateway)
		return
	}
	defer upConn.Close()

	// 劫持客户端连接
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hj.Hijack()
	if err != nil {
		return
	}
	defer clientConn.Close()

	// 构建上游请求（剥离 token 前缀，替换 Host）
	var buf strings.Builder
	fmt.Fprintf(&buf, "%s %s HTTP/1.1\r\n", r.Method, targetPath)
	fmt.Fprintf(&buf, "Host: %s\r\n", upstreamAddr)
	for k, vs := range r.Header {
		if strings.EqualFold(k, "Host") {
			continue
		}
		for _, v := range vs {
			fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
		}
	}
	buf.WriteString("\r\n")
	upConn.Write([]byte(buf.String()))

	// 双向转发
	done := make(chan struct{})
	go func() {
		io.Copy(upConn, clientConn)
		if tc, ok := upConn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
		close(done)
	}()
	io.Copy(clientConn, upConn)
	<-done
}

// ───────────────────── JSON 响应重写 ─────────────────────

// rewriteResponse 将 Chrome 返回的 ws://127.0.0.1:19222 替换为 wss://host/<token>
func rewriteResponse(resp *http.Response) error {
	// 只处理 /json 开头的路径
	if !strings.HasPrefix(resp.Request.URL.Path, "/json") {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	host, _ := resp.Request.Context().Value(ctxHostKey).(string)
	if host == "" {
		host = "localhost:9222"
	}

	// 构建替换目标
	scheme := "wss"
	if disableHTTPS {
		scheme = "ws"
	}
	var newBase string
	if token != "" {
		newBase = fmt.Sprintf("%s://%s/%s", scheme, host, token)
	} else {
		newBase = fmt.Sprintf("%s://%s", scheme, host)
	}

	s := string(body)
	s = strings.ReplaceAll(s, "ws://127.0.0.1:19222", newBase)
	s = strings.ReplaceAll(s, "ws://localhost:19222", newBase)

	resp.Body = io.NopCloser(strings.NewReader(s))
	resp.ContentLength = int64(len(s))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(s)))
	return nil
}

// ───────────────────── 主入口 ─────────────────────

func main() {
	disableHTTPS = os.Getenv("CDP_DISABLE_HTTPS") == "true"
	token = os.Getenv("CDP_TOKEN")
	certFile = envOr("CDP_TLS_CERT", "/config/ssl/cdp-cert.pem")
	keyFile = envOr("CDP_TLS_KEY", "/config/ssl/cdp-key.pem")

	if !disableHTTPS {
		if err := ensureTLSCerts(); err != nil {
			log.Fatalf("❌ TLS setup failed: %v", err)
		}
	}

	// 反向代理
	upstream, _ := url.Parse("http://" + upstreamAddr)
	proxy := httputil.NewSingleHostReverseProxy(upstream)

	defaultDir := proxy.Director
	proxy.Director = func(r *http.Request) {
		defaultDir(r)
		r.URL.Path = stripPrefix(r.URL.Path)
		r.URL.RawPath = ""
		r.Host = upstreamAddr // 确保 Chrome 返回 ws://127.0.0.1:19222
	}
	proxy.ModifyResponse = rewriteResponse

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Token 认证：路径必须以 /<token>/ 开头
		if token != "" {
			pfx := "/" + token
			if !strings.HasPrefix(r.URL.Path, pfx+"/") && r.URL.Path != pfx {
				http.NotFound(w, r)
				return
			}
		}

		// 保存原始 Host 用于 URL 重写
		ctx := context.WithValue(r.Context(), ctxHostKey, r.Host)
		r = r.WithContext(ctx)

		if isWSUpgrade(r) {
			proxyWebSocket(w, r)
			return
		}

		proxy.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:    listenAddr,
		Handler: handler,
	}

	if token != "" {
		if disableHTTPS {
			log.Printf("🚀 CDP Proxy: http://*:9222/%s/ → %s", token, upstreamAddr)
		} else {
			log.Printf("🚀 CDP TLS Proxy: https://*:9222/%s/ → %s", token, upstreamAddr)
		}
	} else {
		if disableHTTPS {
			log.Printf("🚀 CDP Proxy: http://*:9222/ → %s (no token)", upstreamAddr)
		} else {
			log.Printf("🚀 CDP TLS Proxy: https://*:9222/ → %s (no token)", upstreamAddr)
		}
	}

	if disableHTTPS {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
			log.Fatal(err)
		}
	}
}
