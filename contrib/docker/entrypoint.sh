#!/usr/bin/env sh

sysctl net.ipv6.conf.all.disable_ipv6=0 || true
set -e

CONF_DIR="/etc/ruvchain"

if [ ! -f "$CONF_DIR/config.conf" ]; then
  echo "generate $CONF_DIR/config.conf"
  ruvchain --genconf > "$CONF_DIR/config.conf"
fi

ruvchain --useconf < "$CONF_DIR/config.conf"
exit $?
