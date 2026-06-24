package web

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Stride VPN</title>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Roboto:wght@400;500;700&display=swap">
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500&display=swap">
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0,200">
<style>
:root,[data-theme="dark"]{
  --primary:#7C4DFF;
  --on-primary:#FFFFFF;
  --primary-container:#2D235E;
  --on-primary-container:#EADDFF;
  --secondary:#A7C7E7;
  --secondary-container:#18324A;
  --tertiary:#5DD6C7;
  --surface:#111218;
  --surface-dim:#0B0C10;
  --surface-bright:#23242B;
  --surface-container-low:#181920;
  --surface-container:#1D1E26;
  --surface-container-high:#262832;
  --on-surface:#F2EFF7;
  --on-surface-variant:#C9C4D3;
  --outline:#8E8799;
  --outline-variant:#3E3A49;
  --error:#F2B8B5;
  --error-container:#8C1D18;
  --success:#44D07B;
  --success-container:rgba(68,208,123,0.14);
  --shadow-sm:0 1px 2px rgba(0,0,0,0.24);
  --shadow-md:0 4px 14px rgba(0,0,0,0.30);
  --shadow-lg:0 12px 28px rgba(0,0,0,0.42);
  --radius-sm:8px;
  --radius-md:8px;
  --radius-lg:8px;
  --radius-full:50%;
  --font:'Roboto',system-ui,sans-serif;
  --mono:'JetBrains Mono',monospace;
}
[data-theme="light"]{
  --primary:#5A35D6;
  --on-primary:#FFFFFF;
  --primary-container:#E9DDFF;
  --on-primary-container:#1F0A58;
  --secondary:#365F7D;
  --secondary-container:#D3E8FF;
  --tertiary:#006B5F;
  --surface:#FBF8FF;
  --surface-dim:#DDD8E7;
  --surface-bright:#FFFFFF;
  --surface-container-low:#F5F0FA;
  --surface-container:#F0EAF6;
  --surface-container-high:#EAE4F1;
  --on-surface:#1D1B20;
  --on-surface-variant:#4B4658;
  --outline:#7C748A;
  --outline-variant:#CBC3D8;
  --error:#BA1A1A;
  --error-container:#FFDAD6;
  --success:#16a34a;
  --success-container:rgba(22,163,74,0.1);
  --shadow-sm:0 1px 3px rgba(0,0,0,0.08);
  --shadow-md:0 4px 16px rgba(0,0,0,0.12);
  --shadow-lg:0 8px 32px rgba(0,0,0,0.16);
}
*{margin:0;padding:0;box-sizing:border-box}
html,body{height:100%;overflow:hidden}
body{
  font-family:var(--font);
  background:var(--surface-dim);
  color:var(--on-surface);
  -webkit-font-smoothing:antialiased;
}
.flex{display:flex}.mt-8{margin-top:8px}
.material-symbols-outlined{font-variation-settings:'FILL' 0,'wght' 400,'GRAD' 0,'opsz' 24;user-select:none;font-size:24px}

/* LAYOUT */
.app{display:flex;height:100vh;overflow:hidden}

/* SIDEBAR */
.sidebar{
  width:96px;min-width:96px;
  background:var(--surface);
  display:flex;flex-direction:column;align-items:center;
  padding:18px 0;gap:6px;
  border-right:1px solid var(--outline-variant);
}
.sidebar-brand{
  width:56px;height:56px;border-radius:18px;
  background:#050505;
  display:flex;align-items:center;justify-content:center;
  margin-bottom:18px;user-select:none;
  box-shadow:var(--shadow-md);overflow:hidden;
}
.sidebar-brand img{
  width:100%;height:100%;object-fit:cover;display:block;
}
.sidebar-btn{
  width:72px;min-height:64px;border-radius:20px;
  border:none;background:transparent;
  color:var(--on-surface-variant);cursor:pointer;
  display:flex;flex-direction:column;align-items:center;justify-content:center;gap:2px;
  transition:all 0.2s;
  font-family:var(--font);font-size:11px;font-weight:500;
  position:relative;
}
.sidebar-btn:hover{background:var(--surface-container)}
.sidebar-btn.active{color:var(--on-primary-container);background:var(--primary-container)}
.sidebar-btn .material-symbols-outlined{font-size:24px}
.sidebar-spacer{flex:1}
.sidebar-bottom{display:flex;flex-direction:column;align-items:center;gap:4px;padding:0 8px;margin-top:auto}
.sidebar-dot{width:10px;height:10px;border-radius:50%;background:var(--outline);transition:all 0.3s;margin-bottom:8px}
.sidebar-dot.online{background:var(--success);box-shadow:0 0 12px var(--success)}
.sidebar-dot.connecting{background:var(--primary);animation:pulse 0.8s ease-in-out infinite}
@keyframes pulse{0%,100%{opacity:1}50%{opacity:0.3}}

/* MAIN — transparent to let world map show in right-side empty space */
.main{
  flex:1;display:flex;flex-direction:column;overflow:hidden;
  background:transparent;
}

/* SCREENS */
.screen{display:none;flex-direction:column;height:100%}
.screen.active{display:flex}

/* TOPBAR */
.topbar{
  display:flex;align-items:center;justify-content:space-between;
  padding:20px 28px 8px;min-height:64px;flex-shrink:0;
  background:var(--surface);
}
.topbar h1{font-size:24px;font-weight:500;color:var(--on-surface);letter-spacing:0}

/* SCROLL — max-width 980px, map shows in right-side empty space */
.scroll{
  flex:1;overflow-y:auto;padding:8px 28px 120px;max-width:980px;width:100%;
  background:var(--surface);
}
.scroll::-webkit-scrollbar{width:4px}
.scroll::-webkit-scrollbar-thumb{background:var(--outline-variant);border-radius:2px}

/* CARD */
.card{
  background:var(--surface-container);border:1px solid var(--outline-variant);
  border-radius:var(--radius-md);padding:16px;margin-bottom:12px;
  box-shadow:none;
}
.card-title{font-size:12px;font-weight:700;color:var(--on-surface-variant);text-transform:uppercase;letter-spacing:0.4px;margin-bottom:12px}

/* STATUS BAR */
.status-bar{
  display:flex;align-items:center;gap:12px;
  padding:16px;margin-bottom:8px;
  background:var(--surface-container-low);
  border:1px solid var(--outline-variant);
  border-radius:var(--radius-md);
}
.status-led{
  width:14px;height:14px;border-radius:50%;
  background:var(--outline);transition:all 0.4s;flex-shrink:0;
}
.status-led.connected{background:var(--success);box-shadow:0 0 16px var(--success)}
.status-led.connecting{background:var(--primary);animation:pulse 0.8s ease-in-out infinite}
.status-led.error{background:var(--error)}
.status-text{font-size:22px;font-weight:500;color:var(--on-surface);letter-spacing:0}
.status-sec{font-size:13px;color:var(--on-surface-variant);font-family:var(--mono);margin-left:auto}

