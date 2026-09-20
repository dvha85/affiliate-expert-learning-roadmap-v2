# Post-merge PR #489 n8n scope and hosted acceptance — 2026-09-21

Status: `PARTIAL`, bounded repository/fixture/hosted-CI evidence.

## Merge and exact-head verification

PR #489 ([fix: expand n8n CI dependency scope](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/489)) was squash-merged into `main` at
`12cb66f45d4e5e27002d4cbf01e6ec0481436759` from implementation head
`cfa4c6d163d46c33907789907222c5cb17139441`.

On the exact implementation head, `curriculum-gate` and `mission-gate` passed.
The child jobs passed for deterministic runtime, Go/core/mission checks,
learner shards and race, native Windows runtime, validators, offline smokes,
backup/mutation coverage, CodeQL and mission semantics. The
`govulncheck-go-modules` job remained a non-blocking failure under the checked-in
`continue-on-error` policy.

The scoped n8n job executed rather than skipping: hosted setup restored the
pinned Node 24 / n8n 2.38.1 runtime, and the step
`Run M06/M07 engine paths and real M06 Schedule Trigger admission` completed
successfully. This is the first current exact-head evidence for the parser and
Schedule Trigger runtime path after the scope gate was expanded to cover the
offline runner and n8n regression tests.

## Local verification

- Targeted scope, decoder and baseline-allowlist tests: 8 passed.
- Python regression suite: 174 tests passed.
- JSON parsing, `compileall`, `--list` and `git diff --check`: PASS.

The product baseline metadata is now rebound to the actual PR #489 merge. This
does not prove provider/live execution, business outcome, deployment recovery,
clean-machine pilot, distributed locking, power-loss or atomic multi-file
durability; the repository remains `NOT_READY_FOR_PRODUCTION`.
