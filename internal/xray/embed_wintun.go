//go:build windows

package xray

import (
	_ "embed"
	"os"
)

//go:embed wintun.dll
var wintunDLL []byte

func extractEmbeddedWintun(dst string) error {
	if err := os.WriteFile(dst, wintunDLL, 0644); err != nil {
		return err
	}
	return nil
}
