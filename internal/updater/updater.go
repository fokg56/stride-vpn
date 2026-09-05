package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"vpn-client/internal/version"
)

const (
	AssetName    = "stride-vpn.exe"
	CheckTimeout = 30 * time.Second
	// WaitSeconds is how long the apply script sleeps before touching
	// files, giving the current process time to exit cleanly.
	WaitSeconds = 5
)

// Asset is a file attached to a GitHub release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Release is the GitHub "latest release" payload.
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
}

type Updater struct {
	current string
	check   *http.Client
	dl      *http.Client
}

func New(currentVersion string) *Updater {
	return &Updater{
		current: currentVersion,
		check:   &http.Client{Timeout: CheckTimeout},
		dl:      &http.Client{},
	}
}

func (u *Updater) CurrentVersion() string {
	return u.current
}

// Check queries GitHub for the latest release.
// Returns the release (nil if there are no releases yet),
// whether an update is available, and an error.
func (u *Updater) Check() (*Release, bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest",
		version.RepoOwner, version.RepoName)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "stride-vpn-updater")

	resp, err := u.check.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GitHub API: HTTP %d", resp.StatusCode)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, false, err
	}
	if rel.TagName == "" {
		return &rel, false, nil
	}

	return &rel, CompareVersions(u.current, rel.TagName) < 0, nil
}

// FindAsset returns the stride-vpn.exe asset of the release, if any.
func (u *Updater) FindAsset(rel *Release) *Asset {
	for i := range rel.Assets {
		if rel.Assets[i].Name == AssetName {
			return &rel.Assets[i]
		}
	}
	return nil
}

// Download streams the asset to dest, reporting progress (done/total).
func (u *Updater) Download(asset *Asset, dest string, progress func(done, total int64)) error {
	req, err := http.NewRequest(http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "stride-vpn-updater")

	resp, err := u.dl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if progress != nil && resp.ContentLength > 0 {
		var done int64
		buf := make([]byte, 128*1024)
		for {
			n, rerr := resp.Body.Read(buf)
			if n > 0 {
				if _, werr := out.Write(buf[:n]); werr != nil {
					return werr
				}
				done += int64(n)
				progress(done, resp.ContentLength)
			}
			if rerr == io.EOF {
				break
			}
			if rerr != nil {
				return rerr
			}
		}
		return nil
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

// VerifySHA256 checks dest against a "<asset>.sha256" release asset.
// If the release does not publish a checksum, verification is skipped.
func (u *Updater) VerifySHA256(rel *Release, dest string) error {
	var asset *Asset
	for i := range rel.Assets {
		if rel.Assets[i].Name == AssetName+".sha256" {
			asset = &rel.Assets[i]
			break
		}
	}
	if asset == nil {
		return nil
	}

	req, err := http.NewRequest(http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "stride-vpn-updater")

	resp, err := u.check.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("checksum download: HTTP %d", resp.StatusCode)
	}

	var expected string
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		// "sha256:hex  filename" or bare "hex"
		expected = strings.TrimPrefix(fields[0], "sha256:")
		if expected != "" {
			break
		}
	}
	expected = strings.ToLower(expected)
	if expected == "" {
		return fmt.Errorf("checksum file is empty")
	}

	f, err := os.Open(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := hex.EncodeToString(h.Sum(nil))

	if actual != expected {
		return fmt.Errorf("checksum mismatch (expected %s, got %s)", expected, actual)
	}
	return nil
}

// Apply installs the downloaded file next to the running executable and
// schedules the swap + restart through a detached cmd script.
// The caller should exit the process promptly after Apply succeeds.
func (u *Updater) Apply(downloadPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exeAbs, err := filepath.Abs(exe)
	if err != nil {
		return err
	}
	dir := filepath.Dir(exeAbs)

	newPath := filepath.Join(dir, filepath.Base(exeAbs)+".new")
	if err := moveFile(downloadPath, newPath); err != nil {
		return fmt.Errorf("stage update: %w", err)
	}

	oldPath := exeAbs + ".old"
	script := fmt.Sprintf(
		"ping 127.0.0.1 -n %d >nul && "+
			"move /Y \"%s\" \"%s\" && "+
			"move /Y \"%s\" \"%s\" && "+
			"start \"\" \"%s\" & "+
			"del /F /Q \"%s\"",
		WaitSeconds+1, exeAbs, oldPath, newPath, exeAbs, exeAbs, oldPath)

	cmd := exec.Command("cmd", "/c", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	return cmd.Start()
}

// moveFile renames src to dst, falling back to copy+remove across volumes.
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return os.Remove(src)
}

// CompareVersions compares two semver-like versions ("v1.0.3" / "1.0.3").
// Returns -1 if a < b, 0 if equal, 1 if a > b.
func CompareVersions(a, b string) int {
	pa := parseVersion(a)
	pb := parseVersion(b)
	n := len(pa)
	if len(pb) > n {
		n = len(pb)
	}
	for i := 0; i < n; i++ {
		x, y := 0, 0
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		if x < y {
			return -1
		}
		if x > y {
			return 1
		}
	}
	return 0
}

func parseVersion(v string) []int {
	v = strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(v), "v"), "V")
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			n = 0
		}
		out = append(out, n)
	}
	return out
}
