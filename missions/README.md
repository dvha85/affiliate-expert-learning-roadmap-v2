# Chỉ mục Mission — trục thực thi chuẩn

Thứ tự học được quyết định bởi `CURRICULUM.md`.

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
| M08 | Shadow ActionIntent + Policy (ActionIntent chạy bóng + chính sách) | A3-shadow | planned |
| M09 | Durable Approval + Controlled Executor (phê duyệt bền vững + bộ thực thi có kiểm soát) | qua cổng phê duyệt | planned |
| M10 | Governed Canary (canary có quản trị) | tự động RISK0/RISK1 trong giới hạn | planned |
| M11 | Production Closed Loop (vòng kín production có quản trị) | production có quản trị | planned |

```text
O00 → M00 → M01 → M02 → M03 → M04 → M05 → M06 → M07 → M08 → M09 → M10 → M11
```

`ready` nghĩa là lesson + starter + executable eval/runtime cần thiết đã được soạn và CI bảo vệ. Nó **không** có nghĩa người học đã PASS. Learner PASS phụ thuộc evidence thật và operated result theo Mission contract.

M06/M07 có n8n blueprint để học orchestration/Agent nhưng CI kiểm authority boundary bằng `lab/mission-runtime` offline, không cần API key hay dịch vụ ngoài.
