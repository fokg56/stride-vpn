//go:build windows

package xray

import (
	"archive/zip"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const createNoWindow = 0x08000000

func hiddenCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}
	return cmd
}

func IsAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}

func Elevate(connectID string) error {
	exe, _ := os.Executable()
	args := []string{"--admin"}
	if connectID != "" {
		args = append(args, "--connect="+connectID)
	}
	quotedArgs := make([]string, 0, len(args))
	for _, arg := range args {
		quotedArgs = append(quotedArgs, "'"+strings.ReplaceAll(arg, "'", "''")+"'")
	}
	script := fmt.Sprintf("Start-Process -FilePath '%s' -Verb runAs -WindowStyle Hidden -ArgumentList @(%s)",
		strings.ReplaceAll(exe, "'", "''"), strings.Join(quotedArgs, ","))
	cmd := hiddenCommand("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", script)
	return cmd.Start()
}

func EnsureWintunDLL() error {
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	dllPath := filepath.Join(dir, "wintun.dll")

	if _, err := os.Stat(dllPath); err == nil {
		return nil
	}

	if err := extractEmbeddedWintun(dllPath); err == nil {
		return nil
	}

	return downloadWintunDLL(dllPath)
}

func downloadWintunDLL(dllPath string) error {
	tmpDir, err := os.MkdirTemp("", "wintun")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "wintun.zip")
	urls := []string{
		"https://github.com/WireGuard/wintun/releases/download/v0.14.1/wintun-0.14.1.zip",
		"https://www.wintun.net/builds/wintun-0.14.1.zip",
	}
	for _, u := range urls {
		err = downloadFile(zipPath, u)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("download wintun: %w", err)
	}

	arch := runtime.GOARCH
	archDir := "amd64"
	switch arch {
	case "386":
		archDir = "x86"
	case "arm64":
		archDir = "arm64"
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	entryName := fmt.Sprintf("wintun/bin/%s/wintun.dll", archDir)
	for _, f := range r.File {
		if f.Name != entryName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open %s in zip: %w", entryName, err)
		}
		defer rc.Close()

		out, err := os.Create(dllPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", dllPath, err)
		}
		defer out.Close()

		if _, err := io.Copy(out, rc); err != nil {
			return fmt.Errorf("extract %s: %w", entryName, err)
		}
		return nil
	}

	return fmt.Errorf("wintun.dll not found in zip for arch %s", arch)
}

func downloadFile(path, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func SetupTUNAdapter(serverAddr string) error {
	ensureWintunAdapter()

	if err := setTUNIP(); err != nil {
		return fmt.Errorf("set TUN IP: %w", err)
	}

	if err := setTUNDNS(); err != nil {
		return fmt.Errorf("set TUN DNS: %w", err)
	}

	gateway, err := getDefaultGateway()
	if err != nil {
		return fmt.Errorf("get default gateway: %w", err)
	}
	fmt.Fprintf(os.Stderr, "TUN: original gateway = %s\n", gateway)

	if err := addServerRoute(serverAddr, gateway); err != nil {
		return fmt.Errorf("add server route: %w", err)
	}

	if err := addDefaultRouteViaTUN(); err != nil {
		return fmt.Errorf("add default route: %w", err)
	}

	return nil
}

func TeardownTUNAdapter(serverAddr string) {
	removeDefaultRouteViaTUN()
	removeServerRoute(serverAddr)
}

func ensureWintunAdapter() {
	// suppress "element not found" error — adapter may not exist yet;
	// xray-core will create it from the TUN inbound config
	hiddenCommand("netsh", "interface", "ipv4", "show", "interfaces",
		fmt.Sprintf("name=%s", tunName)).Run()
}

func setTUNIP() error {
	cmd := hiddenCommand("netsh", "interface", "ip", "set", "address",
		fmt.Sprintf("name=%s", tunName),
		"source=static",
		fmt.Sprintf("addr=%s", tunIP),
		fmt.Sprintf("mask=%s", tunMask),
		"gateway=none")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("netsh set addr failed: %s: %w", string(out), err)
	}
	return nil
}

func setTUNDNS() error {
	cmd := hiddenCommand("netsh", "interface", "ip", "set", "dns",
		fmt.Sprintf("name=%s", tunName),
		"source=static",
		fmt.Sprintf("addr=%s", tunIP))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("netsh set dns failed: %s: %w", string(out), err)
	}
	return nil
}

