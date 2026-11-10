# Mutual TLS for the Cardano node connection

- `gen-certs.sh` creates a private CA, a server certificate and one client certificate per validator (OpenSSL, ECDSA P-256).
- `stunnel.conf` terminates TLS 1.3 in front of `cardano-node` and requires a client certificate from that CA.

## Certificates

```bash
./gen-certs.sh <out-dir> <server-dns-name> [client-name...]   # e.g. ./gen-certs.sh certs node.example.org validator1 validator2
```

The script never overwrites a file. Each run issues only what `<out-dir>` is missing and then lists every
certificate there with the days it has left, so it is safe to run again:

- **First run**: creates `ca.crt`/`ca.key`, `server.crt`/`server.key` and one `<name>.crt`/`<name>.key` per client.
- **Add a validator**: run it again with the new name; the existing CA signs it. `ca.key` must be in `<out-dir>` for that.
- **Check expiry**: run it with no client names; it issues nothing and prints the days left.
- **Re-issue a certificate**: delete its `.crt` and `.key` and run again; the CA stays the same.

It stops without changing anything when only one file of a pair exists, when `server.crt` was issued for
another `<server-dns-name>`, or when `ca.crt` is missing but other certificates are present.

Keep `ca.key` offline (it is only needed to issue certificates) and give each validator its own
`<name>.crt`/`<name>.key` plus `ca.crt`. Certificates expire after 397 days, the CA after 3650; re-issue
before that.

## Node side

Copy `ca.crt`, `server.crt`, `server.key` to `/etc/stunnel/node-tls/` and run `stunnel stunnel.conf`.
It listens on a **separate port (5522)** and forwards to the node on `127.0.0.1:3001`, so the node's own
relay port keeps serving Cardano peers. Firewall port 5522 to the validator IPs.

## Validator side

In the Cardano chain config (`networkTls`), with `networkAddress` set to `<server-dns-name>:5522`:

```json
"networkTls": {"caFile": "ca.crt", "certFile": "validator1.crt", "keyFile": "validator1.key", "serverName": "node.example.org"}
```

`serverName` must match the server certificate SAN (the `<server-dns-name>` given to `gen-certs.sh`).
`generate-configs cardano-chain` takes the same four values as `--network-tls-ca-file`,
`--network-tls-cert-file`, `--network-tls-key-file` and `--network-tls-server-name`.
