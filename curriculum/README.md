# Đường học — chương trình theo Mission

Đây là đường học hiện hành. Không dùng lesson ID legacy (ID bài học cũ) từ repo cũ để quyết định bài tiếp theo.

## Bắt đầu

1. `BOOT.0` nếu máy chưa có environment/tooling tối thiểu.
2. `BOOT.1` nếu chưa từng chạy/sửa/test Bot.
3. Chạy `O00.1` để nhìn toàn hệ thống bằng dữ liệu mô phỏng; O00 không tạo PASS.
4. `M00.1 → M00.2 → M00.3`.
5. Khi M00 PASS: `M01.1 → M01.2 → M01.3 → M01.4`.
6. Khi M01 PASS: `M02.1 → M02.2 → M02.3 → M02.4`.
7. Khi M02 PASS: `M03.1 → M03.2 → M03.3`.
8. Khi M03 PASS: `M04.1 → M04.2 → M04.3`.
9. Khi M04 PASS: `M05.1 → M05.2 → M05.3`.
10. Khi M05 PASS: `M06.1 → M06.2 → M06.3`.
11. Khi M06 PASS: `M07.1 → M07.2 → M07.3`.
12. Khi M07 PASS: `M08.1 → M08.2 → M08.3`.
13. Khi M08 PASS: `M09.1 → M09.2 → M09.3`.
14. Khi M09 PASS: `M10.1 → M10.2 → M10.3`.
15. Khi M10 PASS: `M11.1 → M11.2 → M11.3`.

M01–M11 hiện **authoring ready và learner-operable (sẵn sàng về nội dung và đường thực hành)**. M08 chỉ shadow; M09 machine execution cần approval từng lần; M10 mở governed canary; M11 mở finite production lease với trusted health/cost, DEGRADE read-only và sticky STOP. CI/sandbox không tự tạo Reality/Operated PASS.

## Vòng học của learner

```text
TRY → OBSERVE GAP → PULL SMALL KNOWLEDGE → BUILD/APPLY → TEST FAILURE CASE → SAVE EVIDENCE → EXPLAIN LIMITS → NEXT MEASUREMENT
```

## Continuity Gate — một Bot, không phải chuỗi demo

Từ M03 tới M11, checkpoint riêng của Mission **luôn đi kèm** `starter-kits/CONTINUITY-CHECKPOINT.md`.

```text
lab/affiliate-bot
= learner baseline/continuity anchor

lab/mission-runtime
= conformance oracle/harness
!= Bot thứ hai

lab/n8n
= adapter/orchestration reference
!= canonical state owner
```

Reality/Operated PASS của M03–M11 phải ghi Integration Evidence: learner Bot version/commit, previous artifact refs, capability entrypoint, canonical contracts, conformance cases và authority ceiling. Chạy demo ở `lab/mission-runtime` một mình chỉ tạo Capability proof.

Chi tiết: `docs/architecture/LEARNER-BOT-CONTINUITY.md`.

## Trục Mission

```text
O00 → M00 → M01 → M02 → M03 → M04 → M05 → M06 → M07 → M08 → M09 → M10 → M11
```

Authority tăng dần; capability Mission sau không được vượt evidence/safety gate Mission trước.

## Boundary quan trọng M03–M11

```text
M03: external action đầu tiên do human_only
M04: AI advisory; không write tool
M05: proposal + human review; không auto-apply
M06: automatic read-only watcher; watcher cache != canonical history
M07: read-only Agent; tool output untrusted
M08: shadow ActionIntent + policy; ALLOW != execution permission
M09: mỗi machine execution cần human ApprovalRecord
M10: CanaryGrant cho bounded RISK0/RISK1; RISK2 không auto
M11: finite ProductionLease + trusted health/cost + sticky STOP + reviewed improvement; không self-widen authority
```
