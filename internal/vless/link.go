package vless

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type Security string

const (
	SecurityNone    Security = "none"
	SecurityTLS     Security = "tls"
	SecurityReality Security = "reality"
)

type Flow string

const (
	FlowNone          Flow = ""
	FlowXTLSPRXVision Flow = "xtls-rprx-vision"
)

type AddrType byte

const (
	AddrIPv4   AddrType = 1
	AddrDomain AddrType = 2
	AddrIPv6   AddrType = 3
)

type Config struct {
	UUID        uuid.UUID
	Host        string   // server address
	Port        uint16
	Security    Security
	Encryption  string
	Flow        Flow
	SNI         string
	Fingerprint string
	PublicKey   string
	ShortID     string
	SpiderX     string
	Type        string   // transport: tcp, ws, grpc, kcp, quic, http
	HeaderType  string   // for tcp+http header obfuscation
	RequestHost string   // Host header for WS/HTTP transport
	Path        string   // path for WS/HTTP/gRPC transport
	ServiceName string   // serviceName for gRPC
	Remark      string
}

func ParseVlessLink(link string) (*Config, error) {
	if !strings.HasPrefix(link, "vless://") {
		return nil, fmt.Errorf("неверный формат: ожидается vless://")
	}

	trimmed := strings.TrimPrefix(link, "vless://")

	var rawURL string
	var remark string
	if idx := strings.LastIndex(trimmed, "#"); idx != -1 {
		remark = trimmed[idx+1:]
		rawURL = trimmed[:idx]
	} else {
		rawURL = trimmed
	}

	remark, _ = url.QueryUnescape(remark)

	atIdx := strings.Index(rawURL, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("неверный формат: отсутствует @")
	}

	uuidStr := rawURL[:atIdx]
	hostPortAndQuery := rawURL[atIdx+1:]

	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, fmt.Errorf("некорректный UUID: %w", err)
	}

	var host string
	var port uint16
	var queryPart string

	qIdx := strings.Index(hostPortAndQuery, "?")
	if qIdx != -1 {
		queryPart = hostPortAndQuery[qIdx+1:]
		hostPort := hostPortAndQuery[:qIdx]
		host, port, err = parseHostPort(hostPort)
		if err != nil {
			return nil, err
		}
	} else {
		host, port, err = parseHostPort(hostPortAndQuery)
		if err != nil {
			return nil, err
		}
	}

	cfg := &Config{
		UUID:        parsedUUID,
		Host:        host,
		Port:        port,
		Security:    SecurityNone,
		Encryption:  "none",
		Flow:        FlowNone,
		SNI:         host,
		Remark:      remark,
		Fingerprint: "chrome",
	}

	if queryPart != "" {
		values, err := url.ParseQuery(queryPart)
		if err != nil {
			return nil, fmt.Errorf("ошибка парсинга query: %w", err)
		}

		if v := values.Get("security"); v != "" {
			cfg.Security = Security(v)
		}
		if v := values.Get("encryption"); v != "" {
			cfg.Encryption = v
		}
		if v := values.Get("flow"); v != "" {
			cfg.Flow = Flow(v)
		}
		if v := values.Get("sni"); v != "" {
			cfg.SNI = v
		}
		if v := values.Get("fp"); v != "" {
			cfg.Fingerprint = v
		}
		if v := values.Get("pbk"); v != "" {
			cfg.PublicKey = v
		}
		if v := values.Get("sid"); v != "" {
			cfg.ShortID = v
		}
		if v := values.Get("spx"); v != "" {
			cfg.SpiderX = v
		}
		if v := values.Get("type"); v != "" {
			cfg.Type = v
		}
		if v := values.Get("headerType"); v != "" {
			cfg.HeaderType = v
		}
		if v := values.Get("path"); v != "" {
			cfg.Path = v
		}
		if v := values.Get("serviceName"); v != "" {
			cfg.ServiceName = v
		}
		if v := values.Get("host"); v != "" {
			cfg.RequestHost = v
		}
	}

	// default path for ws/xhttp if server didn't provide one
	if (cfg.Type == "ws" || cfg.Type == "xhttp") && cfg.Path == "" {
		cfg.Path = "/"
	}

	return cfg, nil
}

func parseHostPort(s string) (string, uint16, error) {
	if strings.Contains(s, "[") && strings.Contains(s, "]") {
		closeB := strings.LastIndex(s, "]")
		if closeB == -1 {
			return "", 0, fmt.Errorf("неверный IPv6 адрес")
		}
		host := s[1:closeB]
		rest := s[closeB+1:]
		if strings.HasPrefix(rest, ":") {
			portStr := rest[1:]
			p, err := strconv.ParseUint(portStr, 10, 16)
			if err != nil {
				return "", 0, fmt.Errorf("некорректный порт: %w", err)
			}
			return host, uint16(p), nil
		}
		return host, 443, nil
	}

	colonIdx := strings.LastIndex(s, ":")
	if colonIdx == -1 {
		return s, 443, nil
	}

	portStr := s[colonIdx+1:]
	p, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return s, 443, nil
	}
	return s[:colonIdx], uint16(p), nil
}

func (c *Config) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("vless://%s@%s:%d", c.UUID.String(), c.Host, c.Port))

	params := make([]string, 0, 4)
	if c.Encryption != "" && c.Encryption != "none" {
		params = append(params, "encryption="+c.Encryption)
	}
	if c.Security != SecurityNone {
		params = append(params, "security="+string(c.Security))
	}
	if c.Flow != FlowNone {
		params = append(params, "flow="+string(c.Flow))
	}
	if c.SNI != "" && c.SNI != c.Host {
		params = append(params, "sni="+c.SNI)
	}
	if c.Fingerprint != "" {
		params = append(params, "fp="+c.Fingerprint)
	}
	if c.PublicKey != "" {
		params = append(params, "pbk="+c.PublicKey)
	}
	if c.ShortID != "" {
		params = append(params, "sid="+c.ShortID)
	}
	if c.SpiderX != "" {
		params = append(params, "spx="+url.QueryEscape(c.SpiderX))
	}
	if c.Type != "" {
		params = append(params, "type="+c.Type)
	}
	if c.HeaderType != "" {
		params = append(params, "headerType="+c.HeaderType)
	}
	if c.RequestHost != "" {
		params = append(params, "host="+c.RequestHost)
	}
	if c.Path != "" {
		params = append(params, "path="+url.QueryEscape(c.Path))
	}
	if c.ServiceName != "" {
		params = append(params, "serviceName="+c.ServiceName)
	}

	if len(params) > 0 {
		sb.WriteString("?")
		sb.WriteString(strings.Join(params, "&"))
	}

	if c.Remark != "" {
		sb.WriteString("#")
		sb.WriteString(url.QueryEscape(c.Remark))
	}

	return sb.String()
}