func getDefaultGateway() (string, error) {
	out, err := hiddenCommand("powershell", "-NoProfile", "-Command",
		"Get-NetRoute -DestinationPrefix '0.0.0.0/0' | Where-Object {$_.NextHop -and $_.NextHop -ne '0.0.0.0'} | Sort-Object RouteMetric,InterfaceMetric | Select-Object -First 1 -ExpandProperty NextHop").Output()
	if err == nil {
		gw := string(out)
		gw = trimString(gw)
		if net.ParseIP(gw) != nil {
			return gw, nil
		}
	}

	out, err = hiddenCommand("route", "print", "0.0.0.0").Output()
	if err != nil {
		return "", fmt.Errorf("route print: %w", err)
	}

	lines := string(out)
	for _, line := range splitLines(lines) {
		if len(line) < 80 {
			continue
		}
		parts := splitFields(line)
		if len(parts) >= 3 && parts[0] == "0.0.0.0" && parts[1] == "0.0.0.0" {
			gateway := parts[2]
			if net.ParseIP(gateway) != nil {
				return gateway, nil
			}
		}
	}
	return "", fmt.Errorf("default gateway not found")
}

func trimString(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r' || s[len(s)-1] == ' ') {
		s = s[:len(s)-1]
	}
	for len(s) > 0 && (s[0] == '\n' || s[0] == '\r' || s[0] == ' ') {
		s = s[1:]
	}
	return s
}

func addServerRoute(serverAddr, gateway string) error {
	ip := net.ParseIP(serverAddr)
	if ip == nil {
		ips, _ := net.LookupHost(serverAddr)
		if len(ips) == 0 {
			return fmt.Errorf("cannot resolve server address")
		}
		serverAddr = ips[0]
	}
	removeServerRoute(serverAddr)
	cmd := hiddenCommand("route", "add", serverAddr, "mask", "255.255.255.255", gateway, "metric", "1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("route add server: %s: %w", string(out), err)
	}
	return nil
}

func findTUNInterfaceIndex() string {
	// retry up to 5 times with backoff — adapter may not be fully enumerated by Windows yet
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			time.Sleep(400 * time.Millisecond)
		}
		idx := queryTUNInterfaceIndex()
		if idx != "" {
			return idx
		}
	}
	return ""
}

func queryTUNInterfaceIndex() string {
	// method 1: netsh by name (fastest)
	idx := netshQueryIndex(tunName)
	if idx != "" {
		return idx
	}
	// method 2: powershell Get-NetAdapter fallback (handles name mismatches)
	idx = psQueryIndex(tunName)
	if idx != "" {
		return idx
	}
	return ""
}

func netshQueryIndex(name string) string {
	out, err := hiddenCommand("netsh", "interface", "ipv4", "show", "interfaces",
		fmt.Sprintf("name=%s", name)).Output()
	if err != nil {
		return ""
	}
	lines := splitLines(string(out))
	for _, line := range lines {
		if strings.Contains(line, name) {
			parts := splitFields(line)
			if len(parts) >= 1 {
				return parts[0]
			}
		}
	}
	return ""
}

func psQueryIndex(name string) string {
	script := fmt.Sprintf(
		"@(Get-NetAdapter -Name '%s' -ErrorAction SilentlyContinue | Get-NetIPInterface -ErrorAction SilentlyContinue).InterfaceIndex",
		name,
	)
	out, err := hiddenCommand("powershell", "-NoProfile", "-Command", script).Output()
	if err != nil {
		return ""
	}
	idx := trimString(string(out))
	if idx != "" {
		return idx
	}
	return ""
}

func addDefaultRouteViaTUN() error {
	removeDefaultRouteViaTUN()
	ifIdx := findTUNInterfaceIndex()
	if ifIdx != "" {
		cmd := hiddenCommand("route", "add", "0.0.0.0", "mask", "0.0.0.0", tunGateway, "metric", "1", "IF", ifIdx)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("route add default: %s: %w", string(out), err)
		}
		return nil
	}
	cmd := hiddenCommand("route", "add", "0.0.0.0", "mask", "0.0.0.0", tunGateway, "metric", "1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("route add default: %s: %w", string(out), err)
	}
	return nil
}

func removeDefaultRouteViaTUN() {
	hiddenCommand("route", "delete", "0.0.0.0", "mask", "0.0.0.0", tunGateway).Run()
	hiddenCommand("route", "delete", "0.0.0.0", "mask", "0.0.0.0", tunGateway, "metric", "1").Run()
}

func removeServerRoute(serverAddr string) {
	ip := net.ParseIP(serverAddr)
	if ip == nil {
		ips, _ := net.LookupHost(serverAddr)
		for _, resolved := range ips {
			if net.ParseIP(resolved) != nil {
				hiddenCommand("route", "delete", resolved, "mask", "255.255.255.255").Run()
			}
		}
		return
	}
	hiddenCommand("route", "delete", serverAddr, "mask", "255.255.255.255").Run()
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' || s[i] == '\r' {
			if i > start {
				lines = append(lines, s[start:i])
			}
			if s[i] == '\r' && i+1 < len(s) && s[i+1] == '\n' {
				i++
			}
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitFields(s string) []string {
	var fields []string
	start := -1
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' {
			if start >= 0 {
				fields = append(fields, s[start:i])
				start = -1
			}
		} else if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		fields = append(fields, s[start:])
	}
	return fields
}
