package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	utls "github.com/refraction-networking/utls"
	"vpn-client/internal/vless"
)

type HandshakeError struct {
	ServerAddr string
	WSPath     string
	HTTPStatus int
	Body       string
	UseTLS     bool
}

func (e *HandshakeError) Error() string {
	scheme := "ws"
	if e.UseTLS {
		scheme = "wss"
	}
	switch e.HTTPStatus {
	case 404:
		return fmt.Sprintf("Сбой рукопожатия: сервер не знает этот путь.\n  %s://%s%s → HTTP 404", scheme, e.ServerAddr, e.WSPath)
	case 502, 503:
		return fmt.Sprintf("Сервер недоступен (HTTP %d). Проверьте Xray на сервере.", e.HTTPStatus)
	case 200:
		return fmt.Sprintf("%s://%s%s → HTTP 200 (не проксируется Nginx). Сервер не настроен на этот путь.", scheme, e.ServerAddr, e.WSPath)
	default:
		body := e.Body
		if len(body) > 120 {
			body = body[:120] + "..."
		}
		return fmt.Sprintf("WebSocket handshake failed: HTTP %d (%s), expected 101", e.HTTPStatus, body)
	}
}

func CheckWebSocketHandshake(host string, port int, wsPath string, sni string, useTLS bool, timeout time.Duration) error {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	var conn net.Conn
	var err error

	if useTLS {
		tlsCfg := &tls.Config{
			ServerName:         sni,
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		}
		dialer := &net.Dialer{Timeout: timeout}
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsCfg)
		if err != nil {
			return fmt.Errorf("TLS connection failed: %w", err)
		}
	} else {
		dialer := &net.Dialer{Timeout: timeout}
		conn, err = dialer.Dial("tcp", addr)
		if err != nil {
			return fmt.Errorf("TCP connection failed: %w", err)
		}
	}
	defer conn.Close()

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s%s", host, wsPath), nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "upgrade")
	req.Header.Set("Host", sni)
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	conn.SetWriteDeadline(time.Now().Add(timeout))
	if err := req.Write(conn); err != nil {
		return fmt.Errorf("write request: %w", err)
	}

	conn.SetReadDeadline(time.Now().Add(timeout))
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := strings.TrimSpace(string(body))

	if resp.StatusCode == http.StatusSwitchingProtocols {
		return nil
	}

	return &HandshakeError{
		ServerAddr: host,
		WSPath:     wsPath,
		HTTPStatus: resp.StatusCode,
		Body:       bodyStr,
		UseTLS:     useTLS,
	}
}

func CheckRealityHandshake(cfg *vless.Config, timeout time.Duration) error {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))

	dialer := &net.Dialer{Timeout: timeout}
	tcpConn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("TCP connection failed: %w", err)
	}
	defer tcpConn.Close()

	helloID := utls.HelloChrome_Auto
	fp := strings.ToLower(cfg.Fingerprint)
	switch fp {
	case "firefox":
		helloID = utls.HelloFirefox_Auto
	case "safari", "ios":
		helloID = utls.HelloIOS_Auto
	case "edge", "edge_auto":
		helloID = utls.HelloEdge_Auto
	case "chrome":
		helloID = utls.HelloChrome_Auto
	}

	uConn := utls.UClient(tcpConn, &utls.Config{
		ServerName:         cfg.SNI,
		InsecureSkipVerify: true,
	}, helloID)
	if err := uConn.Handshake(); err != nil {
		return fmt.Errorf("REALITY handshake: %w", err)
	}
	uConn.Close()
	return nil
}

func ValidateServerConfig(cfg *vless.Config) error {
	if cfg.UUID == uuid.Nil {
		return fmt.Errorf("invalid UUID: all zeros")
	}
	if cfg.Host == "" {
		return fmt.Errorf("server host is empty")
	}
	if cfg.Port == 0 {
		return fmt.Errorf("server port is 0")
	}
	if cfg.Security == vless.SecurityReality && cfg.PublicKey == "" {
		return fmt.Errorf("REALITY requires public key")
	}
	if (cfg.Type == "ws" || cfg.Type == "xhttp") && cfg.Path == "" {
		cfg.Path = "/"
	}
	if cfg.Type == "grpc" && cfg.ServiceName == "" {
		cfg.ServiceName = "."
	}
	return nil
}

func UseTLS(cfg *vless.Config) bool {
	return cfg.Security == vless.SecurityTLS || cfg.Security == vless.SecurityReality
}
