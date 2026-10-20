# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Go (1.24) implementation of the Apex Bridge off-chain validator components. A single binary
(`main.go` → `cli.NewRootCommand()`) exposes cobra subcommands: running the validator components
and the relayer, plus operator tooling (wallet creation, chain registration, config generation,
contract/program deployment, bridge admin, sending test bridging txs). README.md holds the
operator recipes for each command.

## Commands

```shell
make build        # go build -o apex-bridge with versioning ldflags (needs git)
make lint         # golangci-lint run --config .golangci.yml
make unit-test    # go test -tags testenv -race -shuffle=on ... over `go list ./... | grep -v e2e`
make swag-generate  # regenerate docs/oracle-api swagger from api/api.go
```

Single test: `go test -tags testenv -race -run TestName ./batcher/batcher/`

- The `testenv` build tag matters: `testenv/batcher_testenv.go` (tag on) replaces
  `testenv/batcher.go` (tag off) to change batcher TTL behaviour for tests. Run tests with the tag
  as CI does.
- Some tests need external tools CI installs: cardano-cli/cardano tools, `solc`/`solcjs`
  (see `.github/workflows/unit-test.yml` and `setup-ci.sh`).
- `github.com/Ethernal-Tech/solana-infrastructure` is private; fetch with
  `GOPRIVATE=github.com/Ethernal-Tech/solana-infrastructure` and a git `insteadOf` SSH rewrite
  (README "How to go get private repo").

## Lint rules

`.golangci.yml` (golangci-lint v1 config) enables ruleguard with `gorules/rules.go`. Project
rules include: no manual `panic(...)`, no `0777`/`os.ModePerm` file modes, and `time.Now()` must be
`time.Now().UTC()`. New code must pass these.

## Architecture

Chains: Cardano-family (Prime, Vector, other Cardano chains), EVM (Nexus) and Solana. The
"bridge" chain is an EVM chain running the Bridge smart contract where validators submit claims
and get batches agreed. Two run modes (`common.VCRunMode`): `reactor` and `skyline`
(`common.ReactorMode` / `common.SkylineMode`); many code paths branch on it.

Flow of a bridging request:

1. **Oracles** (`oracle_cardano`, `oracle_eth`, `oracle_solana`, shared code in `oracle_common`)
   index source-chain txs (Cardano via `cardano-infrastructure` indexer, EVM via
   `blockchain-event-tracker`, Solana via `solana-infrastructure`), parse bridging metadata,
   validate it in `processor/` (per-tx-type processors under a common `txs_processor`) and
   submit claims to the bridge contract through `oracle_common/bridge` (`BridgeSubmitter`).
2. **Batchers** (`batcher/`): `batcher_manager` runs one batcher per destination chain;
   `batcher/batcher/*_chain_operations.go` hold per-chain logic for building and signing batch
   txs (Cardano multisig, EVM BLS, Solana) from confirmed claims.
3. **Relayer** (`relayer/`): picks up signed batches from the bridge contract and submits them
   to the destination chain (`relayer_manager` + per-chain operations, bbolt state in
   `database_access`).

`validatorcomponents/validatorcomponents/validatorcomponents.go` (`NewValidatorComponents`) is the
composition root: it opens the bbolt DBs, builds the bridge contract wrappers (`eth/`), the
oracles, `batcher_manager`, a `RelayerImitator` (tracks batch execution without being the
relayer), the bridging-request state manager, the bridging-addresses manager/coordinator
(`bridging_addresses_manager`, `bridging_addresses_coordinator`: multiple bridging addresses per
Cardano chain) and the optional HTTP API (`api/`, controllers in
`validatorcomponents/api`). The relayer runs as a separate command (`cli/relayer`).

## Layout

- `common/`: shared types, bridging metadata (de)serialization, chain IDs, secrets, request states.
- `cardano/`, `eth/`, `solana/`: chain helpers, tx builders, contract wrappers.
- `contractbinding/`: abigen bindings for Bridge and Gateway contracts (regeneration steps in
  README).
- `exchange_rate_service/`: fetchers for exchange rates used in fee calculations.
- `telemetry/`: Prometheus/OTel metrics; `validatorcomponents/.../telemetry_worker.go` reports
  chain/balance metrics.
- `cli/generateconfigs`: produces the validator-components and relayer JSON configs from flags
  (per chain type params files).

## Stack

Languages: Go.

- Go 1.24.1 (toolchain go1.24.2)

## Verification

- Before a push: `make lint` (the fast gate).
- Before a PR: `make lint build` (the full gate).
- Build (go): `go vet ./...` (or `golangci-lint run`), `go test ./...` and `govulncheck ./...`,
  each with `GOTOOLCHAIN=local`; a `net/http` service also gets API E2E tests through
  `httptest` against a real database (testcontainers-go), then scripted HTTP requests.
- Then ask Claude to run its verify skill: a Scenario, Layer, Outcome table from real output.
- Results come from real output; anything not run is reported as not run.

## Rules

- Conventional commits, `<type>(<scope>): <summary>`. Why: the squash commit takes the PR
  title, and a readable history is the project's record of changes.
- Claude verifies a task, then leaves its changes unstaged and names the files; the engineer
  commits them. Why: the engineer owns every commit.
- New environment variables go into `.env.example` with a name and no value. Why: secrets stay
  out of git and out of the model's context.
- Adopt the newest stable release of a dependency; fix the code instead of pinning back. Why:
  old versions pile up security fixes that are hard to catch up on.

## Don't

- Don't commit, push or open a PR unless asked; never push to `main` or force-push.
- Don't merge with a red or missing required check.
- Don't read, print or commit `.env` files.
- Don't describe unverified behaviour as verified.
