<img width="323" height="350" alt="stride logo" src="https://github.com/user-attachments/assets/6a39a824-c889-46d0-9d6f-4f9548f99165" />


# Stride VPN

**Windows VPN-клиент на базе xray-core.** Один exe-файл, веб-интерфейс вместо сложных настроек: вставил ссылку — подключился.

| | |
|---|---|
| Версия | v1.0.2 |
| Платформы | Windows (основная), Linux (порт) |
| Язык | Go 1.26, xray-core |
| Лицензия | см. раздел «Почему это сделано» |

---

## Возможности

- 🔌 **Два режима подключения** — TUN (весь трафик системы) и прокси (SOCKS5 + HTTP на `127.0.0.1`).
- 📥 **Импорт по подписке** — вставьте URL — внутренние `vless://` ссылки разберутся автоматически, обновление раз в 60 секунд.
- 🔗 **Импорт ссылок** — поддерживаются как целая подписка, так и одиночные `vless://`.
- 🗂 **Группировка серверов по подпискам** — каждый источник подписок отображается отдельной секцией.
- ⚡ **Пинг серверов** — задержка до каждой точки одним кликом или кнопкой «Ping all».
- 🛡 **TLS-fingerprint** — настройка User Agent (chrome / firefox / safari / edge) влияет на отпечаток руки utls.
- 🌐 **DNS** — свои Primary/Secondary DNS (по умолчанию 8.8.8.8 / 1.0.0.1).
- 🚦 **Маршрутизация** — переключатель «использовать маршрутизацию по умолчанию»: выкл — весь трафик через прокси.
- 🎨 **Тёмная и светлая темы** — переключаются без перезапуска.
- 🔄 **Автопереподключение** — при потере соединения клиент следит за сетью и переподключается с экспоненциальной задержкой (1–30 с).

---

## Быстрый старт

1. Скачайте `stride-vpn.exe` (одним файлом).
2. Запустите — откроется окно приложения с веб-интерфейсом (`localhost:8080`).
3. Вставьте ссылку на подписку или `vless://` в поле импорта.
4. Нажмите на сервер → «Подключиться».

> При первом подключении через TUN Windows запросит права администратора (UAC). Один раз за запуск — при переподключениях окно не появляется.

---

## Как это работает

Клиент поднимает локальный веб-сервер (по умолчанию `0.0.0.0:8080`) и открывает его в отдельном окне браузера (Chrome/Edge в режиме `--app`). Вся логика — в одном бинарнике, UI встроен в него: внешних ресурсов и установки не требуется.

### Режимы

| Режим | Как работает | Права |
|---|---|---|
| **TUN** | Виртуальный сетевой адаптер перехватывает весь трафик системы | Нужен администратор (UAC) |
| **Прокси** | SOCKS5 и HTTP-прокси на `127.0.0.1` (порт выбирается свободный, обычно 10808), системный прокси настраивается автоматически | Не нужны |

### Поддерживаемые транспорты

WebSocket (`ws`), XHTTP (`xhttp`, бывший splithttp), gRPC (`grpc`), mKCP (`kcp`), TCP, HTTP/2 (`h2`), QUIC.

Защита соединения — `tls`, `reality`, либо без защиты (`none`). REALITY работает через [utls](https://github.com/refraction-networking/utls).

### Конфигурация

Всё хранится рядом с exe-файлом:
- `configs.json` — импортированные серверы;
- `config/settings.json` — настройки приложения (DNS, fingerprint, тема, порты).

---

## Что в новой версии (v1.0.2)

- Новые настройки **DNS** (Primary/Secondary раздельно) и **Routing Rules**:
  - переключатель «Use routing by default» (по умолчанию выключен — весь трафик идёт через прокси);
  - поле **User Agent** — задаёт TLS-отпечаток (по умолчанию `chrome(windows)`).
- **Группировка серверов** по подписке, из которой они импортированы.
- Кнопка **«Ping all»** — пинг всех серверов разом.
- Убран скриншот-карта мира — интерфейс стал проще.

<img width="903" height="718" alt="Снимок экрана 2026-08-11 070925" src="https://github.com/user-attachments/assets/d5182d71-63f1-4c98-a482-8839c71816e6" />

---



## Сборка из исходников

Требуется **Go 1.26+**. Все зависимости (включая wintun) вшиваются в бинарник.

### Windows

```ps
go build -ldflags="-s -w -H windowsgui" -o stride-vpn.exe ./cmd
```

> `-H windowsgui` скрывает окно консоли. Без него приложение работает так же, но запускается с терминалом.

### Linux (кросс-сборка)

Порт из репозитория собирается из того же кода:

```sh
GOOS=linux GOARCH=amd64 go build -o stride-vpn ./cmd
```

---

## Linux-порт

Тот же код, собранный под Linux. Отличия по платформе:

- **Интерфейс** — открывается через `xdg-open` (тоже в режиме приложения, если есть Chrome/Chromium/Firefox).
- **TUN** — настройка через `ip tuntap` + `ip route`, права запрашиваются через `pkexec` (или `sudo`).
- **DNS** — применяется через `resolvectl`.
- **Системный прокси** — настраивается через GNOME `gsettings`.

---

## Структура проекта

```
cmd/                    точка входа, запуск браузерного окна
internal/
  client/               менеджер подключений, настройки, импорт, healthcheck
  vless/                парсер vless:// и клиентская реализация протокола
  xray/                 генерация конфига xray, TUN, wintun
  web/                  веб-сервер, REST/WebSocket API, интерфейс
  proxy/                системный прокси (Windows / Linux)
  storage/              хранение конфигов и настроек
```

---

## Почему это сделано

Существующие клиенты для xray-core перегружены: десятки кнопок, графиков и настроек, в которых теряется обычный пользователь. Здесь всё наоборот — скачал файл, запустил, вставил ссылку, подключился. Остальное спрятано в настройки и работает по умолчанию разумно.

Если что-то хочется изменить — код открыт, правите сами.

## MIT License  

Copyright (c) 2026 <str3l0k(fokg/56)>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
