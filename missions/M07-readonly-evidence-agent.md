---
mission_id: M07
title: Read-only Evidence Agent
status: ready
minimum_evidence: E4
authority: A2-RO
external_side_effects: false
runtime: lab/mission-runtime/ + lab/n8n/
eval_pack: evals/M07-readonly-evidence-agent/
---

# Mission M07 — Read-only Evidence Agent (Agent thu thập bằng chứng chỉ đọc)

## Contract bàn giao

Read-only Tool Registry; Agent proposal grounding; prompt-injection boundary; n8n Agent blueprint.

## Authority ceiling

Không machine execution có side effect trong Mission này.

## PASS

`Capability + Reality + Operated`. Synthetic/offline eval chỉ tạo capability proof; learner vẫn cần evidence phù hợp Mission để claim Reality/Operated PASS.
