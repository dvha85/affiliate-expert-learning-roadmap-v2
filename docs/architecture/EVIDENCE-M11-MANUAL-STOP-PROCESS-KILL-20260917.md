# Evidence — M11 direct STOP process termination

- Scope: local learner Bot, synthetic/read-only runtime, POSIX process boundary.
- Implementation commit: `8b3bc43fe17034f2299a8de34cacc3fedbd56589`.
- Test: `TestM11ManualStopProcessKillRequiresLockedRecovery` in
  `lab/affiliate-bot/cmd/bot/m11_manual_stop_process_kill_test.go`.
- Production remains `NOT_READY_FOR_PRODUCTION`; this evidence does not claim
  power-loss durability, atomic multi-file transactions, distributed locking,
  live execution, provider operation, business outcome, pilot or deployment.

## Scenario

The test starts a real learner Bot test-binary child and runs `m11-stop`. A
test-only hook sends POSIX `SIGKILL` after the atomic rename boundary for each
of the three direct-STOP files:

1. `m11-manual-stop-journal.json` is visible;
2. `mission-state.json` is visible with durable STOP;
3. `STOP` is visible with durable STOP.

The journal remains the recovery authority in all three cases. A fresh Bot
process returns `RECOVERY_REQUIRED` and cannot expose a partial direct STOP.
The next locked writer replays the exact journal, retains the STOP state and
marker, removes the journal once, and returns `STOPPED`. A subsequent stop with
a different reason leaves both canonical files byte-identical.

## Verification

```text
GOWORK=off go test ./cmd/bot -run 'TestM11ManualStopProcessKillRequiresLockedRecovery|TestM11ManualStopProcessTerminationChild' -count=1 -v
PASS
  after-journal-rename
  after-state-rename
  after-marker-rename
```

The same learner test package is covered by the deterministic learner Bot
shards in `.github/workflows/curriculum-ci.yml`; the workflow command is
`python scripts/run_learner_bot_test_shard.py --shard-index ...`.
