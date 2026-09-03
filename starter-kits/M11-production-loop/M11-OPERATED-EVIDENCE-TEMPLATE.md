# M11 Operated Evidence — mẫu E6 production closed loop

## 1. Promotion từ E5

- E5 canary evidence refs (tham chiếu bằng chứng E5):
- Source CanaryGrant ID/version/hash (CanaryGrant nguồn — ID/version/hash):
- ProductionLeaseApproval ID + reviewer:
- Validated risk classes (lớp rủi ro đã xác nhận):
- Promotion review ref (tham chiếu review nâng cấp):

## 2. ProductionLease

- Lease ID/version/hash (ID/version/hash của lease):
- Policy version:
- Validity window (khoảng hiệu lực):
- Risk/action/host/executor scope (phạm vi rủi ro/hành động/host/executor):
- Total/rate/cost budgets (ngân sách tổng/tần suất/chi phí):
- Max pending outcomes:
- Failure/outcome-age/health-age thresholds (ngưỡng lỗi/tuổi kết quả/tuổi health):

## 3. Health + recovery controls

- Health snapshot sources/provenance (nguồn gốc):
- Kill-switch test:
- DEGRADE drill (bài thử hạ chế độ):
- STOP drill + sticky restart evidence (bài thử STOP + bằng chứng restart vẫn dừng):
- Recovery runbook:
- Lease/version mới sau STOP nếu resume:

## 4. Observation window thật

- Window start/end:
- Số closed production cycles:
- Observation/Evidence refs:
- Production ledger trước/sau:
- Cost/execution/outcome totals (tổng chi phí/lần thực thi/kết quả):

## 5. Cycle linkage

Lặp cho ít nhất 3 cycle:

- ProductionCycleRecord ID:
- Decision + ActionIntent ID/hash (ID/hash của Decision + ActionIntent):
- ProductionGateDecision ID/reason:
- ExecutionAuthorization ID/mode:
- ExecutionRecord ID:
- OutcomeRecord ID/source:
- EvaluationRecord ID:

## 6. Failure/recovery evidence (bằng chứng lỗi/phục hồi)

- Stale/tampered health:
- Compliance/failure/reconciliation stop (STOP do tuân thủ/lỗi/đối soát):
- Missing durable ledger/restart (mất ledger bền vững/restart):
- Wrong outcome→execution link (liên kết outcome→execution sai):
- Duplicate/idempotency:
- Bằng chứng RISK2 không auto:

## 7. Reviewed improvement

- Evaluation IDs:
- ImprovementProposal ID/version + `auto_apply=false`:
- Human ReviewRecord ID/decision (ID/quyết định ReviewRecord của người):
- Nếu thay policy/lease: version mới + approval mới:
- Rollback/recovery note:

## 8. Authority audit

- Có code/path nào tự renew lease không?: KHÔNG
- Có code/path nào tự tăng budget/risk/scope không?: KHÔNG
- Có code/path nào tự clear STOP/reconciliation không?: KHÔNG
- Có code/path nào auto-apply improvement không?: KHÔNG
