package web

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"vpn-client/internal/client"
	"vpn-client/internal/vless"
)

func ImportSubscription(url string, mgr *client.Manager) ([]string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch subscription: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	text := string(body)

	imported := tryParseJSON(text, mgr)
	if len(imported) > 0 {
		return imported, nil
	}

	imported = tryParseBase64(text, mgr)
	if len(imported) > 0 {
		return imported, nil
	}

	imported = tryParseRegex(text, mgr)
	if len(imported) > 0 {
		return imported, nil
	}

	return nil, fmt.Errorf("no valid configs found in subscription")
}

func tryParseJSON(text string, mgr *client.Manager) []string {
	var links []string
	if err := json.Unmarshal([]byte(text), &links); err == nil {
		return importLinks(links, mgr)
	}
	var obj struct {
		Configs []string `json:"configs"`
	}
	if err := json.Unmarshal([]byte(text), &obj); err == nil && len(obj.Configs) > 0 {
		return importLinks(obj.Configs, mgr)
	}
	return nil
}

func tryParseBase64(text string, mgr *client.Manager) []string {
	decoded, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(text)
	}
	if err != nil {
		return nil
	}
	decodedStr := string(decoded)

	links := extractVlessLinks(decodedStr)
	if len(links) > 0 {
		return importLinks(links, mgr)
	}

	var links2 []string
	if err := json.Unmarshal(decoded, &links2); err == nil {
		return importLinks(links2, mgr)
	}

	return nil
}

func tryParseRegex(text string, mgr *client.Manager) []string {
	links := extractVlessLinks(text)
	if len(links) > 0 {
		return importLinks(links, mgr)
	}
	return nil
}

func extractVlessLinks(text string) []string {
	re := regexp.MustCompile(`vless://[^\s"']+`)
	matches := re.FindAllString(text, -1)
	unique := make(map[string]bool)
	result := make([]string, 0, len(matches))
	for _, m := range matches {
		m = strings.TrimSpace(m)
		if !unique[m] {
			unique[m] = true
			result = append(result, m)
		}
	}
	return result
}

func importLinks(links []string, mgr *client.Manager) []string {
	imported := make([]string, 0, len(links))
	for _, link := range links {
		link = strings.TrimSpace(link)
		if link == "" {
			continue
		}
		cfg, err := vless.ParseVlessLink(link)
		if err != nil {
			continue
		}
		id := mgr.AddConfig(cfg)
		imported = append(imported, id)
	}
	return imported
}
