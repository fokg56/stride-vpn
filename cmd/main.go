package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"vpn-client/internal/client"
	"vpn-client/internal/storage"
	"vpn-client/internal/web"
)

func main() {
	addr := flag.String("addr", "0.0.0.0:8080", "Web server address")
	configPath := flag.String("config", "", "Path to config file (default: configs.json next to exe)")
	isAdmin := flag.Bool("admin", false, "Run as admin helper with TUN")
	connectID := flag.String("connect", "", "Auto-connect to this config (admin mode)")
	noBrowser := flag.Bool("no-browser", false, "Do not open the browser automatically")
	flag.Parse()

	store := storage.NewStore(*configPath)
	routing := client.NewRoutingManager("")
	settings := client.NewSettingsManager("")
	manager := client.NewManager(store, routing, settings)

	if *isAdmin {
		if *connectID != "" {
			go func() {
				time.Sleep(500 * time.Millisecond)
				if err := manager.Connect(*connectID); err != nil {
					fmt.Fprintf(os.Stderr, "Connect error: %v\n", err)
				}
			}()
		}
	}

	server := web.NewServer(*addr, manager, routing, settings)
	server.SetIdleHandler(func() {
		manager.Disconnect()
		os.Exit(0)
	})

	go func() {
		if err := server.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}()

	if !*isAdmin && !*noBrowser {
		time.Sleep(500 * time.Millisecond)

		host := *addr
		if strings.HasPrefix(host, "0.0.0.0:") {
			host = "localhost:" + strings.TrimPrefix(host, "0.0.0.0:")
		}
		url := fmt.Sprintf("http://%s", host)

		if runtime.GOOS == "windows" {
			cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
			cmd.Start()
		} else {
			exec.Command("xdg-open", url).Start()
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	manager.Disconnect()
	os.Exit(0)
}
