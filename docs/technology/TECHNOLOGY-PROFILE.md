# Technology Profile — current implementation reference

**Last reviewed:** 2026-09-03  
**Status:** Reference/adoption guide, không phải curriculum authority.

## 1. Rule chung

```text
Tool available != tool adopted
Tool adopted != tool owns truth
Framework capability != Bot authority
Tool changes != Curriculum changes
```

Mọi tool chỉ được adopt khi giải quyết bottleneck đã quan sát được và không vượt authority ceiling của Mission.

## 2. Current profile

| Capability | Primary/reference | Candidate/comparison | Earliest meaningful Mission | Default |
|---|---|---|---|---|
| Deterministic core | contracts + Go reference | visual rule engine; OPA later | M01 / M08+ | Go/reference first when it reduces ambiguity |
| Orchestration | n8n | Windmill | M06 | n8n primary |
| AgentRuntime | n8n AI Agent visual-first | OpenAI Agents SDK; Hermes Agent | M07 | reuse current runtime first |
| Tool boundary | explicit Tool Registry | MCP | M07 | MCP preferred interoperability candidate |
| Browser acquisition | HTTP/API/manual | Playwright | M06 | browser only when required |
| Observability | correlation/audit contract | OpenTelemetry | M06 | adopt across runtime boundaries |
| AI eval backend | repo fixtures/evals | Langfuse | M04 | optional after repo baseline |
| Policy engine | deterministic rules/contracts | OPA | M08/M09 | adopt on real policy complexity |
| Durable workflow | n8n + canonical persisted state | Temporal | M09 | only on real durability pain |
| Orchestration alternative | n8n | Windmill | M06+ | compare only on measured ops benefit |

## 3. Mission timing

### M00

Không cần Go, n8n, Agent, MCP hoặc AI. Chỉ E1 evidence + Human DecisionPacket.

### M01

Go là deterministic reference/fallback cho Bot v0.1. Không dùng Agent để thay deterministic semantics.

### M02

History/replay trước. Canonical history không phụ thuộc observability vendor.

### M03

Human external action + measurement context. Không machine-execute.

### M04

Grounded AI advisory. Langfuse có thể hỗ trợ experiment/eval sau khi repo fixture/eval baseline tồn tại; Langfuse score không trở thành evidence hoặc policy input mặc định.

### M05

Reviewed improvement/proposal only. Không auto-merge behavioral/policy change.

### M06

Automatic read-only watcher. n8n là orchestration reference. OpenTelemetry được khuyến nghị cho correlation xuyên runtime. Playwright chỉ khi HTTP/API/public manual baseline không đủ.

### M07

Read-only Evidence Agent. Ưu tiên n8n AI Agent nếu runtime hiện có đáp ứng Safe Profile. MCP là tool interoperability candidate với allowlist, least privilege, timeout/audit và tool output untrusted. OpenAI Agents SDK/Hermes chỉ compare khi có measured bottleneck.

### M08

Shadow `ActionIntent` + deterministic policy. Có thể spike visual rule engine hoặc OPA nhưng output chỉ shadow; no execution.

### M09

Approval-gated execution. Temporal chỉ spike nếu wait/resume/retry/recovery thực sự vượt persisted-state+n8n baseline. OPA phù hợp nếu policy complexity cần policy-as-code riêng.

### M10–M11

Bounded governed automation rồi production closed loop. Technology không được tự mở authority; live activation vẫn cần evidence/policy/audit/kill-switch/recovery gate.

## 4. OpenTelemetry

Role: telemetry/correlation protocol, không phải canonical audit/business state.

Gate:

```text
correlation_id mapping exists
+ redaction
+ exporter outage does not break deterministic core
+ sampling cannot remove mandatory audit
```

OpenTelemetry Go hiện có traces và metrics stable; logs còn beta tại review 2026-09-03.

Official ref: https://opentelemetry.io/docs/languages/go/

## 5. MCP

MCP 2026-07-28 chuyển core protocol sang stateless request/response và tăng authorization hardening. Đây là interoperability protocol, không phải permission model thay thế.

