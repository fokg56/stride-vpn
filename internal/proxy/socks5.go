package proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type AddrType byte

const (
	SocksAddrIPv4   AddrType = 1
	SocksAddrDomain AddrType = 3
	SocksAddrIPv6   AddrType = 4
)

type TargetAddr struct {
	Host string
	Port uint16
}

type HandlerFunc func(conn net.Conn, target TargetAddr) error

type Socks5Server struct {
	addr     string
	handler  HandlerFunc
	listener net.Listener
	running  bool
	mu       sync.Mutex
	logFn    func(msg string)
}

func NewSocks5Server(addr string, handler HandlerFunc) *Socks5Server {
	return &Socks5Server{
		addr:    addr,
		handler: handler,
	}
}

func (s *Socks5Server) SetLogFunc(fn func(msg string)) {
	s.logFn = fn
}

func (s *Socks5Server) log(msg string) {
	if s.logFn != nil {
		s.logFn(msg)
	}
}

func (s *Socks5Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("SOCKS5 сервер уже запущен")
	}

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("ошибка запуска SOCKS5: %w", err)
	}

	s.listener = listener
	s.running = true
	s.log(fmt.Sprintf("SOCKS5 сервер запущен на %s", s.addr))

	go s.acceptLoop()
	return nil
}

func (s *Socks5Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.running = false
	if s.listener != nil {
		s.listener.Close()
	}
	s.log("SOCKS5 сервер остановлен")
	return nil
}

func (s *Socks5Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Socks5Server) acceptLoop() {
	for {
		s.mu.Lock()
		ln := s.listener
		s.mu.Unlock()

		if ln == nil {
			return
		}

		conn, err := ln.Accept()
		if err != nil {
			if !s.IsRunning() {
				return
			}
			continue
		}

		go s.handleConn(conn)
	}
}

func (s *Socks5Server) handleConn(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	target, err := s.socks5Handshake(conn)
	if err != nil {
		s.log(fmt.Sprintf("SOCKS5 ошибка: %v", err))
		return
	}

	conn.SetDeadline(time.Time{})

	if s.handler != nil {
		if err := s.handler(conn, *target); err != nil {
			s.log(fmt.Sprintf("Обработка соединения: %v", err))
		}
	}
}

func (s *Socks5Server) socks5Handshake(conn net.Conn) (*TargetAddr, error) {
	buf := make([]byte, 257)

	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return nil, fmt.Errorf("чтение приветствия: %w", err)
	}

	nMethods := buf[1]
	if nMethods == 0 {
		return nil, fmt.Errorf("нет методов аутентификации")
	}

	if _, err := io.ReadFull(conn, buf[:nMethods]); err != nil {
		return nil, fmt.Errorf("чтение методов: %w", err)
	}

	if _, err := conn.Write([]byte{5, 0}); err != nil {
		return nil, fmt.Errorf("ответ на приветствие: %w", err)
	}

	if _, err := io.ReadFull(conn, buf[:4]); err != nil {
		return nil, fmt.Errorf("чтение запроса: %w", err)
	}

	if buf[0] != 5 {
		return nil, fmt.Errorf("неверная версия SOCKS")
	}

	if buf[1] != 1 {
		conn.Write([]byte{5, 7, 0, 1, 0, 0, 0, 0, 0, 0})
		return nil, fmt.Errorf("поддерживается только TCP connect")
	}

	addrType := AddrType(buf[3])
	var host string

	switch addrType {
	case SocksAddrIPv4:
		if _, err := io.ReadFull(conn, buf[:4]); err != nil {
			return nil, fmt.Errorf("чтение IPv4: %w", err)
		}
		host = net.IP(buf[:4]).String()

	case SocksAddrDomain:
		if _, err := io.ReadFull(conn, buf[:1]); err != nil {
			return nil, fmt.Errorf("чтение длины домена: %w", err)
		}
		domainLen := buf[0]
		if _, err := io.ReadFull(conn, buf[:domainLen]); err != nil {
			return nil, fmt.Errorf("чтение домена: %w", err)
		}
		host = string(buf[:domainLen])

	case SocksAddrIPv6:
		if _, err := io.ReadFull(conn, buf[:16]); err != nil {
			return nil, fmt.Errorf("чтение IPv6: %w", err)
		}
		host = net.IP(buf[:16]).String()

	default:
		conn.Write([]byte{5, 8, 0, 1, 0, 0, 0, 0, 0, 0})
		return nil, fmt.Errorf("неподдерживаемый тип адреса: %d", addrType)
	}

	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return nil, fmt.Errorf("чтение порта: %w", err)
	}
	port := binary.BigEndian.Uint16(buf[:2])

	bnd := []byte{5, 0, 0, 1, 127, 0, 0, 1}
	bnd = binary.BigEndian.AppendUint16(bnd, 0)
	if _, err := conn.Write(bnd); err != nil {
		return nil, fmt.Errorf("отправка ответа: %w", err)
	}

	if strings.HasPrefix(host, "0.0.0.0") {
		host = "127.0.0.1"
	}

	return &TargetAddr{Host: host, Port: port}, nil
}
