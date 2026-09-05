# BR-03c.1 — Boundary JSON M03

Phạm vi: `ActionRecord`/`OutcomeRecord` qua lệnh chỉ đọc `m03-check` trong mission-runtime. Đây là phần đầu BR-03c, không đóng cả BR-03 và không thay tích hợp learner Bot BR-08/BR-10.

Trạng thái IN_REVIEW: [PR #28](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/28), code/tests commit `a7ebfed`; người thực hiện Codex, reviewer chủ repo (chưa review). PR độc lập từ main, không chứa diff BR-06a #27 và chưa merge.

## Luồng kiểm

JSON gốc → canonical schema nhúng → exact typed decode → kiểm action → kiểm effect link/cửa sổ đo → kiểm schema của artifact serialize → xuất envelope.

- Dùng package `contracts` đã có từ BR-03b, pin jsonschema/v6 v6.0.2; mission-runtime dùng local replace `../../contracts` và go.sum. Clone cả repo rồi chạy `go mod download` trong module để tải dependency lần đầu. Schema không tải từ mạng khi lệnh chạy.
- Required field thiếu, null sai kiểu, enum sai, date-time sai, số âm, key trùng ở mọi cấp, field lạ, case alias và trailing JSON đều bị từ chối trước semantic checks. `metrics:{}` hợp lệ; `metrics:null`, metric null hoặc thiếu metrics không hợp lệ. Không tự thêm số 0.
- `DecodeM03Action`/`DecodeM03Outcome` chỉ trả VALID về cấu trúc. `CheckM03Pair` mới kiểm semantics của cặp: compliance=false → HUMAN_REVIEW; effect sai → BROKEN_LINK; trước action → OUTCOME_BEFORE_ACTION; zero trước end → MEASUREMENT_WINDOW_OPEN. Whitespace ID và window đảo vẫn bị semantic validator từ chối.
- Action `performed_by:agent` bị schema const chặn bằng INVALID_SCHEMA. Test typed cũ vẫn mong REJECT_MACHINE_EXECUTION; eval E02 ghi riêng `raw_expected`, không xóa hoặc nới kiểm quyền cũ.
- `OutcomeRecord.ActionID` là alias nội bộ của M11, không nhận từ JSON M03 và không serialize. Không refactor M11 hoặc đổi global typed validators trong phần này.

## Chạy từ repo root

```text
cd lab/mission-runtime
go mod download
go run ./cmd/demo m03-check testdata/m03-action.json testdata/m03-outcome.json
go test ./cmd/demo -run 'TestM03' -count=1 -v
```

Kỳ vọng envelope chứa `action`, `outcome`, `result:"VALID"`; valid_orders giữ số 0, EffectRef giữ `syn-a`. Schema chỉ áp dụng từng artifact, không gán schema ActionRecord/OutcomeRecord cho cả envelope.

Hai fixture trên hoàn toàn synthetic. `compliance_reviewed:true` chỉ mô phỏng nhánh đã review, không chứng minh người đã review thật. Không dùng boolean đó để cấp quyền, không sửa fixture BR-06 đang để false. BR-06 và cặp kiểm M03 này là hai bài kiểm khác nhau, không giả là một pipeline đã nối.

Để thực hành, copy hai file vào thư mục bài tập của bạn. Đổi observed_at trong outcome về `2026-09-03T11:00:00+07:00`: stderr phải có MEASUREMENT_WINDOW_OPEN và exit 1, stdout trống. Đổi status sang PENDING: hợp lệ vì đang quan sát tạm. Đổi metrics thành null: INVALID_SCHEMA trước kiểm thời gian. Chỉ sửa bản copy; không ghi đè dữ liệu học viên.

Mọi lỗi đọc file/schema/semantic đều exit 1, không xuất success envelope; lỗi ghi stdout được truyền về caller. Lệnh không persist, sửa file, fetch hoặc execute. `VALID` chỉ nói cặp file vượt qua các kiểm hiện có; không chứng minh decision_id tồn tại trong store, tính đầy đủ của báo cáo, attribution, source truth hoặc Reality/Operated PASS. ID linkage tới DecisionPacket/store thuộc BR-08/BR-10.

## Evidence và giới hạn

- Red/green: wrapper ban đầu dùng json.Unmarshal đơn thuần làm 15/16 raw regression cases FAIL; sau nối schema + strict decode cùng tests PASS. Đây là probe decoder, không phải tuyên bố mọi case đều đã qua semantic validators cũ.
- Test mọi required field bị bỏ/null, nested duplicate/unknown, missing/zero, schema trước compliance, offset tương đương, window đảo, orphan/wrong-kind, read-only files, không có envelope khi reject và writer failure.
- Eval M03 được chạy cả đường typed cũ và JSON gốc; expected semantics vẫn độc lập, chỉ E02 có expected riêng cho schema-first.
- Output thực của handler được decode/kiểm canonical schema trong test; CI thêm smoke lệnh thật với fixture đã commit.
- M05–M11 vẫn cần audit/boundary riêng. Demo hard-code M03 cũ vẫn chỉ là conformance minh họa, không được gọi là raw JSON boundary. File input chưa có giới hạn kích thước bổ sung; lệnh local chỉ dành cho file người vận hành chọn, chưa phải endpoint nhận input từ mạng.

Kiểm local đợt này: Go test/vet cả ba module PASS, 8 Python validators và 10 Python regressions PASS, git diff --check PASS. Lệnh m03-check chạy trên cặp fixture đã commit xuất VALID, giữ valid_orders=0 và effect_id=syn-a. GitHub CI cần kiểm theo head PR, không suy từ kết quả local.
