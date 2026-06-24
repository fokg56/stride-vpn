# Build

```powershell
cd D:\vpn-client
go build -ldflags="-H windowsgui" -o vpn-client.exe ./cmd/
```

- `-H windowsgui` hides the console window (GUI app mode)
- Without it, a console window appears alongside the systray

# Architecture

- `internal/client/routing.go`: `RoutingProfile` (domain strategy, DNS, rules), `RoutingManager` (CRUD, active profile)
- `internal/xray/config.go`: `RouteRules` struct maps profile to xray routing rules (direct/proxy/block, domain strategy)
- `internal/xray/xray.go`: `NewWithRouting(cfg, *RouteRules)` creates xray instance with routing
- `internal/web/server.go`: API `/api/import/subscription` fetches URL, handles base64/JSON/plaintext subscription formats

# Critical Details

- `routeRulesFromProfile()` in routing.go converts `RoutingProfile` -> `*xray.RouteRules` for xray config
- When `GlobalProxy=false`: default outbound is `direct` (freedom), proxy rules override for specific sites
- Subscription handler tries: JSON array -> base64 decode + regex -> raw regex for vless:// links
- Button animation: `::before` pseudo-element with conic-gradient + rotate animation for connecting spinner; full green background + glow for connected
