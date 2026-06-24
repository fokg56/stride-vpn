package proxy

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
)

var savedProxyValue string
var savedProxyEnable uint32

func EnableSystemProxy(socksAddr string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("не удалось открыть реестр: %w", err)
	}
	defer k.Close()

	val, _, _ := k.GetIntegerValue("ProxyEnable")
	savedProxyEnable = uint32(val)
	savedProxyValue, _, _ = k.GetStringValue("ProxyServer")

	if err := k.SetDWordValue("ProxyEnable", 1); err != nil {
		return fmt.Errorf("не удалось включить прокси: %w", err)
	}

	proxyStr := "socks=" + socksAddr
	if err := k.SetStringValue("ProxyServer", proxyStr); err != nil {
		k.SetDWordValue("ProxyEnable", 0)
		return fmt.Errorf("не удалось установить прокси: %w", err)
	}

	return nil
}

func DisableSystemProxy() error {
	k, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("не удалось открыть реестр: %w", err)
	}
	defer k.Close()

	if savedProxyEnable != 0 {
		k.SetDWordValue("ProxyEnable", uint32(savedProxyEnable))
	} else {
		k.SetDWordValue("ProxyEnable", 0)
	}

	if savedProxyValue != "" {
		k.SetStringValue("ProxyServer", savedProxyValue)
	} else {
		k.DeleteValue("ProxyServer")
	}

	savedProxyValue = ""
	savedProxyEnable = 0
	return nil
}
