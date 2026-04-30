# CLI

## Name

`crabbox`

One-liner: lease shared remote test boxes, sync local work, run commands, and clean up.

## Usage

```text
crabbox [global flags] <command> [args]
```

Global flags:

```text
-h, --help
--version
--config <path>
--fleet-config <path-or-url>
--profile <name>
--coordinator <url>
--json
--plain
-q, --quiet
-v, --verbose
--no-color
--no-input
```

Primary output goes to stdout. Progress, diagnostics, and errors go to stderr. JSON output is stable enough for scripts.

## Commands

```text
crabbox login
crabbox doctor
crabbox config show
crabbox profiles list
crabbox pool list
crabbox warmup [--provider hetzner|aws] [--profile <name>] [--ttl <duration>]
crabbox run [--provider hetzner|aws] [--profile <name>] [--ttl <duration>] [--class <name>] -- <command...>
crabbox shell [--id <lease-id>]
crabbox sync [--id <lease-id>]
crabbox download --id <lease-id> <remote-path> <local-path>
crabbox stop <lease-id>
crabbox machine list
crabbox machine drain <machine-id>
crabbox machine cleanup [--dry-run]
```

## Common Flows

One-shot run:

```sh
crabbox run --profile openclaw-check -- pnpm check:changed
```

AWS EC2 Spot run:

```sh
crabbox run --provider aws --class beast -- pnpm check:changed
```

Warm a box, then reuse it:

```sh
crabbox warmup --profile openclaw-check --ttl 90m
crabbox run --id cbx_123 -- pnpm test:changed
crabbox shell --id cbx_123
crabbox stop cbx_123
```

Inspect pool:

```sh
crabbox pool list
crabbox pool list --json
```

Debug config:

```sh
crabbox doctor
crabbox config show --plain
```

## `run`

`crabbox run` is the main command.

Behavior:

1. Load config.
2. Acquire a lease unless `--id` is provided.
3. Sync current repo.
4. Run command over SSH.
5. Stream remote output.
6. Heartbeat in the background.
7. Release lease unless `--keep` is set.
8. Exit with the remote command exit code.

Flags:

```text
--id <lease-id>          reuse an existing lease
--provider <name>        hetzner or aws
--profile <name>        profile to run on
--class <name>          machine class override
--type <name>           provider server or instance type override
--ttl <duration>        lease TTL, default from profile
--workdir <path>        remote workdir override
--env <name>            include one env var by exact name
--env-prefix <prefix>   include env vars by prefix
--no-sync               run without syncing
--sync-only             sync and exit
--keep                  keep lease after command exits
--dry-run               print acquisition/sync/run plan only
```

Secrets must not be accepted as flag values. Env forwarding is name-based only.

## Exit Codes

```text
0   success
1   generic Crabbox failure
2   invalid usage or config
3   auth failure
4   no capacity
5   provisioning failure
6   sync failure
7   SSH failure
8   lease expired
10+ remote command exit code when available
```

If the remote command exits with a code, `crabbox run` should return that code unless Crabbox itself failed first.

## Config Files

Repo-local `crabbox.yaml`:

```yaml
version: 1
defaults:
  profile: openclaw-check
  ttl: 90m
sync:
  exclude:
    - node_modules
    - .turbo
    - .git/lfs
env:
  allow:
    - OPENCLAW_*
    - NODE_OPTIONS
```

User config:

```yaml
identity:
  github: steipete
  sshKey: ~/.ssh/id_ed25519
defaults:
  coordinator: https://crabbox.openclaw.ai
```

## Environment Variables

```text
CRABBOX_COORDINATOR
CRABBOX_PROFILE
CRABBOX_CONFIG
CRABBOX_FLEET_CONFIG
CRABBOX_SSH_KEY
CRABBOX_NO_COLOR
CRABBOX_LOG
```

Provider/deploy variables live outside normal CLI operation:

```text
CRABBOX_COORDINATOR
CRABBOX_COORDINATOR_TOKEN
CRABBOX_CLOUDFLARE_API_TOKEN
CRABBOX_CLOUDFLARE_ACCOUNT_ID
CRABBOX_CLOUDFLARE_ZONE_ID
HCLOUD_TOKEN
GITHUB_TOKEN
```

## Output Rules

Human output:

```text
acquiring lease profile=openclaw-check ttl=90m
leased cbx_abc123 machine=hz-ccx33-01 expires=2026-04-30T17:30:00Z
syncing 184 files -> /work/crabbox/cbx_abc123/openclaw
running pnpm check:changed
...
released cbx_abc123
```

JSON output:

```json
{
  "leaseId": "cbx_abc123",
  "machineId": "hz-ccx33-01",
  "state": "released",
  "exitCode": 0
}
```

No progress bars when stdout is not a TTY.
