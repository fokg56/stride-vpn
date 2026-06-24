package proxy

import (
    "golang.org/x/sys/windows/registry"
    "fmt"
)

func SetSystemProxy(host string, port int) error {
    k, err := registry.OpenKey(registry.CURRENT_USER,
        `Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
        registry.SET_VALUE)
    if err != nil { return fmt.Errorf("open registry: %w", err) }
    defer k.Close()

    httpPort := port + 1000
    proxyStr := fmt.Sprintf("http=%s:%d;socks=%s:%d", host, httpPort, host, port)
    if err := k.SetStringValue("ProxyServer", proxyStr); err != nil {
        return fmt.Errorf("set ProxyServer: %w", err)
    }
    if err := k.SetDWordValue("ProxyEnable", 1); err != nil {
        return fmt.Errorf("set ProxyEnable: %w", err)
    }
    if err := k.SetStringValue("ProxyOverride", "<local>"); err != nil {
        return fmt.Errorf("set ProxyOverride: %w", err)
    }
    return nil
}

func ClearSystemProxy() error {
    k, err := registry.OpenKey(registry.CURRENT_USER,
        `Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
        registry.SET_VALUE)
    if err != nil { return fmt.Errorf("open registry: %w", err) }
    defer k.Close()
    return k.SetDWordValue("ProxyEnable", 0)
}
