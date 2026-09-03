# Roadmap — Affiliate Intelligence Bot v2

`CURRICULUM.md` là authority. File này chỉ là projection dễ đọc của Mission spine.

## Giai đoạn A — Ground truth trước automation

- **O00** — Safe synthetic walkthrough.
- **M00** — First Real Evidence Packet + Human DecisionPacket.
- **M01** — Smallest Deterministic Bot v0.1.
- **M02** — Trustworthy History + Replay v0.2.
- **M03** — First Tracked Human Action + Outcome context.

## Giai đoạn B — AI có grounding nhưng chưa có quyền hành động

- **M04** — Grounded AI Advisor v0.4.
- **M05** — First Reviewed Improvement.

## Giai đoạn C — Read-only automation và Agent

- **M06** — Reliable Automatic Watcher.
- **M07** — Read-only Evidence Agent.

## Giai đoạn D — Controlled action

- **M08** — Shadow ActionIntent + deterministic policy.
- **M09** — Durable Approval + Controlled Executor.
- **M10** — Governed Canary.
- **M11** — Production Closed Loop.

## Nguyên tắc chuyển Mission

Không chuyển Mission chỉ vì code đã viết xong. Phải đạt đúng evidence + authority + operated gate của Mission hiện tại.

```text
Capability PASS
+ Reality PASS khi required
+ Operated PASS khi required
→ Mission PASS
```

Technology không quyết định thứ tự học. Go/n8n/Agent/MCP/Temporal/OPA chỉ được đưa vào khi Mission hiện tại có nhu cầu và adoption gate đạt.