```text
MCP tool visible != tool permitted
MCP call succeeded != result trusted
MCP auth succeeded != action authorized
```

Gate:

- explicit server/tool allowlist;
- least-privilege credentials;
- read-only ceiling ở M07;
- token/secret không model-visible;
- correlation/audit metadata;
- timeout/retry/failure semantics;
- sensitive/write tools đi qua separate deterministic policy + approval.

Official ref: https://modelcontextprotocol.io/specification/2026-07-28

## 6. OpenAI Agents SDK

Comparison runtime, không phải Core dependency. SDK hiện có tools, guardrails, handoffs, sessions, tracing, MCP integration và human-in-the-loop pause/approve/resume.

Gate:

```text
same Safe Profile/eval set
+ explicit tool filters
+ least privilege
+ HITL/guardrails tested
+ tool result untrusted until deterministic validation
+ deterministic fallback
+ trace maps to canonical correlation/audit
+ measured benefit > added code/ops burden
```

Official refs:
- https://openai.github.io/openai-agents-python/
- https://openai.github.io/openai-agents-python/human_in_the_loop/
- https://openai.github.io/openai-agents-python/mcp/

## 7. Langfuse

Optional backend cho traces/experiments/evaluation. Langfuse hiện hỗ trợ experiments qua OpenTelemetry.

Không đặt vào Langfuse:
- canonical evidence/history;
- final grounded truth;
- authorization;
- mandatory audit record duy nhất.

Gate:

```text
repo eval baseline exists first
+ datasets/labels versioned outside vendor-only state
+ private data redacted
+ backend outage does not break deterministic core
+ measured debugging/eval value > ops cost
```

Official refs:
- https://langfuse.com/docs/evaluation/overview
- https://langfuse.com/docs/evaluation/experiments/experiments-via-opentelemetry

## 8. Playwright

Controlled browser acquisition candidate. Dùng khi source public cần browser rendering/interaction mà HTTP/API đơn giản không đủ.

Gate:
- URL/domain allowlist;
- bounded navigation/timeout/rate limit;
- no arbitrary form submit/upload/account mutation ở read-only Missions;
- provenance gồm source URL + observed_at;
- platform terms/compliance reviewed;
- version/browser binaries được pin/review trong CI.

Official ref: https://playwright.dev/docs/browsers

## 9. OPA

OPA là general-purpose policy engine tách policy decision khỏi enforcement. Phù hợp khi policy complexity vượt simple rules/visual table.

Gate:

```text
policy complexity proven
+ structured canonical inputs
+ versioned/reviewed Rego bundle
+ parity tests with current policy baseline
+ timeout/error/undefined fail closed
+ decision log/reason audit
+ no canonical state mutation
```

Official refs:
- https://www.openpolicyagent.org/docs
- https://www.openpolicyagent.org/docs/integration

## 10. Windmill

Comparison orchestrator khi cần code-friendly scripts/Git sync hoặc operational model khác. Windmill hiện hỗ trợ bidirectional Git sync trong các cấu hình được tài liệu hóa.

Adopt thay n8n chỉ khi cùng use case chứng minh:

```text
same contracts + same authority ceiling
+ Git-reviewable artifacts
+ retry/idempotency rõ
+ secret handling đạt
+ correlation/audit không kém
+ measured operational burden thấp hơn
```

Official ref: https://www.windmill.dev/docs/advanced/git_sync

## 11. Temporal

Durable execution candidate cho M09+ khi có real long-running wait/resume/recovery pain. Không thêm chỉ vì workflow có approval.

Gate:
- documented durability/recovery need;
- idempotent/dedup activities;
- canonical business state không bị khóa vào workflow history;
- resume revalidates approval/policy/kill switch;
- operational complexity justified.

Official ref: https://docs.temporal.io/

## 12. Development Agent

Coding agent/Codex/Copilot/Claude thuộc development plane, không thuộc runtime authority.

```text
issue/spec
→ agent implementation + tests
→ PR
→ CI/security checks
→ human review
→ merge/reject
```

Agent-authored code không được auto-merge consequential policy/runtime changes chỉ vì CI xanh hoặc model confidence cao.
