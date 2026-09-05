# BR-03c.7 — M10: kiểm từng artifact canary

Cập nhật: #34 đã merge `1d6594c`. Xem [BR-03c.7b audit chain](M10-CHAIN-AUDIT.md) cho CLI chín file mới; giới hạn bên dưới mô tả riêng `m10-check` của #34.

Trạng thái IN_REVIEW: [PR #34](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/34), commit triển khai `95979c1`. [CI theo head PR](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/34/checks); chưa merge.

Evidence local: tests/vet ba module, 8 Python validators, 10 Python regressions PASS; CLI fixture trả ba flag false. #33 đã review/merge `6fce165` sau CI 4/4 trên head `1333585`. Kết quả local không thay CI theo head PR M10.

Phạm vi: schema-first và exact typed decode cho `grant`, `approval`, `cost`, `ledger`, `gate`, `authorization`, `execution`. Dùng bảy canonical schema tương ứng trong `contracts/`. Authorization/execution chỉ nhận profile canary; không nhận lẫn field M09/M11. Đây là **kiểm từng file**, không phải audit chain M10 đầy đủ hoặc đường thực thi.

## Bài offline

Từ repo root:

```text
cd lab/mission-runtime
go run ./cmd/demo m10-check approval testdata/m10-approval.json
go test ./cmd/demo -run 'TestM10(Raw|LedgerWire|Audit)' -count=1 -v
```

CLI nhận `m10-check KIND FILE.json`; KIND là một trong bảy tên ở trên. Exit 0 xuất `ARTIFACT_VALID_UNVERIFIED`, `artifact_type` và ba flag false: `provenance_authenticated`, `chain_validated`, `execution_permitted`. Exit 1 khi sai args/read/schema/hash/semantic hoặc writer error; lỗi input không xuất summary. Không echo artifact quyền, không gọi gate/authorize/executor, không normalize/persist ledger, không gọi network. Chạy lại file chỉ là kiểm lại, không chứng minh chống replay.

Fixture approval là synthetic với hash placeholder đúng định dạng; cố ý không có grant tương ứng và không phải approval tin cậy. Không đưa fixture vào executor. `approved_by:human`, `APPROVE`, hoặc `execution_authorized:true` trong file chỉ là khai báo. Không có chương trình/kênh affiliate thật và không có E5/live proof từ bài này.

## Những gì được kiểm

- Raw required/null/type/enum/const/unknown/duplicate/trailing JSON trước typed decode. Integer vào int/int64 phải biểu diễn chính xác; không qua float64 hoặc tự làm tròn.
- Grant/cost: tính lại hash bằng thuật toán runtime, không seal hoặc sửa input; kiểm cửa sổ thời gian dương. Grant phải được approved không muộn hơn valid_from. So sánh instant, hỗ trợ timezone offset.
- Authorization: window dương và GOVERNED_CANARY. Gate giữ execution_authorized=false và điều kiện per-action approval của schema.
- Ledger: đếm pending khớp danh sách, không vượt executions total, số lần trong window/success keys không vượt total; outcome/execution links không trùng hoặc đồng thời pending; timestamp không sau updated_at. Đây không phải chứng minh ledger đầy đủ hoặc ngân sách còn đủ.
- Output ledger runtime rỗng dùng array `[]` thay vì `null` qua MarshalJSON không mutate state. Raw `null` vẫn bị chặn. Không thay thuật toán gate hoặc logic persistence; reader persistence cũ chưa được nối decoder này.

Tests serialize grant/cost/ledger từ base runtime và gate/authorization từ AuthorizeCanary; execution là CANCELLED/NOT_PERFORMED từ helper runtime, không gọi executor trong test mới. Test mutations gồm required/null/type/stage/authority, hash tamper, integer lớn hơn 2^53/overflow, time/offset, ledger, CLI read-only/replay/error. Regression M10 hiện hữu vẫn kiểm executor/restart/budget/reconciliation trong sandbox local.

## Phần còn mở

Chưa resolve liên kết giữa bảy file và M08 intent/policy, chưa kiểm grant scope/budget với một execution, chưa xác thực approval_ref/source_ref/trusted ledger/revocation/clock hiện tại. File khác nhau có thể đều PASS riêng dù chain sai. Không dùng summary như kết quả EvaluateCanaryGate hoặc quyền AuthorizeCanary. Cần tích hợp boundary vào trusted runtime/persistence và audit chain riêng trước khi tuyên bố M10 conformance đầy đủ. BR-03 tổng thể, M11 và live evidence vẫn mở; không cập nhật PROGRESS của học viên.
