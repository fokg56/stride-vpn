package xray

import (
	"encoding/json"
	"fmt"
	"net"

	"vpn-client/internal/vless"
)

type xrayConfig struct {
	Log       logConfig        `json:"log"`
	Inbounds  []interface{}    `json:"inbounds"`
	Outbounds []outboundConfig `json:"outbounds"`
	Routing   *routingConfig   `json:"routing,omitempty"`
	DNS       *dnsConfig       `json:"dns,omitempty"`
}

type logConfig struct {
	Loglevel string `json:"loglevel"`
}

type sniffingConfig struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride"`
	MetadataOnly bool     `json:"metadataOnly"`
}

type inboundSocks struct {
	Port     int                  `json:"port"`
	Listen   string               `json:"listen"`
	Protocol string               `json:"protocol"`
	Settings inboundSocksSettings `json:"settings"`
	Sniffing *sniffingConfig      `json:"sniffing,omitempty"`
}

type inboundSocksSettings struct {
	UDP bool `json:"udp"`
}

type inboundHTTP struct {
	Port     int                 `json:"port"`
	Listen   string              `json:"listen"`
	Protocol string              `json:"protocol"`
	Settings inboundHTTPSettings `json:"settings"`
}

type inboundHTTPSettings struct {
	AllowTransparent bool `json:"allowTransparent"`
}

type inboundTun struct {
	Protocol string             `json:"protocol"`
	Tag      string             `json:"tag"`
	Settings inboundTunSettings `json:"settings"`
}

type inboundTunSettings struct {
	Name      string   `json:"name"`
	MTU       uint32   `json:"mtu"`
	UserLevel uint32   `json:"userLevel"`
	IP        []string `json:"ip,omitempty"`
}

type outboundConfig struct {
	Tag            string          `json:"tag"`
	Protocol       string          `json:"protocol"`
	Settings       json.RawMessage `json:"settings"`
	StreamSettings json.RawMessage `json:"streamSettings,omitempty"`
	Mux            *muxConfig      `json:"mux,omitempty"`
}

type muxConfig struct {
	Enabled     bool `json:"enabled"`
	Concurrency int  `json:"concurrency"`
}

type vlessSettings struct {
	Vnext []vnextEntry `json:"vnext"`
}

type vnextEntry struct {
	Address string      `json:"address"`
	Port    int         `json:"port"`
	Users   []vnextUser `json:"users"`
}

type vnextUser struct {
	ID         string `json:"id"`
	Flow       string `json:"flow,omitempty"`
	Encryption string `json:"encryption"`
}

type freedomSettings struct{}

type streamSettings struct {
	Network         string           `json:"network"`
	Security        string           `json:"security"`
	TLSSettings     *tlsSettings     `json:"tlsSettings,omitempty"`
	RealitySettings *realitySettings `json:"realitySettings,omitempty"`
	TCPSettings     *tcpSettings     `json:"tcpSettings,omitempty"`
	KCSettings      *kcSettings      `json:"kcpSettings,omitempty"`
	WSSettings      *wsSettings      `json:"wsSettings,omitempty"`
	HTTPSettings    *httpSettings    `json:"httpSettings,omitempty"`
	QUICSettings    *quicSettings    `json:"quicSettings,omitempty"`
	GRPCSettings    *grpcSettings    `json:"grpcSettings,omitempty"`
	XHTTPSettings   *xhttpSettings   `json:"xhttpSettings,omitempty"`
}

type tlsSettings struct {
	ServerName    string   `json:"serverName"`
	AllowInsecure bool     `json:"allowInsecure"`
	Fingerprint   string   `json:"fingerprint,omitempty"`
	ALPN          []string `json:"alpn,omitempty"`
}

type realitySettings struct {
	Fingerprint string `json:"fingerprint"`
	ServerName  string `json:"serverName"`
	PublicKey   string `json:"publicKey"`
	ShortID     string `json:"shortId"`
	SpiderX     string `json:"spiderX"`
}

type tcpSettings struct {
	Header *tcpHeaderConfig `json:"header,omitempty"`
}

type tcpHeaderConfig struct {
	Type     string        `json:"type"`
	Request  *httpRequest  `json:"request,omitempty"`
	Response *httpResponse `json:"response,omitempty"`
}

type httpRequest struct {
	Version string      `json:"version"`
	Method  string      `json:"method"`
	Path    []string    `json:"path"`
	Headers httpHeaders `json:"headers"`
}

type httpResponse struct {
	Version string      `json:"version"`
	Status  string      `json:"status"`
	Reason  string      `json:"reason"`
	Headers httpHeaders `json:"headers"`
}

type httpHeaders struct {
	Host           []string `json:"Host,omitempty"`
	UserAgent      []string `json:"User-Agent,omitempty"`
	AcceptEncoding []string `json:"Accept-Encoding,omitempty"`
	Connection     []string `json:"Connection,omitempty"`
	Pragma         []string `json:"Pragma,omitempty"`
}

