# BR-09 — Từ packet M00 sang cùng Bot M01/M02

Hai packet t1/t2 là **synthetic**, không phải E1, báo cáo affiliate hay bài thực hành thật của học viên. Chưa có chương trình/tài khoản/kênh được chọn. Không sửa `evidence/` hoặc PROGRESS để dùng ví dụ này.

## Phạm vi và bảng map

`bot evidence import PACKET.json` đọc bản chép JSON có version `m00-input/v1` từ nội dung M00, không parse Markdown tự do, không fetch nguồn và không ghi store. Stdout là JSON envelope: command, status, execution_permitted=false, artifact là array input M01/M02 nếu VALID; lỗi không có artifact. Exit 0/1/2 tương ứng thành công/lỗi input hoặc I/O/usage. stderr chứa chẩn đoán. Không redirect vào chính file đầu vào hoặc history.

| M00 / JSON profile | M01/M02 | Quy tắc |
|---|---|---|
| question | JSON trong transformation_or_method | Câu hỏi nguyên gốc; không tự tạo supported_facts |
| product.subject_id | subject_id = product_id | Một subject ổn định mỗi sản phẩm, mỗi packet chỉ một snapshot/subject |
| product.observation_id | observation_id aggregate | ID mới của phép nhóm; không tái dùng ID field hoặc product_id |
| product_name/currency | product_name/currency | Context do người nhập khai báo; không được converter coi là fact đã xác minh; rate là tỷ lệ 0–1, không phải phần trăm 0–100 |
| fields[].field_or_claim/value | price/commission_rate | V1 chỉ nhận hai field này; mỗi field đúng một lần, không tự chọn giữa hai nguồn mâu thuẫn |
| state/claim_kind | numeric hoặc null | Chỉ observed và không unknown mới dùng numeric, kể cả 0; missing/pending/not_yet_observable/inconclusive/unknown → null trong projection |
| observation_id/source/time/role/method/state/value từng field | transformation_or_method | JSON có version/question/product chứa toàn bộ field đã kiểm, sắp theo tên field; trạng thái và value gốc giữ nguyên, không biến pending thành missing trong provenance |
| observed_at | observed_at aggregate | Thời điểm mới nhất trong nhóm; không làm mới field cũ. as_of phải không sớm hơn nó |
| evidence_kind | evidence_kind | Giữ real/synthetic; từ chối trộn origin trong một sản phẩm. real yêu cầu source_url nhưng không chứng minh nguồn đáng tin |
| aggregate claim/state | assumption; observed hoặc inconclusive | Đây là phép nhóm phục vụ scenario, không nâng seller claim thành fact |

V1 yêu cầu ghi rõ hai field, kể cả khi thiếu: dùng state=missing, claim_kind=unknown, value=null và provenance giải thích thiếu ở đâu. Metadata, tên field không hỗ trợ, duplicate keys/IDs, số ngoài range và subject không khớp bị từ chối. Thêm currency hoặc field khác cần profile mới/review, không nhét nó vào giá trị khác. Một packet có thể chứa nhiều sản phẩm; không có merge nhiều nguồn tự động.

Numeric literal tối đa 128 ký tự, exponent trong [-400, 400]; số không giữ được nghĩa thập phân qua float64 M02 bị từ chối, kể cả underflow. Đây là giới hạn profile, không làm tròn evidence để tiếp tục.

History dùng giới hạn chung 1 MiB (1.048.576 byte) cho mỗi JSON record, không tính LF/CRLF. Capture từ chối record lớn hơn trước khi ghi, không tạo file mới hay sửa history có sẵn; chia packet theo nhóm sản phẩm nhỏ hơn với record ID riêng khi cần. Import VALID chỉ xác nhận phép chuyển đổi, không hứa mọi artifact vừa giới hạn persistence. Reader hỗ trợ record đến đúng giới hạn, kể cả dòng cuối không newline. Không tự sửa/truncate file cũ vượt giới hạn.

Khi timestamp hai field biểu diễn cùng thời điểm, aggregate chọn chuỗi RFC3339 nhỏ nhất theo thứ tự từ điển để không phụ thuộc thứ tự field; timestamp từng nguồn vẫn giữ nguyên trong provenance. Quy tắc này không chuyển timezone nguồn hoặc thay freshness.

