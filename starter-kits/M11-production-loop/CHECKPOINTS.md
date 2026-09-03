# M11 checkpoints — điểm kiểm production closed loop

- [ ] `ProductionLeaseApproval` bind exact `lease_hash` + source E5 refs + validated risk classes (tham chiếu E5 nguồn + lớp rủi ro đã xác nhận).
- [ ] `ProductionLease` hữu hạn theo expiry/budget/rate/cost/outcome; không wildcard và không `RISK2`.
- [ ] RISK1 chỉ production khi E5 promotion approval đã validate RISK1.
- [ ] Trusted cost bound bind exact intent; Agent cost hint không sở hữu budget.
- [ ] Trusted health snapshot bind exact lease hash và có freshness limit.
- [ ] `DEGRADE` = read-only/no side effect (chỉ đọc/không có tác động bên ngoài).
- [ ] Compliance/reconciliation/failure/outcome-age/kill/revoke tạo `STOP`.
- [ ] `STOP` được persist và vẫn sticky sau restart.
- [ ] Resume sau STOP cần human-reviewed lease version mới; không auto-clear.
- [ ] Executor reload durable ledger + revalidate lease/policy/health/cost/budget ngay trước side effect.
- [ ] Unknown side effect → reconciliation + STOP; không retry mù.
- [ ] Outcome bind đúng execution trước khi release backpressure.
- [ ] `ProductionGateDecision.execution_authorized=false`.
- [ ] Chỉ `ExecutionAuthorization(execution_mode=GOVERNED_PRODUCTION)` mở execution cụ thể.
- [ ] Có 3+ real closed cycles qua observation window cho E6.
- [ ] Có Evaluation → ImprovementProposal(`auto_apply=false`) → Human ReviewRecord.
- [ ] Policy/lease/budget/credential/tool permission không tự widen.
- [ ] `PROGRESS.md` chỉ đổi khi learner thực sự có E6 evidence.
