# PR #412 post-merge evidence

## Scope

PR #412 (`RP-07/RP-08: close offline registry and recovery integrity gaps`)
was merged into `main` as `e197f854a9f396bbe1136004b792b2699dc647e6` from head
`f3e5a213e5b142d23961443c3af9097ae3e5456e`. The merge contains the M10/M11
registry graph envelope checks, reverse-link/evaluation guards, recovery-admission
field integrity, mutation proof and the associated readiness documentation.

This record only proves the repository's post-merge regression gate. It does not
prove provider operation, business outcome, live execution authority, clean-machine
pilot, target-host deployment, distributed locking, power-loss durability or
atomic multi-file recovery. Readiness remains `NOT_READY_FOR_PRODUCTION`.

## Verification

Post-merge GitHub Actions runs for `e197f854a9f396bbe1136004b792b2699dc647e6`
completed successfully:

- Curriculum CI run `35241923424`: success.
- Mission Agent Path CI run `35241923430`: success.
- The associated check-run inventory contains 12 completed successful checks,
  including the learner race, mutation, backup/restore and Node 24/n8n paths.

The local Windows checkout independently passed:

```text
python -m unittest discover -s scripts/tests -v  -> 132 tests, OK
python scripts/audit_readiness.py .              -> NOT_READY_FOR_PRODUCTION
READINESS-MATRIX.json + READINESS-EVIDENCE-GRAPH.json -> JSON_OK=2
git diff --check                                  -> PASS
```

The checkout has Node `v22.23.2`, but no `go.exe` and no `n8n` CLI. Therefore
local Go/n8n execution is not claimed; the post-merge CI runs above are the
authoritative Go race and disposable n8n evidence for this record.
