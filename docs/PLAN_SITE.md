# Подробный план создания сайта «Альфа Юнит-1»

## 1. Архитектура проекта

```
ALFA1/
├── main.go                        # Точка входа, HTTP-сервер
├── go.mod                         # Go модуль
├── go.sum                         # Контрольные суммы
├── .env.example                   # Пример конфигурации
├── Dockerfile                     # Multi-stage Docker сборка
├── docker-compose.yml             # Оркестрация контейнеров
├── nginx.conf                     # Nginx конфигурация (reverse proxy)
├── docs/                          # Документация
├── internal/
│   ├── db/
│   │   └── db.go                  # SQLite подключение и миграции
│   ├── models/
│   │   ├── settings.go            # Настройки сайта
│   │   ├── page.go                # Страницы
│   │   ├── service.go             # Услуги
│   │   └── contact.go             # Заявки
│   ├── handlers/
│   │   ├── site.go                # Публичные страницы
│   │   ├── contact.go             # Форма обратной связи
│   │   └── admin.go               # Admin-панель
│   └── middleware/
│       └── auth.go                # Авторизация admin
├── templates/
│   ├── index.html                 # Главная страница (все 12 разделов)
│   ├── admin/
│   │   ├── login.html
│   │   ├── dashboard.html
│   │   ├── settings.html
│   │   ├── services.html
│   │   ├── seo.html
│   │   └── contacts.html
│   └── partials/
│       └── (переиспользуемые компоненты)
└── static/
    ├── css/style.css              # Кастомные стили
    ├── js/main.js                 # Анимации, интерактивность
    └── images/                    # Изображения
```

## 2. Технологический стек

### Backend (Go)

- **Go 1.22** — компилируемый язык, один бинарный файл
- **net/http** — стандартная библиотека HTTP
- **html/template** — безопасный рендеринг HTML
- **embed** — встраивание статики в бинарник
- **modernc.org/sqlite** — SQLite без CGO
- **golang.org/x/crypto/bcrypt** — хеширование паролей

### Frontend

- **Tailwind CSS** (CDN) — утилитарные CSS-классы
- **GSAP 3.12** + ScrollTrigger — профессиональные анимации
- **CountUp.js** — анимированные счётчики
- **Lenis** — плавный скролл
- **Vanilla JS** — без тяжёлых фреймворков

### Database (SQLite)

```sql
-- Таблицы
settings        -- Настройки сайта (title, description, phone, etc.)
services        -- Услуги (name, description, icon, order)
pages           -- Дополнительные страницы
contacts        -- Заявки с формы
users           -- Пользователи admin-панели
seo_settings    -- SEO-настройки для каждой страницы
```

## 3. Разделы сайта

| #   | Раздел           | Компоненты                                       |
| --- | ---------------- | ------------------------------------------------ |
| 1   | **Hero**         | Видео-фон (CSS), заголовок, CTA-кнопки, лицензии |
| 2   | **О компании**   | Анимированные счётчики, цитаты, ценности         |
| 3   | **Услуги**       | 12 glass-карточек с иконками и hover-эффектом    |
| 4   | **Преимущества** | Иконки + описание, анимированный таймлайн        |
| 5   | **Лицензии**     | Галерея документов, lightbox                     |
| 6   | **История**      | Интерактивный таймлайн 2002–2026                 |
| 7   | **Объекты**      | Карточки кейсов (бизнес-центры, склады, etc.)    |
| 8   | **Команда**      | Корпоративные карточки сотрудников               |
| 9   | **Как работаем** | 6-шаговый процесс с анимацией                    |
| 10  | **FAQ**          | Accordion с анимацией                            |
| 11  | **Контакты**     | Карта, форма, WhatsApp/Telegram                  |
| 12  | **Футер**        | Логотип, навигация, реквизиты                    |

## 4. Admin-панель

### Функциональность

- Авторизация (bcrypt пароли, cookie-сессии)
- Редактирование всех текстов и разделов сайта
- Управление услугами (добавить/изменить/удалить)
- Просмотр заявок с формы
- Настройка домена и сервера
- SEO-инструменты (see PLAN_SEO.md)

### URL структура admin

```
/admin/            → Dashboard
/admin/login       → Авторизация
/admin/settings    → Настройки сайта
/admin/services    → Управление услугами
/admin/contacts    → Заявки
/admin/seo         → SEO инструменты
/admin/logout      → Выход
```

## 5. Производительность (Core Web Vitals)

| Метрика | Цель    | Реализация                               |
| ------- | ------- | ---------------------------------------- |
| LCP     | ≤ 2.5с  | Сервер-сайд рендеринг, нет JS-блокировки |
| INP     | ≤ 200мс | Минимум JS на первом экране              |
| CLS     | ≤ 0.1   | Фиксированные размеры изображений        |
| TTFB    | ≤ 600мс | Go = микросекундный отклик               |

## 6. Безопасность

- HTTPS через nginx (Let's Encrypt)
- CSP заголовки
- CSRF защита форм
- Bcrypt для паролей
- Сессии с HttpOnly cookies
- Rate limiting на форму

## 7. Деплой

```bash
# Запуск через Docker Compose
git clone https://github.com/protasovvalera84-stack/ALFA1.git
cd ALFA1
cp .env.example .env
# Отредактировать .env (пароль admin, домен)
docker-compose up -d

# Сайт доступен на http://localhost:8080
# Admin: http://localhost:8080/admin
```

## 8. Резервное копирование

```bash
# Бэкап базы данных
docker exec alfa1_app cp /app/data/alfa1.db /app/data/backup_$(date +%Y%m%d).db

# Восстановление
docker exec alfa1_app cp /app/data/backup_YYYYMMDD.db /app/data/alfa1.db
```
