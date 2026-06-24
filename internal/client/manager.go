package client

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"vpn-client/internal/proxy"
	"vpn-client/internal/storage"
	"vpn-client/internal/vless"
	"vpn-client/internal/xray"
)

type ConnectionState string

const (
	StateDisconnected ConnectionState = "disconnected"
	StateConnecting   ConnectionState = "connecting"
	StateConnected    ConnectionState = "connected"
	StateError        ConnectionState = "error"
	StateNoNetwork    ConnectionState = "no_network"
)

type ConfigEntry struct {
	ID     string        `json:"id"`
	Remark string        `json:"remark"`
	Config *vless.Config `json:"-"`
}

type ConfigInfo struct {
	UUID        string `json:"uuid"`
	Host        string `json:"host"`
	Port        uint16 `json:"port"`
	Security    string `json:"security"`
	Encryption  string `json:"encryption"`
	Flow        string `json:"flow"`
	SNI         string `json:"sni"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"publicKey"`
	ShortID     string `json:"shortId"`
	Remark      string `json:"remark"`
	Link        string `json:"link"`
}

type SubscriptionItem struct {
	URL      string    `json:"url"`
	Interval int64     `json:"interval"`
	Updated  time.Time `json:"updated"`
}

type Manager struct {
	mu        sync.RWMutex
	configs   []*ConfigEntry
	activeID  string
	state     ConnectionState
	logChan   chan string
	stateChan chan ConnectionState
	store     *storage.Store

	xray      *xray.Xray
	Routing   *RoutingManager
	Settings  *SettingsManager
	startedAt time.Time

	networkStop   chan struct{}
	subsStop      chan struct{}
	networkUp     atomic.Bool
	subscriptions []SubscriptionItem
	subMu         sync.Mutex
	subActive     map[string]chan struct{}

	reconnectBackoff    time.Duration
	reconnectMaxBackoff time.Duration
	reconnectAttempts   int
}

func NewManager(store *storage.Store, routing *RoutingManager, settings *SettingsManager) *Manager {
	m := &Manager{
		logChan:             make(chan string, 200),
		stateChan:           make(chan ConnectionState, 10),
		state:               StateDisconnected,
		store:               store,
		Routing:             routing,
		subActive:           make(map[string]chan struct{}),
		reconnectMaxBackoff: 30 * time.Second,
	}

	if err := m.loadConfigs(); err != nil {
		m.log(fmt.Sprintf("Error loading configs: %v", err))
	}

	m.StartNetworkDetection()
	m.StartSubscriptionUpdater()
	return m
}

func (m *Manager) loadConfigs() error {
	stored, err := m.store.Load()
	if err != nil {
		return err
	}

	for _, sc := range stored {
		if sc.Link == "" {
			continue
		}
		cfg, err := vless.ParseVlessLink(sc.Link)
		if err != nil {
			m.log(fmt.Sprintf("Error parsing saved config: %v", err))
			continue
		}
		entry := &ConfigEntry{
			ID:     sc.ID,
			Remark: cfg.Remark,
			Config: cfg,
		}
		m.configs = append(m.configs, entry)
	}

	m.log(fmt.Sprintf("Loaded configs: %d", len(m.configs)))
	return nil
}

func (m *Manager) saveConfigs() {
	stored := make([]storage.StoredConfig, 0, len(m.configs))
	for _, entry := range m.configs {
		stored = append(stored, storage.StoredConfig{
			ID:     entry.ID,
			Link:   entry.Config.String(),
			Remark: entry.Remark,
		})
	}
	if err := m.store.Save(stored); err != nil {
		m.log(fmt.Sprintf("Error saving configs: %v", err))
	}
}

func configID(cfg *vless.Config) string {
	return fmt.Sprintf("%s@%s:%d", cfg.UUID.String(), cfg.Host, cfg.Port)
}

func (m *Manager) AddConfig(cfg *vless.Config) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := configID(cfg)
	for _, entry := range m.configs {
		if entry.ID == id {
			entry.Config = cfg
			if cfg.Remark != "" {
				entry.Remark = cfg.Remark
			}
			m.saveConfigs()
			return id
		}
	}

	entry := &ConfigEntry{
		ID:     id,
		Remark: cfg.Remark,
		Config: cfg,
	}
	m.configs = append(m.configs, entry)
	m.saveConfigs()
	return id
}

func (m *Manager) RemoveConfig(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, entry := range m.configs {
		if entry.ID == id {
			m.configs = append(m.configs[:i], m.configs[i+1:]...)
			if m.activeID == id {
				go m.Disconnect()
			}
			m.saveConfigs()
			return true
		}
	}
	return false
}

