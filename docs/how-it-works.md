# How Crabbox Works

Read when:

- you are new to Crabbox;
- you want the end-to-end mental model;
- you need to know which component owns a behavior before changing code.

Crabbox is a remote testbox system. The local CLI keeps the developer workflow simple: lease a machine, sync the dirty checkout, run a command, stream output, and clean up. The Cloudflare broker keeps shared capacity safe: it owns provider credentials, lease state, cleanup, usage, and cost guardrails.

## The Pieces

```text
local machine
  crabbox CLI
  repo checkout
  per-lease SSH key
    |
    | HTTPS JSON API
    v
Cloudflare Worker
  Fleet Durable Object
  provider credentials
  lease and usage state
    |
    | provider API
    v
Hetzner Cloud or AWS EC2 Spot
  Ubuntu runner
  SSH on port 2222
  /work/crabbox/<lease>/<repo>

local machine
    |
    | SSH + rsync
    v
runner
```

The CLI talks to the broker over HTTPS, then talks directly to the leased runner over SSH and rsync. The runner does not need broker credentials.

## Ownership

CLI owns:

- config loading and command flags;
- per-lease SSH key creation and local key lookup;
- SSH readiness waits;
- Git seeding and rsync;
- sync fingerprints and sanity checks;
- remote command execution;
- stdout/stderr streaming;
- heartbeat and release calls while it is alive.

Broker owns:

- authentication and owner/org attribution;
- serialized lease state in the Fleet Durable Object;
- provider credentials;
- machine creation and deletion;
- lease expiry;
- pool, status, and inspect data;
- usage statistics;
- active lease and monthly spend guardrails.

Provider owns raw compute. Today that means Hetzner Cloud servers or AWS EC2 Spot instances.

Runner owns nothing durable. It is an Ubuntu machine prepared by cloud-init with SSH, Node 24, pnpm, Docker, Git, rsync, build tools, and `/work/crabbox`.

## What Happens On `crabbox run`

1. CLI loads config from flags, env, repo config, user config, and defaults.
2. CLI creates a temporary lease ID and a per-lease SSH key.
3. CLI sends `POST /v1/leases` to the broker with class, provider, TTL, bootstrap options, and the SSH public key.
4. Worker authenticates the request and forwards it to the Fleet Durable Object.
5. Durable Object checks active-lease limits and monthly reserved spend limits.
6. Worker asks the provider for live pricing when available, unless explicit cost rates are configured.
7. Durable Object reserves the worst-case TTL cost for the month.
8. Worker provisions a Hetzner server or AWS EC2 Spot instance.
9. Worker stores the lease and returns host, SSH user, port, work root, expiry, and lease ID.
10. CLI moves the local key directory if the broker returned a final lease ID different from the provisional one.
11. CLI waits for SSH and `crabbox-ready`.
12. CLI seeds remote Git when possible.
13. CLI compares sync fingerprints and skips rsync when nothing changed.
14. CLI rsyncs the dirty checkout into `/work/crabbox/<lease>/<repo>`.
15. CLI runs sync sanity checks and hydrates the configured base ref.
16. CLI starts heartbeats in the background.
17. CLI runs the command over SSH and streams output.
18. CLI releases the lease unless `--keep` is set.
19. Broker deletes the non-kept machine and provider-side lease keys.

If bootstrap never reaches SSH readiness for a fresh non-kept lease, `crabbox run` can retry once with a new machine. It does not duplicate commands on kept or explicitly reused leases.

## Warm Machines And Reuse

`crabbox warmup` follows the same lease creation path, then keeps the box ready for later use. Reuse is explicit:

```sh
crabbox warmup --profile project-check --idle-timeout 90m
crabbox run --id cbx_... -- pnpm test:changed
crabbox ssh --id cbx_...
crabbox stop cbx_...
```

Heartbeats extend brokered leases while the CLI is using them. If a lease goes stale, the Durable Object alarm expires it and deletes non-kept resources.

## Brokered Path vs Direct Provider Path

Brokered path is normal operation:

```text
CLI -> Cloudflare Worker -> Durable Object -> provider API
CLI -> runner over SSH/rsync
```

Use it when maintainers or agents share infrastructure. It keeps provider secrets out of local machines and gives centralized cleanup, usage, and cost control.

Direct provider path is a debug fallback:

```text
CLI -> provider API
CLI -> runner over SSH/rsync
```

It needs local provider credentials such as AWS credentials or `HCLOUD_TOKEN`. Direct mode has no central usage history and no brokered heartbeat.

## Auth And Identity

The broker accepts bearer-token automation and can also use Cloudflare Access identity when present.

Bearer-token CLI requests send:

```text
Authorization: Bearer <token>
X-Crabbox-Owner: <email>
X-Crabbox-Org: <org>
```

Owner comes from `CRABBOX_OWNER`, Git email env, or `git config user.email`. `CRABBOX_ORG` sets the org. Cloudflare Access email wins when available.

## Sync Model

Crabbox sync is intentionally local-first. It does not require a clean checkout.

The sync layer:

- seeds remote Git from the configured origin/base ref when possible;
- overlays local dirty files with rsync;
- can use checksum mode;
- can skip no-op syncs with fingerprints;
- excludes heavy project directories from repo config;
- checks for suspicious mass tracked deletions;
- hydrates base-ref history for changed-test workflows.

This gives agents and maintainers the same local loop: edit locally, run remotely.

## Cost And Usage

The broker tracks two costs:

```text
estimatedUSD   elapsed runtime cost
reservedUSD    worst-case TTL cost reserved before provisioning
```

Hourly price source order:

```text
1. CRABBOX_COST_RATES_JSON explicit override
2. provider live pricing
3. built-in fallback rates
```

AWS pricing comes from EC2 Spot price history. Hetzner pricing comes from server-type hourly prices and is converted with `CRABBOX_EUR_TO_USD`.

`crabbox usage` queries `GET /v1/usage` and can group by user, org, provider, and server type.

## Failure And Cleanup

Crabbox assumes failures are normal:

- CLI can crash;
- SSH can disconnect;
- cloud-init can fail;
- provider calls can partially succeed;
- Cloudflare can retry requests;
- machines can outlive the local process.

The design response:

- one Durable Object serializes fleet decisions;
- lease creation is idempotent where practical;
- provider resources are tagged/labeled;
- release is safe to call repeatedly;
- stale leases expire by alarm;
- direct cleanup is conservative;
- brokered cleanup is broker-owned.

`crabbox cleanup` refuses to sweep provider resources when a coordinator is configured, because brokered cleanup belongs to the Durable Object.

## Where To Go Next

- [Architecture](architecture.md): component model, API, state, and failure model.
- [Orchestrator](orchestrator.md): broker responsibilities, lifecycle, cleanup, cost, and usage.
- [CLI](cli.md): command surface, config, output, and exit codes.
- [Features](features/README.md): one page per feature area.
- [Commands](commands/README.md): one page per command.
- [Infrastructure](infrastructure.md): Cloudflare, DNS, Hetzner, and AWS setup.
- [Security](security.md): auth, secrets, SSH, cleanup, and trust boundaries.
