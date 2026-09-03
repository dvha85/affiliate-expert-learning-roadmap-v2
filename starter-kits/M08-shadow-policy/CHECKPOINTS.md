# M08 Checkpoints

- [ ] ActionIntent liên kết tới Decision/Evidence có thật.
- [ ] Agent-origin intent có `proposal_ref`; Agent proposal không tự cấp authority.
- [ ] `shadow_only=true` và `dry_run=true`.
- [ ] intent hash đổi khi candidate action bị sửa.
- [ ] expiry, target allowlist và policy version được kiểm.
- [ ] idempotency duplicate/collision semantics được thử.
- [ ] PolicyDecision luôn `execution_authorized=false`, kể cả `ALLOW`.
- [ ] policy unavailable/invalid fail closed.
- [ ] không có write tool/executor trong operated path.
- [ ] Reality + Operated evidence đã lưu.
