# Affiliate Intelligence Bot — lộ trình học v2

Repo chính để học và xây một **Affiliate Intelligence Bot tiến hóa dần từ bằng chứng thật tới tự động hóa có kiểm soát**.

> Repo lịch sử: `dvha85/affiliate-expert-learning-roadmap`.

## Bắt đầu

1. Đọc `CURRICULUM.md` — authority duy nhất về Mission sequence, evidence, autonomy và PASS.
2. Đọc `curriculum/README.md` — learner path hiện hành.
3. Nếu chưa từng chạy/test Bot, học `BOOT.1`.
4. Bắt đầu M00 tại `curriculum/M00/M00.1-affiliate-intelligence-objective.md`.

## Canonical progression

```text
O00
→ M00 real evidence
→ M01 deterministic Bot
→ M02 trustworthy history/replay
→ M03 tracked human action + measurement
→ M04 grounded AI advisor
→ M05 reviewed improvement
→ M06 automatic read-only watcher
→ M07 read-only evidence Agent
→ M08 shadow ActionIntent + policy
→ M09 approval-gated executor
→ M10 governed canary
→ M11 production closed loop
```

## Control invariants

```text
AI confidence != execution permission
Decision != Approval != Execution
Agent proposal != authorized ActionIntent
Tool result != trusted evidence
real evidence != automatic recommendation
```

## Technology principle

```text
MISSION DEFINES CAPABILITY + AUTHORITY
TECHNOLOGY PROFILE DEFINES IMPLEMENTATION

Tool changes != Curriculum changes
```

Go, n8n, MCP, OpenTelemetry, Langfuse, Playwright, OpenAI Agents SDK, Hermes Agent, Windmill, Temporal, OPA và rule engine chỉ được adopt khi giải quyết bottleneck đã quan sát được và không vượt authority ceiling của Mission.

## Repo policy

Repo v2 không mang compatibility/migration layer từ repo cũ. Legacy syllabus, numeric lesson map, duplicate Mission, migration scripts và historical runtime chỉ tồn tại ở repo lịch sử.
