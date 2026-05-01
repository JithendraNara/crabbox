# admin

`crabbox admin` contains trusted operator controls for coordinator-backed leases.

```sh
crabbox admin leases
crabbox admin leases --state active --json
crabbox admin release cbx_...
crabbox admin release cbx_... --delete
crabbox admin delete cbx_... --force
```

Admin commands require a configured coordinator and bearer token. The current coordinator trusts the shared operator token; do not expose it to untrusted users.

## leases

List coordinator lease records.

Flags:

```text
--state <state>     filter by active, released, expired, or failed
--owner <email>     filter by owner
--org <name>        filter by org
--limit <n>         default 100, maximum 500
--json              print JSON
```

## release

Mark a lease released. Add `--delete` to delete the backing server while releasing.

## delete

Delete the backing server for an active lease and mark it released. Requires `--force`.

Related docs:

- [Operations](../operations.md)
- [Auth and admin](../features/auth-admin.md)
