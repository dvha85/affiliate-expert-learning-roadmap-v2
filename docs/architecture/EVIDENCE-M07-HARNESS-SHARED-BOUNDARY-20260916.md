# M07 shared harness boundary (2026-09-16)

Trạng thái: **local automated verification / fixture-only / read-only**.

## Change and verification

- `base_commit`: `ba575f3` (`main` trước khi thay đổi)
- `scope`: mission-runtime M07 harness và `core/m07`
- `implementation`: mission-runtime aliases `ToolSpec`, tool request và agent
  output cho core types; registry validation, strict output decoding, claim/value
  grounding và tool request policy đều đi qua `core/m07`. Model-declared tool
  request không có adapter-registered trace bị từ chối trước grounding.
- `fixture_update`: M07 eval/testdata dùng output contract đầy đủ
  `HUMAN_REVIEW`/claims/evidence/authority/write boundary thay cho legacy
  `PROPOSE` payload không được learner Bot sử dụng.
- `negative_cases`: unknown tool, write method, host/port/userinfo,
  hallucinated evidence ID, duplicate/unknown/null/trailing fields, forged value
  dưới known evidence ID và ambiguous allowlist entries.

Commands run on the real implementation:

```text
(cd core && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...) PASS
(cd lab/mission-runtime && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...) PASS
(cd lab/affiliate-bot && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...) PASS
```

## Boundary

This proves local shared-code conformance and fail-closed fixture behavior. It
does not prove provider traffic, live executor, business outcome, authenticated
ACCESSTRADE capture, clean-machine pilot, deployment or multi-host/power-loss
durability. Overall readiness remains `NOT_READY_FOR_PRODUCTION`.
