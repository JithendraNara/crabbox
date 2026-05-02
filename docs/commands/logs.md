# logs

`crabbox logs` prints the retained remote output for a recorded run.

```sh
crabbox logs run_...
crabbox logs --id run_...
crabbox logs run_... --json
```

The plain form writes the log text to stdout. `--json` returns run metadata plus the log.

Logs are bounded remote stdout/stderr captures. The CLI keeps up to 8 MiB per run and the coordinator stores larger captures in chunks, so failures from noisy parallel runs remain visible without turning run history into unlimited archival storage.

Related docs:

- [history](history.md)
- [events](events.md)
- [attach](attach.md)
- [History and logs](../features/history-logs.md)
