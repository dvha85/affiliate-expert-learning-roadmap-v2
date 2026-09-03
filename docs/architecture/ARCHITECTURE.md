# Kiến trúc chuẩn — Governed Affiliate Intelligence Bot

## 1. Architecture authority

Curriculum quyết định **khi nào** capability/authority được mở. Architecture này quyết định **ranh giới hệ thống**; vendor/framework không sở hữu semantics.

```text
Read-only Watcher / Agent (M06/M07)
→ Observation / Evidence
→ Normalize + Validate
→ Canonical History / Replay
→ Deterministic Domain Core
→ DecisionPacket / Grounded AI Advisor
→ ActionIntent (M08+)
→ Deterministic Policy / Risk
→ DENY / WAIT / GET_MORE_DATA / HUMAN_REVIEW / ALLOW

M09 path:
Human ApprovalRecord
→ deterministic revalidation + kill switch
→ ExecutionAuthorization(APPROVED_LIVE)
→ Controlled Executor
→ ExecutionRecord

M10 path:
Human-approved CanaryGrant
+ CanaryLedger
→ CanaryGateDecision
  ALLOW_CANARY | REQUIRE_APPROVAL | WAIT | DENY
→ ExecutionAuthorization(GOVERNED_CANARY) only when ALLOW_CANARY
→ Controlled Executor reloads durable ledger + revalidates grant/policy/scope/budget/kill switch
→ ExecutionRecord
→ OutcomeRecord
→ ledger backpressure release

→ Outcome → Evaluation → Reviewed Improvement
```

M03 là ngoại lệ có chủ đích trước machine `ActionIntent`: người thực hiện external action thật, hệ thống chỉ ghi/validate `ActionRecord` và `OutcomeRecord`.

## 2. Ownership

### Deterministic Domain / Governance Core

Sở hữu contract/behavior của:

- evidence schema và validation;
- identity/provenance/freshness;
- canonical history/replay;
- deterministic decision states;
- `DecisionPacket`, `ActionRecord`, `OutcomeRecord`, `EvaluationRecord`, `ImprovementProposal`, `ReviewRecord`, `ActionIntent`, `PolicyDecision`;
- `ApprovalRecord`, `ExecutionAuthorization`, `ExecutionRecord` từ M09;
- `CanaryGrant`, `CanaryLedger`, `CanaryGateDecision` từ M10;
- risk/authorization semantics;
- budget/rate/outcome-backpressure semantics;
- audit/correlation invariants;
- cross-artifact linkage khi contract yêu cầu.

### Orchestration

Sở hữu trigger/schedule/integration/retry/approval routing và bounded execution plumbing. Orchestrator không tự trở thành policy authority, approver hoặc grant authority. n8n output phải map về canonical contract thay vì tạo data model song song.

### AgentRuntime

Sở hữu unstructured research/reasoning/proposal trong permission ceiling. Agent không sở hữu truth, approval, CanaryGrant hoặc authorization. Tool output là untrusted data cho tới khi qua deterministic validation/grounding.

## 3. Invariants

```text
Decision != Approval != Execution
AI confidence != execution permission
Tool result != trusted evidence
Agent proposal != authorized ActionIntent
Schema-valid reference != resolved provenance
ApprovalRecord != ExecutionAuthorization
CanaryGrant != blanket approval
CanaryGateDecision != ExecutionAuthorization
Persisted workflow state != permission
Budget check without atomic reservation != safe budget enforcement
UNKNOWN side effect → RECONCILIATION_REQUIRED → no automatic retry
Deterministic Policy unavailable/invalid/unverified
→ no consequential execution
```

Từ M09, resume/retry phải revalidate exact intent hash, policy version, approval/authority source, expiry, executor profile, idempotency và kill switch ngay trước side effect.

Từ M10, executor còn phải reload durable `CanaryLedger` và revalidate exact `CanaryGrant` hash/version, revocation, scope, total/rate/cost budget và outcome backpressure. `RISK2` không được đi qua `GOVERNED_CANARY`.

## 4. Implementation flexibility

```text
DETERMINISTIC CORE FIRST != CODE FIRST
```

Go là deterministic reference/fallback khi code làm behavior rõ hơn. Visual rule engine hoặc OPA có thể implement deterministic semantics nếu parity/version/audit/fail-closed gate đạt. n8n có thể orchestration. Temporal chỉ có ý nghĩa khi durability pain thực tế vượt baseline. Agent runtime có thể thay đổi mà không đổi authority model.

`lab/mission-runtime` là conformance/integration harness (bộ kiểm tương thích/tích hợp), không phải một Bot thứ hai thay thế `lab/affiliate-bot`. Mission sau phải reuse canonical contracts/history hoặc chứng minh adapter rõ ràng.

## 5. External action boundary

External action đầu tiên ở M03 và do human thực hiện. Machine ActionIntent bắt đầu shadow ở M08. Machine execution chỉ mở ở M09 qua human approval + deterministic revalidation + controlled executor. Bounded auto-action không cần approval từng lần chỉ bắt đầu ở M10 và chỉ trong human-approved `CanaryGrant` có scope/budget/time/risk nhỏ; `RISK2` vẫn quay về phê duyệt từng hành động. M11 mới là production closed loop có quản trị.
