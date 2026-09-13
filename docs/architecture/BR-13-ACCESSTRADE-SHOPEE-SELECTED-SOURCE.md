# BR-13 — selected source: ACCESSTRADE Shopee Smartlink (read-only)

Trạng thái: **IMPLEMENTED_OFFLINE boundary / chưa có operated evidence**.

Selected source đầu tiên là campaign **Shopee Việt Nam Smartlink cho tất cả
thiết bị** trên ACCESSTRADE. Profile chỉ quan sát metadata campaign được
allowlist. Nó không tạo Smartlink, không publish, không gọi executor, không
đọc report publisher và không chứng minh attribution, commission hay doanh
thu.

## Điều được phép đưa vào capture

Một capture JSON, được tạo ngoài Bot sau khi người vận hành tự xem page đã
đăng nhập, chỉ có các trường sau:

```json
{
  "version": "accesstrade-shopee-campaign-capture/v1",
  "method": "GET",
  "observed_at": "<RFC3339 UTC timestamp>",
  "correlation_id": "<new stable capture id>",
  "status_code": 200,
  "redirected": false,
  "source_page_sha256": "<sha256 of the reviewed page capture>",
  "campaign_title": "Shopee Việt Nam Smartlink cho tất cả thiết bị",
  "merchant_label": "Shopee",
  "campaign_category": "Thương Mại Điện Tử",
  "campaign_status_label": "<label shown by the page>",
  "campaign_period_label": "<period label shown by the page>"
}
```

Không lưu HTML, cookie, credential, publisher ID, review/comment, raw report,
tracking/deep link, số lượng thành viên, EPC, CVR, giá/commission hay nội dung
khuyến mãi. `source_page_sha256` chỉ là dấu vết provenance, không phải HTML và
không thay evidence gốc do người vận hành giữ riêng.

Profile cố định URL campaign trong code; capture không có URL để n8n, model
hoặc payload từ xa không thể đổi host/path. Chỉ `GET`, status `200`, không
redirect, đúng title/merchant/category và SHA-256 lowercase được nhận. Sai
field, extra field, commission/EPC/CVR hay bất kỳ raw content nào đều bị reject
trước canonical history.

## Chạy local, chỉ sau khi đã có sanitized capture

Khởi động adapter với history mới trong một thư mục riêng:

```bash
cd /Users/hadinh/Documents/ChatGPT/bot/lab/affiliate-bot
go run ./cmd/bot watcher serve /absolute/path/history.jsonl
```

Ở terminal khác, import capture từ file local (không commit file này):

```bash
go run ./cmd/bot watcher accesstrade-shopee-campaign-import \
  /absolute/path/history.jsonl /absolute/path/sanitized-capture.json
go run ./cmd/bot history replay /absolute/path/history.jsonl
```

Kết quả hợp lệ phải có `execution_permitted:false`,
`affiliate_link_created:false` và canonical state `GET_MORE_DATA`. Đó là hành
vi mong muốn: metadata campaign không có price/commission đã được xác minh và
không là business outcome.

Blueprint
[`M06-accesstrade-shopee-readonly.blueprint.json`](../../lab/n8n/M06-accesstrade-shopee-readonly.blueprint.json)
dùng Manual Trigger, inactive mặc định, và chỉ handoff `sanitized_capture_json`
tới loopback adapter. Không cấu hình n8n để fetch page đăng nhập trực tiếp.

`scripts/run_n8n_engine_regression.py` nhập một bản sao disposable của
blueprint vào n8n local và dùng metadata synthetic đã làm sạch. Regression đó
xác nhận append, retry `EXACT_DUPLICATE`, reject field `commission_rate` ngoài
allowlist trước ACK/report, append khi fingerprint đổi và history replay
`MATCH`. Nó không truy cập ACCESSTRADE và không thay operated evidence.

## M07 và điều chưa được suy ra

M07 nhận context từ canonical history như mọi M06 record khác. Hai field
`price` và `commission_rate` vẫn là `unknown`/`missing`; limitation nêu rõ
campaign/operator claim là volatile và không phải independent business truth.
Vì vậy output muốn khẳng định commission hay earnings sẽ bị grounding boundary
từ chối. Một output chỉ có thể abstain hoặc trích nguyên evidence hợp lệ dưới
`HUMAN_REVIEW`; không cấp approval hay execution.

Chỉ khi một operated run độc lập được ghi nhận, với capture được kiểm tra và
quyền/approval đúng, mới có thể thêm external evidence. Kể cả khi đó, nó vẫn
không tự chứng minh outcome kinh doanh, payout hay cho phép tạo link.
