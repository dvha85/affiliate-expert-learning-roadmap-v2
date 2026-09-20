# RP-08 post-merge evidence for PR #457

## Scope

PR #457 added the bounded M11 outcome-link mutation proof and corrected the
core lifecycle regression fixture so the intentional invalid-cycle case does
not contaminate later assertions. The implementation head was
`d9e74ddcd2475bd1e6e1ab4247d237ae75b21319`; GitHub squash-merged it into
`main` as `40f3d76fa163b6aee6d20bd8c6bd24bceca9254f`.

This evidence rebinds the offline mutation/readiness bookkeeping to the merged
main snapshot. It does not establish ledger durability, crash or power-loss
recovery, atomic multi-file publication, distributed or multi-host safety,
Windows traversal parity, provider/live execution, deployment, pilot
acceptance, business outcomes, or production readiness.

## Hosted verification

- Curriculum CI run `35407311675` passed all 10 Curriculum jobs, including:
  Windows runtime job `105799633655`, learner race job `105799633794`, and
  deterministic backup/mutations job `105799633765`.
- Mission Agent Path CI run `35407311697` passed all 3 Mission jobs:
  semantics/blueprints `105799633524`, runtime `105799633521`, and n8n engine
  regression `105799633274`.
- GitHub reported `13 / 13 checks OK` for the exact implementation head.
- The new outcome-link mutation proof failed at the swapped outcome-ID
  assertion after removing only the targeted guard in a disposable copy.

## Local verification and boundary

- Readiness audit: PASS, `NOT_READY_FOR_PRODUCTION`, 147 scoped claims.
- Python audit regression suite: PASS, 121 tests.
- Diff, Python compile and JSON validation: PASS.
- Local Go execution remains unavailable because this workspace has no
  `go.exe`; hosted exact-head CI is the executable Go evidence.

