# Evidence: M11 expiry authority no-mutation boundary

## Scope

- Base checkout before this change: `main` at `163f7f5c...`.
- The real learner Bot command path is exercised against a registered M11
  fixture graph with lease, health, cost, gate, authorization and pending
  reservation artifacts.
- This is local fixture/read-only evidence. It does not prove provider clocks,
  live executor behavior, business outcome, deployment, distributed locking,
  or power-loss/atomic multi-file durability.

## Verification

```text
cd lab/affiliate-bot
GOWORK=off go test ./cmd/bot -run TestMissionM11ExpiryRejectsAuthorityWritesWithoutMutation -count=1
PASS

GOWORK=off go test ./cmd/bot -run TestMissionM11ExpiryRejectsAuthorityWritesWithoutMutation -count=5
PASS
```

The test uses exact boundary timestamps and verifies the real commands reject
without changing `m11-artifacts.jsonl`:

- registering a health snapshot at `lease.ExpiresAt`;
- evaluating a gate at `lease.ExpiresAt`;
- issuing authorization at `lease.ExpiresAt`;
- recording a fixture execution at `authorization.ExpiresAt`.

This complements the existing same-ID lease rebind and activation-expiry
regressions. The registry bytes are compared after every rejected command.
The existing CI command `go test -race ./...` executes this test on pull
requests and pushes to `main`.

## Boundary

This closes the tested learner command expiry/no-mutation seam only. RP-07
remains `PARTIAL`; full production authority, trusted external time, crash or
power-loss durability, distributed locking, provider/business outcome, pilot
and deployment evidence remain open. Overall readiness remains
`NOT_READY_FOR_PRODUCTION`.