func (m *Manager) ListConfigs() []*ConfigEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ConfigEntry, len(m.configs))
	copy(result, m.configs)
	return result
}

func (m *Manager) GetConfigInfo(id string) *ConfigInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, entry := range m.configs {
		if entry.ID == id {
			cfg := entry.Config
			return &ConfigInfo{
				UUID:        cfg.UUID.String(),
				Host:        cfg.Host,
				Port:        cfg.Port,
				Security:    string(cfg.Security),
				Encryption:  cfg.Encryption,
				Flow:        string(cfg.Flow),
				SNI:         cfg.SNI,
				Fingerprint: cfg.Fingerprint,
				PublicKey:   cfg.PublicKey,
				ShortID:     cfg.ShortID,
				Remark:      cfg.Remark,
				Link:        cfg.String(),
			}
		}
	}
	return nil
}

func (m *Manager) GetActiveConfig() *ConfigEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeID == "" {
		return nil
	}
	for _, entry := range m.configs {
		if entry.ID == m.activeID {
			return entry
		}
	}
	return nil
}

func (m *Manager) Ping(id string) (int, error) {
	m.mu.RLock()
	var cfg *vless.Config
	for _, entry := range m.configs {
		if entry.ID == id {
			cfg = entry.Config
			break
		}
	}
	m.mu.RUnlock()

	if cfg == nil {
		return 0, fmt.Errorf("config not found")
	}

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))
	start := time.Now()

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	if cfg.Security == vless.SecurityTLS || cfg.Security == vless.SecurityReality {
		tlsConn := tls.Client(conn, &tls.Config{
			ServerName:         cfg.SNI,
			InsecureSkipVerify: true,
		})
		if err := tlsConn.Handshake(); err != nil {
			return 0, fmt.Errorf("TLS handshake failed: %w", err)
		}
	}

	elapsed := time.Since(start)
	return int(elapsed.Milliseconds()), nil
}