type kcSettings struct {
	MTU              uint32           `json:"mtu"`
	TTI              uint32           `json:"tti"`
	UplinkCapacity   int              `json:"uplinkCapacity"`
	DownlinkCapacity int              `json:"downlinkCapacity"`
	Congestion       bool             `json:"congestion"`
	ReadBufferSize   int              `json:"readBufferSize"`
	WriteBufferSize  int              `json:"writeBufferSize"`
	Header           *kcpHeaderConfig `json:"header,omitempty"`
	Seed             string           `json:"seed,omitempty"`
}

type kcpHeaderConfig struct {
	Type string `json:"type"`
}

type wsSettings struct {
	Path    string            `json:"path"`
	Headers *wsHeaderSettings `json:"headers,omitempty"`
}

type wsHeaderSettings struct {
	Host string `json:"Host,omitempty"`
}

type httpSettings struct {
	Host []string `json:"host,omitempty"`
	Path string   `json:"path,omitempty"`
}

type quicSettings struct {
	Security string           `json:"security"`
	Key      string           `json:"key"`
	Header   *kcpHeaderConfig `json:"header,omitempty"`
}

type grpcSettings struct {
	ServiceName string `json:"serviceName"`
	MultiMode   bool   `json:"multiMode"`
}

type xhttpSettings struct {
	Mode             string            `json:"mode"`
	Path             string            `json:"path"`
	Host             string            `json:"host,omitempty"`
	Headers          *wsHeaderSettings `json:"headers,omitempty"`
	UplinkHTTPMethod string            `json:"uplinkHTTPMethod,omitempty"`
}

type routingConfig struct {
	DomainStrategy string        `json:"domainStrategy"`
	Rules          []routingRule `json:"rules"`
}

type routingRule struct {
	Type        string   `json:"type"`
	OutboundTag string   `json:"outboundTag"`
	Domain      []string `json:"domain,omitempty"`
	IP          []string `json:"ip,omitempty"`
	Source      []string `json:"source,omitempty"`
	Network     string   `json:"network,omitempty"`
	Port        string   `json:"port,omitempty"`
}

type dnsConfig struct {
	Servers []string `json:"servers"`
}

type RouteRules struct {
	DomainStrategy string
	DirectDomains  []string
	DirectIPs      []string
	ProxyDomains   []string
	ProxyIPs       []string
	BlockDomains   []string
	BlockIPs       []string
	GlobalProxy    bool
}

const (
	tunName    = "stride-vpn"
	tunIP      = "198.18.0.1"
	tunMask    = "255.255.0.0"
	tunGateway = "198.18.0.1"
	tunCIDR    = "198.18.0.1/16"
)

