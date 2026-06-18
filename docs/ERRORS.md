# Журнал ошибок и исправлений

---

## Сессия 1 — создание сайта

### Ошибка: `mul` не определена в шаблоне Go
**Файл:** `templates/index.html`
**Ошибка:** `function "mul" not defined`
**Причина:** Go html/template не имеет встроенных арифметических функций
**Исправление:** добавили в `main.go`:
```go
funcMap := template.FuncMap{
    "mul": func(a, b int) int { return a * b },
    "add": func(a, b int) int { return a + b },
}
```
**Статус:** ✅ Исправлено

### Ошибка: `slice` не определена в шаблоне Go
**Файл:** `templates/index.html`
**Ошибка:** `function "slice" not defined`
**Причина:** попытка создать срез прямо в шаблоне
**Исправление:** убрали `{{$advantages := slice "..."}}` из шаблона
**Статус:** ✅ Исправлено

### Ошибка: статические файлы не найдены при embed
**Файл:** `main.go`
**Ошибка:** `pattern static: no matching files found`
**Причина:** папки `static/css/` и `static/js/` существовали как директории (touch создал папки вместо файлов)
**Исправление:** создали файлы через filesystem\_write вместо touch
**Статус:** ✅ Исправлено

### Ошибка: golang.org/x/crypto не в go.mod
**Файл:** `internal/handlers/admin.go`
**Ошибка:** `no required module provides package golang.org/x/crypto/bcrypt`
**Исправление:** `go get golang.org/x/crypto@latest`
**Статус:** ✅ Исправлено

### Ошибка: react-refresh warning в ESLint
**Файл:** `src/routes/__root.tsx`
**Ошибка:** `Fast refresh only works when a file only exports components`
**Причина:** `siteMeta` константа экспортировалась из файла с компонентом Route
**Исправление:** вынесли `siteMeta` в отдельный файл `src/lib/site-meta.ts`
**Статус:** ✅ Исправлено

---

## Сессия 2 — деплой на Ubuntu 24.04

### Ошибка: `docker-compose` не найден
**Сервер:** 186.246.12.46 (Ubuntu 24.04.4 LTS)
**Ошибка:** `Command 'docker-compose' not found, but can be installed with: apt install docker-compose`
**Причина:** Ubuntu 24.04 не включает docker-compose v1 по умолчанию; используется плагин `docker compose` v2
**Исправление:**
- Все команды в docs, scripts, docker-compose.yml обновлены на `docker compose` (с пробелом)
- setup.sh проверяет плагин и устанавливает если нет
**Статус:** ✅ Исправлено

### Ошибка: `git pull` — конфликт локальных изменений
**Ошибка:** `error: Your local changes to the following files would be overwritten by merge: setup.sh. Please commit your changes or stash them before you merge. Aborting`
**Причина:** сервер хранил изменённую версию setup.sh, git pull не мог перезаписать
**Исправление:** setup.sh теперь начинается с `git fetch origin main && git reset --hard origin/main`
**Статус:** ✅ Исправлено

### Ошибка: setup.sh определял IPv6 вместо IPv4
**Ошибка:** `IP сервера: 2a03:6f00:a::2795` (IPv6)
**Ожидалось:** `186.246.12.46` (IPv4)
**Причина:** `curl -s4 https://ifconfig.me` игнорировал флаг `-4` на этом сервере
**Исправление:** заменили на `ip -4 route get 1.1.1.1 | awk '{for(i=1;i<=NF;i++) if($i=="src"){print $(i+1); exit}}'`
**Статус:** ✅ Исправлено

### Ошибка: setup.sh молча падал на шаге создания .env (SIGPIPE)
**Ошибка:** скрипт выводил `▶ Создание конфигурации...` и завершался без сообщения об ошибке
**Причина:** `set -o pipefail` + `LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 16`
Когда `head` читает 16 байт и закрывает pipe, `tr` получает SIGPIPE (exit code 141).
С `set -o pipefail` это вызывает немедленный выход скрипта.
**Исправление:**
- Убрали `-o pipefail` из `set -euo pipefail` (оставили `set -eu`)
- Генерацию пароля заменили на `openssl rand -hex 10` (без pipe)
**Статус:** ✅ Исправлено

### Ошибка: `git clone` не получал setup.sh (неверная ветка по умолчанию)
**Ошибка:** `chmod: cannot access 'setup.sh': No such file or directory`
**Причина:** `origin/HEAD` указывал на `neo/premium-website-go-alfa1-x8k2p` (старая ветка без setup.sh), а не на `main`
**Исправление:**
- Обновили ветку `neo/premium-website-go-alfa1-x8k2p` до актуального кода через force push
- setup.sh теперь клонирует с явным указанием: `git clone -q ... "$INSTALL_DIR"` и работает из `$INSTALL_DIR`
**Статус:** ✅ Исправлено

### Ошибка: несовместимость версии Go в Dockerfile
**Ошибка:** `go: go.mod requires go >= 1.25.0 (running go 1.22.12; GOTOOLCHAIN=local)`
**Причина:** разработка велась с Go 1.26.1, `go mod tidy` автоматически обновил `go.mod` до `go 1.25.0`. Dockerfile использовал `golang:1.22-alpine`
**Исправление:** Dockerfile обновлён с `golang:1.22-alpine` на `golang:alpine` (всегда последняя версия, совместима с любым `go.mod`)
**Статус:** ✅ Исправлено

---

## Правила для будущих изменений

1. **Dockerfile**: всегда использовать `golang:alpine`, не фиксировать минорную версию
2. **go.mod**: не запускать `go mod tidy` на машине с более новым Go без проверки Dockerfile
3. **setup.sh**: не использовать `set -o pipefail` с командами, читающими из `/dev/urandom`
4. **IP определение**: использовать `ip -4 route get` вместо curl к внешним сервисам
5. **git pull**: использовать `git fetch + git reset --hard` вместо `git pull` для деплоя
