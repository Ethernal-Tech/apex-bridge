#!/usr/bin/env bash
# E2E: syncs a real preprod block over mutual TLS through stunnel in Docker and checks that
# clients without a valid certificate are rejected. Needs Docker and internet.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$root"

port=15522
node_connect="${E2E_NODE_CONNECT:-preprod-node.play.dev.cardano.org:3001}"
name="apex-node-tls-e2e-$$-$RANDOM"
certs=""
foreign=""
conf=""

test_pid=""
cleanup() {
  local rc=$?
  if [ -n "$test_pid" ]; then kill "$test_pid" 2>/dev/null || true; fi
  docker rm -f "$name" >/dev/null 2>&1 || true
  for d in "$certs" "$foreign" "$conf"; do
    [ -z "$d" ] || rm -rf "$d"
  done
  exit "$rc"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

fail() { echo "e2e: $*" >&2; exit 1; }

command -v docker >/dev/null 2>&1 || fail "docker not found in PATH"
command -v openssl >/dev/null 2>&1 || fail "openssl not found in PATH"
command -v timeout >/dev/null 2>&1 || fail "timeout not found in PATH"
docker info >/dev/null 2>&1 || fail "docker daemon is not reachable"
if (exec 3<>"/dev/tcp/127.0.0.1/$port") 2>/dev/null; then fail "port $port is already in use"; fi

parent="$(mktemp -d)"
certs="$parent/certs"
foreign="$parent/foreign"
conf="$parent"
bash "$root/scripts/cardano-node-tls/gen-certs.sh" "$certs" node.test validator1 >/dev/null
bash "$root/scripts/cardano-node-tls/gen-certs.sh" "$foreign" node.test validator1 >/dev/null
sed "s|^connect = .*|connect = $node_connect|" "$root/scripts/cardano-node-tls/stunnel.conf" >"$parent/stunnel.conf"

mnt=/etc/stunnel/node-tls
docker run -d --name "$name" -p "127.0.0.1:$port:5522" \
  -v "$certs/ca.crt:$mnt/ca.crt:ro" \
  -v "$certs/server.crt:$mnt/server.crt:ro" \
  -v "$certs/server.key:$mnt/server.key:ro" \
  -v "$parent/stunnel.conf:/etc/stunnel/stunnel.conf:ro" \
  alpine:3.20 sh -c 'apk add --no-cache stunnel >/dev/null && exec stunnel /etc/stunnel/stunnel.conf' >/dev/null

# Docker's proxy accepts before stunnel listens: probe with a real handshake.
ready=0
for _ in $(seq 1 60); do
  if [ "$(docker inspect -f '{{.State.Running}}' "$name" 2>/dev/null)" != "true" ]; then
    docker logs "$name" >&2 || true
    fail "stunnel container exited early"
  fi
  out="$(cd "$certs" && timeout 5 openssl s_client -connect "127.0.0.1:$port" -servername node.test \
    -CAfile ca.crt -cert validator1.crt -key validator1.key </dev/null 2>&1 || true)"
  # a closed Docker proxy connection also prints "Verify return code: 0", so require the certificate too
  if grep -q 'Verify return code: 0' <<<"$out" && grep -q 'BEGIN CERTIFICATE' <<<"$out"; then ready=1; break; fi
  sleep 1
done
if [ "$ready" != 1 ]; then
  docker logs "$name" >&2 || true
  fail "stunnel not ready on 127.0.0.1:$port within 60 s"
fi

# a TLS 1.2-only client must be refused by stunnel (not time out or fail locally)
rc=0
out="$(cd "$certs" && timeout 5 openssl s_client -tls1_2 -connect "127.0.0.1:$port" -servername node.test \
  -CAfile ca.crt -cert validator1.crt -key validator1.key </dev/null 2>&1)" || rc=$?
if [ "$rc" = 0 ] || [ "$rc" = 124 ] || ! grep -Eq 'alert protocol version|wrong version number|handshake failure' <<<"$out"; then
  echo "$out" >&2
  fail "TLS 1.2 client was not refused with a protocol-version alert (exit $rc)"
fi
echo "e2e: TLS 1.2 client refused"

export E2E_NODE_ADDR="127.0.0.1:$port" E2E_TLS_DIR="$certs" E2E_FOREIGN_TLS_DIR="$foreign"
# run in the background and wait, so SIGINT/SIGTERM reach the trap at once
go test -tags e2e -count=1 -v ./scripts/cardano-node-tls/ &
test_pid=$!
rc=0
wait "$test_pid" || rc=$?
if [ "$rc" != 0 ]; then
  docker logs "$name" >&2 || true
  echo "e2e: go test failed (exit $rc)" >&2
  exit "$rc"
fi
