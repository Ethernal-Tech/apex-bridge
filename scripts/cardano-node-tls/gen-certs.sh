#!/usr/bin/env bash
# Creates or completes a private CA, a server certificate and one client certificate per name, for
# mutual TLS in front of a Cardano node. Existing files are never overwritten: whatever is missing is
# issued by the CA in <out-dir>, and every certificate there is listed with the days it has left.
# Usage: gen-certs.sh <out-dir> <server-dns-name> [client-name...]
# Runs on bash 3.2 (macOS) and later, with OpenSSL 3 or LibreSSL.
set -euo pipefail

usage() {
  echo "usage: ${0##*/} <out-dir> <server-dns-name> [client-name...]" >&2
  exit 2
}
die() {
  echo "${0##*/}: $*" >&2
  exit 1
}
lower() { printf '%s' "$1" | tr '[:upper:]' '[:lower:]'; }

[ "$#" -ge 2 ] || usage

out="$1"
server="$2"
shift 2
clients=("$@")

# ---- validate every argument before anything is read or created ----
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
for c in ${clients[@]+"${clients[@]}"}; do
  [[ "$c" =~ ^[A-Za-z0-9][A-Za-z0-9_-]*$ ]] || { echo "invalid client name: $c" >&2; usage; }
  [ "${#c}" -le 64 ] || { echo "client name longer than 64 characters" >&2; usage; }
  lc="$(lower "$c")"
  case "$lc" in ca | server) echo "client name '$c' is reserved" >&2; usage ;; esac
  case "$seen" in *" $lc "*) echo "duplicate client name (case-insensitive): $c" >&2; usage ;; esac
  seen="$seen$lc "
done

command -v openssl >/dev/null 2>&1 || die "openssl not found in PATH"

parent="$(dirname "$out")"
[ -d "$parent" ] || die "parent directory $parent of $out does not exist"
if [ -e "$out" ] || [ -L "$out" ]; then
  [ -d "$out" ] || die "$out exists and is not a directory"
fi

# ---- read what <out-dir> already holds; refuse anything that is half there ----
# missing <name>: true when neither <name>.crt nor <name>.key exists, false when both do.
missing() {
  if [ -e "$out/$1.crt" ] && [ -e "$out/$1.key" ]; then return 1; fi
  if [ ! -e "$out/$1.crt" ] && [ ! -e "$out/$1.key" ]; then return 0; fi
  die "only one of $1.crt and $1.key exists in $out; restore the other or delete both to issue a new pair"
}

new_ca=0
if [ -e "$out/ca.crt" ]; then
  : # ca.key may be kept offline; it is only needed below when something has to be issued
elif [ -e "$out/ca.key" ]; then
  die "ca.key exists in $out without ca.crt"
else
  for f in "$out"/*.crt "$out"/*.key; do
    if [ -e "$f" ]; then die "ca.crt is missing but $out already holds ${f##*/}; refusing to start a new CA"; fi
  done
  new_ca=1
fi

# every pair already in <out-dir> must be complete, not only the ones named on this run
for f in "$out"/*.crt "$out"/*.key; do
  [ -e "$f" ] || continue
  n="${f##*/}"
  n="${n%.*}"
  if [ "$n" != ca ]; then missing "$n" || true; fi
done

issue=()
if missing server; then
  issue+=(server)
else
  san="$(openssl x509 -in "$out/server.crt" -noout -text | awk '/Subject Alternative Name/ { getline; gsub(/[ \t]/, ""); print }')"
  [ "$(lower "$san")" = "dns:$(lower "$server")" ] ||
    die "server.crt in $out is for '$san', not 'DNS:$server'; delete server.crt and server.key to issue a new one"
fi
for c in ${clients[@]+"${clients[@]}"}; do
  if missing "$c"; then issue+=("$c"); fi
done

