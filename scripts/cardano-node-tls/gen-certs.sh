#!/usr/bin/env bash
# Creates a private CA, a server certificate and one client certificate per name, for mutual TLS
# in front of a Cardano node. Usage: gen-certs.sh <out-dir> <server-dns-name> <client-name>...
set -euo pipefail

usage() {
  echo "usage: ${0##*/} <out-dir> <server-dns-name> <client-name>..." >&2
  exit 2
}

[ "$#" -ge 3 ] || usage

out="$1"
server="$2"
shift 2
clients=("$@")

# ---- validate every argument before anything is created ----
[ -n "$out" ] || usage

[[ "$server" =~ ^[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?$ ]] || { echo "invalid server name: $server" >&2; usage; }
[ "${#server}" -le 64 ] || { echo "server name longer than 64 characters" >&2; usage; }
[[ ! "$server" =~ ^[0-9]+(\.[0-9]+){3}$ ]] || { echo "server name must be a DNS name, not an IP address" >&2; usage; }
IFS='.' read -r -a labels <<<"$server"
for label in "${labels[@]}"; do
  [[ "$label" =~ ^[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?$ ]] || { echo "invalid label in server name: $server" >&2; usage; }
  [ "${#label}" -le 63 ] || { echo "label longer than 63 characters in server name" >&2; usage; }
done

seen=" "
for c in "${clients[@]}"; do
  [[ "$c" =~ ^[A-Za-z0-9][A-Za-z0-9_-]*$ ]] || { echo "invalid client name: $c" >&2; usage; }
  [ "${#c}" -le 64 ] || { echo "client name longer than 64 characters" >&2; usage; }
  lower="${c,,}"
  case "$lower" in ca | server) echo "client name '$c' is reserved" >&2; usage ;; esac
  case "$seen" in *" $lower "*) echo "duplicate client name (case-insensitive): $c" >&2; usage ;; esac
  seen="$seen$lower "
done

command -v openssl >/dev/null 2>&1 || { echo "openssl not found in PATH" >&2; exit 1; }

parent="$(dirname "$out")"
[ -d "$parent" ] || { echo "parent directory $parent of $out does not exist" >&2; exit 1; }
if [ -e "$out" ] || [ -L "$out" ]; then
  [ -d "$out" ] || { echo "$out exists and is not a directory" >&2; exit 1; }
  [ -z "$(ls -A "$out")" ] || { echo "$out is not empty, refusing to overwrite" >&2; exit 1; }
fi

# ---- everything is built in a private work dir and moved into <out-dir> only on success ----
umask 077
work="$(mktemp -d)"
created=0
moved=()
done_ok=0
cleanup() {
  rm -rf "$work"
  if [ "$done_ok" != 1 ]; then
    if [ "$created" = 1 ]; then
      rm -rf "$out"
    else
      for f in ${moved[@]+"${moved[@]}"}; do rm -f "$out/$f"; done
    fi
  fi
}
trap cleanup EXIT

newkey() { openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 -out "$1" 2>/dev/null; }

ca_cn="apex-bridge cardano-node CA $(openssl rand -hex 4)"
newkey "$work/ca.key"
openssl req -x509 -new -key "$work/ca.key" -days 3650 -subj "/CN=$ca_cn" \
  -addext "basicConstraints=critical,CA:TRUE" -addext "keyUsage=critical,keyCertSign,cRLSign" \
  -out "$work/ca.crt"

# sign <name> <common-name> <extensions>
sign() {
  newkey "$work/$1.key"
  openssl req -new -key "$work/$1.key" -subj "/CN=$2" -out "$work/$1.csr"
  printf '%s\n' "basicConstraints=critical,CA:FALSE" "keyUsage=critical,digitalSignature" "$3" >"$work/$1.ext"
  openssl x509 -req -in "$work/$1.csr" -CA "$work/ca.crt" -CAkey "$work/ca.key" \
    -CAserial "$work/ca.srl" -CAcreateserial -days 397 -extfile "$work/$1.ext" -out "$work/$1.crt" 2>/dev/null
}

sign server "$server" "$(printf 'subjectAltName=DNS:%s\nextendedKeyUsage=serverAuth' "$server")"
for c in "${clients[@]}"; do
  sign "$c" "$c" "extendedKeyUsage=clientAuth"
done

if [ ! -e "$out" ]; then
  mkdir "$out"
  created=1
fi
chmod 700 "$out"
files=(ca.key ca.crt server.key server.crt)
for c in "${clients[@]}"; do files+=("$c.key" "$c.crt"); done
for f in "${files[@]}"; do
  mv "$work/$f" "$out/$f"
  moved+=("$f")
  case "$f" in *.crt) chmod 644 "$out/$f" ;; esac
done
done_ok=1
echo "created CA, server and ${#clients[@]} client certificate(s) in $out (keep ca.key offline)"
