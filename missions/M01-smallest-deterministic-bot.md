---
mission_id: M01
title: Smallest Deterministic Bot v0.1
status: ready
minimum_evidence: E0 + E1 support
authority: A0 deterministic
external_side_effects: false
starter: starter-kits/M01-deterministic-bot/
eval_pack: evals/M01-deterministic-bot/
runtime: lab/affiliate-bot/
---

# Mission M01 — Smallest Deterministic Bot v0.1 (Bot tất định nhỏ nhất)

## Mục tiêu bàn giao

```text
canonical observations/context (quan sát/ngữ cảnh chuẩn hóa) + formula_version
→ deterministic formula + stable tie-break
  (công thức tất định + quy tắc phá hòa ổn định)
→ HUMAN_REVIEW > GET_MORE_DATA > RANK_SCENARIO
→ reason + missing/invalid evidence
  (lý do + bằng chứng thiếu/không hợp lệ)
→ KHÔNG có hành động bên ngoài
```

Go là reference implementation (triển khai tham chiếu) hiện hành cho M01, không phải nguồn có thẩm quyền của curriculum.

## Hành vi bắt buộc

- cùng input (đầu vào) + cùng version (phiên bản) → cùng result (kết quả);
- `0 != missing` (0 khác dữ liệu thiếu);
- field (trường) không hợp lệ không được silently default (âm thầm gán mặc định);
- conflict (xung đột) về identity/currency/mixed-origin (định danh/tiền tệ/nguồn gốc trộn) → `HUMAN_REVIEW`;
- missing/invalid evidence (bằng chứng thiếu/không hợp lệ) → `GET_MORE_DATA` nếu không có conflict mạnh hơn;
- real evidence (bằng chứng thật) không tự nâng weak formula (công thức yếu) thành `RECOMMEND`;
- output (đầu ra) luôn giải thích limitation (giới hạn) của formula.

## Các thẻ bài học

`M01.1 → M01.2 → M01.3 → M01.4` trong `curriculum/M01/`.

## PASS

### Capability (năng lực)
Runtime, `go vet`, regression tests (kiểm thử hồi quy) và executable eval pack (bộ ca đánh giá có thể chạy) đạt behavior contract (hợp đồng hành vi).

### Reality (thực tế)
Đối chiếu input boundary (ranh giới đầu vào) với evidence/context thật của M00; fixture (dữ liệu kiểm thử mẫu) chỉ là E0 và không được dùng để claim market reality (khẳng định thực tế thị trường).

### Operated (đã tự vận hành chứng minh)
Người học tự dự đoán, chạy runtime/evals, quan sát failure case (ca lỗi) và explain-back (tự giải thích lại) được limitation/authority ceiling (giới hạn/trần quyền hạn).

```text
M01 PASS = Capability (năng lực) + Reality (thực tế) + Operated (đã vận hành chứng minh)
```

## Kết quả

Bot v0.1 tất định, không có quyền hành động. M02 mới thêm trustworthy history + replay (lịch sử đáng tin + phát lại).
