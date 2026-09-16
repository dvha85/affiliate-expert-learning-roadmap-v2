# M11 authorization → gate exact lineage evidence

## Scope

This record covers one offline graph-validation seam. A production
authorization must retain the exact lease, health-snapshot and cost-bound
artifacts that its `ALLOW_PRODUCTION` gate evaluated, including IDs, hashes and
the evaluated cost minor. A different artifact may be individually
checksum-valid and still must not be accepted as the authorization's evidence
lineage.

This does not prove ledger durability, a live executor, provider operation,
business outcome, crash/power-loss recovery, atomic multi-file commit,
distributed locking, pilot or deployment readiness.

## Implementation

- `core/m11/artifact_registry.go` compares authorization fields with the
  resolved gate before accepting the authorization.
- The comparison covers lease ID/version/hash, health snapshot ID/hash and cost
  bound ID/hash/minor, in addition to the existing canonical links and hashes.

## Verification

```text
cd core
GOWORK=off go test ./m11
ok   github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11
```

`TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks` creates
separate checksum-valid health and cost artifacts, switches an authorization to
those artifacts without changing the gate, and asserts that
`ValidateArtifactGraph` rejects the graph. The positive exact-lineage graph
and the existing ID/timeline/ledger regressions remain covered by the same
test.