func (m *Manager) Connect(id string) error {
	m.waitBackoff()

	m.mu.Lock()
	if m.state == StateConnected || m.state == StateConnecting {
		m.mu.Unlock()
		return fmt.Errorf("already connected")
	}
	m.state = StateConnecting
	m.activeID = id
	m.mu.Unlock()

	m.sendState(StateConnecting)
	m.log("Connecting...")

	var cfg *vless.Config
	m.mu.RLock()
	for _, entry := range m.configs {
		if entry.ID == id {
			cfg = entry.Config
			break
		}
	}
	m.mu.RUnlock()

	if cfg == nil {
		m.mu.Lock()
		m.state = StateDisconnected
		m.activeID = ""
		m.mu.Unlock()
		m.sendState(StateDisconnected)
		return fmt.Errorf("config not found")
	}

	if err := ValidateServerConfig(cfg); err != nil {
		m.mu.Lock()
		m.state = StateDisconnected
		m.activeID = ""
		m.mu.Unlock()
		m.sendState(StateError)
		m.log(fmt.Sprintf("Config invalid: %v", err))
		return fmt.Errorf("config validation: %w", err)
	}

	m.log(fmt.Sprintf("Server: %s:%d [%s]", cfg.Host, cfg.Port, cfg.Security))

	if cfg.Type == "ws" || cfg.Type == "xhttp" {
		useTLS := UseTLS(cfg)
		if useTLS {
			m.log("Checking WebSocket handshake (WSS)...")
		} else {
			m.log("Checking WebSocket handshake (WS)...")
		}
		port := int(cfg.Port)
		if err := CheckWebSocketHandshake(cfg.Host, port, cfg.Path, cfg.SNI, useTLS, 10*time.Second); err != nil {
			m.mu.Lock()
			m.state = StateDisconnected
			m.activeID = ""
			m.mu.Unlock()
			m.sendState(StateError)
			m.log(fmt.Sprintf("Handshake failed: %v", err))
			return err
		}
		m.log("Handshake OK (101)")
	} else if cfg.Security == vless.SecurityReality && cfg.Type != "xhttp" {
		m.log("Checking REALITY handshake...")
		if err := CheckRealityHandshake(cfg, 10*time.Second); err != nil {
			m.mu.Lock()
			m.state = StateDisconnected
			m.activeID = ""
			m.mu.Unlock()
			m.sendState(StateError)
			m.log(fmt.Sprintf("REALITY handshake failed: %v", err))
			return err
		}
		m.log("REALITY handshake OK")
	}

	rules := m.Routing.RouteRules()

	m.log("Validating config...")
	port, _ := xray.FindFreePort()
	if err := xray.ValidateConfig(cfg, rules, port, true); err != nil {
		m.log(fmt.Sprintf("TUN config invalid (%v), trying proxy...", err))
		if err := xray.ValidateConfig(cfg, rules, port, false); err != nil {
			m.mu.Lock()
			m.state = StateDisconnected
			m.activeID = ""
			m.mu.Unlock()
			m.sendState(StateError)
			m.log(fmt.Sprintf("Config validation failed: %v", err))
			return fmt.Errorf("config validation: %w", err)
		}
	}

	isAdmin := xray.IsAdmin()
	settings := &AppSettings{TUNEnabled: true, SystemProxy: true}
	if m.Settings != nil {
		settings = m.Settings.Get()
	}

	var x *xray.Xray
	var err error
	useTUN := false

	tryTUN := settings.TUNEnabled
	if tryTUN && !isAdmin {
		m.log("TUN needs admin, elevating...")
		m.mu.Lock()
		m.state = StateDisconnected
		m.activeID = ""
		m.mu.Unlock()
		m.sendState(StateDisconnected)
		if err := xray.Elevate(id); err != nil {
			return fmt.Errorf("elevate: %w", err)
		}
		os.Exit(0)
		return nil
	}

	x, err = xray.NewWithRouting(cfg, rules, tryTUN)
	if err != nil {
		if tryTUN {
			m.log(fmt.Sprintf("TUN mode not available (%v), falling back to proxy", err))
		}
		x, err = xray.NewWithRouting(cfg, rules, false)
		if err != nil {
			m.mu.Lock()
			m.state = StateDisconnected
			m.activeID = ""
			m.mu.Unlock()
			m.sendState(StateError)
			m.log(fmt.Sprintf("Error creating xray: %v", err))
			return fmt.Errorf("error creating xray: %w", err)
		}
	}

	useTUN = x.IsTUN()
	if err := x.Start(); err != nil {
		if useTUN {
			xray.TeardownTUNAdapter(cfg.Host)
		}
		m.mu.Lock()
		m.state = StateDisconnected
		m.activeID = ""
		m.mu.Unlock()
		m.sendState(StateError)
		m.log(fmt.Sprintf("Error starting xray: %v", err))
		return fmt.Errorf("error starting xray: %w", err)
	}

	m.log("Waiting for core readiness...")
	if err := x.WaitForReady(5 * time.Second); err != nil {
		x.Close()
		if useTUN {
			xray.TeardownTUNAdapter(cfg.Host)
		}
		m.mu.Lock()
		m.state = StateDisconnected
		m.activeID = ""
		m.mu.Unlock()
		m.sendState(StateError)
		m.log(fmt.Sprintf("Core not ready: %v", err))
		return fmt.Errorf("core not ready: %w", err)
	}
	m.log("Core ready")

	if useTUN && isAdmin {
		if err := xray.SetupTUNAdapter(cfg.Host); err != nil {
			x.Close()
			xray.TeardownTUNAdapter(cfg.Host)
			m.mu.Lock()
			m.state = StateDisconnected
			m.activeID = ""
			m.mu.Unlock()
			m.sendState(StateError)
			m.log(fmt.Sprintf("TUN setup failed: %v", err))
			return fmt.Errorf("TUN setup: %w", err)
		}
		m.log("TUN routes configured")
	} else if useTUN {
		m.log("TUN mode (routes from previous setup)")
	}

	m.mu.Lock()
	m.xray = x
	m.state = StateConnected
	m.startedAt = time.Now()
	m.mu.Unlock()

	m.resetBackoff()
	m.networkUp.Store(true)
	m.sendState(StateConnected)

	if !useTUN && settings.SystemProxy {
		if err := proxy.SetSystemProxy("127.0.0.1", x.SocksPort); err != nil {
			m.log(fmt.Sprintf("Set system proxy: %v", err))
		} else {
			m.log("System proxy enabled")
		}
	} else if !useTUN {
		m.log("System proxy disabled by settings")
	}

	if useTUN {
		m.log(fmt.Sprintf("Connected via TUN! SOCKS5 debug on 127.0.0.1:%d", x.SocksPort))
	} else {
		m.log(fmt.Sprintf("Connected! SOCKS5 on 127.0.0.1:%d, HTTP on 127.0.0.1:%d", x.SocksPort, x.SocksPort+1000))
	}
	return nil
}

