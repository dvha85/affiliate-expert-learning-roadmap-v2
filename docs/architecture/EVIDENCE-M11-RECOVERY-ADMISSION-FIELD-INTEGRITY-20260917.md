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
the Python readiness suite and audit remain local checks. PR #412 was merged
as `e197f854a9f396bbe1136004b792b2699dc647e6` from source head
`f3e5a213e5b142d23961443c3af9097ae3e5456e`; Curriculum CI run
`35241923424` and Mission Agent Path CI run `35241923430` completed with all 12
checks successful, including the learner race suite. This confirms the
repository regression gate only; it is not production, provider, live-executor
or pilot evidence, and readiness remains `NOT_READY_FOR_PRODUCTION`. See
`EVIDENCE-PR412-POST-MERGE-20260917.md` for the bounded acceptance record.
