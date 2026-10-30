# Mutual TLS for the Cardano node connection

- `gen-certs.sh` creates a private CA, a server certificate and one client certificate per validator (OpenSSL, ECDSA P-256).
- `stunnel.conf` terminates TLS 1.3 in front of `cardano-node` and requires a client certificate from that CA.

## Certificates

```bash
./gen-certs.sh <out-dir> <server-dns-name> <client-name>...   # e.g. ./gen-certs.sh certs node.example.org validator1 validator2
```

`<out-dir>` must be missing or empty. Keep `ca.key` offline (it is only needed to issue certificates) and
give each validator its own `<name>.crt`/`<name>.key` plus `ca.crt`. Certificates expire after 397 days,
the CA after 3650; re-issue before that.

The node accepts every certificate the CA has signed and there is no revocation list. To revoke one
(e.g. a leaked validator key), create a new CA with `gen-certs.sh` and roll the new certificates out to
the node and every validator; until then the port 5522 firewall allowlist limits the damage.

## Node side

Copy `ca.crt`, `server.crt`, `server.key` to `/etc/stunnel/node-tls/` and run `stunnel stunnel.conf`.
It listens on a **separate port (5522)** and forwards to the node on `127.0.0.1:3001`, so the node's own
relay port keeps serving Cardano peers. Firewall port 5522 to the validator IPs.

## Validator side

In the Cardano chain config (`networkTls`), with `networkAddress` set to `<server-dns-name>:5522`:

```json
"networkTls": {"caFile": "ca.crt", "certFile": "validator1.crt", "keyFile": "validator1.key"}
```

`serverName` is optional: it defaults to the host of `networkAddress`, which must then match the server
certificate SAN (the `<server-dns-name>` given to `gen-certs.sh`). Set it only when connecting by IP or another name.
`generate-configs cardano-chain` takes the same four values as `--network-tls-ca-file`,
`--network-tls-cert-file`, `--network-tls-key-file` and `--network-tls-server-name`.
