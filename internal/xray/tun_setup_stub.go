//go:build !windows

package xray

func IsAdmin() bool {
	return false
}

func Elevate() error {
	return nil
}

func EnsureWintunDLL() error {
	return nil
}

func SetupTUNAdapter(serverAddr string) error {
	return nil
}

func TeardownTUNAdapter(serverAddr string) {}
