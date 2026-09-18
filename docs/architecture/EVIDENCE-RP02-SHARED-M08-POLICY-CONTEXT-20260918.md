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
`proposal_ref` during policy evaluation. `core/m08` now also publishes one
deterministic conformance table covering allow/review/wait/expiry/link,
proposal, target, idempotency and tamper outcomes. Harness tests run every
scenario; learner tests run the scenarios expressible by its trusted
decision/evidence binding and explicitly leave missing-link registry states as
a documented parity boundary.

The persisted `sha256:` intent-hash prefix is now named V1. Unknown prefixes
such as `sha256-v2:` are returned as `UNSUPPORTED_HASH_VERSION` by the core
decoder/evaluator and the harness preserves that fail-closed status. No code
path reseals or rewrites an existing intent hash; a future migration remains a
separate, approval-reviewed operation.

Large JSON numbers are also checked after a real file round trip: the harness
reloads the persisted intent through `DecodeM08Intent`, while the learner
reloads its artifact through the `UseNumber` reader and recomputes the intent
hash. Both retain `9007199254740993` as `json.Number`; this is exact local
restart/fixture evidence, not distributed or power-loss durability evidence.

## Regression coverage

- `core/m08/m08_test.go` covers valid full-context decode plus missing, null,
  duplicate-key, unknown-key, invalid-time, invalid-risk, duplicate-ID and
  blank-idempotency rejects.
- `lab/mission-runtime/cmd/demo/m08_boundary_test.go` compares the harness
  boundary with the canonical core decoder for valid and invalid contexts.
- `lab/affiliate-bot/cmd/bot/mission_command_test.go` covers learner rejection
  of invalid shared context semantics.
- `core/m08/conformance.go` is the shared scenario/expected-result table;
  `core/m08/conformance_test.go`, the harness test and the learner test consume
  it rather than maintaining separate expected reasons.
- `core/m08/m08_test.go` and `lab/mission-runtime/cmd/demo/m08_boundary_test.go`
  reject unsupported hash prefixes without mutating the decoded or persisted
  intent.
- The harness and learner M08 parameter-number regressions now write/read the
  intent artifact before checking the exact number and hash.
- Existing intent/policy tests remain in place for exact JSON numbers,
  `parameters:null`, proposal-only authority and non-authorizing decisions.

## Verification status and boundary

The current worktree has no Go executable, so no local Go test or vet PASS is
claimed. Hosted CI is now green for PR #417:

- [Curriculum CI run 35316418613](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35316418613)
  passed the learner full tests/vet, `go test -race ./...`, deterministic
  runtime/smoke jobs and `windows-runtime` job 105508926281.
- [Mission Agent Path CI run 35316418619](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35316418619)
  passed the mission runtime job 105508926366, including its Go test/vet
  coverage.
- The shared M08-specific learner and harness regressions are included in
  learner jobs in the two hosted workflows; the learner race regression is job
  105508926341.

This is hosted offline/fixture evidence for the decoder and policy boundary;
it does not claim provider access, live execution or production readiness. The
readiness audit remains `NOT_READY_FOR_PRODUCTION`; RP-02, R06 and R07 remain
`PARTIAL` because shared scenario conformance, exact hash-version/migration,
provenance/persistence and external execution requirements are still open.

This change is decoder/policy hardening only. It does not add provider access,
live execution authority, approval migration, proposal persistence, business
outcomes, distributed locking, crash/power-loss atomicity or deployment proof.
