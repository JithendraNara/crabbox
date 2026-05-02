# Broker Auth And Routing

Read when:

- changing coordinator authentication;
- changing Cloudflare routes or Access policy;
- debugging bearer-token automation or GitHub browser login.

The broker is exposed through Cloudflare Workers routes:

```text
https://crabbox.openclaw.ai
https://crabbox-access.openclaw.ai
https://crabbox-coordinator.services-91b.workers.dev
crabbox.clawd.bot/*
```

Normal users run `crabbox login`, which opens GitHub and stores a signed Crabbox user token. The coordinator needs a GitHub OAuth app with callback:

```text
https://crabbox.openclaw.ai/v1/auth/github/callback
```

Self-hosted coordinators need their own GitHub OAuth app. The callback URL on
that app must exactly match the public Worker URL plus
`/v1/auth/github/callback`, and the Worker `CRABBOX_PUBLIC_URL` must use that
same public origin.

Worker secrets:

```text
CRABBOX_GITHUB_CLIENT_ID
CRABBOX_GITHUB_CLIENT_SECRET
CRABBOX_GITHUB_ALLOWED_ORG
CRABBOX_GITHUB_ALLOWED_ORGS
CRABBOX_GITHUB_ALLOWED_TEAMS
CRABBOX_SESSION_SECRET
```

GitHub browser login requires active membership in the allowed GitHub org before
the coordinator mints a Crabbox user token. Set `CRABBOX_GITHUB_ALLOWED_ORG` or
comma-separated `CRABBOX_GITHUB_ALLOWED_ORGS`; if unset, the Worker falls back
to `CRABBOX_DEFAULT_ORG`, then `openclaw`. The OAuth app must request
`read:user user:email read:org`.

Set comma-separated `CRABBOX_GITHUB_ALLOWED_TEAMS` to require membership in at
least one team after org membership passes. Entries are GitHub team slugs. Use
`team-slug` for the selected org or `org/team-slug` when multiple orgs are
allowed.

Trusted automation can still use the shared operator bearer token configured in the CLI and Worker. The CLI sends:

```text
Authorization: Bearer <token>
X-Crabbox-Owner: <email>
X-Crabbox-Org: <org>
```

If the coordinator route is also protected by Cloudflare Access, the CLI can send Access credentials before the Worker receives the request. Configure `CRABBOX_ACCESS_CLIENT_ID` and `CRABBOX_ACCESS_CLIENT_SECRET` for a Cloudflare Access service token, or `CRABBOX_ACCESS_TOKEN` to forward an already minted Access JWT as `cf-access-token`. These Access credentials only satisfy Cloudflare Access; the Worker still requires the Crabbox bearer token or a signed Crabbox user token.

The live Access-protected route is `https://crabbox-access.openclaw.ai`. Its Access app is service-token-only (`non_identity`) and currently allows the local Crabbox CLI service token, so automated clients can prove both layers independently: first Cloudflare Access, then the Worker bearer or signed user token.

Owner selection for bearer-token requests:

```text
CRABBOX_OWNER
GIT_AUTHOR_EMAIL
GIT_COMMITTER_EMAIL
git config user.email
```

`CRABBOX_ORG` sets the org header. When a request comes through Cloudflare Access and Access identity is forwarded, that Access email wins over the CLI-provided owner. Normal `crabbox login` requests use the signed GitHub token identity.

GitHub user tokens are signed by the Worker and are not admin tokens. Admin routes require the shared operator token. The `crabbox.openclaw.ai/*` route is the canonical CLI and browser-login endpoint. `crabbox-access.openclaw.ai/*` is the service-token-protected endpoint. `https://crabbox-coordinator.services-91b.workers.dev` and `crabbox.clawd.bot/*` are fallbacks.

Related docs:

- [Coordinator](coordinator.md)
- [Security](../security.md)
- [Infrastructure](../infrastructure.md)
- [config command](../commands/config.md)
