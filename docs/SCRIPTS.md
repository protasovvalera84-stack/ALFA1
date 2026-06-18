# Все скрипты проекта «Альфа Юнит-1»

## Разработка

```bash
# Запуск в режиме разработки
cd ALFA1
go run main.go

# Сборка бинарника
go build -o alfa1 .

# Запуск тестов
go test ./...

# Линтинг
go vet ./...
gofmt -l .

# Проверка зависимостей
go mod tidy
go mod verify
```

## Docker

```bash
# Собрать образ
docker build -t alfa1:latest .

# Запустить в Docker
docker run -p 8080:8080 -v $(pwd)/data:/app/data alfa1:latest

# Запустить через Compose (рекомендуется)
docker-compose up -d

# Остановить
docker-compose down

# Посмотреть логи
docker-compose logs -f

# Пересобрать после изменений
docker-compose up -d --build

# Проверить статус контейнеров
docker-compose ps
```

## База данных

```bash
# Подключиться к SQLite (внутри контейнера)
docker exec -it alfa1_app sqlite3 /app/data/alfa1.db

# Резервная копия
docker exec alfa1_app sqlite3 /app/data/alfa1.db ".backup /app/data/backup.db"

# Скопировать бэкап на хост
docker cp alfa1_app:/app/data/backup.db ./backup_$(date +%Y%m%d).db

# Просмотр таблиц
sqlite3 data/alfa1.db ".tables"

# Посмотреть заявки
sqlite3 data/alfa1.db "SELECT * FROM contacts ORDER BY created_at DESC LIMIT 10;"
```

## Деплой на сервер Linux

```bash
# Клонировать репозиторий
git clone https://github.com/protasovvalera84-stack/ALFA1.git
cd ALFA1

# Скопировать конфиг
cp .env.example .env
nano .env  # Установить пароль, домен, порт

# Установить Docker (если нет)
curl -fsSL https://get.docker.com | sh
systemctl enable docker
systemctl start docker

# Установить Docker Compose
apt install docker-compose-plugin

# Запустить сайт
docker-compose up -d

# Проверить работу
curl http://localhost:8080

# Автозапуск при перезагрузке (уже настроен через restart: always)
```

## Nginx (SSL, HTTPS)

```bash
# Установить Certbot (Let's Encrypt)
apt install certbot python3-certbot-nginx

# Получить SSL сертификат
certbot --nginx -d alfaunit1.ru -d www.alfaunit1.ru

# Обновить сертификат (автоматически через cron)
certbot renew --dry-run

# Перезагрузить nginx
nginx -t && nginx -s reload
```

## Обновление сайта

```bash
# Получить последние изменения
git pull origin main

# Пересобрать и перезапустить
docker-compose up -d --build

# Проверить статус
docker-compose ps
docker-compose logs --tail=20
```

## Мониторинг

```bash
# Использование ресурсов
docker stats alfa1_app

# Проверка ответа сервера
curl -I https://alfaunit1.ru

# Проверка sitemap
curl https://alfaunit1.ru/sitemap.xml

# Проверка robots.txt
curl https://alfaunit1.ru/robots.txt

# Проверка schema.org
curl https://alfaunit1.ru | grep -o '"@type":"[^"]*"'
```

## SEO Проверки

```bash
# Проверить скорость (через API PageSpeed)
curl "https://www.googleapis.com/pagespeedonline/v5/runPagespeed?url=https://alfaunit1.ru&strategy=mobile"

# Проверить индексацию Google
# Перейти в https://search.google.com/search-console

# Проверить индексацию Яндекс
# Перейти в https://webmaster.yandex.ru/

# Проверить schema.org
# https://validator.schema.org/
# Вставить URL сайта

# Проверить мета-теги
curl -s https://alfaunit1.ru | grep -E "<title>|<meta name="description""
```
