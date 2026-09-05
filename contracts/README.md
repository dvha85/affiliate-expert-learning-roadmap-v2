# Canonical contracts

Các schema trong thư mục này mô tả boundary máy đọc của hệ thống. Contract semantics thuộc Deterministic Core, không thuộc Agent/orchestrator/vendor.

## Artifact spine chuẩn

```text
Observation(observation_id, subject_id, provenance)
→ DecisionPacket(evidence_ids)
→ HistoryRecord(reuse canonical Observation + recorded decision linkage)
→ ActionRecord hoặc ActionIntent
→ Execution
→ EffectRef(HUMAN_ACTION | MACHINE_EXECUTION)
→ OutcomeRecord(effect_ref)
→ EvaluationRecord(same effect_ref)
→ ImprovementProposal + ReviewRecord
```

Một artifact có ID nhưng downstream reference không resolve được vẫn là broken lineage (liên kết đứt), không đủ để claim Reality/Operated PASS. `EffectRef` ngăn một `execution_id` bị giả làm `action_id` khi hệ thống chuyển từ M03 human action sang M09+ machine execution.

## Semantics không cấp quyền

```text
ActionIntent.intent_mode = PROPOSAL_ONLY
ActionIntent.execution_authorized = false
PolicyDecision.policy_mode = NON_AUTHORIZING
PolicyDecision.execution_authorized = false
```

`policy_review_required` chỉ là tín hiệu review của policy layer. Execution approval/authority nằm ở `ApprovalRecord`, grant/lease governance, gate và `ExecutionAuthorization` tương ứng; không dùng một boolean generic để trộn các lớp này.

## Activation theo Mission

- M00: canonical `Observation` + human `DecisionPacket`; DecisionPacket phải bind exact `evidence_ids` và `action=null`.
- M01: deterministic decision behavior trên canonical observation/context; fixture synthetic vẫn chỉ là E0.
- M02: `HistoryRecord` reuse canonical `Observation` cho immutable snapshot + version + replay; recorded decision phải bind `decision_id` + exact `evidence_ids`.
- M03: `ActionRecord` human-only + `OutcomeRecord(effect_ref=HUMAN_ACTION)`.
- M04: `AdvisorOutput` grounded, không write.
- M05: `EvaluationRecord` dùng cùng `EffectRef` với Outcome + `ImprovementProposal` + `ReviewRecord`; `auto_apply=false`.
- M06: watcher normalize về canonical `Observation`.
- M07: `ToolRegistry` read-only; tool output untrusted.
- M08: `ActionIntent(PROPOSAL_ONLY)` + `PolicyDecision(NON_AUTHORIZING)`; không artifact nào tự có execution authority.
- M09: `ApprovalRecord` + `ExecutionAuthorization(APPROVED_LIVE)` + `ExecutionRecord`; outcome machine dùng `EffectRef(MACHINE_EXECUTION)`.
- M10: `CanaryGrantApproval` + `CanaryGrant` + trusted stage-neutral `TrustedCostBound` + `CanaryLedger` + `CanaryGateDecision`; `RISK2` không delegated.
- M11: `ProductionLeaseApproval` + finite `ProductionLease` + `ProductionActivationRecord` + trusted `ProductionHealthSnapshot` + `TrustedCostBound` + `ProductionLedger` + `ProductionGateDecision` + `ProductionCycleRecord` + trusted `ProductionReconciliationResolution`; execution dùng `GOVERNED_PRODUCTION` và `RISK2` vẫn không delegated.

`TrustedCostBound` là một canonical contract dùng chung M10/M11; stage-specific authorization field vẫn giữ prefix `canary_`/`production_` để audit provenance. Không tạo hai cost-bound ontology chỉ vì Mission khác nhau.

Contract tồn tại **không tự cấp authority**. `ApprovalRecord`, `CanaryGrant` hay `ProductionLease` đều không phải execution. Gate decision luôn `execution_authorized=false`; executor chỉ chạy khi có authorization riêng, exact binding và deterministic revalidation. Agent/orchestrator không sở hữu truth, trusted cost/health, promotion, lease renewal hoặc authority widening.

## Schema runtime — BR-03b

Các file `*.schema.json` ở đây vẫn là canonical source. Go package chỉ nhúng chúng bằng `go:embed`, không sao chép schema vào module bot. `lab/affiliate-bot/go.mod` dùng local replace `../../contracts`; phải clone đầy đủ repo, không tách riêng thư mục bot.

## Dependency và kiểm thử

Pin `github.com/santhosh-tekuri/jsonschema/v6 v6.0.2`, cùng dependency gián tiếp trong go.mod/go.sum. Đây là bộ kiểm JSON Schema hỗ trợ Draft 2020-12 và `$ref`/`allOf`; bật `AssertFormat` để date-time/URI không chỉ là annotation. Tham khảo [API chính thức phiên bản đã pin](https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6@v6.0.2).

```bash
cd contracts
go mod download
go test ./...
go vet ./...
```

Lần tải dependency đầu cần mạng; sau khi build, schema nằm trong binary. Loader không cho tải URL/file ngoài bộ đã nhúng, kể cả tên schema do caller đưa vào. Binary phải được build lại khi schema đổi. CI chạy tests package này trong job deterministic-runtime.

## Boundary

`ValidateRaw` kiểm JSON gốc và schema; JSON duplicate key/trailing data bị từ chối. `DecodeStrict` kiểm exact spelling và field mà Go type hỗ trợ, thay vì để decoder bỏ field lạ/case alias. Schema cho extension không có nghĩa Go projection giữ được extension: M02 từ chối field chưa hỗ trợ, không bỏ dữ liệu ngầm.

Giới hạn độ sâu JSON 128; file/JSONL vẫn có giới hạn I/O riêng của caller. Package này không kiểm provenance, freshness, linkage, input hash hoặc quyền thực thi; runtime giữ các kiểm tra đó sau kiểm schema. Compile đủ schema không chứng minh output mọi Mission đã conformance; BR-03c còn phải nối boundary vào từng luồng.

## Tương thích history cũ

M02 trước BR-03b có thể serialize ranking rỗng thành `ranked:null`, dù schema đòi array. Không đổi schema để nhận mọi dạng sai: chỉ tại loader/validator history, khi **field ranked tồn tại và bằng null**, kiểm một view rỗng `[]`; giữ nguyên object/file dùng cho replay. Record thiếu field ranked vẫn bị từ chối. Record mới luôn xuất array. Replay coi null và array rỗng của riêng ranked có cùng nghĩa “không có phần tử”; không normalize các field khác hoặc rewrite history cũ.

Đây là ngoại lệ tương thích minh bạch, có regression test. Hash observations và formula version không đổi; không áp dụng ngoại lệ này cho DecisionPacket hoặc artifact khác.
