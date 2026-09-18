# RP-04 shared canonical evidence context — 2026-09-18

## Scope

This change moves the learner M07 evidence-context builder into
`core/canonical`. M07, M08 intent/policy validation, and the M06 HTTP/CLI
handoffs continue to resolve one stored history record through the existing
replay-`MATCH` boundary; they now consume one versioned
`canonical-evidence-context/v1` envelope for aggregate and original M00 field
evidence.

The shared builder fails closed when recorded aggregate IDs do not exactly
match canonical observations, when the raw source projection is malformed, or
when an original field ID is empty/duplicated/collides with another evidence
ID. Missing values remain `null` in the original field evidence and are never
invented by the context layer. Source role, claim kind and limitation are
copied from the canonical M00 projection.

## Changed paths

- `core/canonical/context.go`: shared versioned context and strict ID/provenance
  builder.
- `core/canonical/context_test.go`: aggregate/field provenance, malformed
  source, forged ID and collision regressions.
- `lab/affiliate-bot/cmd/bot/m07.go`: learner adapter delegates to the shared
  builder instead of reimplementing M00 field projection.
- `lab/affiliate-bot/cmd/bot/m07_test.go`: M07 envelope/version and forged
  recorded-ID regression.

## Verification boundary

The implementation is offline/fixture/read-only evidence. It does not add a
provider fetch, authenticated source, live executor, business outcome,
distributed lock, filesystem crash/power-loss, or multi-file transaction
claim. The resolver remains conservative: a record must resolve exactly once
and replay `MATCH`; `DRIFT` and `UNREPLAYABLE` cannot become context or intent.

The current worktree has no `go.exe`, so local Go test/vet execution is not
claimed. Hosted acceptance is the required gate: full learner `go test -race
./...`, the native Windows runtime test/vet/targeted lock-backup-restore job,
and the Mission Agent Path workflow. Python readiness audit and its unit suite
remain separate structural checks.

## Readiness effect

R02/R03/R04 receive a narrower shared-context implementation and regression
seam, but RP-04 remains `PARTIAL` until the broader source/deployment and
external evidence blockers are addressed. Overall readiness remains
`NOT_READY_FOR_PRODUCTION`.
