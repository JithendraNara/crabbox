# MVP Plan

## Goal

Build Crabbox as a Go CLI plus Cloudflare coordinator that lets trusted OpenClaw maintainers run local worktrees on shared remote machines with the same core feel as Blacksmith Testboxes:

1. Ask for a machine class.
2. Get an idle warm machine or provision a new Hetzner machine.
3. Sync the local dirty tree.
4. Run a command remotely with streamed output.
5. Release or clean up the machine automatically.

The MVP should optimize for a useful maintainer workflow, not generalized cloud scheduling.

## Product Shape

Primary command:

```sh
crabbox run --profile openclaw-check -- pnpm check:changed
```

Expected user experience:

- Human-readable progress by default.
- Machine-readable `--json` for scripts.
- No central project secrets store in MVP.
- Local env allowlist only.
- Shared pool for trusted maintainers.
- Warm machines for fast repeated checks.
- TTL cleanup for abandoned leases.
- Explicit `stop`/`release` for manual cleanup.

## Repositories

Use two repos:

- `openclaw/crabbox`: Go CLI, Worker coordinator, docs, deploy scripts.
- `openclaw/crabbox-fleet`: desired fleet config only.

The fleet repo is not a lock database. It stores profiles, machine classes, default TTLs, sync excludes, and backend declarations. Live lease state belongs in Cloudflare Durable Objects.

## MVP Components

Build in this order:

1. Repo scaffold
   - Go module.
   - `cmd/crabbox`.
   - `worker/` or `services/coordinator/` for Cloudflare Worker code.
   - `docs/`, `configs/`, `scripts/`.
   - CI with build, format, and focused tests.

2. Config loading
   - Flags override env.
   - Env overrides repo-local `crabbox.yaml`.
   - Repo-local config overrides user config.
   - User config overrides shared fleet config.
   - Shared fleet config can be fetched from GitHub raw content or local checkout.

3. Coordinator API
   - Cloudflare Worker validates Cloudflare Access JWT.
   - Durable Object owns lease state and atomic machine selection.
   - Worker calls Hetzner API for create/delete/status.
   - Worker exposes JSON API under `/v1`.

4. Lease lifecycle
   - `POST /v1/leases` acquires or provisions.
   - `POST /v1/leases/{id}/heartbeat` keeps lease alive.
   - `POST /v1/leases/{id}/release` releases or deletes.
   - Durable Object alarm reaps expired leases.
   - Machines have states: `idle`, `leased`, `draining`, `provisioning`, `failed`.

5. SSH runner
   - MVP transport: public SSH to Hetzner, key-only, locked-down `crabbox` user.
   - CLI receives machine address and SSH username from the coordinator.
   - CLI owns rsync, command execution, streaming output, and exit code propagation.
   - Later transport: Cloudflare Tunnel/Access SSH or SSH CA.

6. Sync
   - Use `rsync` for MVP.
   - Preserve local dirty tree, including uncommitted changes.
   - Exclude heavy local folders by profile: `node_modules`, `.turbo`, `.git/lfs`, caches.
   - Sync to `/work/crabbox/<lease-id>/<repo-name>`.
   - Record sync metadata for debugging.

7. Hetzner backend
   - Create machines from configured image.
   - Attach configured SSH key.
   - Apply labels: `crabbox=true`, `profile=...`, `lease=...`, `owner=...`.
   - Support warm static pool and ephemeral overflow.
   - Implement cleanup for stale ephemeral machines.

8. OpenClaw profile
   - `openclaw-check` profile.
   - Linux x64, Docker, Node 22, pnpm, Git.
   - Default TTL: 90 minutes.
   - Default machine class configurable, likely `ccx33` first.
   - Env allowlist: `OPENCLAW_*`, `NODE_OPTIONS`, common model/provider keys only when explicitly configured locally.

9. Access/auth
   - Primary org: GitHub `openclaw`.
   - Cloudflare Access org: `openclaw-crabbox.cloudflareaccess.com`.
   - MVP can use Cloudflare OTP Access first.
   - GitHub IdP comes next with OAuth app callback:
     `https://openclaw-crabbox.cloudflareaccess.com/cdn-cgi/access/callback`.

10. Usability pass
    - `crabbox doctor`.
    - Helpful errors for missing `rsync`, SSH key, config, Access token, or provider token.
    - `--json` for every state-inspecting command.
    - Shell completions.

## Definition Of Done

MVP is done when this works from a local OpenClaw checkout:

```sh
crabbox login
crabbox run --profile openclaw-check -- pnpm check:changed
```

And proves:

- A lease is created.
- A Hetzner machine is selected or provisioned.
- Local files sync.
- Remote command output streams.
- The local exit code matches the remote command exit code.
- Lease is released on success/failure.
- Expired leases are cleaned by TTL.
- Machine pool state is visible through `crabbox pool`.

## Non-Goals For MVP

- No Kubernetes.
- No central secret storage.
- No full autoscaling scheduler.
- No multi-tenant untrusted execution.
- No Windows/macOS workers.
- No Blacksmith backend in the first implementation path.
- No attempt to perfectly hide SSH; make it reliable first.

## Known Current Infra Facts

- Intended primary domain: `crabbox.openclaw.ai`.
- Current Cloudflare-manageable fallback domain: `crabbox.clawd.bot`.
- `openclaw.ai` is currently not visible as a Cloudflare zone in the available account; DNS is on Namecheap nameservers.
- Cloudflare account ID and Crabbox Cloudflare token are available in local and Mac Studio `~/.profile`.
- Cloudflare Access is enabled.
- Current Access IdP is OTP only.
- Hetzner token is available in local and Mac Studio `~/.profile`.
- GitHub org slug is `openclaw`.

## Next Implementation Milestones

1. Add Go module and CLI skeleton.
2. Add config schema and parser.
3. Add Worker API skeleton with local tests.
4. Add Durable Object lease store.
5. Add Hetzner provider implementation.
6. Add SSH/rsync runner.
7. Wire `crabbox run`.
8. Deploy Worker to Cloudflare fallback domain.
9. Provision first warm Hetzner machine.
10. Run OpenClaw `pnpm check:changed` through Crabbox.

