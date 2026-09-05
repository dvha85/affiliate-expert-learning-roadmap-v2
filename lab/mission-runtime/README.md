# Mission Runtime — conformance harness cho O00 và M03–M11

Runtime offline để chứng minh boundary/semantics của các Mission mà không cần API key, n8n instance hay dịch vụ bên ngoài.

```bash
go test ./...
go vet ./...
go run ./cmd/demo O00
go run ./cmd/demo M03
go run ./cmd/demo M04
go run ./cmd/demo M05
go run ./cmd/demo M06
go run ./cmd/demo M07
go run ./cmd/demo M08
go run ./cmd/demo M09
go run ./cmd/demo M10
go run ./cmd/demo M11
```

Runtime này là **conformance oracle/harness (bộ đối chiếu chuẩn)**, không phải Affiliate Bot thứ hai. Từ M03 trở đi learner dùng behavior/failure cases ở đây để kiểm implementation được gắn vào cùng learner Bot/workspace đang tiến hóa từ `lab/affiliate-bot/`.

## Kiểm ActionRecord/OutcomeRecord từ file (BR-03c.1)

M03 có lệnh riêng `go run ./cmd/demo m03-check testdata/m03-action.json testdata/m03-outcome.json` từ module này. Lệnh chỉ đọc, schema-first rồi kiểm liên kết/cửa sổ đo; lỗi schema hoặc semantic đều exit 1 và không xuất envelope. Xem [hướng dẫn và giới hạn BR-03c.1](../../docs/architecture/M03-JSON-BOUNDARY.md). Cặp fixture synthetic mô phỏng compliance=true, không phải review/action thật; không có store hoặc execution.

## Kiểm Evaluation/Proposal/Review từ file (BR-03c.2)

M05 có lệnh chỉ đọc kiểm 5 file: [m05-check và bài tập](../../docs/architecture/M05-JSON-BOUNDARY.md). Schema-first, giữ liên kết/timeline, không apply proposal; result VALID không có nghĩa review APPROVE hoặc execution authorized.

## Kiểm Observation offline (BR-03c.3)

M06 có [m06-check](../../docs/architecture/M06-JSON-BOUNDARY.md) cho response fixture offline. Normalizer xuất synthetic/test và kiểm Observation schema; không fetch hoặc persist. Không dùng nhãn real của output harness cũ làm bằng chứng nguồn thật.

## Kiểm M07 từ file (BR-03c.4)

M07: [m07-check offline](../../docs/architecture/M07-JSON-BOUNDARY.md) kiểm proposal/registry/context IDs từ ba file, không fetch/model và không cấp execution. ToolRegistry kiểm cả mảng canonical; AgentProposal chỉ có shape local.

## Kiểm M08 từ file (BR-03c.5)

M08: [m08-check](../../docs/architecture/M08-JSON-BOUNDARY.md), raw schema trước policy, không seal lại hash/quyền. Exit 0 có thể chứa DENY/WAIT; không execute.

## Kiểm M09 từ file (BR-03c.6)

M09 có [m09-check](../../docs/architecture/M09-JSON-BOUNDARY.md) audit năm file lịch sử, chỉ xuất CONSISTENT_UNVERIFIED với execution_permitted=false. Không gọi AuthorizeM09/executor hoặc xác thực approver từ một chuỗi JSON.

## Kiểm AdvisorOutput từ file (BR-03a)

Từ thư mục `lab/mission-runtime`, chạy fixture không cần API key:

```bash
go run ./cmd/demo advisor-check testdata/m04-advisor-output.json testdata/m04-evidence.json 2026-09-03T01:00:00Z 24
```

Output gồm `advisor_output` và `result=SUPPORTED`. Đây là fixture tổng hợp, không phải evidence thật. Thay hai path bằng bản output/context của bạn để kiểm JSON gốc; không tự bỏ field lạ hay thêm giá trị thiếu trước khi kiểm. Schema sai trả lỗi `INVALID_SCHEMA` trên stderr và exit 1; yêu cầu write trả `REJECT_WRITE_REQUEST`. JSON hợp lệ mới đi qua kiểm liên kết/freshness và xuất envelope; `result` có thể là `REJECT_UNGROUNDED`, `ABSTAIN_STALE`, `ABSTAIN_FUTURE` hoặc `INVALID`, dù exit 0 vì lệnh đã hoàn tất phép kiểm. Luôn đọc result, không lấy exit 0 làm phê duyệt.

`MAX_AGE_HOURS` là số nguyên không âm. Evidence file là mảng các object `evidence_id`, `observed_at`, `source_ref`; đây là context rút gọn, không phải canonical Observation. AdvisorOutput dùng schema riêng; không gán schema đó cho cả envelope. Lệnh không gọi model, fetch nguồn hoặc thực hiện action; tích hợp learner Bot thuộc BR-11.

```text
run mission-runtime demo/test
!= learner integration
!= Reality PASS
!= Operated PASS
```

M03 cần action thật do người thực hiện; M04/M05 cần evidence/evaluation thật; M06/M07 cần learner vận hành workflow/read-only Agent với nguồn được phép; M08–M11 cần đúng authority/evidence gate của Mission. Continuity gate: `docs/architecture/LEARNER-BOT-CONTINUITY.md`.
