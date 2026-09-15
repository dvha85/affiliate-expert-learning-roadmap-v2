# Operated evidence — M06 ACCESSTRADE Shopee Smartlink (2026-09-15)

Trạng thái: **local operated / read-only / chưa phải business outcome**.

## Nguồn và phạm vi

- Nguồn được quan sát trong phiên đăng nhập ACCESSTRADE: campaign `Shopee Việt
  Nam Smartlink cho tất cả thiết bị`.
- Metadata allowlist quan sát được: merchant `Shopee`, category `Thương Mại
  Điện Tử`, trạng thái `Chờ duyệt`, thời hạn `05/05/2023 - Nay`.
- SHA-256 của ảnh chụp full-page giữ riêng trong phiên vận hành:
  `46507a8166cec7849f21e5104b1c89d622193e8236f23b64fe54f1a144e6d0d1`.
- Không lưu HTML, cookie, credential, publisher ID, report, Smartlink/tracking
  link, commission, EPC/CVR, giá hoặc nội dung review/khuyến mãi.

## M06 CLI run

Capture sanitized được import vào một history mới bằng learner Bot thật. Kết quả:

```text
status=APPENDED
record_id=watch-4164382f85dd370210cf54d17e2bb1f0de38222dfc9e2a4d3c76ef2dc986b453
state=GET_MORE_DATA
execution_permitted=false
affiliate_link_created=false
network_fetch_performed=false
replay=MATCH
```

Record giữ `evidence_kind=synthetic`, `claim_kind=assumption`, và để
`price`/`commission_rate` ở trạng thái thiếu. Đây là chủ ý: một sanitized
transcription không tự chứng minh business truth.

## n8n local run

Một bản copy inactive của blueprint
`M06-accesstrade-shopee-readonly.blueprint.json` được import vào n8n local
2.38.1 và chạy qua adapter canonical loopback `127.0.0.1:8787` với cùng
sanitized capture và history mới. Workflow ID:
`accesstrade-operated-final-20260915`.

Kết quả execution:

```text
status=success
finished=true
result=APPENDED
canonical_history_ack=true
canonical_history_persisted=true
state=GET_MORE_DATA
execution_permitted=false
classification=observed_campaign_metadata_not_business_outcome
```

History sau n8n replay trả `MATCH`. Workflow không gọi ACCESSTRADE; HTTP node
chỉ gọi adapter loopback. Không tạo Smartlink, không publish, không gọi
executor và không tạo business outcome.

## Giới hạn chứng cứ

Đây là operated evidence trên local host do phiên này thực hiện, không phải
independent review trên deployment mục tiêu. Capture vẫn được canonical runtime
gắn `synthetic`/unverified; trạng thái campaign `Chờ duyệt` không cấp quyền
tham gia hay quyền thực thi. Provider diversity, live executor, attribution,
commission/payout, clean-machine pilot và deployment recovery vẫn chưa được
chứng minh. Readiness tiếp tục là `NOT_READY_FOR_PRODUCTION`.
