# Все скрипты проекта «Альфа Юнит-1»

## ⚡ Установка (одна команда, Ubuntu 22.04/24.04)

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/protasovvalera84-stack/ALFA1/main/setup.sh)
```

Скрипт сам: установит Docker → скачает код → создаст .env → запустит сайт → покажет пароль.

> **Ubuntu 24.04:** используйте `docker compose` (с пробелом), НЕ `docker-compose` (с дефисом).
> `docker compose` (v2 plugin) входит в Docker Engine. `docker-compose` (v1) не установлен.

---

## Разработка (локально)

```bash
# Запуск без Docker
go run main.go

# Сборка бинарника
go build -o alfa1 .

# Проверка кода
go vet ./...
go build -o /dev/null ./...

# Зависимости
go mod tidy
go mod verify
```

## Docker

```bash
# Запустить через Compose (рекомендуется)
docker compose up -d

# Первый запуск + сборка образа
docker compose up -d --build

# Остановить
docker compose down

# Посмотреть логи
docker compose logs -f

# Логи только приложения
docker compose logs -f app

# Перезапустить приложение
docker compose restart app

# Пересобрать после git pull
docker compose up -d --build

# Проверить статус
docker compose ps
```

## База данных (SQLite)

```bash
# Подключиться к SQLite (внутри контейнера)
docker exec -it alfa1_app sqlite3 /app/data/alfa1.db

# Посмотреть таблицы
sqlite3 data/alfa1.db ".tables"

# Посмотреть заявки
sqlite3 data/alfa1.db "SELECT * FROM contacts ORDER BY created_at DESC LIMIT 10;"

# Резервная копия
docker exec alfa1_app sqlite3 /app/data/alfa1.db ".backup /app/data/backup.db"
docker cp alfa1_app:/app/data/backup.db ./backup_$(date +%Y%m%d).db

# Восстановить из бэкапа
docker cp backup_YYYYMMDD.db alfa1_app:/app/data/backup.db
docker exec alfa1_app cp /app/data/backup.db /app/data/alfa1.db
docker compose restart app
```

## Деплой на сервер (Ubuntu 24.04)

```bash
# Автоматически (рекомендуется):
bash <(curl -fsSL https://raw.githubusercontent.com/protasovvalera84-stack/ALFA1/main/setup.sh)

# Вручную (Docker уже установлен):
git clone https://github.com/protasovvalera84-stack/ALFA1.git
cd ALFA1
cp .env.example .env
nano .env        # установить ADMIN_PASSWORD
docker compose up -d

# Обновить до последней версии:
cd ~/ALFA1
git fetch origin main
git reset --hard origin/main
docker compose up -d --build
```

## SSL (HTTPS, после регистрации домена)

```bash
# 1. Установить Certbot
apt install certbot python3-certbot-nginx

# 2. Получить SSL-сертификат
certbot certonly --standalone -d ВАШИ_ДОМЕН -d www.ВАШИ_ДОМЕН

# 3. Запустить с nginx + SSL
docker compose --profile nginx up -d

# 4. Авто-продление (уже настроено в cron)
certbot renew --dry-run
```

## Nginx (только с профилем nginx)

```bash
# Запустить с nginx
docker compose --profile nginx up -d

# Перезагрузить конфиг nginx
docker exec alfa1_nginx nginx -s reload

# Проверить конфиг nginx
docker exec alfa1_nginx nginx -t
```

## Мониторинг

```bash
# Ресурсы контейнеров
docker stats alfa1_app

# Проверка ответа сайта
curl -I http://186.246.12.46

# Проверка sitemap
curl http://186.246.12.46/sitemap.xml

# Проверка robots.txt
curl http://186.246.12.46/robots.txt
```

## SEO-проверки

```bash
# PageSpeed (через API)
curl "https://www.googleapis.com/pagespeedonline/v5/runPagespeed?url=http://186.246.12.46"

# Проверить schema.org
curl -s http://186.246.12.46 | grep -o '"@type":"[^"]*"'

# Проверить мета-теги
curl -s http://186.246.12.46 | grep -E '<title>|<meta name="description"'
```
