#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════
# Альфа Юнит-1 — полностью автоматическая установка
#
# Запуск (одна команда):
#   bash <(curl -fsSL https://raw.githubusercontent.com/protasovvalera84-stack/ALFA1/main/setup.sh)
#
# Или из папки с репозиторием:
#   sudo bash setup.sh
# ═══════════════════════════════════════════════════════════════
set -eu

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'
CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'

ok()   { echo -e "${GREEN}[OK]${NC}  $*"; }
info() { echo -e "${BLUE}[..]${NC}  $*"; }
step() { echo -e "\n${CYAN}${BOLD}▶ $*${NC}"; }

echo ""
echo -e "${BOLD}${BLUE}══════════════════════════════════════════${NC}"
echo -e "${BOLD}${BLUE}  Альфа Юнит-1 — установка сайта         ${NC}"
echo -e "${BOLD}${BLUE}══════════════════════════════════════════${NC}"
echo ""

# ── Рабочая папка ───────────────────────────────────────────
INSTALL_DIR="/root/ALFA1"

# ── 1. Docker ────────────────────────────────────────────────
step "Проверка Docker..."

if ! command -v docker > /dev/null 2>&1; then
  info "Устанавливаю Docker..."
  apt-get update -qq
  apt-get install -y -qq ca-certificates curl gnupg
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
    | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  chmod a+r /etc/apt/keyrings/docker.gpg
  CODENAME=$(. /etc/os-release && echo "$VERSION_CODENAME")
  ARCH=$(dpkg --print-architecture)
  echo "deb [arch=${ARCH} signed-by=/etc/apt/keyrings/docker.gpg] \
https://download.docker.com/linux/ubuntu ${CODENAME} stable" \
    > /etc/apt/sources.list.d/docker.list
  apt-get update -qq
  apt-get install -y -qq docker-ce docker-ce-cli containerd.io \
    docker-buildx-plugin docker-compose-plugin
  systemctl enable --now docker
fi

ok "Docker $(docker --version | grep -oP '[\d.]+' | head -1)"
ok "Compose $(docker compose version --short)"

# ── 2. Код ───────────────────────────────────────────────────
step "Получение актуального кода..."

if [ -d "$INSTALL_DIR/.git" ]; then
  cd "$INSTALL_DIR"
  git fetch origin main -q
  git reset --hard origin/main -q
  ok "Репозиторий обновлён"
else
  rm -rf "$INSTALL_DIR"
  git clone -q https://github.com/protasovvalera84-stack/ALFA1.git "$INSTALL_DIR"
  ok "Репозиторий загружен"
fi

cd "$INSTALL_DIR"

# ── 3. IPv4 адрес ────────────────────────────────────────────
step "Определение IP-адреса..."

# Получаем IPv4 через таблицу маршрутизации — надёжно, без curl
SERVER_IP=$(ip -4 route get 1.1.1.1 2>/dev/null \
  | awk '{for(i=1;i<=NF;i++) if($i=="src") {print $(i+1); exit}}')

# Запасной вариант
if [ -z "$SERVER_IP" ]; then
  SERVER_IP=$(ip -4 addr show scope global \
    | awk '/inet /{gsub(/\/.*/,"",$2); print $2; exit}')
fi

[ -z "$SERVER_IP" ] && SERVER_IP="localhost"
ok "IPv4: $SERVER_IP"

# ── 4. Конфиг .env ──────────────────────────────────────────
step "Создание конфигурации..."

# Генерируем пароль и секрет через openssl
ADMIN_PASS=$(openssl rand -hex 10)
SESSION_KEY=$(openssl rand -hex 32)

# Если .env уже есть с нормальным паролем — не трогаем
if [ -f ".env" ]; then
  EXISTING=$(grep '^ADMIN_PASSWORD=' .env 2>/dev/null | cut -d'=' -f2- || true)
  if [ -n "$EXISTING" ] \
     && [ "$EXISTING" != "ChangeMe2026!" ] \
     && [ "$EXISTING" != "YourStrongPasswordHere" ] \
     && [ "$EXISTING" != "Alfa2026Admin" ]; then
    ADMIN_PASS="$EXISTING"
    ok ".env уже настроен — оставляю без изменений"
  else
    # Перезаписываем дефолтный .env
    printf 'PORT=8080\nDB_PATH=/app/data/alfa1.db\nSITE_DOMAIN=http://%s\nADMIN_PASSWORD=%s\nSESSION_SECRET=%s\nAPP_ENV=production\n' \
      "$SERVER_IP" "$ADMIN_PASS" "$SESSION_KEY" > .env
    ok ".env создан"
  fi
else
  printf 'PORT=8080\nDB_PATH=/app/data/alfa1.db\nSITE_DOMAIN=http://%s\nADMIN_PASSWORD=%s\nSESSION_SECRET=%s\nAPP_ENV=production\n' \
    "$SERVER_IP" "$ADMIN_PASS" "$SESSION_KEY" > .env
  ok ".env создан"
fi

# ── 5. Запуск ────────────────────────────────────────────────
step "Сборка и запуск (1-3 минуты)..."

docker compose down --remove-orphans 2>/dev/null || true
docker compose up -d --build

# ── 6. Проверка ──────────────────────────────────────────────
step "Проверка работоспособности..."

ALIVE=false
for i in $(seq 1 30); do
  if curl -sf --connect-timeout 2 "http://localhost:8080/" > /dev/null 2>&1; then
    ALIVE=true
    break
  fi
  printf "."
  sleep 3
done
echo ""

if [ "$ALIVE" = "true" ]; then
  ok "Сайт отвечает"
else
  echo "Логи: docker compose -f ${INSTALL_DIR}/docker-compose.yml logs --tail=30"
fi

# ── 7. Сохранить данные ──────────────────────────────────────
printf '═══════════════════════════════════\n  АЛЬФА ЮНИТ-1 — данные входа\n═══════════════════════════════════\n  Сайт:   http://%s\n  Admin:  http://%s/admin/\n  Пароль: %s\n═══════════════════════════════════\n  %s\n' \
  "$SERVER_IP" "$SERVER_IP" "$ADMIN_PASS" "$(date)" \
  > /root/ALFA1_access.txt

# ── Итог ─────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}${GREEN}══════════════════════════════════════════${NC}"
echo -e "${BOLD}${GREEN}  ✓  Готово!                              ${NC}"
echo -e "${BOLD}${GREEN}══════════════════════════════════════════${NC}"
echo ""
echo -e "  Сайт:         ${CYAN}http://${SERVER_IP}${NC}"
echo -e "  Admin-панель: ${CYAN}http://${SERVER_IP}/admin/${NC}"
echo -e "  Пароль admin: ${BOLD}${YELLOW}${ADMIN_PASS}${NC}"
echo ""
echo -e "  Данные сохранены: ${GREEN}/root/ALFA1_access.txt${NC}"
echo ""
echo "  Управление:"
echo "    docker compose -C ${INSTALL_DIR} logs -f      # логи"
echo "    docker compose -C ${INSTALL_DIR} restart app  # перезапуск"
echo "    docker compose -C ${INSTALL_DIR} down         # остановить"
echo ""
