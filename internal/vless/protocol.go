package vless

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	utls "github.com/refraction-networking/utls"
	xnet "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/protocol"
	xuuid "github.com/xtls/xray-core/common/uuid"
	"github.com/xtls/xray-core/proxy/vless"
	vlessenc "github.com/xtls/xray-core/proxy/vless/encoding"
)

type CommandType byte

const (
	CmdTCP CommandType = 1
	CmdUDP CommandType = 2
)

type LogFunc func(format string, args ...interface{})

type Connection struct {
	conn      net.Conn
	config    *Config
	createdAt time.Time
	closed    bool
	logFn     LogFunc
}

func (c *Connection) SetLogger(l LogFunc) {
	c.logFn = l
}

func (c *Connection) logf(format string, args ...interface{}) {
	if c.logFn != nil {
		c.logFn(format, args...)
	}
}

func Dial(config *Config) (*Connection, error) {
	addr := net.JoinHostPort(config.Host, strconv.Itoa(int(config.Port)))
	dialer := &net.Dialer{Timeout: 10 * time.Second}

	var rawConn net.Conn
	var err error

	switch config.Security {
	case SecurityTLS:
		tlsCfg := &tls.Config{
			ServerName:         config.SNI,
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		}
		rawConn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsCfg)
		if err != nil {
			return nil, fmt.Errorf("ошибка TLS: %w", err)
		}

	case SecurityReality:
		tcpConn, err := dialer.Dial("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("ошибка TCP: %w", err)
		}
		helloID := getUtlsHelloID(config.Fingerprint)
		uConn := utls.UClient(tcpConn, &utls.Config{
			ServerName:         config.SNI,
			InsecureSkipVerify: true,
		}, helloID)
		if err := uConn.Handshake(); err != nil {
			tcpConn.Close()
			return nil, fmt.Errorf("REALITY хендшейк: %w", err)
		}
		rawConn = uConn

	case SecurityNone:
		rawConn, err = dialer.Dial("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("ошибка TCP: %w", err)
		}

	default:
		return nil, fmt.Errorf("неизвестный security: %s", config.Security)
	}

	return &Connection{
		conn:      rawConn,
		config:    config,
		createdAt: time.Now(),
	}, nil
}

func getUtlsHelloID(fp string) utls.ClientHelloID {
	switch strings.ToLower(fp) {
	case "firefox":
		return utls.HelloFirefox_Auto
	case "safari", "ios":
		return utls.HelloIOS_Auto
	case "edge", "edge_auto":
		return utls.HelloEdge_Auto
	case "chrome":
		return utls.HelloChrome_Auto
	default:
		return utls.HelloRandomized
	}
}

func (c *Connection) Handshake(targetAddr string, targetPort uint16, cmd CommandType) error {
	var uid [16]byte
	copy(uid[:], c.config.UUID[:])

	account := &vless.MemoryAccount{
		ID:   protocol.NewID(xuuid.UUID(uid)),
		Flow: string(c.config.Flow),
	}

	request := &protocol.RequestHeader{
		Version: vlessenc.Version,
		Command: protocol.RequestCommandTCP,
		Address: xnet.ParseAddress(targetAddr),
		Port:    xnet.Port(targetPort),
		User: &protocol.MemoryUser{
			Account: account,
		},
	}

	var headerBuf bytes.Buffer
	addons := &vlessenc.Addons{
		Flow: string(c.config.Flow),
	}
	if err := vlessenc.EncodeRequestHeader(&headerBuf, request, addons); err != nil {
		return fmt.Errorf("ошибка отправки VLESS заголовка: %w", err)
	}
	data := headerBuf.Bytes()
	if _, err := c.conn.Write(data); err != nil {
		return fmt.Errorf("ошибка записи заголовка: %w", err)
	}
	c.logf("[VLESS] отправлен заголовок (%d байт): %x", len(data), data)

	c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	respAddons, err := vlessenc.DecodeResponseHeader(c.conn, request)
	c.conn.SetReadDeadline(time.Time{})
	if err != nil {
		return fmt.Errorf("сервер отклонил соединение: %w", err)
	}
	c.logf("[VLESS] получен ответ от сервера, addons flow: %q", respAddons.Flow)

	return nil
}

func (c *Connection) Read(p []byte) (int, error) {
	if c.closed {
		return 0, fmt.Errorf("соединение закрыто")
	}
	return c.conn.Read(p)
}

func (c *Connection) Write(p []byte) (int, error) {
	if c.closed {
		return 0, fmt.Errorf("соединение закрыто")
	}
	return c.conn.Write(p)
}

func (c *Connection) Close() error {
	c.closed = true
	return c.conn.Close()
}

func (c *Connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *Connection) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *Connection) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
