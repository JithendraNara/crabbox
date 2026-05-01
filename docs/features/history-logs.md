# History And Logs

Read when:

- changing run recording;
- debugging failed remote commands;
- deciding what belongs in coordinator history.

Coordinator-backed `crabbox run` records a run before the remote command starts. When the command exits, the CLI finishes that run with:

- exit code;
- sync duration;
- command duration;
- total duration;
- owner and org;
- provider, class, and server type;
- retained remote output tail.

Use:

```sh
crabbox history
crabbox history --lease cbx_...
crabbox logs run_...
```

History records live in the Fleet Durable Object. Log text is stored separately from run metadata and intentionally capped to the latest tail so noisy commands cannot exhaust storage.

Direct-provider mode does not have central history. Use shell output or local terminal logs there.

Related docs:

- [history command](../commands/history.md)
- [logs command](../commands/logs.md)
- [Observability](../observability.md)
