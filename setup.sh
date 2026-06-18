#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════
# Альфа Юнит-1 — автоматический скрипт установки
# Проверено: Ubuntu 22.04 / 24.04 LTS
#
# Использование:
#   chmod +x setup.sh && sudo ./setup.sh
# ═══════════════════════════════════════════════════════════════
set -euo pipefail

# ── Цвета ────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; NC='\033[0m' # No Color

info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC}   $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*" >&2; exit 1; }

echo ""
echo -e "${BLUE}════════════════════════════════════════════════${NC}"
echo -e "${BLUE}   Альфа Юнит-1 — установка сайта             ${NC}"
echo -e "${BLUE}════════════════════════════════════════════════${NC}"
echo ""

# ── 1. Проверка root ─────────────────────────────────────────
if [ "$EUID" -ne 0 ]; then
  error "Запустите скрипт от root: sudo ./setup.sh"
fi

# ── 2. Установка Docker (если не установлен) ─────────────────
if ! command -v docker &> /dev/null; then
  info "Устанавливаю Docker..."
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
  success "Docker установлен: $(docker --version)"
else
  success "Docker уже установлен: $(docker --version)"
fi

# ── 3. Проверка docker compose v2 ────────────────────────────
if ! docker compose version &> /dev/null; then
  info "Устанавливаю Docker Compose plugin..."
  apt-get install -y -q docker-compose-plugin
fi
success "Docker Compose: $(docker compose version --short)"

# ── 4. Создание .env (если не существует) ────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if [ ! -f ".env" ]; then
  info "Создаю .env из .env.example..."
  cp .env.example .env

  # Генерируем случайный секрет сессии
  SESSION_SECRET=$(openssl rand -base64 32 | tr -d '\n')
  sed -i "s|replace-this-with-a-very-long-random-secret-key-12345|${SESSION_SECRET}|" .env

  warn ""
  warn "ВАЖНО: отредактируйте .env перед запуском!"
  warn "  nano .env"
  warn "Установите:"
  warn "  ADMIN_PASSWORD=ВашНадёжныйПароль"
  warn "  SITE_DOMAIN=https://ВашДомен.ru"
  warn ""
  read -rp "Нажмите Enter для открытия редактора .env, или Ctrl+C чтобы отменить..."
  ${EDITOR:-nano} .env
else
  success ".env уже существует"
fi

# ── 5. Сборка и запуск ───────────────────────────────────────
info "Собираю и запускаю сайт..."
docker compose up -d --build

# ── 6. Ожидание запуска ──────────────────────────────────────
info "Ожидаю запуска контейнера..."
for i in $(seq 1 30); do
  if docker compose ps | grep -q "healthy\|Up"; then
    break
  fi
  sleep 2
  echo -n "."
done
echo ""

# ── 7. Итоговый статус ───────────────────────────────────────
echo ""
echo -e "${GREEN}════════════════════════════════════════════════${NC}"
echo -e "${GREEN}   Установка завершена!                        ${NC}"
echo -e "${GREEN}════════════════════════════════════════════════${NC}"
echo ""
docker compose ps
echo ""

SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || hostname -I | awk '{print $1}')
echo -e "${GREEN}Сайт:${NC}  http://${SERVER_IP}"
echo -e "${GREEN}Admin:${NC} http://${SERVER_IP}/admin/"
echo ""
echo "Полезные команды:"
echo "  docker compose logs -f          # логи в реальном времени"
echo "  docker compose restart app      # перезапустить"
echo "  docker compose down             # остановить"
echo "  docker compose up -d --build    # пересобрать"
echo ""
