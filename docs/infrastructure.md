# Infrastructure

## Current Intended Setup

Primary public endpoint:

```text
https://crabbox.openclaw.ai
```

Current deployable Cloudflare fallback:

```text
https://crabbox.clawd.bot
```

Reason for the fallback: `openclaw.ai` currently resolves through Namecheap nameservers and is not visible as a Cloudflare zone in the available Cloudflare account. Cloudflare Workers custom domains are simplest when the zone is managed by Cloudflare.

## Cloudflare

Use Cloudflare for:

- HTTPS coordinator.
- Access auth.
- Worker runtime.
- Durable Object lease state.
- DNS/custom domain once the target zone is available.

Known setup:

- Access org: `openclaw-crabbox.cloudflareaccess.com`.
- Access enabled.
- Current IdP: one-time PIN.
- Desired IdP: GitHub restricted to org `openclaw`.

Required env:

```text
CRABBOX_CLOUDFLARE_API_TOKEN
CRABBOX_CLOUDFLARE_ACCOUNT_ID
CRABBOX_CLOUDFLARE_ZONE_ID
CRABBOX_CLOUDFLARE_ZONE_NAME
CRABBOX_DOMAIN
CRABBOX_FALLBACK_DOMAIN
CRABBOX_GITHUB_ALLOWED_ORG
```

GitHub IdP needs a GitHub OAuth app:

```text
Homepage URL: https://crabbox.openclaw.ai
Callback URL: https://openclaw-crabbox.cloudflareaccess.com/cdn-cgi/access/callback
```

Store resulting values outside the repo:

```text
CRABBOX_GITHUB_OAUTH_CLIENT_ID
CRABBOX_GITHUB_OAUTH_CLIENT_SECRET
```

## DNS Decision

Preferred path:

1. Add `openclaw.ai` to Cloudflare.
2. Copy existing DNS records exactly.
3. Add `crabbox.openclaw.ai`.
4. Switch nameservers at registrar.
5. Deploy Worker custom domain.

Temporary path:

1. Deploy Worker under `crabbox.clawd.bot`.
2. Keep `CRABBOX_DOMAIN=crabbox.openclaw.ai` as intended target.
3. Use fallback domain for early testing.
4. Move to `openclaw.ai` once DNS is ready.

## Hetzner

Use Hetzner Cloud for worker machines.

Required env:

```text
HCLOUD_TOKEN
HETZNER_TOKEN
```

MVP defaults:

```yaml
provider: hetzner-main
location: fsn1
serverType: ccx33
image: ubuntu-24.04
sshUser: crabbox
workdir: /work/crabbox
```

Machine labels:

```text
crabbox=true
profile=openclaw-check
class=ccx33
lease=cbx_...
owner=<github-login-or-email>
ttl=<timestamp>
```

## Machine Classes

Fleet config should define machine classes instead of hardcoding Hetzner types:

```yaml
classes:
  standard:
    provider: hetzner-main
    serverType: ccx23
    cpu: 4
    memory: 8gb
  fast:
    provider: hetzner-main
    serverType: ccx33
    cpu: 8
    memory: 32gb
  large:
    provider: hetzner-main
    serverType: ccx43
    cpu: 16
    memory: 64gb
```

Profiles choose a default class, and commands can override with `--class`.

## Fleet Repo

`openclaw/crabbox-fleet` should contain:

```text
fleet.yaml
profiles/openclaw.yaml
bootstrap/cloud-init.yaml
images/README.md
```

It should not contain secrets or live lease data.

Example:

```yaml
version: 1
fleet:
  name: openclaw
  coordinator: https://crabbox.openclaw.ai

profiles:
  openclaw-check:
    labels: [linux, x64, docker, node22]
    defaultClass: fast
    ttl: 90m
    maxTTL: 24h
    sync:
      exclude: [node_modules, .turbo, .git/lfs]
    envAllowlist:
      - OPENCLAW_*
      - NODE_OPTIONS
```

## Deployment

MVP deploy command:

```sh
crabbox admin deploy-coordinator
```

Or a script first:

```sh
scripts/deploy-worker
```

Deployment should:

1. Build Worker.
2. Create/update Durable Object bindings.
3. Set Worker secrets.
4. Deploy Worker.
5. Configure route/custom domain.
6. Verify `/v1/health`.

## Local And Mac Studio

The same required env should exist on both the local machine and Mac Studio. Do not commit these values.

