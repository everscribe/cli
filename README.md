# Everscribe CLI

`es` is the command-line interface for [Everscribe](https://everscribe.io). It supports project management, ingest API key administration, and live tailing of events — all authenticated by a personal access token (PAT) you mint via the browser.

## Install

```sh
go install github.com/everscribe/cli/cmd/es@latest
```

## Quick start

```sh
# Authenticate via the device authorization flow (RFC 8628).
es auth login

# List existing projcts
es projects list
# Or create a new project
es projects create --name my-app

# Set your config to the project of your choice so you don't have to append --project <project-uuid> for everything.
es projects use <project-id-from-the-output-above>
```

## Commands

### Auth

| Command | What it does |
|---|---|
| `es auth login [--no-browser]` | Runs the [RFC 8628](https://datatracker.ietf.org/doc/html/rfc8628) device authorization flow: shows a one-time code, opens the verification URL in your browser, and polls until you sign in and confirm. The resulting PAT is persisted to `~/.config/everscribe/pat.json` (mode 0600). `--no-browser` skips opening the browser (useful over SSH — copy the URL to a browser on any device). |
| `es auth logout` | Revokes the PAT server-side and removes the local session file. Tolerant of stale tokens — local cleanup always runs. |
| `es auth whoami [--format ...]` | Prints the user identity associated with the saved PAT. Never echoes the token itself. |

The login flow looks like:

```
$ es auth login
First copy your one-time code:

    PHJK-7M3X

Then open this URL in your browser (we'll try to open it for you):

    https://everscribe.io/cli/verify

Waiting for authorization... (press Ctrl+C to cancel)
Logged in as alice@example.com.
```

The PAT lifetime is server-controlled (90 days by default). To mint a token with a different expiry, use the [Developer Settings](https://everscribe.io/settings/developer) page in the UI.

### Projects

| Command | What it does |
|---|---|
| `es projects list` | Your projects. |
| `es projects get <id>` | One project. |
| `es projects create --name <name>` | Create. |
| `es projects update <id> --name <new-name>` | Rename. |
| `es projects delete <id>` | Server-side soft delete. |
| `es projects use <id>` | Set the default project so other commands can omit `--project`. Validated by `GET /v1/projects/<id>` before being saved, so a typo'd ID errors immediately. |
| `es projects current` | Print the saved default project ID. |

### API keys

Project-scoped ingest credentials (`evs_…`). Distinct from the PAT the CLI itself uses.

`--project` is optional once `es projects use <id>` has been run; it falls back to the saved default.

| Command | What it does |
|---|---|
| `es keys list [--project <id>]` | Active keys for a project. |
| `es keys create [--project <id>] --name <key-name>` | Mint a new key. The plaintext is shown **once** — copy it before closing the terminal. |
| `es keys revoke [--project <id>] --key <key-id>` | Revoke a key. |

### Events

`--project` is optional once `es projects use <id>` has been run.

| Command | What it does |
|---|---|
| `es events list [--project <id>] [filters] [--limit N] [--all]` | Page through events newest-first. With `--all`, follows `next_cursor` through every page. |
| `es events watch [--project <id>] [filters] [--interval 3s]` | Polls and prints new events as they arrive, deduped by event ID. |
| `es events describe <event-id> [--project <id>] [--format yaml\|json]` | Full event with nested `actor` / `target` / `metadata` / `origin` / `result` / `change` decoded inline. YAML default; no table form. |
| `es events diff <event-id> [--project <id>]` | Unified `+`/`-` diff of `change.before` vs `change.after`. Errors clearly when the event has no change record (most events don't). |

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

| Path | Contents | Written by |
|---|---|---|
| `~/.config/everscribe/pat.json` (mode 0600) | Saved PAT, user identity, expiry. | `es auth login` |
| `~/.config/everscribe/config.json` (mode 0600) | Non-sensitive preferences (currently just `default_project_id`). | `es projects use` |

The two files are kept separate so `es auth logout` doesn't clobber preferences like the saved default project — log back in and `es events list` keeps working without re-typing the project ID.

The API host is hardcoded to `https://api.everscribe.io`; the verification URL for `auth login` comes back from the API itself, so the CLI never needs to know the UI host. `EVERSCRIBE_API_URL_OVERRIDE` retargets the API for local development against a self-hosted monorepo.

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
internal/config/       # pat.json + config.json persistence, ResolveProjectID helper
internal/output/       # table / JSON / YAML renderers, color, age formatter
internal/types/        # wire types mirrored from the monorepo
internal/testutil/     # shared CLI test helpers
```

The device authorization flow has its server-side counterpart in the
[everscribe/monorepo](https://github.com/everscribe/monorepo):
`POST /v1/cli/device-codes` and `POST /v1/cli/device-tokens` on the API
(unauthenticated), `POST /v1/cli/device-codes/approve` (authenticated),
and `/cli/verify` on the UI.
