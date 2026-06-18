#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════
# Альфа Юнит-1 — автоматический скрипт установки
# Проверено: Ubuntu 22.04 / 24.04 LTS
# Использование: sudo ./setup.sh
# ═══════════════════════════════════════════════════════════════
set -eu

# ── Цвета ────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'

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
  echo -e "${RED}Ошибка:${NC} запустите скрипт от root: sudo ./setup.sh"
  exit 1
fi

# ── 2. Установка Docker (если не установлен) ─────────────────
step "Проверка Docker..."

if ! command -v docker > /dev/null 2>&1; then
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
  apt-get install -y -q docker-ce docker-ce-cli containerd.io \
    docker-buildx-plugin docker-compose-plugin
  systemctl enable --now docker
  success "Docker установлен"
else
  success "Docker найден: $(docker --version | grep -oP '\d+\.\d+\.\d+' | head -1)"
fi

if ! docker compose version > /dev/null 2>&1; then
  apt-get install -y -q docker-compose-plugin
fi
success "Docker Compose: $(docker compose version --short)"

# ── 3. Определить IPv4 сервера ────────────────────────────────
step "Определение IP-адреса сервера..."

SERVER_IP=""
# Пробуем получить IPv4 через несколько сервисов
for SVC in "https://ipv4.icanhazip.com" "https://api4.ipify.org" "https://ifconfig.me/ip"; do
  IP=$(curl -s4 --connect-timeout 4 "$SVC" 2>/dev/null || true)
  if echo "$IP" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$'; then
    SERVER_IP="$IP"
    break
  fi
done

# Запасной вариант — локальный IP
if [ -z "$SERVER_IP" ]; then
  SERVER_IP=$(ip -4 addr show scope global | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -1 || echo "localhost")
fi

success "IP сервера: $SERVER_IP"

# ── 4. Создание .env ─────────────────────────────────────────
step "Создание конфигурации .env..."

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

NEED_CREATE=true
if [ -f ".env" ]; then
  CURRENT_PASS=$(grep '^ADMIN_PASSWORD=' .env 2>/dev/null | cut -d'=' -f2- || echo "")
  if [ -n "$CURRENT_PASS" ] && [ "$CURRENT_PASS" != "ChangeMe2026!" ] && [ "$CURRENT_PASS" != "YourStrongPasswordHere" ]; then
    success ".env уже настроен — пересоздавать не нужно"
    ADMIN_PASS="$CURRENT_PASS"
    NEED_CREATE=false
  fi
fi

if [ "$NEED_CREATE" = "true" ]; then
  # Генерируем пароль через openssl (без pipefail-проблем)
  ADMIN_PASS=$(openssl rand -hex 10)
  # Генерируем секрет сессии
  SESSION_KEY=$(openssl rand -hex 32)

  cat > .env << ENVEOF
# =====================================================
# Конфигурация «Альфа Юнит-1» (создано setup.sh)
# Изменить домен и пароль можно здесь, затем:
#   docker compose restart app
# =====================================================

PORT=8080
DB_PATH=/app/data/alfa1.db
SITE_DOMAIN=http://${SERVER_IP}
ADMIN_PASSWORD=${ADMIN_PASS}
SESSION_SECRET=${SESSION_KEY}
APP_ENV=production
ENVEOF

  success ".env создан"
fi

# ── 5. Сборка и запуск ───────────────────────────────────────
step "Сборка Docker-образа и запуск сайта..."
info "Первая сборка занимает 2-5 минут..."

docker compose down 2>/dev/null || true
docker compose up -d --build

# ── 6. Проверка запуска ──────────────────────────────────────
step "Ожидание запуска сайта..."

READY=false
for i in $(seq 1 40); do
  if curl -s --connect-timeout 2 "http://localhost:8080/" > /dev/null 2>&1; then
    READY=true
    break
  fi
  printf "."
  sleep 3
done
echo ""

if [ "$READY" = "false" ]; then
  warn "Сайт не ответил за 2 минуты. Проверьте: docker compose logs -f"
else
  success "Сайт работает"
fi

# ── 7. Сохранить данные и показать итог ──────────────────────
cat > /root/ALFA1_credentials.txt << CREDEOF
═══════════════════════════════════════════════════
  АЛЬФА ЮНИТ-1 — данные для входа
═══════════════════════════════════════════════════
  Сайт:         http://${SERVER_IP}
  Admin-панель: http://${SERVER_IP}/admin/
  Пароль admin: ${ADMIN_PASS}
═══════════════════════════════════════════════════
  Создано: $(date)
  Файл с данными: /root/ALFA1_credentials.txt
CREDEOF

echo ""
echo -e "${BOLD}${GREEN}════════════════════════════════════════════════════${NC}"
echo -e "${BOLD}${GREEN}   ✓  Установка завершена!                          ${NC}"
echo -e "${BOLD}${GREEN}════════════════════════════════════════════════════${NC}"
echo ""
echo -e "  ${CYAN}Сайт:${NC}         http://${SERVER_IP}"
echo -e "  ${CYAN}Admin-панель:${NC} http://${SERVER_IP}/admin/"
echo -e "  ${CYAN}Пароль admin:${NC} ${BOLD}${YELLOW}${ADMIN_PASS}${NC}"
echo ""
echo -e "${GREEN}Данные сохранены в:${NC} /root/ALFA1_credentials.txt"
echo ""
echo "Команды управления:"
echo "  docker compose logs -f          # логи"
echo "  docker compose restart app      # перезапуск"
echo "  docker compose down             # остановить"
echo "  docker compose up -d --build    # обновить"
echo ""
