# BR-13b — Parser fixture và history handoff

Baseline `0e47b30` sau #74. Chỉ profile `br13-offer-fixture/v1`, GET và URL chính xác `https://example.com/br13/offer`. URL là nhãn fixture, không được truy cập. Không nhận allowlist/endpoint/network permission từ input. Core M06 normalize/hash; parser chuyển qua m00.Convert để giữ provenance theo field và kiểm numeric precision, rồi NewHistoryRecord/AppendHistory ghi canonical history của bot.

## Chạy thủ công và đọc lại

Từ `lab/affiliate-bot`, chọn một thư mục lab riêng đã tồn tại; thay `/path/to/lab/history.jsonl` bằng đường dẫn thực, không dùng ledger DeepSeek:

```sh
go run ./cmd/bot watcher fixture-import /path/to/lab/history.jsonl ../../examples/watcher/offer-valid.json
go run ./cmd/bot watcher fixture-import /path/to/lab/history.jsonl ../../examples/watcher/offer-valid.json
go run ./cmd/bot watcher fixture-import /path/to/lab/history.jsonl ../../examples/watcher/offer-missing.json
go run ./cmd/bot history list /path/to/lab/history.jsonl
go run ./cmd/bot history replay /path/to/lab/history.jsonl
```

Expected lần đầu APPENDED/RANK_SCENARIO; retry EXACT_DUPLICATE cùng record_id; fixture thiếu commission APPENDED/GET_MORE_DATA; hai record replay MATCH sau process mới. `offer-error.json` trả FIXTURE_ERROR, persisted=false, không artifact. Input tối đa 64 KiB, regular file không symlink cuối, JSON strict từ chối duplicate/unknown fields. Body là chuỗi JSON một offer, product_id/product_name và currency USD bắt buộc; price/commission thiếu hoặc null được map missing/unknown, không default zero. Giá trị invalid/không biểu diễn chính xác bị từ chối qua M00.

Mọi dữ liệu có evidence_kind=synthetic/use_context=test. Field có giá trị được gắn assumption, không nâng seller claim thành fact. Source URL, observed_at, body hash và correlation_id nằm trong provenance field; history schema được giữ nguyên. Tên sản phẩm không được diễn giải thành instruction.

## Identity, retry và ACK

- Record ID = `watch-` + SHA256(JSON [profile, fixed URL, correlation_id]). Correlation ID là một sự kiện quan sát, không dùng lại cho sự kiện khác hoặc sản phẩm khác.
- Cùng sự kiện + nguyên bytes body + thời điểm tương đương → cùng record; observed_at được chuẩn UTC trước normalize. Ingested_at/as_of trong chế độ fixture bằng thời gian kịch bản, không giả làm wall-clock fetch time.
- Đổi body (kể cả whitespace), thời gian hoặc metadata dưới cùng correlation_id → conflict trong history. Thay đổi hợp lệ mới dùng correlation_id mới; không tự đổi ID để né conflict.
- Observation ID do core M06 tạo từ subject/URL/time/method/correlation/body hash. Không có watcher cache/previous_hash để quyết định bỏ qua persistence. History là nguồn đối chiếu retry.
- `persisted=true` chỉ trả sau AppendHistory thành công hoặc EXACT_DUPLICATE đã xác nhận. Artifact chứa record_id/decision_id/state/observation_ids; normalize thành công chưa là đã lưu. Mọi lỗi trả nonzero/persisted=false/không artifact.
- HANDOFF_ERROR có thể là conflict, store hỏng hoặc lỗi ghi; không tự retry/reset. Nếu output bị mất hoặc ghi thất bại một phần, đọc history/replay và kiểm file trước. Không cam kết transaction/rollback khi disk lỗi; persisted là ACK mức adapter, không bảo đảm fsync chịu mất điện.

## Kiểm thử và giới hạn

Từ repo root: `python3 scripts/smoke_br13b.py`. Smoke build binary trong thư mục tạm, kiểm fixture hợp lệ/thiếu field/lỗi, retry không đổi bytes, failed sink không ACK, list/replay hai record từ process mới; CI chạy cùng script. Unit tests thêm timezone-equivalent retry, event conflict/new event, source/method/status, malformed/unknown/precision/negative input, alias và corrupt sink.

Một writer, thư mục tin cậy, kế thừa BR-09 JSONL adapter; không khóa liên tiến trình/transaction, không chống local rollback/hostile path replacement. Final history không symlink/nonregular; alias history/input bị chặn. Không thay schema/store cũ hoặc sửa dữ liệu hỏng tự động.

BR-13b chỉ hoàn thành handoff fixture. BR-13c vẫn phải chọn nguồn public/được phép và parser thực, quyền truy cập cùng giới hạn HTTP/SSRF/redirect/retry, timestamp thực và bằng chứng fetch. Example.com không phải nguồn affiliate đã nghiệm thu. BR-13 cha IN_PROGRESS; BR-14 n8n chưa được đóng bởi smoke CLI này.
