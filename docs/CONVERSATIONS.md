# История переписки с AI-агентом Pulumi Neo

---

## Сессия 1 — Июнь 2026 (создание сайта)

### Задача
Создать премиальный корпоративный сайт уровня 2026 года для охранной компании «Альфа Юнит-1» с переносом всего контента с alfaunit1.ru, admin-панелью, SEO-инструментами и Docker-деплоем.

### Технические решения
- **Backend**: Go 1.22+ + net/http + html/template + embed
- **БД**: SQLite (modernc.org/sqlite, чистый Go, без CGO)
- **Frontend**: HTML/template + Tailwind CSS CDN + GSAP + AOS + vanilla JS
- **React**: TanStack Start + React 19 + shadcn/ui (Lovable-совместимый)
- **Admin**: собственная CMS с сессионной авторизацией (bcrypt)
- **Deploy**: Docker multi-stage + docker compose + nginx

### Созданные файлы
| Файл | Описание |
|------|----------|
| `main.go` | Go HTTP-сервер, роутинг, embed, security headers |
| `internal/db/db.go` | SQLite: инициализация, миграции, seed 13 услуг |
| `internal/models/` | settings, service, contact, seo |
| `internal/handlers/site.go` | главная, форма, sitemap.xml, robots.txt |
| `internal/handlers/admin.go` | CRUD: настройки, услуги, заявки, SEO |
| `internal/middleware/auth.go` | bcrypt + SQLite сессии |
| `templates/index.html` | 12 разделов сайта (Go template) |
| `templates/sections_7_12.html` | разделы 7–12 с реальными клиентами |
| `templates/admin/*.html` | login, dashboard, settings, services, contacts, seo |
| `static/css/style.css` | glassmorphism, dark #05070A, gold #D4AF37 + AOS fallback |
| `static/js/main.js` | GSAP, счётчики, FAQ, форма, back-to-top |
| `src/routes/index.tsx` | React: полная страница, 12 разделов |
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
- Hostname: `msk-1-vm-4rkn`

### Проблемы деплоя (все решены)
1. `docker-compose` не найден → заменили на `docker compose` (v2)
2. `git pull` конфликт → `git reset --hard origin/main`
3. IPv6 вместо IPv4 → `ip -4 route get 1.1.1.1`
4. setup.sh падал на `.env` → SIGPIPE при `tr | head`, заменили на `openssl rand`
5. HEAD на старой ветке → force push обеих веток
6. Go 1.22 vs go.mod 1.25 → Dockerfile `golang:alpine`
7. Внешний firewall VPS → `iptables -I INPUT 1 -p tcp --dport 80 -j ACCEPT`

---

## Сессия 3 — Июнь 2026 (контент с alfaunit1.ru)

### Перенесённый контент
- **11 реальных клиентов** с alfaunit1.ru/partners/: Советская Звезда, АО Трест №68, ВК Сервис, Северная Корона, Технология холода, Ведомственная парковка, ГК Алгоритм, Ренейссанс Констракшн, Медиэстетик, ЭнергоСвязьСтрой, Венчурный Капитал
- **13 реальных услуг** с alfaunit1.ru/services/: 10 видов физической охраны + 2 сопровождения грузов + 3 вида сигнализации
- **Реальный текст** «О компании»: ЧОО Альфа Юнит-1 + Альфа Безопасность, Ассоциация ветеранов антитеррора «Альфа», Закон РФ от 11.03.1992 г.

### Проблемы контента (все решены)
1. AOS с unpkg.com блокировался → переключили на cdnjs.cloudflare.com
2. Весь контент скрыт после hero → AOS CSS скрывал элементы без JS, добавили CSS fallback
3. Страница обрывалась после заголовка услуг → `template.HTML` передавался в `safeHTML(string)`, убрали лишний `safeHTML`

### Итоговая команда переустановки
```bash
cd ~/ALFA1 && docker compose down -v && git fetch origin main && git reset --hard origin/main && docker compose up -d --build
```

---

*Файл обновляется при каждой рабочей сессии.*
