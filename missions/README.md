# Chỉ mục Mission — trục thực thi chuẩn

Thứ tự học được quyết định bởi `CURRICULUM.md`.

`missions/manifest.json` là chỉ mục machine-readable duy nhất cho structural
spine O00→M11. `scripts/validate_missions.py` kiểm common assets/path; các
validator khác là semantic plug-ins theo boundary, không sở hữu một spine riêng.

| Mission | Kết quả | Quyền hạn | Trạng thái |
|---|---|---|---|
| O00 | Safe synthetic walkthrough (mô phỏng tổng thể an toàn), không PASS | không side effect | ready orientation |
| M00 | First Real Evidence Packet + Human DecisionPacket (gói bằng chứng thật đầu tiên + gói quyết định do người lập) | người/read-only | ready |
| M01 | Smallest Deterministic Bot v0.1 (Bot tất định nhỏ nhất) | A0 tất định | ready |
| M02 | Trustworthy History + Replay v0.2 (lịch sử đáng tin + phát lại) | A0 tất định | ready |
| M03 | First Tracked Human Action + Outcome context (hành động thật đầu tiên do người làm + ngữ cảnh kết quả) | người thực thi | ready |
| M04 | Grounded AI Advisor v0.4 (AI tư vấn dựa trên bằng chứng) | A1 tư vấn | ready |
| M05 | First Reviewed Improvement (cải tiến đầu tiên có review) | A1 chỉ đề xuất | ready |
| M06 | Reliable Automatic Watcher (bộ theo dõi tự động chỉ đọc đáng tin) | tự động chỉ đọc | ready |
| M07 | Read-only Evidence Agent (Agent bằng chứng chỉ đọc) | A2-RO | ready |
| M08 | Shadow ActionIntent + Policy (ActionIntent chạy bóng + chính sách) | A3-shadow | ready |
| M09 | Durable Approval + Controlled Executor (phê duyệt bền vững + bộ thực thi có kiểm soát) | qua cổng phê duyệt | ready |
| M10 | Governed Canary (canary có quản trị) | tự động RISK0/RISK1 trong grant giới hạn; RISK2 phê duyệt từng lần | ready |
| M11 | Production Closed Loop (vòng kín production có quản trị) | finite production lease cho RISK0/RISK1; RISK2 phê duyệt từng lần | ready |

```text
O00 → M00 → M01 → M02 → M03 → M04 → M05 → M06 → M07 → M08 → M09 → M10 → M11
```

`ready` nghĩa là lesson + Mission contract + starter README + checkpoints + operated-evidence template + executable eval/runtime + CI guard đã được soạn. `ready` **không** có nghĩa người học đã PASS; learner PASS phụ thuộc evidence thật và operated result.

Từ M03–M11, `ready` còn giả định learner áp dụng `starter-kits/CONTINUITY-CHECKPOINT.md`: capability mới phải gắn vào cùng learner Bot/workspace, previous artifact refs resolve được và `lab/mission-runtime` chỉ được dùng như conformance oracle chứ không thay Integration/Reality evidence.

M06/M07 có n8n workflow để learner vận hành thật. M06 n8n static data chỉ là watcher cache/change-detection state; canonical Observation/History thuộc Deterministic Core. M08 chỉ shadow. M09 dùng human approval từng execution. M10 dùng human-approved CanaryGrant + durable canary ledger. M11 dùng finite human-reviewed ProductionLease + trusted health/cost + sticky STOP; local sandbox/CI chỉ chứng minh Capability, không tự tạo E6.
