---
mission_id: M04
title: Grounded AI Advisor v0.4
status: ready
minimum_evidence: E3
authority: A1 advisory
external_side_effects: false
runtime: lab/mission-runtime/
eval_pack: evals/M04-grounded-ai-advisor/
---

# Mission M04 — Grounded AI Advisor v0.4 (AI tư vấn dựa trên bằng chứng)

## Contract bàn giao

Grounded AdvisorOutput; evidence IDs; freshness; abstention; no write tool.

## Authority ceiling

Không machine execution có side effect trong Mission này.

## PASS

`Capability + Reality + Operated`. Synthetic/offline eval chỉ tạo capability proof; learner vẫn cần evidence phù hợp Mission để claim Reality/Operated PASS.