/* POWER BUTTON */
.power-wrap{
  display:flex;align-items:center;justify-content:center;
  padding:26px 0 24px;position:relative;
}
.power-btn{
  width:112px;height:112px;border-radius:32px;border:none;
  background:var(--primary-container);
  color:var(--on-primary-container);
  cursor:pointer;position:relative;
  box-shadow:var(--shadow-md);
  transition:all 0.3s cubic-bezier(0.4,0,0.2,1);
  display:flex;align-items:center;justify-content:center;
  border:1px solid color-mix(in srgb,var(--primary) 34%,var(--outline-variant));
}
.power-btn:hover{box-shadow:var(--shadow-lg);transform:translateY(-1px)}
.power-btn:active{transform:scale(0.95)}
.power-btn .material-symbols-outlined{font-size:44px;transition:all 0.3s;font-variation-settings:'FILL' 1,'wght' 500,'GRAD' 0,'opsz' 40}
.power-btn.connected{
  background:var(--success);border-color:var(--success);
  color:#fff;box-shadow:0 0 32px rgba(34,197,94,0.3);
}
.power-btn.connected .material-symbols-outlined{transform:rotate(0deg)}
.power-btn.connecting{animation:pulse-btn 1.5s ease-in-out infinite}
.power-btn.connecting .material-symbols-outlined{animation:spin 0.8s linear infinite}
@keyframes pulse-btn{0%,100%{box-shadow:0 0 16px var(--primary)}50%{box-shadow:0 0 40px var(--primary)}}
@keyframes spin{from{transform:rotate(0deg)}to{transform:rotate(360deg)}}
.power-ring{
  position:absolute;inset:-4px;border-radius:50%;
  border:2px solid transparent;transition:all 0.5s;
}
.power-btn.connected .power-ring{
  border-color:var(--success);opacity:0.5;
  animation:ring-pulse 2s ease-in-out infinite;
}
@keyframes ring-pulse{0%,100%{opacity:0.3;transform:scale(1)}50%{opacity:0.6;transform:scale(1.05)}}

/* INPUT */
.input{
  width:100%;padding:12px 14px;
  background:var(--surface-container-high);border:1px solid var(--outline-variant);
  border-radius:var(--radius-sm);color:var(--on-surface);
  font-family:var(--font);font-size:14px;outline:none;transition:border 0.15s;
}
.input:focus{border-color:var(--primary);box-shadow:0 0 0 2px color-mix(in srgb,var(--primary) 22%,transparent)}
.input::placeholder{color:var(--on-surface-variant)}
.input-group{display:flex;gap:8px}
.input-group .input{flex:1}

/* BUTTONS */
.btn{
  display:inline-flex;align-items:center;justify-content:center;gap:6px;
  padding:0 20px;height:40px;border:none;border-radius:20px;
  font-family:var(--font);font-size:14px;font-weight:500;cursor:pointer;
  transition:all 0.15s;white-space:nowrap;user-select:none;
}
.btn:active{transform:scale(0.95)}
.btn:disabled{opacity:0.4;pointer-events:none}
.btn-primary{background:var(--primary);color:var(--on-primary)}
.btn-primary:hover{box-shadow:var(--shadow-sm);filter:brightness(1.05)}
.btn-secondary{background:var(--secondary-container);color:var(--on-surface)}
.btn-secondary:hover{filter:brightness(1.05)}
.btn-danger{background:var(--error-container);color:var(--error)}
.btn-ghost{background:transparent;color:var(--on-surface-variant)}
.btn-ghost:hover{background:var(--surface-bright)}
.btn-sm{height:36px;padding:0 16px;font-size:13px}
.btn-icon{width:40px;padding:0;border-radius:50%}
.btn-icon.btn-sm{width:36px;height:36px;min-width:36px}

