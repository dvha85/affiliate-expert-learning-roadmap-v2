# Registry graph envelope mutation evidence

## Scope

The Curriculum CI mutation job removes the direct `ArtifactEntry` envelope
check from a disposable copy of each public M10 and M11 graph validator. The
matching real core regression must then fail on forged `artifact_id`,
`content_hash`, or non-canonical artifact bytes. This proves the regression is
connected to the runtime guard rather than only asserting the guard's source
text.

This is bounded offline/read-only mutation evidence. It does not prove
filesystem durability, power-loss recovery, atomic multi-file commit,
distributed locking, provider operation, live execution, business outcome,
pilot or deployment readiness. The repository remains
`NOT_READY_FOR_PRODUCTION`.

## Verification

`scripts/mutate_registry_graph_envelope_integrity.py` copies `core` and
`contracts` to a temporary workspace, removes the graph-only envelope check for
one module at a time, and runs the corresponding real graph regression with
`GOWORK=off`. The required CI command is:

```text
python scripts/mutate_registry_graph_envelope_integrity.py
```

The current Windows worktree has no `go.exe`, so the mutation command cannot
be claimed PASS locally. PR #412 was merged as
`e197f854a9f396bbe1136004b792b2699dc647e6` from source head
`f3e5a213e5b142d23961443c3af9097ae3e5456e`; Curriculum CI run
`35241923424` and Mission Agent Path CI run `35241923430` completed with all 12
checks successful, including the learner race and this mutation proof. This
remote CI result is still bounded fixture/read-only evidence and does not
change `NOT_READY_FOR_PRODUCTION`. See
`EVIDENCE-PR412-POST-MERGE-20260917.md` for the complete acceptance record.
