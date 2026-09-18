# RP-02 shared M08 policy-context decoder/policy evidence

## Scope

This snapshot records the first RP-02 seam after the PR #416 post-merge
baseline. `core/m08` now owns the full policy-context decoder and semantic
validation boundary. The mission-runtime M08 harness calls that decoder, while
the learner's intentionally smaller policy-input contract calls the same
semantic validator after it binds decision, evidence and proposal IDs.

The shared boundary rejects missing/null fields, duplicate or unknown/case
variant keys, invalid RFC3339 time, duplicate identity lists, unknown risk
classes and blank idempotency entries before policy evaluation. A human context
may have no proposal set; an agent intent still requires a resolvable
`proposal_ref` during policy evaluation.

## Regression coverage

- `core/m08/m08_test.go` covers valid full-context decode plus missing, null,
  duplicate-key, unknown-key, invalid-time, invalid-risk, duplicate-ID and
  blank-idempotency rejects.
- `lab/mission-runtime/cmd/demo/m08_boundary_test.go` compares the harness
  boundary with the canonical core decoder for valid and invalid contexts.
- `lab/affiliate-bot/cmd/bot/mission_command_test.go` covers learner rejection
  of invalid shared context semantics.
- Existing intent/policy tests remain in place for exact JSON numbers,
  `parameters:null`, proposal-only authority and non-authorizing decisions.

## Verification status and boundary

The current worktree has no Go executable, so no local Go test or vet PASS is
claimed. Hosted CI is now green for PR #417:

- [Curriculum CI run 35314331386](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35314331386)
  passed the learner full tests/vet, `go test -race ./...`, deterministic
  runtime/smoke jobs and `windows-runtime` job 105502656378.
- [Mission Agent Path CI run 35314331378](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35314331378)
  passed the mission runtime job 105502656431, including its Go test/vet
  coverage.
- The shared M08-specific learner and harness regressions are included in
  learner jobs 105502656457/105502656411 and mission job 105502656431; the
  learner race regression is job 105502656474.

This is hosted offline/fixture evidence for the decoder and policy boundary;
it does not claim provider access, live execution or production readiness. The
readiness audit remains `NOT_READY_FOR_PRODUCTION`; RP-02, R06 and R07 remain
`PARTIAL` because shared scenario conformance, exact hash-version/migration,
provenance/persistence and external execution requirements are still open.

This change is decoder/policy hardening only. It does not add provider access,
live execution authority, approval migration, proposal persistence, business
outcomes, distributed locking, crash/power-loss atomicity or deployment proof.