func BuildConfig(cfg *vless.Config, rules *RouteRules, socksPort int, useTUN bool) (string, error) {
	id := cfg.UUID.String()

	flow := string(cfg.Flow)
	enc := cfg.Encryption
	if enc == "" {
		enc = "none"
	}

	sni := cfg.SNI
	if sni == "" {
		sni = cfg.Host
	}

	sec := string(cfg.Security)
	if sec == "" {
		sec = "none"
	}

	network := cfg.Type
	if network == "" {
		network = "tcp"
	}

	ss := streamSettings{
		Network:  network,
		Security: sec,
	}

	switch cfg.Security {
	case vless.SecurityTLS:
		fp := cfg.Fingerprint
		if fp == "" {
			fp = "chrome"
		}
		ss.TLSSettings = &tlsSettings{
			ServerName:    sni,
			AllowInsecure: false,
			Fingerprint:   fp,
			ALPN:          []string{"h2", "http/1.1"},
		}
	case vless.SecurityReality:
		spx := cfg.SpiderX
		if spx == "" {
			spx = "/"
		}
		fp := cfg.Fingerprint
		if fp == "" {
			fp = "chrome"
		}
		ss.RealitySettings = &realitySettings{
			Fingerprint: fp,
			ServerName:  sni,
			PublicKey:   cfg.PublicKey,
			ShortID:     cfg.ShortID,
			SpiderX:     spx,
		}
	}

	switch network {
	case "ws":
		wsPath := cfg.Path
		if wsPath == "" {
			wsPath = "/"
		}
		ws := &wsSettings{Path: wsPath}
		if cfg.RequestHost != "" {
			ws.Headers = &wsHeaderSettings{Host: cfg.RequestHost}
		}
		ss.WSSettings = ws

	case "xhttp":
		xhPath := cfg.Path
		if xhPath == "" {
			xhPath = "/"
		}
		xh := &xhttpSettings{
			Mode: "auto",
			Path: xhPath,
		}
		xh.Host = cfg.RequestHost
		if cfg.RequestHost != "" {
			xh.Headers = &wsHeaderSettings{Host: cfg.RequestHost}
		}
		ss.XHTTPSettings = xh

	case "grpc":
		ss.GRPCSettings = &grpcSettings{
			ServiceName: cfg.ServiceName,
			MultiMode:   false,
		}

	case "kcp":
		kc := &kcSettings{
			MTU:              1350,
			TTI:              20,
			UplinkCapacity:   5,
			DownlinkCapacity: 20,
			Congestion:       false,
			ReadBufferSize:   1,
			WriteBufferSize:  1,
		}
		if cfg.HeaderType != "" {
			kc.Header = &kcpHeaderConfig{Type: cfg.HeaderType}
		}
		ss.KCSettings = kc

	case "http", "h2":
		h2 := &httpSettings{Path: "/"}
		if cfg.RequestHost != "" {
			h2.Host = []string{cfg.RequestHost}
		}
		if cfg.Path != "" {
			h2.Path = cfg.Path
		}
		ss.HTTPSettings = h2

	case "quic":
		qc := &quicSettings{Security: "none", Key: ""}
		if cfg.HeaderType != "" {
			qc.Header = &kcpHeaderConfig{Type: cfg.HeaderType}
		}
		ss.QUICSettings = qc

	default: // tcp
		if cfg.HeaderType == "http" {
			headers := httpHeaders{
				UserAgent:      []string{"Mozilla/5.0"},
				AcceptEncoding: []string{"gzip, deflate"},
				Connection:     []string{"keep-alive"},
			}
			if cfg.RequestHost != "" {
				headers.Host = []string{cfg.RequestHost}
			}
			ss.TCPSettings = &tcpSettings{
				Header: &tcpHeaderConfig{
					Type: "http",
					Request: &httpRequest{
						Version: "1.1",
						Method:  "GET",
						Path:    []string{"/"},
						Headers: headers,
					},
					Response: &httpResponse{
						Version: "1.1",
						Status:  "200",
						Reason:  "OK",
						Headers: headers,
					},
				},
			}
		}
	}

	ssRaw, _ := json.Marshal(ss)

	portInt := int(cfg.Port)

	enableMux := network != "grpc" && cfg.Security != vless.SecurityReality && cfg.Flow == vless.FlowNone
	var mux *muxConfig
	if enableMux {
		mux = &muxConfig{Enabled: true, Concurrency: 8}
	}

	vlessOut := outboundConfig{
		Tag:      "proxy-out",
		Protocol: "vless",
		Settings: mustJSON(vlessSettings{
			Vnext: []vnextEntry{
				{
					Address: cfg.Host,
					Port:    portInt,
					Users: []vnextUser{
						{
							ID:         id,
							Flow:       flow,
							Encryption: enc,
						},
					},
				},
			},
		}),
		StreamSettings: ssRaw,
		Mux:            mux,
	}

	directOut := outboundConfig{
		Tag:      "direct-out",
		Protocol: "freedom",
		Settings: mustJSON(freedomSettings{}),
	}

	blockOut := outboundConfig{
		Tag:      "block-out",
		Protocol: "blackhole",
		Settings: mustJSON(struct{}{}),
	}

	// block rules must be at the VERY TOP — copied from Nekoray sing-box model
	rulesList := []routingRule{
		{
			// block UDP broadcast storm ports (NetBIOS 137, LLMNR 5355)
			Type: "field", OutboundTag: "block-out",
			Network: "udp",
			Port:    "137,5355",
		},
		{
			// block common STUN/TURN ports to reduce WebRTC local-IP leaks
			Type: "field", OutboundTag: "block-out",
			Network: "udp",
			Port:    "3478-3481,5349",
		},
		{
			// block all IPv4 multicast + subnet broadcast (224.0.0.0/3 covers 224.0.0.0–255.255.255.255)
			Type: "field", OutboundTag: "block-out",
			IP: []string{"224.0.0.0/3", "255.255.255.255", "198.18.255.255"},
		},
		{
			// block all IPv6 multicast
			Type: "field", OutboundTag: "block-out",
			IP: []string{"ff00::/8"},
		},
	}

	if useTUN {
		// also block source multicast to prevent reflected loop storms
		rulesList = append(rulesList, routingRule{
			Type: "field", OutboundTag: "block-out",
			Source: []string{"224.0.0.0/3", "ff00::/8"},
		})
	}

	rulesList = append(rulesList, routingRule{
		Type: "field", OutboundTag: "direct-out",
		Domain: []string{cfg.Host},
	})
	if ip := net.ParseIP(cfg.Host); ip != nil {
		rulesList = append(rulesList, routingRule{
			Type: "field", OutboundTag: "direct-out",
			IP: []string{cfg.Host},
		})
	}

	hasProxy := false
	hasDirect := false

	if rules != nil {
		if rules.DomainStrategy == "" {
			rules.DomainStrategy = "IPIfNonMatch"
		}

		for _, d := range rules.BlockDomains {
			rulesList = append(rulesList, routingRule{Type: "field", OutboundTag: "block-out", Domain: []string{d}})
		}
		for _, ip := range rules.BlockIPs {
			rulesList = append(rulesList, routingRule{Type: "field", OutboundTag: "block-out", IP: []string{ip}})
		}

		for _, d := range rules.DirectDomains {
			rulesList = append(rulesList, routingRule{Type: "field", OutboundTag: "direct-out", Domain: []string{d}})
			hasDirect = true
		}
		for _, ip := range rules.DirectIPs {
			rulesList = append(rulesList, routingRule{Type: "field", OutboundTag: "direct-out", IP: []string{ip}})
			hasDirect = true
		}

		for _, d := range rules.ProxyDomains {
			rulesList = append(rulesList, routingRule{Type: "field", OutboundTag: "proxy-out", Domain: []string{d}})
			hasProxy = true
		}
		for _, ip := range rules.ProxyIPs {
			rulesList = append(rulesList, routingRule{Type: "field", OutboundTag: "proxy-out", IP: []string{ip}})
			hasProxy = true
		}
	}

	rulesList = append(rulesList, routingRule{
		Type: "field", OutboundTag: "direct-out",
		IP: []string{
			"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10",
			"127.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12",
			"192.0.0.0/24", "192.0.2.0/24", "192.168.0.0/16",
			"198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24",
			"::1/128", "fc00::/7", "fe80::/10",
		},
	})

	if rules != nil && rules.GlobalProxy {
		rulesList = append(rulesList, routingRule{
			Type:        "field",
			OutboundTag: "proxy-out",
			Port:        "53",
		})
		rulesList = append(rulesList, routingRule{
			Type: "field", OutboundTag: "proxy-out",
			Domain: []string{"regexp:.*"},
		})
		rulesList = append(rulesList, routingRule{
			Type: "field", OutboundTag: "proxy-out",
			IP: []string{"0.0.0.0/0", "::/0"},
		})
	} else if hasProxy && !hasDirect {
		rulesList = append(rulesList, routingRule{
			Type: "field", OutboundTag: "direct-out",
			Domain: []string{"regexp:.*"},
		})
	}

	outbounds := []outboundConfig{vlessOut, directOut, blockOut}

	defaultTag := "proxy-out"
	if rules != nil && !rules.GlobalProxy {
		defaultTag = "direct-out"
	}
	if defaultTag == "direct-out" && !hasDirect {
		outbounds = []outboundConfig{directOut, vlessOut, blockOut}
	}

	ds := "IPIfNonMatch"
	if rules != nil && rules.DomainStrategy != "" {
		ds = normalizeDomainStrategy(rules.DomainStrategy)
	}

	sniff := &sniffingConfig{
		Enabled:      true,
		DestOverride: []string{"http", "tls", "quic"},
		MetadataOnly: false,
	}

	httpPort := socksPort + 1000
	inbounds := []interface{}{
		inboundSocks{
			Port: socksPort, Listen: "127.0.0.1",
			Protocol: "socks",
			Settings: inboundSocksSettings{UDP: true},
			Sniffing: sniff,
		},
		inboundHTTP{
			Port: httpPort, Listen: "127.0.0.1",
			Protocol: "http",
			Settings: inboundHTTPSettings{AllowTransparent: true},
		},
	}

	if useTUN {
		inbounds = append(inbounds, inboundTun{
			Protocol: "tun",
			Tag:      "tun-in",
			Settings: inboundTunSettings{
				Name:      tunName,
				MTU:       1500,
				UserLevel: 0,
				IP:        []string{tunCIDR},
			},
		})
	}

	xc := xrayConfig{
		Log:       logConfig{Loglevel: "warning"},
		Inbounds:  inbounds,
		Outbounds: outbounds,
		Routing: &routingConfig{
			DomainStrategy: ds,
			Rules:          rulesList,
		},
		DNS: &dnsConfig{
			Servers: []string{"https://1.1.1.1/dns-query", "https://8.8.8.8/dns-query", "1.1.1.1", "8.8.8.8"},
		},
	}

	data, err := json.Marshal(xc)
	if err != nil {
		return "", fmt.Errorf("marshal config: %w", err)
	}

	return string(data), nil
}

func mustJSON(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

func normalizeDomainStrategy(strategy string) string {
	switch strategy {
	case "asis", "AsIs":
		return "AsIs"
	case "ipifnonmatch", "IPIfNonMatch":
		return "IPIfNonMatch"
	case "ipondemand", "IPOnDemand":
		return "IPOnDemand"
	default:
		return strategy
	}
}
