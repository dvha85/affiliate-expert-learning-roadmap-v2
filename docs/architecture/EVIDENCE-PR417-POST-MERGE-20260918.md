# Post-merge evidence — PR #417 / RP-02 shared M08 decoder and policy

## Snapshot

- PR: [#417](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/417)
- Merged `main`: `c3d10f8`
- PR head before squash: `6e8ad1a`
- Readiness status: `RP-02=PARTIAL`; overall `NOT_READY_FOR_PRODUCTION`.

This record rebinds the readiness snapshot after the RP-02 squash merge. It
does not claim provider access, live execution, proposal migration,
cross-store transactionality, distributed persistence, business outcomes,
pilot or deployment readiness.

## Verification

The merged commit contains the shared M08 decoder/policy seam, conformance
table, intent hash-version fail-closed boundary, exact-number reload checks and
M07 proposal provenance resolver regression.

- [Curriculum CI run 35318212387](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35318212387)
  passed all 9 jobs, including `windows-runtime` job
  [105514514976](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35318212387/job/105514514976)
  and learner race job
  [105514515032](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35318212387/job/105514515032).
- [Mission Agent Path CI run 35318212517](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35318212517)
  passed all 4 jobs, including mission runtime and semantics/blueprints.
- The merge event reports 13 checks passed. Local post-merge bookkeeping is
  bounded to readiness audit, JSON parsing, Python auditor tests and
  `git diff --check`; no local Go executable is available.

## Boundary

The M08 decoder and semantic validator are shared by mission-runtime and the
learner's smaller policy-input contract. The policy table and provenance
resolver regressions remain offline/fixture evidence. Hash migration,
cross-store crash atomicity, distributed persistence, approval/ledger
execution-chain hardening and provider/live execution remain open.
