# Registry graph envelope integrity evidence

## Scope

The canonical M10 and M11 graph validators now verify each `ArtifactEntry`
envelope before resolving lifecycle links. A graph caller cannot replace an
entry's `artifact_id` or `content_hash` while retaining the canonical artifact
bytes and then make a gate point at forged parent metadata. The check also
requires the stored artifact bytes to be the canonical serialization produced
by the shared artifact builder.

This is a bounded offline/read-only API integrity guard. It does not prove
filesystem durability, power-loss recovery, atomic multi-file commit,
distributed locking, live execution, provider operation, business outcome,
pilot or deployment readiness. The repository remains
`NOT_READY_FOR_PRODUCTION`.

## Implementation and regression

- `core/m10/artifact_registry.go` and `core/m11/artifact_registry.go` rebuild
  every entry with `NewArtifactEntry` before inserting it into the graph maps,
  then compare artifact ID, content hash and canonical bytes.
- `core/m10/cost_bound_test.go` and `core/m11/artifact_test.go` pass forged
  artifact IDs and content hashes to the public graph validators and require
  rejection even for otherwise standalone valid artifacts.
- Learner loaders already validate each JSONL envelope before invoking the
  graph; the core regression closes the direct graph API boundary as well.

## Verification

The acceptance command is wired to the Curriculum CI race suite:

```text
go test -race ./...
```

The current Windows worktree has no `go.exe`, so local Go test/vet execution
is unavailable. The Python baseline suite and readiness audit remain local
checks; no local Go PASS is claimed. The merged PR #412 acceptance runs
`35241923424` and `35241923430` completed with all 12 checks successful on
commit `e197f854a9f396bbe1136004b792b2699dc647e6`; this remains bounded
repository regression evidence. See
`EVIDENCE-PR412-POST-MERGE-20260917.md` for the exact run record.