/* SERVER LIST */
.slist{list-style:none}
.sitem{
  display:flex;align-items:center;gap:12px;
  padding:12px;border-radius:var(--radius-sm);
  cursor:pointer;transition:all 0.15s;
  margin-bottom:2px;
}
.sitem:hover{background:var(--surface-container-high)}
.sitem.active{background:var(--primary-container);outline:1px solid color-mix(in srgb,var(--primary) 45%,transparent)}
.sicon{
  width:40px;height:40px;border-radius:var(--radius-sm);
  background:var(--surface-container-high);display:flex;align-items:center;justify-content:center;
  flex-shrink:0;color:var(--on-surface-variant);
}
.sitem.active .sicon{background:var(--primary);color:var(--on-primary)}
.sbody{flex:1;min-width:0}
.sname{font-size:15px;font-weight:500;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;letter-spacing:0}
.shost{font-size:12px;color:var(--on-surface-variant);margin-top:2px}
.sbadges{display:flex;gap:4px;margin-top:4px}
.badge{
  font-size:10px;font-weight:600;padding:2px 6px;border-radius:4px;
  text-transform:uppercase;letter-spacing:0.3px;
}
.badge-tls{background:var(--primary-container);color:var(--primary)}
.badge-reality{background:rgba(168,85,247,0.15);color:#a855f7}
.badge-none{background:var(--surface-dim);color:var(--on-surface-variant)}
.badge-active{background:var(--success-container);color:var(--success)}
.sping{
  font-family:var(--mono);font-size:11px;font-weight:500;
  padding:4px 8px;border-radius:6px;background:var(--surface-dim);
  cursor:pointer;transition:all 0.15s;white-space:nowrap;
  display:flex;align-items:center;gap:4px;flex-shrink:0;
}
.sping:hover{filter:brightness(1.1)}
.sping.loading{pointer-events:none;opacity:0.6}
.sping .material-symbols-outlined{font-size:14px}
.sping.green{color:var(--success);background:var(--success-container)}
.sping.amber{color:#f59e0b;background:rgba(245,158,11,0.12)}
.sping.red{color:var(--error);background:var(--error-container)}
.sping.none{color:var(--on-surface-variant)}
.sacts{display:flex;gap:4px;flex-shrink:0}

/* EMPTY */
.empty{
  display:flex;flex-direction:column;align-items:center;
  justify-content:center;padding:40px 0;color:var(--on-surface-variant);gap:8px;
}
.empty .material-symbols-outlined{font-size:48px;opacity:0.4}

/* TOGGLE */
.trow{
  display:flex;align-items:center;justify-content:space-between;
  padding:12px 0;border-bottom:1px solid var(--outline-variant);
}
.trow:last-child{border-bottom:none}
.tlbl{font-size:14px;font-weight:500}
.tdesc{font-size:12px;color:var(--on-surface-variant);margin-top:2px}
.toggle{position:relative;width:52px;height:32px;flex-shrink:0;cursor:pointer}
.toggle input{opacity:0;width:0;height:0}
.tslider{
  position:absolute;inset:0;
  background:var(--surface-dim);border-radius:16px;border:2px solid var(--outline);
  transition:all 0.2s;
}
.tslider::before{
  content:'';position:absolute;left:2px;top:2px;
  width:24px;height:24px;border-radius:50%;
  background:var(--on-surface-variant);transition:all 0.2s;
}
.toggle input:checked+.tslider{background:var(--primary);border-color:var(--primary)}
.toggle input:checked+.tslider::before{background:#fff;transform:translateX(20px)}

/* LOGS */
.logs-box{
  background:var(--surface-container-low);border-radius:var(--radius-sm);
  padding:12px;height:calc(100vh - 140px);overflow-y:auto;
  font-family:var(--mono);font-size:11px;line-height:1.7;color:var(--on-surface-variant);
}
.logs-box::-webkit-scrollbar{width:4px}
.logs-box::-webkit-scrollbar-thumb{background:var(--outline-variant);border-radius:2px}
.log-entry{padding:0}
.log-entry.ok{color:var(--success)}
.log-entry.err{color:var(--error)}
.log-entry.warn{color:#f59e0b}
.log-time{opacity:0.4;margin-right:8px}

/* MODAL */
.moverlay{
  position:fixed;inset:0;background:rgba(0,0,0,0.5);
  z-index:100;display:none;align-items:center;justify-content:center;padding:24px;
  backdrop-filter:blur(4px);
}
.moverlay.visible{display:flex}
.modal{
  background:var(--surface-container);border:1px solid var(--outline-variant);
  border-radius:var(--radius-lg);padding:24px;width:100%;max-width:480px;
  max-height:80vh;overflow-y:auto;
}
.modal h2{font-size:20px;font-weight:500;margin-bottom:16px}
.mclose{
  float:right;background:transparent;border:none;color:var(--on-surface-variant);cursor:pointer;padding:4px;
}
.mclose:hover{color:var(--on-surface)}
.igrid{display:grid;gap:8px}
.iitem{display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid var(--outline-variant);font-size:13px}
.iitem.full{flex-direction:column;gap:4px}
.ilbl{color:var(--on-surface-variant)}
.ival{color:var(--on-surface);font-weight:500;text-align:right;word-break:break-all}
.iitem.full .ival{text-align:left}
.mfooter{display:flex;justify-content:flex-end;gap:8px;margin-top:16px}

/* TOAST */
.toast{
  position:fixed;bottom:140px;left:50%;transform:translateX(-50%);
  background:var(--surface-container-high);color:var(--on-surface);border:1px solid var(--outline-variant);
  padding:12px 24px;border-radius:var(--radius-lg);font-size:13px;font-weight:500;
  box-shadow:var(--shadow-lg);z-index:200;
  opacity:0;transition:all 0.3s;pointer-events:none;
}
.toast.visible{opacity:1}
.toast.success{border-color:var(--success);color:var(--success)}
.toast.error{border-color:var(--error);color:var(--error)}

/* ABOUT */
.aoverlay{
  position:fixed;inset:0;background:rgba(0,0,0,0.5);
  z-index:100;display:none;align-items:center;justify-content:center;
  backdrop-filter:blur(4px);
}
.aoverlay.visible{display:flex}
.acard{
  background:var(--surface-container);border:1px solid var(--outline-variant);
  border-radius:var(--radius-lg);padding:32px;text-align:center;max-width:360px;
}
.acard h2{font-size:22px;font-weight:500;margin:12px 0 4px}
.acard .ver{font-size:13px;color:var(--on-surface-variant);margin-bottom:12px}
.acard p{font-size:14px;color:var(--on-surface-variant);margin-bottom:16px;line-height:1.5}
@media (max-width:720px){
  .app{flex-direction:column}
  .sidebar{width:100%;min-width:0;height:76px;flex-direction:row;padding:8px 10px;border-right:none;border-bottom:1px solid var(--outline-variant)}
  .sidebar-brand{width:44px;height:44px;margin:0 8px 0 0;border-radius:14px}
  .sidebar-btn{width:64px;min-height:56px;font-size:10px;border-radius:18px}
  .sidebar-spacer{display:none}
  .sidebar-bottom{flex-direction:row;margin-left:auto;margin-top:0;padding:0}
  .topbar{padding:16px 16px 6px}
  .scroll{padding:8px 16px 96px;max-width:none}
  .input-group{flex-direction:column}
  .sitem{align-items:flex-start}
  .sacts{flex-direction:column}
  .sping{display:none}
}

/* WORLD MAP STYLES */
.world-map{
  position:fixed;
  top:0;
  left:0;
  width:100vw;
  height:100vh;
  z-index:0;
  pointer-events:none;
  opacity:0.25;
  filter:blur(0.8px);
  transition:opacity 0.3s ease, filter 0.3s ease;
  overflow:visible;
}

.world-map.connected{
  opacity:1;
  filter:blur(0);
}

.world-map .map-inner{
  transition:transform 1.5s cubic-bezier(0.22,1,0.36,1);
  transform-origin:center center;
}

.world-map svg{
  width:100%;
  height:100%;
  overflow:visible;
}

.country{
  stroke:rgba(255,255,255,0.08);
  stroke-width:0.3;
  fill:rgba(220,225,235,0.12);
  transition:fill 0.5s ease, stroke 0.5s ease;
}

.country.highlighted{
  fill:#7C4DFF !important;
  stroke:rgba(124,77,255,0.5) !important;
  stroke-width:0.6;
  filter:drop-shadow(0 0 20px rgba(124,77,255,0.5));
  animation:country-pulse 3s ease-in-out infinite;
}

@keyframes country-pulse{
  0%,100%{opacity:1;filter:drop-shadow(0 0 20px rgba(124,77,255,0.5))}
  50%{opacity:0.85;filter:drop-shadow(0 0 30px rgba(124,77,255,0.7))}
}

[data-theme="light"] .world-map{
  opacity:0.18;
}

/* LIGHT THEME OVERRIDES */
[data-theme="light"] .country{
  stroke:rgba(50,60,80,0.08);
  fill:rgba(50,60,80,0.12);
}
[data-theme="light"] .country.highlighted{
  fill:#5A35D6 !important;
  stroke:rgba(90,53,214,0.5) !important;
  filter:drop-shadow(0 0 20px rgba(90,53,214,0.4));
}
/* COUNTRY TITLE */
.country-title{
  position:fixed;
  right:10%;
  bottom:120px;
  z-index:10;
  text-align:right;
  opacity:0;
  transition:opacity 1s ease 1.5s;
  pointer-events:none;
}

.country-title.visible{
  opacity:1;
}

.country-name{
  font-size:56px;
  font-weight:800;
  letter-spacing:4px;
  color:rgba(255,255,255,0.92);
  text-shadow:0 2px 12px rgba(0,0,0,0.3);
  white-space:nowrap;
}

.country-coords{
  margin-top:4px;
  font-size:16px;
  opacity:0.5;
  color:rgba(255,255,255,0.7);
  font-weight:400;
}

[data-theme="light"] .country-name{
  color:rgba(0,0,0,0.92);
}

[data-theme="light"] .country-coords{
  color:rgba(0,0,0,0.7);
}
</style>
</head>
<body>

<!-- WORLD MAP BACKGROUND -->
<div class="world-map" id="worldMap">
  <svg class="map-inner" viewBox="0 0 800 400" preserveAspectRatio="xMidYMid slice" id="mapInner">
    <g id="mapCountries"></g>
  </svg>
</div>

<!-- COUNTRY TITLE -->
<div class="country-title" id="countryTitle">
  <div class="country-name" id="countryName"></div>
  <div class="country-coords" id="countryCoords"></div>
</div>

<div class="app">

<!-- SIDEBAR -->
<nav class="sidebar">
  <div class="sidebar-brand"><img src="/assets/logo.png" alt="Stride VPN"></div>
  <button class="sidebar-btn active" data-screen="servers">
    <span class="material-symbols-outlined">dns</span>
    Servers
  </button>
  <button class="sidebar-btn" data-screen="settings">
    <span class="material-symbols-outlined">settings</span>
    Settings
  </button>
  <button class="sidebar-btn" data-screen="logs">
    <span class="material-symbols-outlined">terminal</span>
    Logs
  </button>
  <div class="sidebar-spacer"></div>
  <div class="sidebar-bottom">
    <button class="sidebar-btn" id="themeBtn" title="Toggle theme">
      <span class="material-symbols-outlined" id="themeIcon">dark_mode</span>
      Theme
    </button>
    <div class="sidebar-dot" id="statusDot"></div>
  </div>
</nav>

<!-- MAIN -->
<div class="main">

<!-- SERVERS -->
<section id="screen-servers" class="screen active">
  <div class="topbar">
    <h1>Servers</h1>
    <div class="flex" style="align-items:center;gap:8px">
      <span style="font-size:12px;color:var(--on-surface-variant)" id="configCount"></span>
      <button class="btn btn-ghost btn-icon btn-sm" id="aboutBtn" title="About"><span class="material-symbols-outlined">info</span></button>
    </div>
  </div>
  <div class="scroll">

    <!-- STATUS -->
    <div class="status-bar">
      <div class="status-led" id="statusLed"></div>
      <span class="status-text" id="statusLabel">Disconnected</span>
      <span class="status-sec" id="statusTimer"></span>
    </div>

    <!-- POWER -->
    <div class="power-wrap">
      <button class="power-btn" id="powerBtn">
        <div class="power-ring"></div>
        <span class="material-symbols-outlined">power_settings_new</span>
      </button>
    </div>

    <!-- IMPORT -->
    <div class="card">
      <div class="card-title">Import Server</div>
      <div class="input-group">
        <input class="input" id="importInput" placeholder="vless:// link or subscription URL">
        <button class="btn btn-primary btn-sm" id="importBtn">Add</button>
      </div>
      <div class="flex mt-8" style="gap:8px;margin-top:8px">
        <button class="btn btn-secondary btn-sm" id="subImportBtn">
          <span class="material-symbols-outlined" style="font-size:18px">link</span> Subscription
        </button>
      </div>
    </div>

    <!-- SERVER LIST -->
    <div class="card" style="padding-bottom:8px">
      <div class="card-title">Servers</div>
      <ul class="slist" id="serverList"></ul>
      <div class="empty" id="emptyServers">
        <span class="material-symbols-outlined">dns_off</span>
        <p>No servers configured</p>
        <p style="font-size:12px">Paste a vless:// link or subscription URL above</p>
      </div>
    </div>

  </div>
</section>

<!-- SETTINGS -->
<section id="screen-settings" class="screen">
  <div class="topbar"><h1>Settings</h1></div>
  <div class="scroll">
    <div class="card">
      <div class="card-title">Connection</div>
      <div class="trow">
        <div><div class="tlbl">TUN Mode</div><div class="tdesc">Full traffic interception</div></div>
        <label class="toggle"><input type="checkbox" id="tunToggle"><span class="tslider"></span></label>
      </div>
      <div class="trow">
        <div><div class="tlbl">System Proxy</div><div class="tdesc">Auto-configure Windows proxy</div></div>
        <label class="toggle"><input type="checkbox" id="sysProxyToggle"><span class="tslider"></span></label>
      </div>
    </div>
    <div class="card">
      <div class="card-title">Network</div>
      <div class="trow">
        <div><div class="tlbl">SOCKS5 Port</div></div>
        <input class="input" id="socksPortInput" type="number" value="1080" style="width:80px;text-align:center">
      </div>
      <div class="trow">
        <div><div class="tlbl">DNS Servers</div></div>
        <input class="input" id="dnsInput" type="text" value="8.8.8.8,1.1.1.1" style="width:180px">
      </div>
    </div>
    <button class="btn btn-primary" id="saveSettingsBtn" style="width:100%">Save Settings</button>
    <div style="margin-top:32px;text-align:center;padding:16px">
      <div style="font-size:12px;color:var(--on-surface-variant)">Stride VPN v1.0.0</div>
    </div>
  </div>
</section>

<!-- LOGS -->
<section id="screen-logs" class="screen">
  <div class="topbar">
    <h1>Logs</h1>
    <button class="btn btn-ghost btn-icon btn-sm" id="clearLogsBtn"><span class="material-symbols-outlined">delete_sweep</span></button>
  </div>
  <div class="scroll" style="padding-bottom:16px">
    <div class="logs-box" id="logsBox"></div>
  </div>
</section>

</div>
</div>

<!-- MODALS -->
<div class="moverlay" id="subModal">
  <div class="modal">
    <button class="mclose" data-close="subModal"><span class="material-symbols-outlined">close</span></button>
    <h2>Import Subscription</h2>
    <input class="input" id="subUrlInput" placeholder="https://example.com/sub">
    <div class="mfooter">
      <button class="btn btn-ghost" data-close="subModal">Cancel</button>
      <button class="btn btn-primary" id="subConfirmBtn">Import</button>
    </div>
  </div>
</div>

<div class="moverlay" id="cfgModal">
  <div class="modal">
    <button class="mclose" data-close="cfgModal"><span class="material-symbols-outlined">close</span></button>
    <h2 id="cfgModalTitle">Config</h2>
    <div class="igrid" id="cfgModalBody"></div>
    <div class="mfooter">
      <button class="btn btn-ghost" data-close="cfgModal">Close</button>
      <button class="btn btn-primary" id="cfgConnectBtn">Connect</button>
    </div>
  </div>
</div>

<div class="aoverlay" id="aboutOverlay">
  <div class="acard">
    <img src="/assets/logo.png" alt="Stride VPN" style="width:72px;height:72px;border-radius:18px;object-fit:cover;display:block;margin:0 auto;box-shadow:var(--shadow-md)">
    <h2>Stride VPN</h2>
    <div class="ver">v1.0.0</div>
    <p>VPN client powered by xray-core<br>VLESS + REALITY + TUN</p>
    <button class="btn btn-primary" id="aboutCloseBtn">Close</button>
  </div>
</div>

<div class="toast" id="toast"></div>

<script>
const WS=location.protocol==='https:'?'wss:':'ws:';
let ws=null,state='disconnected',activeId=null,timerInterval=null,timerSec=0,pingCtrl=null;
let wsBackoff=1000,wsReconnectTimer=null,isClosing=false;

function debounce(fn,ms){
  let t=null;
  return function(...args){if(t)clearTimeout(t);t=setTimeout(()=>{t=null;fn.apply(this,args)},ms)};
}

function wsConnect(){
  ws=new WebSocket(WS+'//'+location.host+'/ws');
  ws.onopen=()=>{wsBackoff=1000};
  ws.onmessage=e=>{const m=JSON.parse(e.data);if(m.type==='log')debouncedAddLog(m.data,m.tag);else if(m.type==='state')debouncedUpdateState(m.data,m.mode)};
  ws.onclose=()=>{
    if(isClosing)return;
    if(wsReconnectTimer)clearTimeout(wsReconnectTimer);
    wsReconnectTimer=setTimeout(wsConnect,wsBackoff);
    wsBackoff=Math.min(wsBackoff*2,10000);
  };
  ws.onerror=()=>ws&&ws.close();
}
const debouncedAddLog=debounce(addLog,100);
const debouncedUpdateState=debounce(updateState,300);

// NAV
document.querySelectorAll('.sidebar-btn[data-screen]').forEach(b=>{
  b.addEventListener('click',()=>{
    document.querySelectorAll('.sidebar-btn.active').forEach(x=>x.classList.remove('active'));
    b.classList.add('active');
    document.querySelectorAll('.screen.active').forEach(x=>x.classList.remove('active'));
    document.getElementById('screen-'+b.dataset.screen).classList.add('active');
  });
});

// TOAST
function toast(m,t){const e=document.getElementById('toast');e.textContent=m;e.className='toast visible'+(t?' '+t:'');setTimeout(()=>e.className='toast',3000)}

// LOGS
const logsEl=document.getElementById('logsBox');
function addLog(msg,tag){
  const e=document.createElement('div');e.className='log-entry'+(tag?' '+tag:'');
  const m=msg.match(/^\[(.*?)\]\s*(.*)/);
  if(m){const s=document.createElement('span');s.className='log-time';s.textContent='['+m[1]+']';e.appendChild(s);e.appendChild(document.createTextNode(m[2]))}
  else e.textContent=msg;
  logsEl.appendChild(e);logsEl.scrollTop=logsEl.scrollHeight;
}
document.getElementById('clearLogsBtn').onclick=()=>logsEl.innerHTML='';

// STATE
function updateState(s,mode){
  state=s;
  const led=document.getElementById('statusLed'),lbl=document.getElementById('statusLabel');
  const btn=document.getElementById('powerBtn'),dot=document.getElementById('statusDot');
  led.className='status-led';dot.className='sidebar-dot';
  btn.className='power-btn';
  if(s==='connected'){
    led.classList.add('connected');dot.classList.add('online');
    lbl.textContent='Connected';btn.classList.add('connected');
    btn.querySelector('.material-symbols-outlined').textContent='power_settings_new';
    document.getElementById('statusTimer').textContent='';
    startTimer();
  }else if(s==='connecting'){
    led.classList.add('connecting');dot.classList.add('connecting');
    lbl.textContent='Connecting...';btn.classList.add('connecting');
    btn.querySelector('.material-symbols-outlined').textContent='sync';
    document.getElementById('statusTimer').textContent='';
  }else{
    lbl.textContent='Disconnected';
    btn.querySelector('.material-symbols-outlined').textContent='power_settings_new';
    document.getElementById('statusTimer').textContent='';
    stopTimer();
    if(s==='error'){led.classList.add('error')}
  }
  if(mode&&s==='connected'){
    document.getElementById('statusTimer').textContent=mode==='TUN'?'TUN • ':'Proxy • ';
  }
  refresh();
}

// TIMER
function startTimer(){
  timerSec=0;stopTimer();
  timerInterval=setInterval(()=>{
    timerSec++;
    const h=Math.floor(timerSec/3600),m=Math.floor((timerSec%3600)/60),s=timerSec%60;
    const ts=document.getElementById('statusTimer');
    const mode=ts.textContent.includes('TUN')?'TUN • ':'Proxy • ';
    ts.textContent=mode+String(h).padStart(2,'0')+':'+String(m).padStart(2,'0')+':'+String(s).padStart(2,'0');
  },1000);
}
function stopTimer(){if(timerInterval){clearInterval(timerInterval);timerInterval=null}}

// POWER
document.getElementById('powerBtn').onclick=()=>{
  if(state==='connected'){
    const a=document.querySelector('.sitem.active');
    if(a&&a.dataset.id)disc(a.dataset.id);
  }else if(state==='disconnected'||state==='error'){
    const a=document.querySelector('.sitem.active');
    if(a&&a.dataset.id)conn(a.dataset.id);
    else toast('Select a server first','error');
  }
};

// CONN/DISC
function conn(id){
  updateState('connecting');activeId=id;
  fetch('/api/v1/connect',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({id})})
    .then(r=>{if(!r.ok)return r.text().then(e=>{throw new Error(e)});refresh()})
    .catch(e=>{updateState('disconnected');toast('Connect: '+e.message,'error')});
}
function disc(id){
  if(pingCtrl){pingCtrl.abort();pingCtrl=null}
  updateState('disconnected');
  fetch('/api/v1/disconnect',{method:'POST'}).catch(()=>{});
}

// RENDER
function esc(t){const d=document.createElement('div');d.textContent=t;return d.innerHTML}

let cachedServers=null;
function renderServers(list){
  const ul=document.getElementById('serverList'),empty=document.getElementById('emptyServers');
  const cnt=document.getElementById('configCount');
  if(!list){
    if(!cachedServers)return;
    list=cachedServers;
  }
  cachedServers=list;
  ul.innerHTML='';
  if(list.length===0){empty.style.display='flex';cnt.textContent='';return}
  empty.style.display='none';cnt.textContent=list.length+' server'+(list.length>1?'s':'');
  list.forEach(c=>{
    const li=document.createElement('li');
    li.className='sitem'+(c.active?' active':'');li.dataset.id=c.id;
    const sec=(c.security||'none').toLowerCase();
    const b='<span class="badge badge-'+sec+'">'+(sec==='tls'?'TLS':sec==='reality'?'REALITY':sec)+'</span>'+(c.active?' <span class="badge badge-active">ACTIVE</span>':'');
    li.innerHTML=
      '<div class="sicon"><span class="material-symbols-outlined">dns</span></div>'+
      '<div class="sbody"><div class="sname">'+esc(c.remark||c.host)+'</div><div class="shost">'+c.host+':'+c.port+'</div><div class="sbadges">'+b+'</div></div>'+
      '<div class="sping none" data-id="'+c.id+'" title="Ping"><span class="material-symbols-outlined">network_check</span><span class="ping-val">—</span></div>'+
      '<div class="sacts">'+
        '<button class="btn btn-ghost btn-icon btn-sm info-btn"><span class="material-symbols-outlined">info</span></button>'+
        (c.active?'<button class="btn btn-danger btn-icon btn-sm disc-btn"><span class="material-symbols-outlined">power_off</span></button>'
        :'<button class="btn btn-primary btn-icon btn-sm conn-btn"><span class="material-symbols-outlined">play_arrow</span></button>')+
        '<button class="btn btn-ghost btn-icon btn-sm del-btn"><span class="material-symbols-outlined">delete</span></button>'+
      '</div>';
    ul.appendChild(li);
    const pe=li.querySelector('.sping');pe.onclick=e=>{e.stopPropagation();doPing(c.id,pe)};
    li.querySelector('.info-btn').onclick=e=>{e.stopPropagation();openInfo(c.id)};
    const cb=li.querySelector('.conn-btn');if(cb)cb.onclick=e=>{e.stopPropagation();conn(c.id)};
    const db=li.querySelector('.disc-btn');if(db)db.onclick=e=>{e.stopPropagation();disc(c.id)};
    li.querySelector('.del-btn').onclick=e=>{e.stopPropagation();doDel(c.id)};
    li.onclick=()=>openInfo(c.id);
  });
}

// PING
async function doPing(id,el){
  if(pingCtrl)pingCtrl.abort();
  const ctrl=new AbortController();pingCtrl=ctrl;
  el.classList.add('loading');el.className='sping none';el.querySelector('.ping-val').textContent='...';
  try{
    const r=await fetch('/api/v1/server/ping/'+encodeURIComponent(id),{signal:ctrl.signal});
    if(!r.ok)throw new Error('fail');
    const d=await r.json();
    if(d.status==='ok'&&d.latency!=null){
      const ms=d.latency;el.className='sping '+(ms<150?'green':ms<400?'amber':'red');
      el.querySelector('.ping-val').textContent=ms+'ms';
    }else throw new Error('bad');
  }catch(e){
    if(e.name==='AbortError')return;
    el.className='sping red';el.querySelector('.ping-val').textContent='ERR';
  }
  el.classList.remove('loading');pingCtrl=null;
}

// INFO
async function openInfo(id){
  try{
    const r=await fetch('/api/v1/config/'+encodeURIComponent(id));
    if(!r.ok){toast('Failed to load','error');return}
    const inf=await r.json();
    document.getElementById('cfgModalTitle').textContent=inf.remark||inf.host;
    const items=[
      {l:'UUID',v:inf.uuid,f:1},{l:'Server',v:inf.host+':'+inf.port},{l:'Security',v:inf.security},
      {l:'Encryption',v:inf.encryption||'none'},{l:'Flow',v:inf.flow||'—'},{l:'SNI',v:inf.sni},
      {l:'Fingerprint',v:inf.fingerprint||'—'},{l:'PublicKey',v:inf.publicKey||'—',f:1},{l:'ShortID',v:inf.shortId||'—'},
    ];
    document.getElementById('cfgModalBody').innerHTML=items.map(i=>
      '<div class="iitem'+(i.f?' full':'')+'"><span class="ilbl">'+i.l+'</span><span class="ival">'+esc(i.v)+'</span></div>'
    ).join('');
    document.getElementById('cfgConnectBtn').textContent=state==='connected'?'Switch':'Connect';
    document.getElementById('cfgConnectBtn').onclick=()=>{
      document.getElementById('cfgModal').classList.remove('visible');
      if(state==='connected')disc(inf.id);
      setTimeout(()=>conn(inf.id),300);
    };
    document.getElementById('cfgModal').classList.add('visible');
  }catch(e){toast('Error: '+e.message,'error')}
}

// DELETE
async function doDel(id){
  if(!confirm('Delete this server?'))return;
  try{
    const r=await fetch('/api/v1/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({id})});
    if(r.ok){toast('Deleted','success');refresh()}else{const e=await r.text();toast('Error: '+e,'error')}
  }catch(e){toast('Error: '+e.message,'error')}
}

// IMPORT
document.getElementById('importBtn').onclick=async()=>{
  const link=document.getElementById('importInput').value.trim();
  if(!link){toast('Enter a link','error');return}
  // Try subscription first, else single link
  try{
    const r=await fetch('/api/v1/import/subscription',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({url:link})});
    if(r.ok){const d=await r.json();toast('Imported '+d.count+' configs','success');document.getElementById('importInput').value='';refresh();return}
  }catch(e){}
  try{
    const r=await fetch('/api/v1/import',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({link})});
    if(r.ok){const d=await r.json();toast('Imported: '+(d.remark||d.host),'success');document.getElementById('importInput').value='';refresh()}
    else{const e=await r.text();toast('Error: '+e,'error')}
  }catch(e){toast('Error: '+e.message,'error')}
};
document.getElementById('importInput').onkeydown=e=>{if(e.key==='Enter')document.getElementById('importBtn').click()};
document.getElementById('subImportBtn').onclick=()=>{document.getElementById('subUrlInput').value='';document.getElementById('subModal').classList.add('visible')};
document.getElementById('subConfirmBtn').onclick=async()=>{
  const url=document.getElementById('subUrlInput').value.trim();
  if(!url){toast('Enter a URL','error');return}
  try{
    const r=await fetch('/api/v1/import/subscription',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({url})});
    if(r.ok){const d=await r.json();toast('Imported '+d.count+' configs','success');document.getElementById('subModal').classList.remove('visible');refresh()}
    else{const e=await r.text();toast('Error: '+e,'error')}
  }catch(e){toast('Error: '+e.message,'error')}
};

// SETTINGS
document.getElementById('saveSettingsBtn').onclick=async()=>{
  const s={
    socks_port:parseInt(document.getElementById('socksPortInput').value)||1080,
    dns_servers:document.getElementById('dnsInput').value.split(',').map(s=>s.trim()).filter(Boolean),
    auto_start:false,system_proxy:document.getElementById('sysProxyToggle').checked,
    tun_enabled:document.getElementById('tunToggle').checked,routing_mode:'custom'
  };
  try{
    const r=await fetch('/api/v1/settings/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(s)});
    if(r.ok)toast('Settings saved','success');else{const e=await r.text();toast('Error: '+e,'error')}
  }catch(e){toast('Error: '+e.message,'error')}
};
async function loadSettings(){
  try{
    const r=await fetch('/api/v1/settings');const s=await r.json();
    if(s.socks_port!=null)document.getElementById('socksPortInput').value=s.socks_port;
    if(s.dns_servers)document.getElementById('dnsInput').value=s.dns_servers.join(',');
    document.getElementById('sysProxyToggle').checked=!!s.system_proxy;
    document.getElementById('tunToggle').checked=!!s.tun_enabled;
  }catch(e){}
}

// REFRESH
async function refresh(){
  try{
    const r=await fetch('/api/v1/configs',{cache:'no-store'});
    if(!r.ok)throw new Error('configs failed');
    const l=await r.json();
    if(Array.isArray(l))renderServers(l);
  }catch(e){renderServers(null)}
}

// MODALS
document.querySelectorAll('[data-close]').forEach(e=>{e.addEventListener('click',()=>{document.getElementById(e.dataset.close).classList.remove('visible')})});
document.querySelectorAll('.moverlay').forEach(ov=>{ov.addEventListener('click',e=>{if(e.target===ov)ov.classList.remove('visible')})});
document.querySelectorAll('.aoverlay').forEach(ov=>{ov.addEventListener('click',e=>{if(e.target===ov)ov.classList.remove('visible')})});
document.onkeydown=e=>{
  if(e.key==='Escape'){document.querySelectorAll('.moverlay.visible,.aoverlay.visible').forEach(m=>m.classList.remove('visible'))}
};
window.addEventListener('pagehide',()=>{isClosing=true;if(ws){try{ws.close()}catch(e){}}});

// ABOUT
document.getElementById('aboutBtn').onclick=()=>document.getElementById('aboutOverlay').classList.add('visible');
document.getElementById('aboutCloseBtn').onclick=()=>document.getElementById('aboutOverlay').classList.remove('visible');

// THEME
const savedTheme=localStorage.getItem('theme')||'dark';
document.documentElement.setAttribute('data-theme',savedTheme);
document.getElementById('themeIcon').textContent=savedTheme==='dark'?'dark_mode':'light_mode';
document.getElementById('themeBtn').addEventListener('click',()=>{
  const cur=document.documentElement.getAttribute('data-theme');
  const next=cur==='dark'?'light':'dark';
  document.documentElement.setAttribute('data-theme',next);
  localStorage.setItem('theme',next);
  document.getElementById('themeIcon').textContent=next==='dark'?'dark_mode':'light_mode';
});

// ──────────────────────────────────────────────────
// WORLD MAP
// ──────────────────────────────────────────────────
// Convert lat/lon to SVG coords: x=(lon+180)/360*800, y=(90-lat)/180*400
function ll(lon,lat){return[(lon+180)/360*800,(90-lat)/180*400]}
const countryCoords={
  US:ll(-100,39),CA:ll(-100,55),GB:ll(-3,55),DE:ll(10,51),FR:ll(2,47),
  NL:ll(5,52),SE:ll(15,60),NO:ll(10,62),DK:ll(10,56),FI:ll(26,62),
  CH:ll(8,47),AT:ll(14,48),IT:ll(12,42),ES:ll(-4,40),PT:ll(-9,40),
  BE:ll(4,51),IE:ll(-8,53),PL:ll(20,52),CZ:ll(15,50),SK:ll(19,49),
  HU:ll(19,47),RO:ll(25,46),BG:ll(25,43),GR:ll(22,39),HR:ll(15,45),
  RS:ll(21,44),RU:ll(40,60),UA:ll(31,49),TR:ll(35,39),
  JP:ll(138,36),KR:ll(128,37),HK:ll(114,22),SG:ll(104,1),TW:ll(120,24),
  IN:ll(77,20),CN:ll(105,35),AU:ll(134,-25),NZ:ll(174,-41),ID:ll(115,-5),
  MY:ll(102,4),TH:ll(101,15),VN:ll(107,16),PH:ll(122,12),AE:ll(55,24),
  IL:ll(35,31),ZA:ll(26,-30),AR:ll(-64,-36),BR:ll(-52,-14),CL:ll(-70,-30),
  CO:ll(-74,4),MX:ll(-102,23),PE:ll(-76,-10),VE:ll(-66,8),EG:ll(30,27),
  NG:ll(8,8),KE:ll(38,0),MA:ll(-7,32)
};

const countryNames={
  US:'United States',CA:'Canada',GB:'United Kingdom',DE:'Germany',
  FR:'France',NL:'Netherlands',SE:'Sweden',NO:'Norway',DK:'Denmark',
  FI:'Finland',CH:'Switzerland',AT:'Austria',IT:'Italy',ES:'Spain',
  PT:'Portugal',BE:'Belgium',IE:'Ireland',PL:'Poland',CZ:'Czech Republic',
  SK:'Slovakia',HU:'Hungary',RO:'Romania',BG:'Bulgaria',GR:'Greece',
  HR:'Croatia',RS:'Serbia',RU:'Russia',UA:'Ukraine',
  TR:'Turkey',JP:'Japan',KR:'South Korea',HK:'Hong Kong',SG:'Singapore',
  TW:'Taiwan',IN:'India',CN:'China',AU:'Australia',NZ:'New Zealand',
  ID:'Indonesia',MY:'Malaysia',TH:'Thailand',VN:'Vietnam',PH:'Philippines',
  AE:'UAE',IL:'Israel',ZA:'South Africa',AR:'Argentina',BR:'Brazil',
  CL:'Chile',CO:'Colombia',MX:'Mexico',PE:'Peru',VE:'Venezuela',
  EG:'Egypt',NG:'Nigeria',KE:'Kenya',MA:'Morocco'
};

// Country → continent mapping for highlight
const countryContinent={
  US:'NA',CA:'NA',MX:'NA',
  AR:'SA',BR:'SA',CL:'SA',CO:'SA',PE:'SA',VE:'SA',
  GB:'EU',DE:'EU',FR:'EU',NL:'EU',SE:'EU',NO:'EU',DK:'EU',FI:'EU',
  CH:'EU',AT:'EU',IT:'EU',ES:'EU',PT:'EU',BE:'EU',IE:'EU',
  PL:'EU',CZ:'EU',SK:'EU',HU:'EU',RO:'EU',BG:'EU',GR:'EU',HR:'EU',RS:'EU',
  RU:'EU',UA:'EU',TR:'EU',
  ZA:'AF',EG:'AF',NG:'AF',KE:'AF',MA:'AF',
  JP:'AS',KR:'AS',HK:'AS',SG:'AS',TW:'AS',IN:'AS',CN:'AS',
  TH:'AS',VN:'AS',ID:'AS',MY:'AS',PH:'AS',AE:'AS',IL:'AS',
  AU:'OC',NZ:'OC'
};

let mapCountries={};

async function loadMapData(){
  try{
    const r=await fetch('/api/v1/map-data');
    if(!r.ok)return;
    const data=await r.json();
    const g=document.getElementById('mapCountries');
    g.innerHTML='';
    data.countries.forEach(c=>{
      if(!c.path)return;
      const p=document.createElementNS('http://www.w3.org/2000/svg','path');
      p.setAttribute('d',c.path);
      p.setAttribute('class','country other');
      p.dataset.country=c.name;
      g.appendChild(p);
      mapCountries[c.name]=p;
    });
  }catch(e){}
}

let connectedCountry='';
let isAnimating=false;

function updateMapState(newState,countryCode){
  const map=document.getElementById('worldMap');
  const inner=document.getElementById('mapInner');
  const title=document.getElementById('countryTitle');
  const nameEl=document.getElementById('countryName');
  const coordsEl=document.getElementById('countryCoords');

  if(!map||!inner)return;

  if(newState==='connected'&&countryCode){
    connectedCountry=countryCode;

    // Calculate zoom transform
    const coords=countryCoords[countryCode]||[400,200];
    const mapW=800,mapH=400;
    const targetX=coords[0],targetY=coords[1];

    // We want the country on the right side: country at ~65% of viewport
    const viewW=window.innerWidth,viewH=window.innerHeight;
    const targetViewX=viewW*0.65;
    const targetViewY=viewH*0.45;

    // Scale to make country ~35% of screen width
    const screenSz=Math.min(viewW,viewH)*0.35;
    const scale=screenSz/120;

    // Calculate translate
    const cx=mapW/2,cy=mapH/2;
    const tx=targetViewX-scale*targetX;
    const ty=targetViewY-scale*targetY;

    // Animate
    isAnimating=true;
    map.classList.add('connected');
    inner.style.transform='translate('+tx+'px,'+ty+'px) scale('+scale+')';
    map.style.opacity='1';
    map.style.filter='blur(0)';

    // Highlight continent
    const continent=countryContinent[countryCode]||'';
    document.querySelectorAll('.country').forEach(p=>{
      p.classList.remove('highlighted');
      if(continent&&p.dataset.country===continent){
        p.classList.add('highlighted');
      }else{
        p.classList.add('other');
      }
    });

    // Show title after animation
    setTimeout(()=>{
      if(title){
        nameEl.textContent=countryNames[countryCode]||countryCode;
        coordsEl.textContent='VPN Connected';
        title.classList.add('visible');
      }
      isAnimating=false;
    },1800);
  }else{
    // Disconnected: reset to world view
    connectedCountry='';
    map.classList.remove('connected');
    inner.style.transform='translate(0,0) scale(1)';
    map.style.opacity='0.2';
    map.style.filter='blur(1px)';

    document.querySelectorAll('.country').forEach(p=>{
      p.classList.remove('highlighted');
      p.classList.add('other');
    });

    if(title)title.classList.remove('visible');
  }
}

// Patch state update to trigger map animation
const origUpdateState=updateState;
updateState=function(newState,msg){
  origUpdateState(newState,msg);
  const stateMap={'connected':'connected','connecting':'connecting','disconnected':'disconnected','error':'disconnected','no_network':'disconnected'};
  const mapped=stateMap[newState]||'disconnected';

  let country='';
  if(mapped==='connected'&&cachedServers&&cachedServers.length>0){
    const active=cachedServers.find(s=>s.id===activeId);
    if(active&&active.host){
      // Extract country code from remark or host
      country=active.remark||active.host;
    }
    // Simple country code detection from hostname/remark
    if(!country||!countryCoords[country]){
      const host=active?active.host:'';
      country=guessCountry(host,active?active.remark:'');
    }
  }
  updateMapState(mapped,country);
};

function guessCountry(host,remark){
  const text=(host+' '+remark).toLowerCase();
  // Common country indicators in hostnames
  const map={
    'united states':'US','united states of america':'US','usa':'US','new york':'US','los angeles':'US','chicago':'US','miami':'US','dallas':'US',
    'germany':'DE','deutschland':'DE','frankfurt':'DE','berlin':'DE',
    'united kingdom':'GB','uk':'GB','london':'GB','england':'GB','manchester':'GB',
    'france':'FR','paris':'FR','marseille':'FR',
    'netherlands':'NL','holland':'NL','amsterdam':'NL',
    'japan':'JP','tokyo':'JP','osaka':'JP',
    'singapore':'SG','sg':'SG',
    'australia':'AU','sydney':'AU','melbourne':'AU',
    'canada':'CA','toronto':'CA','vancouver':'CA','montreal':'CA',
    'switzerland':'CH','zurich':'CH',
    'sweden':'SE','stockholm':'SE',
    'norway':'NO','oslo':'NO',
    'denmark':'DK','copenhagen':'DK',
    'finland':'FI','helsinki':'FI',
    'italy':'IT','milan':'IT','rome':'IT',
    'spain':'ES','madrid':'ES','barcelona':'ES',
    'portugal':'PT','lisbon':'PT',
    'austria':'AT','vienna':'AT',
    'poland':'PL','warsaw':'PL',
    'czech':'CZ','prague':'CZ',
    'brazil':'BR','sao paulo':'BR','riodejaneiro':'BR',
    'mexico':'MX','mexico city':'MX',
    'india':'IN','mumbai':'IN','bangalore':'IN',
    'south korea':'KR','korea':'KR','seoul':'KR',
    'hong kong':'HK','hk':'HK',
    'taiwan':'TW','taipei':'TW','taichung':'TW',
    'turkey':'TR','istanbul':'TR',
    'russia':'RU','moscow':'RU',
    'south africa':'ZA','johannesburg':'ZA','capetown':'ZA',
    'argentina':'AR','buenos aires':'AR',
    'chile':'CL','santiago':'CL',
    'colombia':'CO','bogota':'CO','medellin':'CO',
    'peru':'PE','lima':'PE',
    'uae':'AE','dubai':'AE','abudhabi':'AE',
    'israel':'IL','telaviv':'IL'
  };
  for(const[key,code]of Object.entries(map)){
    if(text.includes(key))return code;
  }
  // Try to match 2-letter country code (e.g., "🇩🇪 DE" or "-DE-")
  const m=text.match(/\b([a-z]{2})\b/g);
  if(m){
    for(const c of m){
      if(c==='us')continue; // skip 'us' (too many false positives)
      const code=c.toUpperCase();
      if(countryCoords[code])return code;
    }
  }
  return '';
}

// INIT
loadMapData();
wsConnect();loadSettings();refresh();setInterval(refresh,5000);
</script>
</body>
</html>`
