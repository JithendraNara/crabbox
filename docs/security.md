# Security

## Trust Model

MVP is for trusted OpenClaw maintainers, not arbitrary untrusted users.

Assumptions:

- Users can run arbitrary commands on leased machines.
- Machines may see forwarded local env values.
- Users are trusted not to attack other users intentionally.
- Bugs and crashes still happen, so cleanup must be defensive.

## Authentication

Cloudflare Access protects the coordinator.

MVP:

- One-time PIN Access is acceptable for early testing.
- Coordinator validates `Cf-Access-Jwt-Assertion`.
- Coordinator maps Access identity to lease owner.

Target:

- GitHub IdP.
- Require membership in GitHub org `openclaw`.
- Optional team allowlist for admin commands.

## Authorization

Roles:

```text
user: acquire, heartbeat, release own leases, list own leases
maintainer: shared warm pool access
admin: drain machines, cleanup, view all leases, deploy
```

Until GitHub teams are wired, admin identity can be an explicit allowlist in Worker config.

## Secrets

No central project secret store in MVP.

Rules:

- Secrets stay local.
- CLI forwards env only by allowlist.
- Users can opt in additional env names with `--env`.
- Never accept secret values as command-line flag values.
- Never log env values.
- Redact known secret-looking strings in diagnostics.

Profile allowlist example:

```yaml
envAllowlist:
  - OPENCLAW_*
  - NODE_OPTIONS
```

## SSH

MVP SSH posture:

- Public SSH allowed only for worker machines.
- Key-only authentication.
- Dedicated `crabbox` user.
- No password login.
- No root login.
- Work happens under `/work/crabbox`.
- Machines are disposable or cleanable.

Later hardening:

- Cloudflare Tunnel or Access SSH.
- SSH CA with short-lived certs.
- Per-lease Unix users.
- Per-lease workdir ownership and cleanup.

## Cleanup

Cleanup is security-sensitive.

Required:

- Lease TTL.
- Heartbeat deadline.
- Explicit release.
- Durable Object alarm cleanup.
- Provider label sweep for orphan machines.
- Boot-time cleanup of stale `/work/crabbox/*` dirs.

Release must be idempotent. Delete must tolerate already-deleted provider resources.

## Data Retention

Store only operational metadata:

- lease ID.
- owner identity.
- machine ID.
- profile.
- timestamps.
- state transitions.
- command string, unless disabled.

Do not store:

- stdout/stderr logs in the coordinator for MVP.
- env values.
- file contents.
- SSH keys.

## Audit Trail

Durable Object events should record:

```text
lease.created
machine.provisioned
lease.heartbeat
lease.extended
lease.released
lease.expired
machine.drained
machine.deleted
provider.error
```

The audit trail is for debugging and cleanup, not compliance.

