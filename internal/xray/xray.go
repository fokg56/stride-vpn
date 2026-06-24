package xray

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/xtls/xray-core/main/distro/all"

	"github.com/xtls/xray-core/common/platform"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"

	"vpn-client/internal/vless"
)

var (
	assetDir string
)

func init() {
	exe, err := os.Executable()
	if err != nil {
		assetDir = "assets"
	} else {
		assetDir = filepath.Join(filepath.Dir(exe), "assets")
	}
	os.MkdirAll(assetDir, 0755)
	os.Setenv(platform.AssetLocation, assetDir)
	os.Setenv(platform.CertLocation, assetDir)

	assets := []string{"geoip.dat", "geosite.dat"}
	for _, a := range assets {
		p := filepath.Join(assetDir, a)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			go downloadAsset(a, p)
		}
	}
}

func downloadAsset(name, path string) {
	url := fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/latest/download/%s", name)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}
	out, err := os.Create(path)
	if err != nil {
		return
	}
	defer out.Close()
	io.Copy(out, resp.Body)
}

type Xray struct {
	server    *core.Instance
	config    *vless.Config
	SocksPort int
	useTUN    bool
	startedAt time.Time
}

func New(cfg *vless.Config) (*Xray, error) {
	return NewWithRouting(cfg, nil, false)
}

func NewWithRouting(cfg *vless.Config, rules *RouteRules, useTUN bool) (*Xray, error) {
	if useTUN {
		if err := EnsureWintunDLL(); err != nil {
			return nil, fmt.Errorf("ensure wintun: %w", err)
		}
	}

	port, err := FindFreePort()
	if err != nil {
		port = 10808
	}

	jsonCfg, err := BuildConfig(cfg, rules, port, useTUN)
	if err != nil {
		return nil, fmt.Errorf("build config: %w", err)
	}

	pbConfig, err := serial.LoadJSONConfig(bytes.NewReader([]byte(jsonCfg)))
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	server, err := core.New(pbConfig)
	if err != nil {
		return nil, fmt.Errorf("create xray: %w", err)
	}

	return &Xray{
		server:    server,
		config:    cfg,
		SocksPort: port,
		useTUN:    useTUN,
		startedAt: time.Now(),
	}, nil
}

func (x *Xray) Start() error {
	return x.server.Start()
}

func (x *Xray) Close() error {
	done := make(chan error, 1)
	go func() {
		done <- x.server.Close()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(3 * time.Second):
		return fmt.Errorf("close timed out (forced)")
	}
}

func (x *Xray) WaitForReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if x.useTUN {
			ifaces, err := net.Interfaces()
			if err == nil {
				for _, iface := range ifaces {
					if iface.Name == "stride-vpn" && iface.Flags&net.FlagUp != 0 {
						return nil
					}
				}
			}
		} else {
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", x.SocksPort), 200*time.Millisecond)
			if err == nil {
				conn.Close()
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}

	if x.useTUN {
		return fmt.Errorf("TUN interface not ready within %v", timeout)
	}
	return fmt.Errorf("SOCKS5 listener not ready within %v", timeout)
}

func (x *Xray) IsTUN() bool {
	return x.useTUN
}

func (x *Xray) ServerAddr() string {
	if x.config != nil {
		return x.config.Host
	}
	return ""
}

func (x *Xray) Uptime() time.Duration {
	return time.Since(x.startedAt)
}

func ValidateConfig(cfg *vless.Config, rules *RouteRules, port int, useTUN bool) error {
	jsonCfg, err := BuildConfig(cfg, rules, port, useTUN)
	if err != nil {
		return fmt.Errorf("build config: %w", err)
	}

	_, err = serial.LoadJSONConfig(bytes.NewReader([]byte(jsonCfg)))
	if err != nil {
		return fmt.Errorf("config validation: %w", err)
	}
	return nil
}

func FindFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}
