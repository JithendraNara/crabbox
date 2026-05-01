# usage

`crabbox usage` shows orchestrator usage and estimated cost.

```sh
crabbox usage
crabbox usage --scope org --org openclaw
crabbox usage --scope user --user peter@example.com --month 2026-05
crabbox usage --scope all --json
```

Usage requires a configured coordinator. Direct-provider mode has no central history to query.

Lease ownership comes from Cloudflare Access when available. In bearer-token mode, the CLI sends `CRABBOX_OWNER`, Git email env, or local `git config user.email`; set `CRABBOX_ORG` to group leases under an org.

Scopes:

```text
user    one owner email; default
org     one organization
all     whole fleet
```

Flags:

```text
--scope user|org|all
--user <email>
--org <name>
--month YYYY-MM
--json
```

Cost values are estimates. The orchestrator fetches provider pricing when possible, uses explicit configured rates as overrides, and falls back to built-in defaults if provider pricing is unavailable. It reserves lease TTL cost before provisioning; summaries report elapsed runtime and reserved worst-case cost.
