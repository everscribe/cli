# Everscribe CLI

`es` is the command-line interface for [Everscribe](https://everscribe.io). It supports project management, ingest API key administration, and live tailing of events — all authenticated by a personal access token (PAT) you mint via the browser.

## Install

```sh
go install github.com/everscribe/cli/cmd/es@latest
```

## Quick start

```sh
# Authenticate via browser-loopback (mints a PAT scoped to your user).
es auth login

# Create a project, mint an ingest key for it.
es projects create --name my-app
es keys create --project <project-id> --name production

# Tail events as they arrive.
es events watch --project <project-id>
```

## Commands

### Auth

| Command | What it does |
|---|---|
| `es auth login [--expires-in-days N] [--no-browser]` | Opens the browser to the Everscribe UI, you authorize, the resulting PAT is persisted to `~/.config/everscribe/pat.json` (mode 0600). `--no-browser` prints the URL instead of opening it (useful over SSH). |
| `es auth logout` | Revokes the PAT server-side and removes the local session file. Tolerant of stale tokens — local cleanup always runs. |
| `es auth whoami [--format ...]` | Prints the user identity associated with the saved PAT. Never echoes the token itself. |

### Projects

| Command | What it does |
|---|---|
| `es projects list` | Your projects. |
| `es projects get <id>` | One project. |
| `es projects create --name <name>` | Create. |
| `es projects update <id> --name <new-name>` | Rename. |
| `es projects delete <id>` | Server-side soft delete. |

### API keys

Project-scoped ingest credentials (`evs_…`). Distinct from the PAT the CLI itself uses.

| Command | What it does |
|---|---|
| `es keys list --project <id>` | Active keys for a project. |
| `es keys create --project <id> --name <key-name>` | Mint a new key. The plaintext is shown **once** — copy it before closing the terminal. |
| `es keys revoke --project <id> --key <key-id>` | Revoke a key. |

### Events

| Command | What it does |
|---|---|
| `es events list --project <id> [filters] [--limit N] [--all]` | Page through events newest-first. With `--all`, follows `next_cursor` through every page. |
| `es events watch --project <id> [filters] [--interval 3s]` | Polls and prints new events as they arrive, deduped by event ID. |
| `es events describe <event-id> --project <id> [--format yaml\|json]` | Full event with nested `actor` / `target` / `metadata` / `origin` / `result` / `change` decoded inline. YAML default; no table form. |
| `es events diff <event-id> --project <id>` | Unified `+`/`-` diff of `change.before` vs `change.after`. Errors clearly when the event has no change record (most events don't). |

Shared filter flags on `list` and `watch`:
`--since`, `--before`, `--action`, `--actor`, `--actor-type`, `--target-type`, `--tenant`.

## Output formats

Most commands accept `--format table | json | yaml` (default `table`). Special cases:

- `es events describe` defaults to `yaml`; `--format json` for JSON.
- `es events diff` has no `--format` — always renders a unified diff.
- `es events watch` table mode uses fixed-width columns so rows stay aligned across rolling appends.

JSON output for list endpoints is a bare array (no `{"projects": [...]}` envelope); JSON for single-resource endpoints is a bare object. Friendlier for piping through `jq` and `yq`.

ANSI color is auto-enabled when stdout is a TTY and disabled when [`NO_COLOR`](https://no-color.org) is set.

## Time filters

`--since` and `--before` accept either:

- An RFC3339 timestamp — `2026-05-09T12:00:00Z`
- A Go duration interpreted as "this far ago" — `30m`, `1h`, `24h`
- `Nd` for days — `7d`, `30d`

## Configuration

| Path | Contents |
|---|---|
| `~/.config/everscribe/pat.json` (mode 0600) | Saved PAT, user identity, expiry. Written by `es auth login`. |

The API and UI hosts are hardcoded to `https://api.everscribe.io` and `https://everscribe.io`. Two undocumented environment variables — `EVERSCRIBE_API_URL_OVERRIDE` and `EVERSCRIBE_UI_URL_OVERRIDE` — exist only for the test suite.

## Development

```sh
make test     # go test ./...
make fmt      # goimports with -local github.com/everscribe/cli grouping
make install  # go install ./cmd/es
```

Layout:

```
cmd/es/                # CLI entrypoint
internal/client/       # HTTP client: bearer auth, error envelope, resource methods
internal/cmds/         # cobra commands (auth, projects, keys, events)
internal/config/       # PAT persistence (~/.config/everscribe/pat.json)
internal/output/       # table / JSON / YAML renderers, color, age formatter
internal/types/        # wire types mirrored from the monorepo
internal/testutil/     # shared CLI test helpers
```

The browser-loopback login flow has its own server-side counterpart in the
[everscribe/monorepo](https://github.com/everscribe/monorepo): `POST /v1/pats/cli`
on the API and `/cli/auth` on the UI.
