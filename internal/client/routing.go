package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"vpn-client/internal/xray"
)

type RoutingProfile struct {
	Name          string   `json:"name"`
	DomainStrategy string  `json:"domainStrategy"`
	DNS           []string `json:"dns,omitempty"`
	DirectDomains []string `json:"directDomains,omitempty"`
	ProxyDomains  []string `json:"proxyDomains,omitempty"`
	BlockDomains  []string `json:"blockDomains,omitempty"`
	DirectIPs     []string `json:"directIPs,omitempty"`
	ProxyIPs      []string `json:"proxyIPs,omitempty"`
	BlockIPs      []string `json:"blockIPs,omitempty"`
	GlobalProxy   bool     `json:"globalProxy"`
}

type RoutingManager struct {
	mu       sync.RWMutex
	profiles []*RoutingProfile
	active   string
	path     string
}

func NewRoutingManager(path string) *RoutingManager {
	if path == "" {
		exe, _ := os.Executable()
		dir := filepath.Dir(exe)
		path = filepath.Join(dir, "config", "routing.json")
	}
	rm := &RoutingManager{
		profiles: defaultProfiles(),
		active:   "global",
		path:     path,
	}
	rm.load()
	return rm
}

func defaultProfiles() []*RoutingProfile {
	return []*RoutingProfile{
		{
			Name:          "global",
			DomainStrategy: "asis",
			GlobalProxy:   true,
		},
		{
			Name:          "bypass",
			DomainStrategy: "ipifnonmatch",
			ProxyDomains: []string{
				"github.com",
				"google.com",
				"youtube.com",
				"regexp:.*\\.com$",
			},
		},
		{
			Name:           "proxy-selective",
			DomainStrategy:  "ipifnonmatch",
			DirectDomains: []string{
				"regexp:.*\\.ru$",
				"regexp:.*\\.рф$",
				"yandex.ru",
				"mail.ru",
				"vk.com",
			},
		},
	}
}

func (rm *RoutingManager) Profiles() []*RoutingProfile {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	result := make([]*RoutingProfile, len(rm.profiles))
	copy(result, rm.profiles)
	return result
}

func (rm *RoutingManager) ActiveProfile() *RoutingProfile {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	for _, p := range rm.profiles {
		if p.Name == rm.active {
			return p
		}
	}
	return rm.profiles[0]
}

func (rm *RoutingManager) SetActive(name string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	for _, p := range rm.profiles {
		if p.Name == name {
			rm.active = name
			rm.save()
			return nil
		}
	}
	return fmt.Errorf("profile %q not found", name)
}

func (rm *RoutingManager) AddProfile(p *RoutingProfile) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	for _, existing := range rm.profiles {
		if existing.Name == p.Name {
			return fmt.Errorf("profile %q already exists", p.Name)
		}
	}
	rm.profiles = append(rm.profiles, p)
	rm.save()
	return nil
}

func (rm *RoutingManager) RemoveProfile(name string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if name == "global" {
		return fmt.Errorf("cannot remove default profile")
	}
	for i, p := range rm.profiles {
		if p.Name == name {
			rm.profiles = append(rm.profiles[:i], rm.profiles[i+1:]...)
			if rm.active == name {
				rm.active = "global"
			}
			rm.save()
			return nil
		}
	}
	return fmt.Errorf("profile %q not found", name)
}

func (rm *RoutingManager) RouteRules() *xray.RouteRules {
	p := rm.ActiveProfile()
	return &xray.RouteRules{
		DomainStrategy: p.DomainStrategy,
		DirectDomains:  p.DirectDomains,
		ProxyDomains:   p.ProxyDomains,
		BlockDomains:   p.BlockDomains,
		DirectIPs:      p.DirectIPs,
		ProxyIPs:       p.ProxyIPs,
		BlockIPs:       p.BlockIPs,
		GlobalProxy:    p.GlobalProxy,
	}
}

func (rm *RoutingManager) load() {
	data, err := os.ReadFile(rm.path)
	if err != nil {
		return
	}
	var stored struct {
		Profiles []*RoutingProfile `json:"profiles"`
		Active   string            `json:"active"`
	}
	if err := json.Unmarshal(data, &stored); err != nil {
		return
	}
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if len(stored.Profiles) > 0 {
		rm.profiles = stored.Profiles
	}
	for _, p := range rm.profiles {
		if p.Name == stored.Active {
			rm.active = stored.Active
			break
		}
	}
}

func (rm *RoutingManager) save() {
	data, err := json.MarshalIndent(map[string]interface{}{
		"profiles": rm.profiles,
		"active":   rm.active,
	}, "", "  ")
	if err != nil {
		return
	}
	dir := filepath.Dir(rm.path)
	os.MkdirAll(dir, 0755)
	os.WriteFile(rm.path, data, 0644)
}
