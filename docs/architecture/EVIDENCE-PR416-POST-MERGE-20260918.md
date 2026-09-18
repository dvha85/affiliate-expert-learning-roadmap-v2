# Post-merge evidence — PR #416 / RP-01 Windows ancestor race

## Snapshot

- PR: [#416](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/416)
- Merged `main`: `cefdb758f70ce36c08bc822ee658737149600bcd`
- PR implementation head before squash: `bbd8aec14e56ba0b484376014b2ab5ec775ce3ec`
- Readiness status: `RP-01=PARTIAL`; overall `NOT_READY_FOR_PRODUCTION`.

This record synchronizes the readiness snapshot after the squash merge. It does
not claim a new provider, live-executor, pilot, deployment, distributed-locking,
power-loss, or multi-file atomicity result.

## Verification

The merged commit contains the PR implementation and its reviewed CI evidence:

- [Curriculum CI run 35306191800](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35306191800)
  passed the Windows runtime job [105478750919](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35306191800/job/105478750919)
  (`go test ./...`, `go vet ./...`, and targeted lock/backup/restore tests).
- The same Curriculum workflow passed learner race job
  [105478751048](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35306191800/job/105478751048)
  (`go test -race ./...`).
- [Mission Agent Path CI run 35306191794](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35306191794)
  passed all three jobs.
- The local post-merge bookkeeping checks are the readiness audit, its Python
  regression suite, JSON parsing, and `git diff --check`; their results are
  recorded in the commit that updates this evidence snapshot.

## Boundary

The Windows hardening uses native parent handles opened with
`OPEN_REPARSE_POINT` and no delete sharing, pins existing ancestors, and creates
missing components one at a time. Controlled junction/reparse-point replacement
after preflight fails closed without writing the external tree for backup/restore,
stable reader/append, and `internal/store` JSONL reader paths.

Windows does not provide a portable `openat`/`mkdirat` equivalent for arbitrary
multi-component traversal in this boundary. The supported claim is therefore a
conservative local single-host fail-closed boundary, not POSIX traversal parity
or a distributed/power-loss/multi-file atomicity guarantee.