History giữ nguyên JSONL/hash/formula cũ; provenance nằm trong trường transformation đã có và được hash. Khi capture, projection import được tính lại để phát hiện lệch provenance. Reuse aggregate ID hoặc source field ID với nội dung khác bị CONFLICT; lần quan sát t2 phải có ID mới. DecisionPacket.evidence_ids trỏ tới aggregate tồn tại trong history; mở transformation JSON để resolve tiếp exact IDs từng field. Không giả vờ DecisionPacket trỏ trực tiếp raw field IDs.

## Thực hành t1/t2, restart, quyết định

Chạy từ root repo, cần Go theo go.mod và Python 3. Tạo workspace tạm riêng; không sử dụng ledger cá nhân. Ví dụ shell macOS/Linux:

```sh
work=$(mktemp -d)
(cd lab/affiliate-bot && GOWORK=off go build -o "$work/bot" ./cmd/bot)
for n in 1 2; do
  "$work/bot" evidence import "examples/m00-import/packet-t$n.json" > "$work/envelope-t$n.json"
  python3 -c 'import json,sys; p=json.load(open(sys.argv[1])); assert p["status"]=="VALID"; print(json.dumps(p["artifact"]))' "$work/envelope-t$n.json" > "$work/input-t$n.json"
  "$work/bot" "$work/input-t$n.json"
  "$work/bot" history capture "$work/history.jsonl" "$work/input-t$n.json" "br09-t$n" "2026-09-0${n}T01:00:00Z" "2026-09-0${n}T02:00:00Z"
done
"$work/bot" history list "$work/history.jsonl"
"$work/bot" history replay "$work/history.jsonl"
"$work/bot" history decision "$work/history.jsonl" br09-t1 lab/affiliate-bot/data/m02-decision-context.json
"$work/bot" history decision "$work/history.jsonl" br09-t2 lab/affiliate-bot/data/m02-decision-context.json
```

Mỗi invocation là process mới, dùng lại history trên đĩa. Context quyết định trên là synthetic với supported_facts rỗng; khi nhập packet riêng phải tự review context phù hợp, không dùng nguyên context này cho dữ liệu thật. Lệnh decision dùng adapter BR-03 hiện có, không sinh approval/action.

Để kiểm đầy đủ và xem output tóm tắt có assertion độc lập:

```sh
python3 scripts/smoke_br09.py
```

Output đầy đủ của smoke:

```text
t1: import VALID; M01/M02 GET_MORE_DATA; duplicate EXACT_DUPLICATE; decision IDs resolved
t2: import VALID; M01/M02 RANK_SCENARIO; duplicate EXACT_DUPLICATE; decision IDs resolved
restart/list/replay: 2 MATCH; source-ID conflict and projection tamper rejected; history unchanged
BR-09 SMOKE PASS (synthetic; no live proof)
```

`recorded_result.decision_id` phải bằng record_id; t1 commission_rate=null, nguồn gốc vẫn pending; t2 score=8 (100 × 0.08), không phải lợi nhuận đã kiếm được. DecisionPacket action=null. Output JSON đầy đủ nằm ở envelope/history trong workspace của bài tập; không commit các file cá nhân này. Smoke tự cleanup workspace riêng.

## Bài tập và giới hạn

Đổi t1 pending thành observed/assumption/value=0: phải phân biệt với null. Tạo hai field cùng tên hoặc đổi subject: importer phải lỗi. Giữ ID nguồn t2 nhưng sửa price rồi capture dưới record_id mới: phải CONFLICT. Tạo quan sát t3 với ID mới, sau đó list/replay và tự giải thích chain ID.

Không có credential, live evidence, trusted clock, multi-writer lock hay chứng minh freshness/business truth. Có thể sửa hoặc giả mạo cả input/provenance; kiểm projection/hash không là chữ ký hay xác thực nguồn. Giới hạn JSONL một writer và kích thước record hiện có vẫn áp dụng. Không thêm action/outcome store (BR-10). Không nâng M00 E1/BR-06b hoặc năng lực học viên thành DONE từ smoke.