func (m *Manager) Disconnect() error {
	return m.disconnect(false)
}

func (m *Manager) disconnect(preserveActive bool) error {
	m.mu.Lock()
	wasConnected := m.state == StateConnected
	wasTUN := false
	var serverAddr string
	var x *xray.Xray
	if m.xray != nil {
		wasTUN = m.xray.IsTUN()
		if wasTUN {
			serverAddr = m.xray.ServerAddr()
		}
		x = m.xray
		m.xray = nil
	}
	m.state = StateDisconnected
	if !preserveActive {
		m.activeID = ""
	}
	m.mu.Unlock()

	defer func() {
		if wasConnected {
			if wasTUN {
				xray.TeardownTUNAdapter(serverAddr)
				m.log("TUN adapter removed")
			}
			if m.Settings != nil {
				s := m.Settings.Get()
				if !wasTUN && s.SystemProxy {
					proxy.ClearSystemProxy()
					m.log("System proxy disabled")
				}
			}
			m.sendState(StateDisconnected)
			m.log("Disconnected")
		}
	}()

	if x == nil {
		return nil
	}

	m.log("Stopping xray gracefully...")
	if err := x.Close(); err != nil {
		m.log(fmt.Sprintf("xray close error: %v", err))
	}

	return nil
}

func (m *Manager) GetState() ConnectionState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

func (m *Manager) CurrentMode() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.xray != nil && m.xray.IsTUN() {
		return "TUN"
	}
	return "Proxy"
}

func (m *Manager) SocksPort() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.xray != nil {
		return m.xray.SocksPort
	}
	return 0
}

func (m *Manager) Uptime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.state != StateConnected {
		return 0
	}
	return time.Since(m.startedAt)
}

func (m *Manager) GetLogChan() <-chan string {
	return m.logChan
}

func (m *Manager) GetStateChan() <-chan ConnectionState {
	return m.stateChan
}

// ─── Subscription updater ──────────────────────────────────────────

func (m *Manager) SetSubscriptions(items []SubscriptionItem) {
	m.subMu.Lock()
	defer m.subMu.Unlock()
	m.subscriptions = items
}

func (m *Manager) GetSubscriptions() []SubscriptionItem {
	m.subMu.Lock()
	defer m.subMu.Unlock()
	result := make([]SubscriptionItem, len(m.subscriptions))
	copy(result, m.subscriptions)
	return result
}

func (m *Manager) StartSubscriptionUpdater() {
	m.subsStop = make(chan struct{})
	go m.subscriptionLoop()
}

func (m *Manager) StopSubscriptionUpdater() {
	if m.subsStop != nil {
		close(m.subsStop)
	}
}

func (m *Manager) subscriptionLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.subsStop:
			return
		case <-ticker.C:
			m.checkSubscriptions()
		}
	}
}

func (m *Manager) checkSubscriptions() {
	items := m.GetSubscriptions()
	now := time.Now()

	for _, item := range items {
		if item.URL == "" || item.Interval <= 0 {
			continue
		}

		elapsed := now.Sub(item.Updated)
		if elapsed.Seconds() < float64(item.Interval) {
			continue
		}

		m.subMu.Lock()
		if _, running := m.subActive[item.URL]; running {
			m.subMu.Unlock()
			continue
		}
		done := make(chan struct{})
		m.subActive[item.URL] = done
		m.subMu.Unlock()

		go func(url string) {
			defer func() {
				m.subMu.Lock()
				delete(m.subActive, url)
				m.subMu.Unlock()
				close(done)
			}()

			m.log(fmt.Sprintf("Auto-updating subscription: %s", url))
			if err := m.refreshSubscription(url); err != nil {
				m.log(fmt.Sprintf("Subscription update failed: %v", err))
			} else {
				m.log(fmt.Sprintf("Subscription updated: %s", url))
			}
		}(item.URL)
	}
}

func (m *Manager) refreshSubscription(url string) error {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	text := string(body)
	imported := 0

	imported += m.importFromText(text)

	if imported > 0 {
		m.subMu.Lock()
		for i := range m.subscriptions {
			if m.subscriptions[i].URL == url {
				m.subscriptions[i].Updated = time.Now()
				break
			}
		}
		m.subMu.Unlock()
	}

	return nil
}

