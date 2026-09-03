# Kiến trúc chuẩn — Governed Affiliate Intelligence Bot

## 1. Architecture authority

Curriculum quyết định **khi nào** capability/authority được mở. Architecture quyết định **ranh giới hệ thống**; vendor/framework không sở hữu semantics.

```text
Read-only Watcher / Agent (M06/M07)
→ Observation / Evidence
→ Normalize + Validate
→ Canonical History / Replay
→ Deterministic Domain Core
→ DecisionPacket / Grounded AI Advisor
→ ActionIntent (M08+)
→ Deterministic Policy / Risk

M09:
Human ApprovalRecord → revalidation → ExecutionAuthorization(APPROVED_LIVE) → Executor → ExecutionRecord

M10:
Human-approved CanaryGrant + CanaryLedger + trusted cost
→ CanaryGateDecision
→ ExecutionAuthorization(GOVERNED_CANARY)
→ Executor → ExecutionRecord → OutcomeRecord

M11:
E5 promotion review → ProductionLeaseApproval → finite ProductionLease
+ trusted ProductionHealthSnapshot + trusted cost + durable ProductionLedger
→ ProductionGateDecision
  ALLOW_PRODUCTION | DEGRADE | STOP | REQUIRE_APPROVAL | WAIT | DENY
→ ExecutionAuthorization(GOVERNED_PRODUCTION) only when ALLOW_PRODUCTION
→ Executor reloads durable state + revalidates exact lease/policy/health/cost/budget/kill switch
→ ExecutionRecord → OutcomeRecord → EvaluationRecord
→ ImprovementProposal(auto_apply=false) → Human ReviewRecord
↺
```

## 2. Ownership

### Deterministic Domain / Governance Core

Sở hữu canonical evidence/history, deterministic decision/policy/risk, approval/authorization, CanaryGrant/ProductionLease semantics, trusted cost/health bindings, budget/rate/outcome backpressure, sticky STOP/reconciliation, audit/correlation và cross-artifact linkage.

### Orchestration

Sở hữu trigger/schedule/integration/retry/routing. Orchestrator không tự trở thành policy/approval/grant/lease authority và không tự clear STOP.

### AgentRuntime

Sở hữu unstructured research/reasoning/proposal trong permission ceiling. Agent không sở hữu truth, trusted cost/health, approval, CanaryGrant, ProductionLease hoặc authorization.

## 3. Invariants

```text
Decision != Approval != Execution
AI confidence != execution permission
Tool result != trusted evidence
Agent proposal != authorized ActionIntent
Schema-valid reference != resolved provenance
ApprovalRecord != ExecutionAuthorization
CanaryGrant != blanket approval
ProductionLease != infinite authority
GateDecision != ExecutionAuthorization
Budget check without atomic reservation != safe budget enforcement
DEGRADE = read-only / no side effect
STOP = sticky until human-reviewed recovery/new lease version
UNKNOWN side effect → RECONCILIATION_REQUIRED → STOP → no automatic retry
ImprovementProposal.auto_apply = false
Automation may narrow/stop authority; it may not widen its own authority
```

## 4. Implementation flexibility

Go là deterministic reference khi code giảm ambiguity. n8n có thể orchestration. OPA chỉ adopt khi policy complexity/parity gate biện minh. Temporal chỉ khi durability pain thật vượt persisted-state baseline. OpenTelemetry hỗ trợ telemetry/correlation nhưng không thay canonical audit/health authority.

`lab/mission-runtime` là conformance/integration harness, không phải Bot thứ hai.

## 5. External action boundary

M03 human action → M08 shadow intent → M09 per-action approved execution → M10 governed canary → M11 finite governed production. `RISK2` không được auto ở M10/M11. Promotion, renewal, scope/budget/risk widening và recovery sau sticky STOP đều cần human governance path.
