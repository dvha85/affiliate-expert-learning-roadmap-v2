# Learner path — Mission-based curriculum

Đây là đường học hiện hành. Không dùng lesson ID legacy từ repo cũ để quyết định bài tiếp theo.

## Bắt đầu

1. `BOOT.1` nếu chưa từng chạy/sửa/test Bot.
2. `M00.1 → M00.2 → M00.3`.
3. Hoàn thành evidence/artifact của M00 trước khi sang M01.
4. Khi M00 PASS: `M01.1 → M01.2 → M01.3 → M01.4`.

M01 đã **authoring ready**, nhưng điều này không bỏ qua M00 learner gate.

M02 đang được author theo path dự kiến `M02.1 → M02.2 → M02.3 → M02.4`, nhưng vẫn là **planned** cho tới khi lesson + starter + executable eval/runtime + CI gate cùng PASS. Khi M02 chuyển `ready`, learner chỉ được vào M02 sau M01 PASS.

## Learner loop

```text
TRY
→ OBSERVE GAP
→ PULL 1–3 SMALL KNOWLEDGE CARDS
→ BUILD / APPLY
→ TEST FAILURE CASE
→ SAVE EVIDENCE
→ EXPLAIN LIMITS
→ NEXT MEASUREMENT
```

Mỗi lesson/card phải phục vụ một câu hỏi cụ thể của Mission hiện tại. Không đọc theory chỉ để hoàn thành checklist.

## Mission sequence

```text
O00 → M00 → M01 → M02 → M03 → M04 → M05 → M06 → M07 → M08 → M09 → M10 → M11
```

Authority tăng dần. Capability của Mission sau không được dùng để vượt evidence/safety gate của Mission trước.
