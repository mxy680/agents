#!/usr/bin/env bash
# Server setup script for agents.markshteyn.com
# Run as root on the Hetzner server (178.156.139.74)
set -euo pipefail

echo "=== Agents Portal Setup ==="

# 1. Create deployment directory
echo "[1/6] Creating /opt/agents..."
mkdir -p /opt/agents
chown deploy:deploy /opt/agents

# 2. Install systemd services
echo "[2/6] Installing systemd services..."
cp /opt/agents/deploy/agents-portal.service /etc/systemd/system/
cp /opt/agents/deploy/agents-ws.service /etc/systemd/system/
cp /opt/agents/deploy/agents-cron.service /etc/systemd/system/
cp /opt/agents/deploy/agents-cron.timer /etc/systemd/system/

# 3. Update Caddy config (preserve reef docker-compose, just update Caddyfile)
echo "[3/6] Updating Caddyfile..."
cp /opt/agents/deploy/Caddyfile /opt/reef/Caddyfile

# 4. Add extra_hosts to reef docker-compose if not present
echo "[4/6] Updating docker-compose for host.docker.internal..."
if ! grep -q "host.docker.internal" /opt/reef/docker-compose.yml; then
    cp /opt/agents/deploy/docker-compose.yml /opt/reef/docker-compose.yml
    echo "  Updated docker-compose.yml"
else
    echo "  Already has host.docker.internal, skipping"
fi

# 5. Reload Caddy
echo "[5/6] Reloading Caddy..."
cd /opt/reef
docker compose up -d --force-recreate caddy

# 6. Enable and start services
echo "[6/6] Enabling systemd services..."
systemctl daemon-reload
systemctl enable agents-portal agents-ws agents-cron.timer
echo "  Services enabled. Start with:"
echo "    systemctl start agents-portal"
echo "    systemctl start agents-ws"
echo "    systemctl start agents-cron.timer"

echo ""
echo "=== Setup complete ==="
echo "Remaining manual steps:"
echo "  1. Authenticate Doppler: su - deploy -c 'doppler login'"
echo "  2. Build portal: cd /opt/agents/portal && doppler run -p agents -c prd -- npm run build"
echo "  3. Start services: systemctl start agents-portal agents-ws agents-cron.timer"
