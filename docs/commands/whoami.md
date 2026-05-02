# whoami

`crabbox whoami` verifies broker auth and prints the identity the coordinator sees.

```sh
crabbox whoami
crabbox whoami --json
```

Human output:

```text
user=steipete@gmail.com org=openclaw auth=github broker=https://crabbox.openclaw.ai
```

Identity normally comes from the signed GitHub login token. Shared bearer-token automation reports owner/org from `X-Crabbox-Owner` and `X-Crabbox-Org`; the CLI fills those from `CRABBOX_OWNER`, Git email env, `git config user.email`, and `CRABBOX_ORG`. Raw Cloudflare Access identity headers are ignored; only a verified Access JWT email can become the bearer-token owner. JSON output also reports the forwarded auth mode, such as `github` or `bearer`.

Related docs:

- [login](login.md)
- [Broker auth and routing](../features/broker-auth-routing.md)
