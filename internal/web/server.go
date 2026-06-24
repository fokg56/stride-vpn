package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"vpn-client/internal/client"
	"vpn-client/internal/vless"
)

type Server struct {
	addr      string
	manager   *client.Manager
	routing   *client.RoutingManager
	settings  *client.SettingsManager
	upgrader  websocket.Upgrader
	wsClients map[*websocket.Conn]bool
	onIdle    func()
	idleTimer *time.Timer
	mu        sync.RWMutex
}

func NewServer(addr string, manager *client.Manager, routing *client.RoutingManager, settings *client.SettingsManager) *Server {
	return &Server{
		addr:     addr,
		manager:  manager,
		routing:  routing,
		settings: settings,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		wsClients: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) SetIdleHandler(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onIdle = fn
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handleIndex)
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(assetFS())))

	// Legacy API routes
	mux.HandleFunc("/api/configs", s.handleConfigs)
	mux.HandleFunc("/api/config/", s.handleConfigInfo)
	mux.HandleFunc("/api/connect", s.handleConnect)
	mux.HandleFunc("/api/disconnect", s.handleDisconnect)
	mux.HandleFunc("/api/import", s.handleImport)
	mux.HandleFunc("/api/delete", s.handleDelete)
	mux.HandleFunc("/api/stats", s.handleStats)

	// v1 API routes
	mux.HandleFunc("/api/v1/configs", s.handleConfigs)
	mux.HandleFunc("/api/v1/config/", s.handleConfigInfo)
	mux.HandleFunc("/api/v1/connect", s.handleConnect)
	mux.HandleFunc("/api/v1/disconnect", s.handleDisconnect)
	mux.HandleFunc("/api/v1/import", s.handleImport)
	mux.HandleFunc("/api/v1/delete", s.handleDelete)
	mux.HandleFunc("/api/v1/stats", s.handleStats)
	mux.HandleFunc("/api/v1/import/subscription", s.handleSubscriptionImport)
	mux.HandleFunc("/api/v1/settings", s.handleSettings)
	mux.HandleFunc("/api/v1/settings/save", s.handleSettingsSave)
	mux.HandleFunc("/api/v1/routing/profiles", s.handleRoutingProfiles)
	mux.HandleFunc("/api/v1/routing/profile", s.handleRoutingProfile)
	mux.HandleFunc("/api/v1/routing/active", s.handleRoutingActive)
	mux.HandleFunc("/api/v1/server/ping/", s.handleServerPing)
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/v1/map-data", handleMapData)

	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/api/v1/logs/stream", s.handleWebSocket)

	log.Printf("Server started on http://%s", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

type ImportRequest struct {
	Link string `json:"link"`
}

func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	cfg, err := vless.ParseVlessLink(req.Link)
	if err != nil {
		jsonError(w, fmt.Sprintf("Parse error: %v", err), http.StatusBadRequest)
		return
	}

	id := s.manager.AddConfig(cfg)

	s.broadcastMessage(map[string]interface{}{
		"type": "log",
		"data": fmt.Sprintf("Imported: %s", cfg.Remark),
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"id":     id,
		"remark": cfg.Remark,
		"host":   cfg.Host,
		"port":   cfg.Port,
	})
}

func (s *Server) handleSubscriptionImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		jsonError(w, "URL is required", http.StatusBadRequest)
		return
	}

	imported, err := ImportSubscription(req.URL, s.manager)
	if err != nil {
		jsonError(w, fmt.Sprintf("Subscription error: %v", err), http.StatusBadRequest)
		return
	}

	s.broadcastMessage(map[string]interface{}{
		"type": "log",
		"data": fmt.Sprintf("Subscription imported: %d configs", len(imported)),
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ok",
		"imported": imported,
		"count":    len(imported),
	})
}

