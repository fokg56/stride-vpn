package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type AppSettings struct {
	AutoStart   bool     `json:"auto_start"`
	SocksPort   int      `json:"socks_port"`
	DNSServers  []string `json:"dns_servers,omitempty"`
	RoutingMode string   `json:"routing_mode"`
	Theme       string   `json:"theme"`
	SystemProxy bool     `json:"system_proxy"`
	TUNEnabled  bool     `json:"tun_enabled"`
}

type SettingsManager struct {
	mu       sync.RWMutex
	path     string
	settings *AppSettings
}

func NewSettingsManager(path string) *SettingsManager {
	if path == "" {
		exe, _ := os.Executable()
		dir := filepath.Dir(exe)
		path = filepath.Join(dir, "config", "settings.json")
	}
	sm := &SettingsManager{
		path: path,
		settings: &AppSettings{
			AutoStart:   false,
			SocksPort:   1080,
			DNSServers:  []string{"8.8.8.8", "1.1.1.1"},
			RoutingMode: "global",
			Theme:       "light",
			SystemProxy: true,
			TUNEnabled:  true,
		},
	}
	sm.load()
	return sm
}

func (sm *SettingsManager) Get() *AppSettings {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	cp := *sm.settings
	return &cp
}

func (sm *SettingsManager) Update(s *AppSettings) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.settings = s
	return sm.save()
}

func (sm *SettingsManager) load() {
	data, err := os.ReadFile(sm.path)
	if err != nil {
		return
	}
	var s AppSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return
	}
	sm.mu.Lock()
	sm.settings = &s
	sm.mu.Unlock()
}

func (sm *SettingsManager) save() error {
	data, err := json.MarshalIndent(sm.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("serialize settings: %w", err)
	}
	dir := filepath.Dir(sm.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(sm.path, data, 0644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}
	return nil
}
