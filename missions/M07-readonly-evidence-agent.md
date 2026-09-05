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

Tool Registry (sổ đăng ký tool) chỉ đọc; Agent proposal (đề xuất của Agent) phải bám bằng chứng; prompt injection (chèn chỉ dẫn độc hại) không được nâng quyền; có n8n Agent blueprint (bản thiết kế Agent trên n8n).

## Authority ceiling

Kiểm offline [m07-check](../docs/architecture/M07-JSON-BOUNDARY.md) không cấp quyền gọi tool. SUPPORTED chỉ kiểm ID/registry/call hiện có, không xác minh nội dung claim hoặc live enforcement.

Không machine execution có side effect trong Mission này.

## PASS

`Capability + Reality + Operated`. Synthetic/offline eval chỉ tạo capability proof; learner vẫn cần evidence phù hợp Mission để claim Reality/Operated PASS.
