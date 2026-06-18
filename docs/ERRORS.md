# Журнал ошибок и исправлений

---

## Сессия 1 — создание сайта

### `mul` не определена в шаблоне
**Ошибка:** `function "mul" not defined`
**Исправление:** добавили в funcMap: `"mul": func(a, b int) int { return a * b }`
**Статус:** ✅

### `slice` не определена в шаблоне
**Ошибка:** `function "slice" not defined`
**Исправление:** убрали `{{$advantages := slice "..."}}` из шаблона
**Статус:** ✅

### Статические файлы не найдены при embed
**Ошибка:** `pattern static: no matching files found`
**Причина:** `touch` создал папки вместо файлов
**Исправление:** файлы созданы через filesystem\_write
**Статус:** ✅

### golang.org/x/crypto не в go.mod
**Ошибка:** `no required module provides package golang.org/x/crypto/bcrypt`
**Исправление:** `go get golang.org/x/crypto@latest`
**Статус:** ✅

### react-refresh warning в ESLint
**Ошибка:** `Fast refresh only works when a file only exports components`
**Исправление:** вынесли `siteMeta` в `src/lib/site-meta.ts`
**Статус:** ✅

---

## Сессия 2 — деплой Ubuntu 24.04

### `docker-compose` не найден
**Ошибка:** `Command 'docker-compose' not found`
**Причина:** Ubuntu 24.04 использует `docker compose` (plugin v2), не `docker-compose` (v1)
**Исправление:** все команды обновлены на `docker compose` (с пробелом)
**Статус:** ✅

### `git pull` — конфликт
**Ошибка:** `Your local changes would be overwritten by merge: setup.sh`
**Исправление:** setup.sh делает `git fetch origin main && git reset --hard origin/main`
**Статус:** ✅

### IPv6 вместо IPv4
**Ошибка:** IP определялся как `2a03:6f00:a::2795` вместо `186.246.12.46`
**Причина:** `curl -s4` игнорировал флаг на этом сервере
**Исправление:** `ip -4 route get 1.1.1.1 | awk '{for(i=1;i<=NF;i++) if($i=="src"){print $(i+1); exit}}'`
**Статус:** ✅

### setup.sh падал молча на шаге .env (SIGPIPE)
**Ошибка:** скрипт выводил `▶ Создание конфигурации...` и завершался без ошибки
**Причина:** `set -o pipefail` + `tr < /dev/urandom | head -c 16` = SIGPIPE (exit 141)
**Исправление:** убрали `-o pipefail`, пароль через `openssl rand -hex 10`
**Статус:** ✅

### `git clone` не получал setup.sh
**Ошибка:** `chmod: cannot access 'setup.sh': No such file or directory`
**Причина:** `origin/HEAD` указывал на старую ветку без setup.sh
**Исправление:** force push актуального кода в обе ветки
**Статус:** ✅

### Несовместимость версии Go в Dockerfile
**Ошибка:** `go: go.mod requires go >= 1.25.0 (running go 1.22.12)`
**Причина:** `go mod tidy` на Go 1.26 обновил go.mod до `go 1.25.0`, Dockerfile использовал `golang:1.22-alpine`
**Исправление:** Dockerfile → `golang:alpine` (всегда последняя версия)
**Статус:** ✅

### Внешний firewall VPS блокировал порт 80
**Ошибка:** `ERR_CONNECTION_TIMED_OUT` в браузере, но `curl localhost:80` работает
**Причина:** ufw был настроен, но VPS-провайдер имеет отдельный сетевой firewall
**Исправление:** `iptables -I INPUT 1 -p tcp --dport 80 -j ACCEPT` + удаление ufw
**Статус:** ✅

---

## Сессия 3 — контент и отображение

### Контент не отображался после hero-секции (AOS CDN)
**Ошибка:** видны только hero + часть about, остальное скрыто
**Причина:** AOS CDN с `unpkg.com` работает медленно/нестабильно на RU-серверах.
AOS CSS скрывает элементы с `data-aos` через `opacity: 0`. Если JS не загрузился — контент остаётся скрытым навсегда.
**Исправление 1:** AOS CDN заменён на `cdnjs.cloudflare.com` (надёжнее)
**Исправление 2:** CSS fallback в style.css:
```css
html:not(.aos-initialized) [data-aos] {
  opacity: 1 !important;
  transform: none !important;
}
```
**Статус:** ✅

### Страница обрывалась на заголовке «НАШИ УСЛУГИ»
**Ошибка из логов:**
```
template: index.html:593:26: executing "index.html" at <safeHTML>:
wrong type for value; expected string; got template.HTML
```
**Причина:** поле `Service.Icon` объявлено как `template.HTML` в модели Go.
В шаблоне использовался `{{$s.Icon | safeHTML}}`, где `safeHTML` ожидает `string`.
Go template engine при ошибке останавливает рендеринг на том месте где произошла ошибка.
Поэтому сервер отправлял только частичный HTML: всё до первой иконки услуги.
**Исправление:** `{{$s.Icon | safeHTML}}` → `{{$s.Icon}}`
(поле уже `template.HTML` — safeHTML не нужен)
**Строка:** `templates/index.html:593`
**Статус:** ✅

---

## Правила для будущих изменений

1. **Dockerfile**: использовать `golang:alpine`, не фиксировать версию
2. **go.mod**: не запускать `go mod tidy` на машине с более новым Go без проверки Dockerfile
3. **setup.sh**: не использовать `set -o pipefail` с командами через `/dev/urandom`
4. **IP определение**: `ip -4 route get` вместо curl к внешним сервисам
5. **git pull на сервере**: `git fetch + git reset --hard` вместо `git pull`
6. **CDN**: использовать `cdnjs.cloudflare.com` вместо `unpkg.com` для RU-серверов
7. **template.HTML vs string**: если поле типа `template.HTML` — рендерить напрямую `{{.Field}}`, не через `safeHTML`
8. **AOS fallback**: всегда добавлять CSS-fallback для AOS на случай блокировки CDN
