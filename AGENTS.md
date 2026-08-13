# AGENTS.md

Go CLI that scans Tor relay lists and outputs working bridges as a torrc config.

## Build and verify

- Build: `go build -o tor-scanner cmd/main.go` (binary is gitignored — never commit it).
- Entrypoint is `cmd/main.go` (package `main` lives under `cmd/`, not repo root).
- There are no tests, linter, or Makefile. `go vet ./...` is the only sanity check.

## Architecture

- `cmd/main.go` — CLI (leaanthony/clir), flag parsing, threads/chunking, collects results until `--count` reached. Workers are tracked with a `sync.WaitGroup`; `alive` channel closes when all chunks are scanned, so the run ends gracefully (warning + partial output) instead of deadlocking when fewer than `--count` bridges match. Errors out if nothing was found.
- `internal/checker/relays.go` — fetches/parses input. `Relays` struct fields: `fingerprint`, `or_addresses`, `country`, plus `type`, `extra` for obfs4 lines and `ping` (measured dial time, ms). `BridgeLine()` renders the torrc output line.
- `internal/checker/check.go` — raw TCP dial to `or_addresses[0]` with the `--ping` timeout; country filter is a comma-split, case-sensitive match. `--type obfs4` sources have no `country` field, so country is resolved via reverse DNS (PTR) with a `sync.Map` cache. Console `Found` lines print `- <Country> <ping>ms`.
- Output file is torrc lines: `Bridge <ip:port> <fingerprint>` (onionoo) or `Bridge obfs4 <ip:port> <fingerprint> cert=... iat-mode=...` (obfs4) + `UseBridges 1`.
- `example.json` is a sample of the onionoo JSON input format.

## Gotchas

- `--type` selects the input format: empty = Onionoo JSON (ValdikSS tor-onionoo-mirror raw URL), `obfs4` = line-based list from scriptzteam Tor-Bridges-Collector (`bridges-obfs4`). Default `--input` is resolved at runtime based on `--type` (constants in `cmd/main.go:29`). Console shows the type with a `vanilla` label when `--type` is empty.
- `--input` treats anything not prefixed with `https://` as a local file path — plain `http://` URLs will fail as file reads.
- PTR country detection: IPs without a reverse-DNS record never match `--country`; a 2s lookup timeout applies per unique IP.
- Worker count is capped at `len(relays)` and floored at 1 (`cmd/main.go:70`) so `--threads` larger than the relay count no longer hangs; chunks are distributed to cover all relays (remainder spread over the first workers).
- Version string is hardcoded (`v0.0.3` at `cmd/main.go:35`) — bump it manually when releasing.
- Releases build 5 cross-platform binaries from `cmd/main.go` on `v*` tags (`.github/workflows/release.yaml`); Go versions: go.mod `1.25.4`, CI `1.25.6`.

## Docs

Docs are bilingual: Russian is primary (`README.md`, `docker/howto.md`), English mirrors are `README-en.md` and `docker/howto-en.md`. Update both when touching docs. Renovate bumps dependencies automatically.
