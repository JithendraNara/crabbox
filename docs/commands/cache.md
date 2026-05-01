# cache

`crabbox cache` inspects, purges, or warms caches on a leased box.

```sh
crabbox cache stats --id cbx_...
crabbox cache stats --id cbx_... --json
crabbox cache warm --id cbx_... -- pnpm install --frozen-lockfile
crabbox cache purge --id cbx_... --kind pnpm --force
```

Cache kinds:

```text
pnpm
npm
docker
git
all
```

`cache warm` runs a command in the synced repo workdir for that lease. On boxes prepared by `crabbox actions hydrate`, it uses the hydrated `$GITHUB_WORKSPACE` and sources the workflow env handoff like `crabbox run`.

Repo `cache.pnpm`, `cache.npm`, `cache.docker`, and `cache.git` toggles control which kinds `stats` reports and which kinds `purge --kind all` removes.

Related docs:

- [Performance](../performance.md)
- [Cache controls](../features/cache.md)
