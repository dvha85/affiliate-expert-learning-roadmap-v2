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

Watcher (bộ theo dõi) chỉ đọc dùng GET/HEAD trên allowlist (danh sách cho phép), phân biệt NEW/UNCHANGED/CHANGED, có correlation (liên kết lần chạy) và n8n blueprint (bản thiết kế n8n).

## Authority ceiling

Không machine execution có side effect trong Mission này.

## PASS

`Capability + Reality + Operated`. Synthetic/offline eval chỉ tạo capability proof; learner vẫn cần evidence phù hợp Mission để claim Reality/Operated PASS.
