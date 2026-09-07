# BR-13c — Fetch HTTPS fixture đã ghim

Baseline `e0896de` sau BR-13b #75. `watcher fetch-fixture HISTORY` là một canh cắt network thật nhưng dữ liệu giả lập: URL, ref commit, host, method GET và SHA-256 đều cố định trong binary. Không nhận URL, host, proxy, redirect, header, key hoặc input body từ CLI.

URL hiện tại trỏ tới file fixture trong repo tại ref `9732ea1...`, hash `f70e18b8...7817b`. HTTP client không proxy, không redirect, timeout tổng 20s, header/response 64KiB, mã hóa content bị từ chối; DNS chỉ chấp nhận địa chỉ global-unicast của `raw.githubusercontent.com` và dial đúng IP đã kiểm tra. Status phải 200. Đây là giảm rủi ro SSRF cho profile cố định, không phải capability fetch arbitrary hay bảo đảm chống mọi DNS/route tấn công.

Sau khi fetch, body được parser qua M06/M00 và ghi history canonical bằng `AppendHistory`. Artifact chỉ trả sau append/exact duplicate; output luôn `network_fetch_attempted`, `persisted`, `execution_permitted=false`. Retry cùng correlation/event trả EXACT_DUPLICATE; history replay sau process mới phải MATCH. Digest mismatch, redirect, status/encoding, timeout, parser hoặc sink lỗi đều trả nonzero và không artifact; không retry tự động.

Chạy trong môi trường có network:

```sh
cd lab/affiliate-bot
go run ./cmd/bot watcher fetch-fixture /tmp/br13c-history.jsonl
go run ./cmd/bot history list /tmp/br13c-history.jsonl
go run ./cmd/bot history replay /tmp/br13c-history.jsonl
```

Smoke CI `python3 scripts/smoke_br13c.py` build binary trong thư mục tạm, gọi URL ghim một lần rồi kiểm retry/replay, source URL và hash; không gọi chương trình affiliate, không dùng key, không ghi repo. Nếu CI/network không truy cập được GitHub, phải ghi FETCH_ERROR; không đổi URL/hash để làm xanh. Fixture có `evidence_kind=synthetic`, `use_context=test`; thời gian `scenario_observed_at` khác `fetch_started/completed_at`, không giả là thời gian thị trường.

Giới hạn: upstream Git/ref có thể bị xóa hoặc thay đổi quyền; hash bảo vệ nội dung tại thời điểm chạy nhưng không chứng minh ý nghĩa kinh doanh. History vẫn một writer, append chưa transaction/fsync chịu mất điện; không watcher scheduler/cache, không affiliate attribution, không seller authority. Nguồn public thật và parser format affiliate vẫn OPEN; BR-13c chỉ chứng minh fetch pipeline trên synthetic fixture, không chốt BR-13 cha.
