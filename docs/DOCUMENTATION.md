# Техническая документация сайта «Альфа Юнит-1»

## Обзор системы

**Альфа Юнит-1** — корпоративный сайт охранной компании, реализованный в виде Go-веб-приложения с встроенной системой управления контентом (CMS) и SEO-инструментами.

**Ключевые принципы:**

- Один бинарный файл Go — никаких внешних рантаймов
- SQLite — никаких отдельных серверов баз данных
- Docker — одна команда для запуска на любом сервере
- Server-Side Rendering — максимальная скорость и SEO

---

## Алгоритм работы сервера

```
HTTP Request
    │
    ▼
main.go → mux.ServeHTTP()
    │
    ├─► /static/*     → http.FileServer (встроенные файлы)
    ├─► /robots.txt   → handlers.RobotsTxt()
    ├─► /sitemap.xml  → handlers.Sitemap()
    ├─► /             → handlers.Home()
    ├─► /contact      → handlers.Contact() [POST]
    └─► /admin/*      → middleware.Auth() → handlers.Admin*()
                              │
                              ▼
                        Проверка сессии в cookie
                        SQLite запрос к users
                        ├─ OK → выполнить handler
                        └─ Fail → redirect /admin/login
```

## Структура базы данных

### Таблица `settings`

```sql
CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
-- Ключи: site_title, site_description, phone1, phone2, email,
--        address, working_hours, admin_password_hash, domain
```

### Таблица `services`

```sql
CREATE TABLE services (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    description TEXT NOT NULL,
    icon        TEXT NOT NULL,    -- SVG иконка (inline)
    sort_order  INTEGER DEFAULT 0,
    active      BOOLEAN DEFAULT 1
);
```

### Таблица `seo_pages`

```sql
CREATE TABLE seo_pages (
    slug        TEXT PRIMARY KEY,  -- '/', '/services', etc.
    title       TEXT NOT NULL,
    description TEXT NOT NULL,
    og_image    TEXT,
    schema_json TEXT              -- JSON-LD структура
);
```

### Таблица `contacts`

```sql
CREATE TABLE contacts (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    phone      TEXT NOT NULL,
    email      TEXT,
    message    TEXT,
    service    TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    read       BOOLEAN DEFAULT 0
);
```

### Таблица `sessions`

```sql
CREATE TABLE sessions (
    token      TEXT PRIMARY KEY,
    user_id    INTEGER,
    expires_at DATETIME NOT NULL
);
```

---

## API Endpoints

### Публичные

| Метод | URL            | Описание              |
| ----- | -------------- | --------------------- |
| GET   | `/`            | Главная страница      |
| POST  | `/contact`     | Отправка формы (JSON) |
| GET   | `/sitemap.xml` | Карта сайта XML       |
| GET   | `/robots.txt`  | Правила для ботов     |
| GET   | `/static/*`    | Статические файлы     |

### Admin (требует авторизации)

| Метод    | URL               | Описание            |
| -------- | ----------------- | ------------------- |
| GET/POST | `/admin/login`    | Авторизация         |
| GET      | `/admin/`         | Dashboard           |
| GET/POST | `/admin/settings` | Настройки сайта     |
| GET/POST | `/admin/services` | Управление услугами |
| GET      | `/admin/contacts` | Заявки с формы      |
| GET/POST | `/admin/seo`      | SEO настройки       |
| GET      | `/admin/logout`   | Выход               |

---

## Компоненты frontend

### CSS (static/css/style.css)

```
:root
  └── CSS переменные (цвета, шрифты, отступы)

.glass          → glassmorphism карточка
.btn-primary    → золотая кнопка (#D4AF37)
.btn-outline    → прозрачная кнопка с рамкой
.section        → стандартная секция
.container      → контейнер с max-width

.hero           → Hero-блок (100vh, видео-фон)
.services-grid  → сетка карточек услуг (3 колонки)
.timeline       → временная линия истории
.accordion      → FAQ аккордеон
.counter-item   → анимированный счётчик
```

### JavaScript (static/js/main.js)

```
Modules:
├── initLenis()          → плавный скролл
├── initGSAP()           → ScrollTrigger анимации
├── initCounters()       → CountUp при появлении в viewport
├── initFAQ()            → аккордеон FAQ
├── initMobileMenu()     → мобильное меню
├── initContactForm()    → AJAX отправка формы
└── initLightbox()       → просмотр лицензий
```

---

## Конфигурация (.env)

```env
# Порт сервера
PORT=8080

# Путь к базе данных SQLite
DB_PATH=./data/alfa1.db

# Начальный пароль admin (заменить сразу после установки!)
ADMIN_PASSWORD=ChangeMe2026!

# Домен сайта (для sitemap.xml и canonical URLs)
SITE_DOMAIN=https://alfaunit1.ru

# Секрет для сессий (32+ символа)
SESSION_SECRET=your-very-long-random-secret-key-here
```

---

## Безопасность

### Авторизация admin

1. Пользователь вводит пароль
2. Go сравнивает с bcrypt-хешем в БД (`bcrypt.CompareHashAndPassword`)
3. При успехе — создаётся UUID токен сессии в таблице `sessions`
4. Токен записывается в cookie: `HttpOnly=true, SameSite=Lax, Secure=true`
5. Каждый запрос к `/admin/*` — проверка токена в БД

### Защита форм

- CSRF-токен в скрытом поле
- Rate limiting: максимум 5 отправок в минуту с одного IP
- Валидация и санитизация всех входных данных

### HTTP заголовки безопасности

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'; ...
```

---

## Адаптивность (Responsive Design)

| Устройство | Ширина      | Поведение                    |
| ---------- | ----------- | ---------------------------- |
| Мобильный  | < 640px     | 1 колонка, меню-бургер       |
| Планшет    | 640–1024px  | 2 колонки                    |
| ПК         | 1024–1440px | 3 колонки                    |
| ТВ/широкий | > 1440px    | 4 колонки, max-width: 1400px |

---

## SEO-алгоритм

При каждом запросе страницы:

1. Загрузить SEO-настройки для URL из `seo_pages`
2. Вставить `<title>`, `<meta description>`, OG-теги в `<head>`
3. Вставить Schema.org JSON-LD из `schema_json`
4. Добавить `<link rel="canonical">` с правильным URL
5. Вставить hreflang (если есть мультиязычность)

Sitemap генерируется динамически:

```go
// При GET /sitemap.xml
w.Header().Set("Content-Type", "application/xml")
tmpl.Execute(w, pages) // pages из БД + статические URL
```

---

## Производительность

**Почему Go быстрее Node.js/PHP для этого сайта:**

- Компилируемый язык — нет интерпретации
- Конкурентная обработка запросов (goroutines)
- Server-side rendering — браузер получает готовый HTML
- Нет тяжёлого JS-фреймворка на клиенте

**Оценочные показатели на VPS 1 CPU / 1GB RAM:**

- TTFB: ~15-30 мс
- 1000+ одновременных пользователей без замедления
- Потребление памяти: ~30-50 MB

---

## Резервное копирование и восстановление

SQLite — это файл. Бэкап = копирование файла.

```bash
# Создать бэкап
cp data/alfa1.db data/alfa1_backup_$(date +%Y%m%d_%H%M).db

# Восстановить
cp data/alfa1_backup_20260601_1200.db data/alfa1.db
# Перезапустить сервер
docker-compose restart
```

---

_Документация актуальна на момент создания сайта — Июнь 2026._
