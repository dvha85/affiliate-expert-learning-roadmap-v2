# M11 outcome process-kill recovery — 2026-09-16

## Scope

This record covers both visible outcome-side points of the M11 replay
boundary. It uses the real learner Bot command path and a separately executed
Go test-binary child. One scenario terminates the child with POSIX `SIGKILL`
immediately before the new ledger append; a second terminates it immediately
after the outcome append and before journal cleanup.

The fresh read-only process must expose `RECOVERY_REQUIRED` in both scenarios.
A new locked writer then replays or confirms the journal's exact
outcome/ledger transition. The exact retry must not append a second outcome,
must leave the journal removed after recovery, and must leave the ledger head
equal to the journal transition.

This is local synthetic/read-only evidence for process termination. It is not
power-loss or filesystem-crash proof, does not establish atomic multi-file
durability, Windows native-lock behavior, distributed/multi-host locking,
provider operation, business outcome, pilot, or deployment readiness.

## Verification

- test: `TestM11OutcomeProcessKillAfterJournalBeforeLedgerAppend`
- test: `TestM11OutcomeProcessKillAfterOutcomeAppend`
- child: `TestM11ProcessTerminationChild`
- injected termination: POSIX `SIGKILL` before ledger append and after outcome append
- fresh read-only result: `RECOVERY_REQUIRED`
- locked exact retry: `EXACT_DUPLICATE`
- persisted outcome count after recovery: exactly one
- recovered ledger: byte-equivalent to the journal transition
- result: PASS on local macOS arm64

Command:

```text
cd lab/affiliate-bot
GOWORK=off go test ./cmd/bot -run TestM11OutcomeProcessKillAfterJournalBeforeLedgerAppend -count=1
```

The broader learner test shard remains the CI authority; this focused command
is the local evidence capture for the newly isolated outcome path.
