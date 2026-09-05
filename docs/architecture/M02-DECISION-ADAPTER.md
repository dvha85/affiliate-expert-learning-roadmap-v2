# BR-03b — Xuất DecisionPacket từ history M02

`HistoryRecord.recorded_result` tiếp tục là projection deterministic ranking. Adapter tạo artifact mới, không sửa projection, lịch sử hoặc formula version. Context do người cung cấp; bot không tự bịa câu hỏi/fact/assumption để điền required fields.

## Lệnh chạy bằng fixture

Từ `lab/affiliate-bot`, dùng đường dẫn history mới riêng cho bài tập:

```bash
go run ./cmd/bot history capture /tmp/br03b-demo.jsonl data/m02-sample-observations.json br03b-demo 2026-09-01T01:00:00Z 2026-09-03T10:00:00Z
go run ./cmd/bot history replay /tmp/br03b-demo.jsonl
go run ./cmd/bot history decision /tmp/br03b-demo.jsonl br03b-demo data/m02-decision-context.json
```

Trên Windows thay `/tmp/br03b-demo.jsonl` bằng một path ghi được như `br03b-demo.jsonl`; hướng dẫn này không claim đã smoke Windows. Không dùng đường dẫn trùng dữ liệu thật. Lệnh cuối in JSON packet ra stdout, lỗi trả exit 1 trên stderr và không in packet. Có thể lưu stdout ra **file mới khác history** sau khi kiểm kết quả; không redirect vào file đầu vào (shell có thể xóa nội dung trước khi bot đọc).

Fixture là synthetic, không phải E1/Reality/Operated proof. Packet hợp lệ không chứng minh fact đúng hoặc cấp quyền publish.

## Mapping được dùng

| Packet field | Nguồn / quy tắc |
|---|---|
| decision_id, evidence_ids | Giữ exact từ recorded_result; linkage tới observations đã được kiểm |
| state | Giữ RANK_SCENARIO/GET_MORE_DATA/HUMAN_REVIEW, không nâng thành RECOMMEND |
| reason | Nối các reasons hiện có bằng newline; không tự thêm lý do kinh doanh |
| missing_evidence | Copy từ projection; nil chuyển array rỗng trong packet mới |
| question, supported_facts, assumptions, unknowns, next_measurement | Context JSON do người cung cấp; required fields phải có, arrays phải tường minh, không null |
| action | Luôn null; không nhận override từ context |
| ranked, formula_version, evidence_mode | Ở lại history, không lọt sang canonical packet |

Hàm `HistoryDecisionPacket` chỉ xuất khi replay=MATCH; DRIFT/UNREPLAYABLE/INTEGRITY_ERROR phải được xử lý trước, không xuất packet nhìn như đã xác minh. Context không được thêm field lạ/duplicate hoặc override ID/state/action. Schema và adapter không tự chứng minh một supported_fact có nội dung được evidence hỗ trợ; người vẫn phải rà từng phát biểu và nguồn.

## Input/output và bảo toàn

Capture kiểm JSON gốc bằng observation subschema của history (tái sử dụng `$ref`/`allOf`), rồi strict decode trước khi hash/lưu. Load kiểm history raw schema, strict fields, rồi ID/hash/time semantics. Append kiểm record serialize trước khi ghi. Packet output được kiểm theo `decision-packet.schema.json` trước khi in.

Các field extension mà schema cho phép nhưng Go projection chưa hỗ trợ bị từ chối, không âm thầm bỏ. Không tự sửa file nhập. Các file history hỏng hoặc không canonical trước đây từng lọt qua cần rà/chuyển đổi riêng; chỉ ngoại lệ `ranked:null` lịch sử được hỗ trợ có chủ đích trong [schema runtime](../../contracts/README.md).

Tests bao gồm thiếu/null/sai kiểu/enum/ID trùng, timestamp sai, trường lạ/case alias, context xấu, legacy ranked null, packet serialization, read-only export, replay drift và hash tamper. M01 evaluator giữ nguyên behavior; boundary M02 chặt hơn vì artifact được lưu lại.
