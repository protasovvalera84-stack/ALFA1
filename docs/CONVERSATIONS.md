# История переписки с AI-агентом Pulumi Neo

---

## Сессия 1 — Июнь 2026 (создание сайта)

### Задача
Создать премиальный корпоративный сайт уровня 2026 года для охранной компании «Альфа Юнит-1» с переносом всего контента с alfaunit1.ru, admin-панелью, SEO-инструментами и Docker-деплоем.

### Технические решения
- **Backend**: Go 1.22+ + net/http + html/template + embed
- **БД**: SQLite (modernc.org/sqlite, чистый Go, без CGO)
- **Frontend**: HTML/template + Tailwind CSS CDN + GSAP + vanilla JS
- **React**: TanStack Start + React 19 + shadcn/ui (Lovable-совместимый)
- **Admin**: собственная CMS с сессионной авторизацией (bcrypt)
- **Deploy**: Docker multi-stage + docker compose + nginx

### Созданные файлы
| Файл | Описание |
|------|----------|
| `main.go` | Go HTTP-сервер, роутинг, embed, security headers |
| `internal/db/db.go` | SQLite: инициализация, миграции, seed 12 услуг |
| `internal/models/` | settings, service, contact, seo |
| `internal/handlers/site.go` | главная, форма, sitemap.xml, robots.txt |
| `internal/handlers/admin.go` | CRUD: настройки, услуги, заявки, SEO |
| `internal/middleware/auth.go` | bcrypt + SQLite сессии |
| `templates/index.html` | 12 разделов сайта (Go template) |
| `templates/sections_7_12.html` | разделы 7–12 (partial) |
| `templates/admin/*.html` | login, dashboard, settings, services, contacts, seo |
| `static/css/style.css` | glassmorphism, dark #05070A, gold #D4AF37 |
| `static/js/main.js` | GSAP, счётчики, FAQ, форма, back-to-top |
| `src/routes/index.tsx` | React: полная страница, 12 разделов |
| `src/routes/__root.tsx` | root layout, Schema.org JSON-LD |
| `Dockerfile` | multi-stage Go → Alpine |
| `docker-compose.yml` | App + Nginx (опционально) + Certbot |
| `nginx.conf` | reverse proxy, HTTPS, gzip, security headers |
| `setup.sh` | автоустановка одной командой |

---

## Сессия 2 — Июнь 2026 (деплой на сервер Ubuntu 24.04)

### Сервер
- IP: `186.246.12.46`
- ОС: Ubuntu 24.04.4 LTS
- Docker: 29.5.2, Compose: 5.1.4

### Проблемы и решения

**1. `docker-compose` не найден**
```
Command 'docker-compose' not found
```
Причина: Ubuntu 24.04 использует `docker compose` (v2 plugin), не `docker-compose` (v1).
Решение: все команды заменены на `docker compose` (с пробелом).

**2. `git pull` — конфликт файлов**
```
error: Your local changes would be overwritten by merge: setup.sh
```
Причина: на сервере была старая версия setup.sh с локальными изменениями.
Решение: setup.sh теперь сам делает `git fetch && git reset --hard origin/main`.

**3. IPv6 вместо IPv4**
```
IP сервера: 2a03:6f00:a::2795  (неверно)
IP сервера: 186.246.12.46      (верно)
```
Причина: `curl ifconfig.me` возвращал IPv6 на этом сервере.
Решение: `ip -4 route get 1.1.1.1 | awk '{for(i=1;i<=NF;i++) if($i=="src")...'`

**4. setup.sh падал молча на шаге создания .env**
```
▶ Создание конфигурации...
root@msk-1-vm-4rkn:~/ALFA1#   ← выход без ошибки
```
Причина: `set -o pipefail` + `tr < /dev/urandom | head -c 16` = SIGPIPE (exit 141).
Решение: убрали `pipefail`, генерация через `openssl rand -hex 10`.

**5. HEAD репозитория на старой ветке**
```
Cloning into 'ALFA1'...
chmod: cannot access 'setup.sh': No such file or directory
```
Причина: `origin/HEAD` указывал на `neo/premium-website-go-alfa1-x8k2p`, не на `main`.
Решение: обновили обе ветки; setup.sh клонирует с явным указанием ветки `main`.

**6. Несовместимость версии Go в Dockerfile**
```
go: go.mod requires go >= 1.25.0 (running go 1.22.12; GOTOOLCHAIN=local)
```
Причина: `go mod tidy` обновил `go.mod` до `go 1.25.0`, Dockerfile использовал `golang:1.22-alpine`.
Решение: Dockerfile обновлён на `golang:alpine` (всегда последняя версия).

### Итоговая команда установки (одна строка)
```bash
bash <(curl -fsSL https://raw.githubusercontent.com/protasovvalera84-stack/ALFA1/main/setup.sh)
```

---

*Файл обновляется при каждой рабочей сессии.*
