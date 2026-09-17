# M11 recovery-admission field-integrity evidence

## Scope

The canonical M11 artifact decoder now rejects whitespace-only identity and
lineage fields in `PRODUCTION_RECOVERY_ADMISSION`, even though the JSON Schema
minimum-length rule permits a single whitespace character. This keeps a
recovery handoff from carrying an empty-looking prior or new runtime/lease,
approval, or resolution reference into graph validation.

The guard does not infer or validate the prior runtime's external state; the
learner admission path still performs the separate stopped-runtime and human
resolution checks. This is bounded offline/read-only input integrity, not
cross-runtime atomicity, power-loss recovery, distributed locking, provider
operation, live execution, business outcome, pilot or deployment readiness.
The repository remains `NOT_READY_FOR_PRODUCTION`.

## Regression

`core/m11/artifact_test.go` takes a valid non-authorizing recovery admission,
clears each required identity/lineage field to whitespace, and requires
`DecodeArtifact` to reject it before graph/runtime use. The existing approval,
non-authorizing, and prior-identity regressions remain in the same test.

The acceptance command is the required Curriculum CI race suite:

```text
go test -race ./...
```

The current Windows worktree has no `go.exe`, so no local Go PASS is claimed;
the Python readiness suite and audit remain local checks. PR #412 head
`9b9f01d21dec22c0b7e85f57b4b7846085086747` completed the Ubuntu Curriculum CI
run `35239183309` and Mission Agent Path CI run `35239183371`; all 12 required
remote checks passed, including the learner race suite. This confirms the
repository regression gate only; it is not production, provider, live-executor
or pilot evidence, and readiness remains `NOT_READY_FOR_PRODUCTION`.