func (s *Server) handleConfigs(w http.ResponseWriter, r *http.Request) {
	configs := s.manager.ListConfigs()
	type configView struct {
		ID       string `json:"id"`
		Remark   string `json:"remark"`
		Host     string `json:"host"`
		Port     uint16 `json:"port"`
		Active   bool   `json:"active"`
		Security string `json:"security"`
	}

	activeCfg := s.manager.GetActiveConfig()
	var activeID string
	if activeCfg != nil {
		activeID = activeCfg.ID
	}

	views := make([]configView, 0, len(configs))
	for _, entry := range configs {
		views = append(views, configView{
			ID:       entry.ID,
			Remark:   entry.Remark,
			Host:     entry.Config.Host,
			Port:     entry.Config.Port,
			Active:   entry.ID == activeID,
			Security: string(entry.Config.Security),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(views)
}

func (s *Server) handleConfigInfo(w http.ResponseWriter, r *http.Request) {
	var id string
	if strings.HasPrefix(r.URL.Path, "/api/v1/config/") {
		id, _ = url.QueryUnescape(r.URL.Path[len("/api/v1/config/"):])
	} else {
		id, _ = url.QueryUnescape(r.URL.Path[len("/api/config/"):])
	}
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	info := s.manager.GetConfigInfo(id)
	if info == nil {
		http.Error(w, "Config not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := s.manager.Connect(req.ID); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "connected"})
}

func (s *Server) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.manager.Disconnect()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "disconnected"})
}

func (s *Server) handleServerPing(w http.ResponseWriter, r *http.Request) {
	id, _ := url.QueryUnescape(strings.TrimPrefix(r.URL.Path, "/api/v1/server/ping/"))
	if id == "" {
		jsonError(w, "ID required", http.StatusBadRequest)
		return
	}

	ms, err := s.manager.Ping(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"latency": ms,
	})
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if !s.manager.RemoveConfig(req.ID) {
		http.Error(w, "Config not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	state := s.manager.GetState()
	configs := s.manager.ListConfigs()
	uptime := s.manager.Uptime()

	activeCfg := s.manager.GetActiveConfig()
	var activeHost, activeRemark string
	if activeCfg != nil {
		activeHost = activeCfg.Config.Host
		activeRemark = activeCfg.Remark
	}

	socksPort := s.manager.SocksPort()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"state":         string(state),
		"totalConfigs":  len(configs),
		"activeHost":    activeHost,
		"activeRemark":  activeRemark,
		"proxyAddress":  fmt.Sprintf("127.0.0.1:%d", socksPort),
		"uptimeSeconds": int(uptime.Seconds()),
		"socksPort":     socksPort,
		"mode":          s.manager.CurrentMode(),
	})
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		settings := s.settings.Get()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(settings)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleSettingsSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var settings client.AppSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		jsonError(w, fmt.Sprintf("Invalid settings: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.settings.Update(&settings); err != nil {
		jsonError(w, fmt.Sprintf("Save error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}

func (s *Server) handleRoutingProfiles(w http.ResponseWriter, r *http.Request) {
	profiles := s.routing.Profiles()
	active := s.routing.ActiveProfile()

	type profileView struct {
		*client.RoutingProfile
		IsActive bool `json:"isActive"`
	}

	views := make([]profileView, 0, len(profiles))
	for _, p := range profiles {
		views = append(views, profileView{
			RoutingProfile: p,
			IsActive:       p.Name == active.Name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(views)
}

func (s *Server) handleRoutingProfile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var p client.RoutingProfile
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			jsonError(w, fmt.Sprintf("Invalid profile: %v", err), http.StatusBadRequest)
			return
		}
		if err := s.routing.AddProfile(&p); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "created"})

	case http.MethodDelete:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := s.routing.RemoveProfile(req.Name); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRoutingActive(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		p := s.routing.ActiveProfile()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)

	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := s.routing.SetActive(req.Name); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "active"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket error: %v", err)
		return
	}

	s.mu.Lock()
	if s.idleTimer != nil {
		s.idleTimer.Stop()
		s.idleTimer = nil
	}
	s.wsClients[conn] = true
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	go func() {
		logCh := s.manager.GetLogChan()
		stateCh := s.manager.GetStateChan()
		for {
			select {
			case <-done:
				return
			case logMsg, ok := <-logCh:
				if !ok {
					return
				}
				tag := "info"
				if strings.Contains(logMsg, "Error") || strings.Contains(logMsg, "error") {
					tag = "error"
				} else if strings.Contains(logMsg, "Connected") || strings.Contains(logMsg, "Imported") {
					tag = "ok"
				} else if strings.Contains(logMsg, "Warning") {
					tag = "warn"
				}
				if err := conn.WriteJSON(map[string]interface{}{
					"type": "log",
					"data": logMsg,
					"tag":  tag,
				}); err != nil {
					return
				}
			case state, ok := <-stateCh:
				if !ok {
					return
				}
				if err := conn.WriteJSON(map[string]interface{}{
					"type": "state",
					"data": string(state),
					"mode": s.manager.CurrentMode(),
				}); err != nil {
					return
				}
			}
		}
	}()

	<-done
	s.mu.Lock()
	delete(s.wsClients, conn)
	s.scheduleIdleShutdownLocked()
	s.mu.Unlock()
	conn.Close()
}

func (s *Server) scheduleIdleShutdownLocked() {
	if len(s.wsClients) != 0 || s.onIdle == nil {
		return
	}
	if s.idleTimer != nil {
		s.idleTimer.Stop()
	}
	s.idleTimer = time.AfterFunc(2500*time.Millisecond, func() {
		s.mu.RLock()
		empty := len(s.wsClients) == 0
		onIdle := s.onIdle
		s.mu.RUnlock()
		if empty && onIdle != nil {
			onIdle()
		}
	})
}

func (s *Server) broadcastMessage(msg map[string]interface{}) {
	s.mu.RLock()
	clients := make([]*websocket.Conn, 0, len(s.wsClients))
	for conn := range s.wsClients {
		clients = append(clients, conn)
	}
	s.mu.RUnlock()

	for _, conn := range clients {
		if err := conn.WriteJSON(msg); err != nil {
			s.mu.Lock()
			delete(s.wsClients, conn)
			s.mu.Unlock()
			conn.Close()
		}
	}
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
