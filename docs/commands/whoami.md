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

Identity normally comes from the signed GitHub login token. Shared bearer-token automation reports owner/org from `X-Crabbox-Owner` and `X-Crabbox-Org`; the CLI fills those from `CRABBOX_OWNER`, Git email env, `git config user.email`, and `CRABBOX_ORG`. If a fallback route forwards Cloudflare Access identity, that Access email wins over shared-token owner headers. JSON output also reports the forwarded auth mode, such as `github` or `bearer`.

Related docs:

- [login](login.md)
- [Broker auth and routing](../features/broker-auth-routing.md)
