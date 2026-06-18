#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════
# Альфа Юнит-1 — автоматический скрипт установки
# Проверено: Ubuntu 22.04 / 24.04 LTS
#
# Использование:
#   chmod +x setup.sh
#   sudo ./setup.sh
#
# Скрипт автоматически:
#   1. Установит Docker (если не установлен)
#   2. Сгенерирует пароль администратора и секретный ключ
#   3. Запустит сайт через docker compose
#   4. Выведет все данные для входа
# ═══════════════════════════════════════════════════════════════
set -euo pipefail

# ── Цвета ────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()    { echo -e "${BLUE}[INFO]${NC}  $*"; }
success() { echo -e "${GREEN}[OK]${NC}    $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC}  $*"; }
step()    { echo -e "\n${CYAN}${BOLD}▶ $*${NC}"; }

echo ""
echo -e "${BOLD}${BLUE}════════════════════════════════════════════════════${NC}"
echo -e "${BOLD}${BLUE}   Альфа Юнит-1 — автоматическая установка сайта   ${NC}"
echo -e "${BOLD}${BLUE}════════════════════════════════════════════════════${NC}"
echo ""

# ── 1. Проверка root ─────────────────────────────────────────
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}Ошибка:${NC} запустите скрипт от root:"
  echo "  sudo ./setup.sh"
  exit 1
fi

# ── 2. Установка Docker (если не установлен) ─────────────────
step "Проверка Docker..."

if ! command -v docker &> /dev/null; then
  info "Docker не найден. Устанавливаю..."
  apt-get update -q
  apt-get install -y -q ca-certificates curl gnupg
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
    | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  chmod a+r /etc/apt/keyrings/docker.gpg
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
    > /etc/apt/sources.list.d/docker.list
  apt-get update -q
  apt-get install -y -q docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  systemctl enable --now docker
  success "Docker установлен"
else
  success "Docker найден: $(docker --version | cut -d' ' -f3 | tr -d ',')"
fi

# Проверить docker compose v2
if ! docker compose version &>/dev/null; then
  info "Устанавливаю docker-compose-plugin..."
  apt-get install -y -q docker-compose-plugin
fi
success "Docker Compose: $(docker compose version --short)"

# ── 3. Определить IP сервера ─────────────────────────────────
step "Определение IP-адреса сервера..."

SERVER_IP=""
# Попробовать несколько источников
SERVER_IP=$(curl -s --connect-timeout 3 https://ifconfig.me 2>/dev/null || true)
if [ -z "$SERVER_IP" ]; then
  SERVER_IP=$(curl -s --connect-timeout 3 https://api.ipify.org 2>/dev/null || true)
fi
if [ -z "$SERVER_IP" ]; then
  SERVER_IP=$(hostname -I | awk '{print $1}' | tr -d '\n')
fi

success "IP сервера: $SERVER_IP"

# ── 4. Генерация конфигурации .env ───────────────────────────
step "Создание конфигурации .env..."

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if [ -f ".env" ]; then
  # .env уже существует — проверить, не дефолтный ли пароль
  CURRENT_PASSWORD=$(grep '^ADMIN_PASSWORD=' .env 2>/dev/null | cut -d'=' -f2 || echo "")
  if [ "$CURRENT_PASSWORD" = "ChangeMe2026!" ] || [ -z "$CURRENT_PASSWORD" ]; then
    info ".env существует но пароль не изменён — генерирую новый..."
    REGENERATE=true
  else
    success ".env уже настроен"
    REGENERATE=false
  fi
else
  REGENERATE=true
fi

if [ "$REGENERATE" = "true" ]; then
  # Генерируем надёжный пароль (16 символов: буквы + цифры)
  ADMIN_PASS=$(LC_ALL=C tr -dc 'A-Za-z0-9@#!%' < /dev/urandom | head -c 16)
  # Генерируем секрет сессии (32 байта base64)
  SESSION_KEY=$(openssl rand -base64 32 | tr -d '\n/=+')

  cat > .env <<EOF
# =====================================================
# Конфигурация «Альфа Юнит-1» (создано setup.sh)
# =====================================================

# Порт HTTP сервера
PORT=8080

# Путь к файлу базы данных SQLite
DB_PATH=/app/data/alfa1.db

# Домен сайта (для sitemap.xml и canonical URLs)
# Без слеша в конце!
SITE_DOMAIN=http://${SERVER_IP}

# Пароль администратора (автоматически сгенерирован)
ADMIN_PASSWORD=${ADMIN_PASS}

# Секрет для подписи сессий (автоматически сгенерирован)
SESSION_SECRET=${SESSION_KEY}

# Режим работы
APP_ENV=production
EOF

  success ".env создан с автоматическими учётными данными"
else
  # Если .env уже был — всё равно прочитаем пароль из него
  ADMIN_PASS=$(grep '^ADMIN_PASSWORD=' .env | cut -d'=' -f2)
fi

# ── 5. Сборка и запуск ───────────────────────────────────────
step "Сборка Docker-образа и запуск сайта..."
info "Это займёт 1–3 минуты при первом запуске..."

docker compose down 2>/dev/null || true
docker compose up -d --build

# ── 6. Ожидание запуска контейнера ───────────────────────────
step "Ожидание готовности сайта..."

READY=false
for i in $(seq 1 40); do
  if curl -s --connect-timeout 2 "http://localhost:8080/" > /dev/null 2>&1; then
    READY=true
    break
  fi
  printf "."
  sleep 2
done
echo ""

if [ "$READY" = "false" ]; then
  warn "Сайт не ответил в течение 80 секунд."
  warn "Проверьте логи: docker compose logs -f"
else
  success "Сайт запущен и отвечает"
fi

# ── 7. Итог ──────────────────────────────────────────────────
echo ""
echo -e "${BOLD}${GREEN}════════════════════════════════════════════════════${NC}"
echo -e "${BOLD}${GREEN}   ✓ Установка завершена успешно!                   ${NC}"
echo -e "${BOLD}${GREEN}════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${BOLD}Данные для входа:${NC}"
echo -e "  ${CYAN}Сайт:${NC}        http://${SERVER_IP}"
echo -e "  ${CYAN}Admin-панель:${NC} http://${SERVER_IP}/admin/"
echo -e "  ${CYAN}Пароль admin:${NC} ${BOLD}${YELLOW}${ADMIN_PASS}${NC}"
echo ""
echo -e "${BOLD}СОХРАНИТЕ ПАРОЛЬ!${NC} Он также записан в файл: ${SCRIPT_DIR}/.env"
echo ""
echo -e "${BOLD}Полезные команды:${NC}"
echo "  docker compose logs -f          # логи в реальном времени"
echo "  docker compose restart app      # перезапустить сайт"
echo "  docker compose down             # остановить"
echo "  docker compose up -d --build    # обновить после git pull"
echo ""

# Сохраним данные в файл для удобства
cat > /root/ALFA1_credentials.txt <<EOF
═══════════════════════════════════════════
АЛЬФА ЮНИТ-1 — данные для входа
═══════════════════════════════════════════
Сайт:        http://${SERVER_IP}
Admin-панель: http://${SERVER_IP}/admin/
Пароль admin: ${ADMIN_PASS}
═══════════════════════════════════════════
Создано: $(date)
EOF

echo -e "${GREEN}Данные также сохранены в: /root/ALFA1_credentials.txt${NC}"
echo ""
