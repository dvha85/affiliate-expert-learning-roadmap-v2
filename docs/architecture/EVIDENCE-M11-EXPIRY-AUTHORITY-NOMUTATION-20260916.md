# Evidence: M11 expiry authority no-mutation boundary

## Scope

- Base checkout before this change: `main` at `01c27fb...`.
- The real learner Bot command path is exercised against a fully populated
  M08-M10 learner fixture and a registered M11 graph with lease, health, cost,
  activation, ledger, gate, authorization and reservation checkpoints.
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

The test creates backup snapshots before gate, authorization, reservation and
execution, restores each into a fresh runtime, then invokes a newly compiled
Bot binary at the exact boundary. The real commands reject without changing
the restored canonical runtime for:

- evaluating a gate at `lease.ExpiresAt`;
- issuing authorization at `lease.ExpiresAt`;
- reserving against `authorization.ExpiresAt`;
- recording a fixture execution at `authorization.ExpiresAt`.

The assertion includes `mission-state.json`, M10/cost registries and
`m11-artifacts.jsonl`. Health registration is intentionally not presented as
an expiry authority rejection: it is an evidence append, not a production
authority decision. This complements the existing same-ID lease rebind and
activation-expiry regressions.
The existing CI command `go test -race ./...` executes this test on pull
requests and pushes to `main`.

## Boundary

This closes the tested learner command expiry/no-mutation seam only. RP-07
remains `PARTIAL`; full production authority, trusted external time, crash or
power-loss durability, distributed locking, provider/business outcome, pilot
and deployment evidence remain open. Overall readiness remains
`NOT_READY_FOR_PRODUCTION`.
