---
mission_id: M06
title: Reliable Automatic Watcher
status: ready
minimum_evidence: E4
authority: automatic read-only
external_side_effects: false
runtime: lab/mission-runtime/ + lab/n8n/
eval_pack: evals/M06-readonly-watcher/
---

# Mission M06 — Reliable Automatic Watcher (bộ theo dõi tự động chỉ đọc đáng tin)

## Contract bàn giao

Read-only watcher; GET/HEAD allowlist; change detection; correlation; n8n blueprint.

## Authority ceiling

Không machine execution có side effect trong Mission này.

## PASS

`Capability + Reality + Operated`. Synthetic/offline eval chỉ tạo capability proof; learner vẫn cần evidence phù hợp Mission để claim Reality/Operated PASS.