func (m *Manager) importFromText(text string) int {
	count := 0

	var links []string
	if err := json.Unmarshal([]byte(text), &links); err == nil {
		for _, link := range links {
			link = strings.TrimSpace(link)
			if link == "" {
				continue
			}
			cfg, err := vless.ParseVlessLink(link)
			if err != nil {
				continue
			}
			m.AddConfig(cfg)
			count++
		}
		return count
	}

	decoded, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		decoded, _ = base64.RawStdEncoding.DecodeString(text)
	}
	if err == nil {
		decodedStr := string(decoded)
		re := regexp.MustCompile(`vless://[^\s"']+`)
		matches := re.FindAllString(decodedStr, -1)
		seen := make(map[string]bool)
		for _, link := range matches {
			link = strings.TrimSpace(link)
			if seen[link] {
				continue
			}
			seen[link] = true
			cfg, err := vless.ParseVlessLink(link)
			if err != nil {
				continue
			}
			m.AddConfig(cfg)
			count++
		}
		return count
	}

	re := regexp.MustCompile(`vless://[^\s"']+`)
	matches := re.FindAllString(text, -1)
	seen := make(map[string]bool)
	for _, link := range matches {
		link = strings.TrimSpace(link)
		if seen[link] {
			continue
		}
		seen[link] = true
		cfg, err := vless.ParseVlessLink(link)
		if err != nil {
			continue
		}
		m.AddConfig(cfg)
		count++
	}
	return count
}

// ─── Network Detection ─────────────────────────────────────────────

func (m *Manager) StartNetworkDetection() {
	m.networkStop = make(chan struct{})
	go m.networkDetectionLoop()
}

func (m *Manager) StopNetworkDetection() {
	if m.networkStop != nil {
		close(m.networkStop)
	}
}

func (m *Manager) networkDetectionLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.networkStop:
			return
		case <-ticker.C:
			m.checkNetworkHealth()
		}
	}
}

func (m *Manager) checkNetworkHealth() {
	up := isAnyRealInterfaceUp()
	wasUp := m.networkUp.Load()

	if up && !wasUp {
		m.networkUp.Store(true)
		m.mu.RLock()
		wasConnected := m.state == StateConnected || m.state == StateConnecting
		activeID := m.activeID
		m.mu.RUnlock()

		if !wasConnected && activeID != "" {
			m.log("Network recovered, reconnecting...")
			go func(id string) {
				if err := m.Connect(id); err != nil {
					m.log(fmt.Sprintf("Auto-reconnect failed: %v", err))
				}
			}(activeID)
		}
	}

	if !up && wasUp {
		m.networkUp.Store(false)
		m.mu.RLock()
		isConnected := m.state == StateConnected
		m.mu.RUnlock()

		if isConnected {
			m.log("Network lost, disconnecting...")
			m.resetBackoff()
			go m.disconnect(true)
		}
	}
}

func isAnyRealInterfaceUp() bool {
	interfaces, err := net.Interfaces()
	if err != nil {
		return false
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		name := iface.Name
		if strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "veth") ||
			strings.HasPrefix(name, "br-") || strings.HasPrefix(name, "lo") ||
			name == "stride-vpn" {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil {
					return true
				}
			}
		}
	}

	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.Dial("tcp", "8.8.8.8:53")
	if err == nil {
		conn.Close()
		return true
	}
	return false
}

func (m *Manager) log(msg string) {
	select {
	case m.logChan <- fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg):
	default:
		m.drainLog(50)
		select {
		case m.logChan <- fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg):
		default:
		}
	}
}

func (m *Manager) drainLog(max int) {
	for i := 0; i < max; i++ {
		select {
		case <-m.logChan:
		default:
			return
		}
	}
}

func (m *Manager) sendState(state ConnectionState) {
	select {
	case m.stateChan <- state:
	default:
	}
}

func (m *Manager) nextBackoff() time.Duration {
	if m.reconnectAttempts == 0 {
		m.reconnectBackoff = 1 * time.Second
	} else {
		m.reconnectBackoff *= 2
	}
	if m.reconnectBackoff > m.reconnectMaxBackoff {
		m.reconnectBackoff = m.reconnectMaxBackoff
	}
	m.reconnectAttempts++

	jitter := time.Duration(rand.Int63n(int64(m.reconnectBackoff / 4)))
	backoff := m.reconnectBackoff - jitter

	m.log(fmt.Sprintf("Reconnect backoff: %v (attempt %d)", backoff, m.reconnectAttempts))
	return backoff
}

func (m *Manager) resetBackoff() {
	m.reconnectBackoff = 0
	m.reconnectAttempts = 0
}

func (m *Manager) waitBackoff() {
	if m.reconnectAttempts == 0 {
		return
	}
	backoff := m.nextBackoff()
	time.Sleep(backoff)
}