if [ "$new_ca" = 0 ] && [ "${#issue[@]}" -gt 0 ] && [ ! -e "$out/ca.key" ]; then
  die "ca.key is not in $out; put it back to issue: ${issue[*]}"
fi

# ---- new files are built in a private work dir and moved into <out-dir> only on success ----
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

# run <command...>: stays quiet on success, shows the command's stderr and stops on failure.
run() {
  local err
  err="$("$@" 2>&1 >/dev/null)" || {
    [ -z "$err" ] || printf '%s\n' "$err" >&2
    die "failed: $*"
  }
}

newkey() { run openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 -out "$1"; }

files=()
if [ "$new_ca" = 1 ]; then
  newkey "$work/ca.key"
  run openssl req -x509 -new -key "$work/ca.key" -days 3650 -subj "/CN=apex-bridge cardano-node CA $(openssl rand -hex 4)" \
    -addext "basicConstraints=critical,CA:TRUE" -addext "keyUsage=critical,keyCertSign,cRLSign" \
    -out "$work/ca.crt"
  ca_crt="$work/ca.crt"
  ca_key="$work/ca.key"
  files+=(ca.key ca.crt)
else
  ca_crt="$out/ca.crt"
  ca_key="$out/ca.key"
fi

# sign <name> <common-name> <extensions>
sign() {
  newkey "$work/$1.key"
  run openssl req -new -key "$work/$1.key" -subj "/CN=$2" -out "$work/$1.csr"
  printf '%s\n' "basicConstraints=critical,CA:FALSE" "keyUsage=critical,digitalSignature" "$3" >"$work/$1.ext"
  run openssl x509 -req -in "$work/$1.csr" -CA "$ca_crt" -CAkey "$ca_key" \
    -CAserial "$work/ca.srl" -CAcreateserial -days 397 -extfile "$work/$1.ext" -out "$work/$1.crt"
  files+=("$1.key" "$1.crt")
}

for n in ${issue[@]+"${issue[@]}"}; do
  if [ "$n" = server ]; then
    sign server "$server" "$(printf 'subjectAltName=DNS:%s\nextendedKeyUsage=serverAuth' "$server")"
  else
    sign "$n" "$n" "extendedKeyUsage=clientAuth"
  fi
done

if [ ! -e "$out" ]; then
  mkdir "$out"
  chmod 700 "$out"
  created=1
fi
for f in ${files[@]+"${files[@]}"}; do
  mv "$work/$f" "$out/$f"
  moved+=("$f")
  case "$f" in *.crt) chmod 644 "$out/$f" ;; esac
done
done_ok=1

# ---- list every certificate in <out-dir> with the days it has left ----
now="$(date -u +%s)"
report() {
  local end epoch state
  end="$(openssl x509 -in "$out/$1" -noout -enddate)"
  end="${end#notAfter=}"
  # GNU date first, then BSD date (macOS)
  epoch="$(date -u -d "$end" +%s 2>/dev/null || date -j -u -f '%b %e %T %Y %Z' "$end" +%s)"
  state="existing"
  case " ${moved[*]-} " in *" $1 "*) state="new" ;; esac
  if [ "$epoch" -le "$now" ]; then
    printf '%-24s %-8s EXPIRED on %s\n' "$1" "$state" "$end"
  else
    printf '%-24s %-8s valid %d more days (until %s)\n' "$1" "$state" "$(((epoch - now) / 86400))" "$end"
  fi
}

report ca.crt
report server.crt
for f in "$out"/*.crt; do
  f="${f##*/}"
  case "$f" in ca.crt | server.crt) ;; *) report "$f" ;; esac
done

if [ "${#issue[@]}" -eq 0 ] && [ "$new_ca" = 0 ]; then
  echo "nothing to issue in $out"
else
  echo "issued ${#moved[@]} new file(s) in $out"
fi
if [ -e "$out/ca.key" ]; then echo "keep ca.key offline; it is only needed to issue certificates"; fi
