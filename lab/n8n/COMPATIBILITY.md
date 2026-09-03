# Sổ tương thích n8n — n8n compatibility ledger

Blueprint JSON hợp lệ chỉ chứng minh cấu trúc file. Khả năng import và giữ đúng semantics khi chạy là thuộc tính của **engine n8n cụ thể**, nên phải được chứng minh trên instance thật.

## Baseline node đã khai báo

| Blueprint | Node | `typeVersion` đang khai báo |
|---|---|---|
| M06 | Schedule Trigger / Set / HTTP Request / Code | 1.2 / 3.4 / 4.2 / 2 |
| M07 | AI Agent / OpenAI Chat Model / HTTP Request Tool / Code | 3.1 / 1.2 / 1.2 / 2 |

Các version trên là version được lưu trong blueprint, **không phải tuyên bố rằng đó là version mới nhất hoặc đã được test trên engine hiện hành**.

## Trạng thái upstream cần review

Tại lần review 2026-09-03, source upstream n8n vẫn cho thấy AI Agent có `defaultVersion=3.1`, còn standalone HTTP Request đã có `defaultVersion=4.5`. HTTP Request Tool hiện được n8n triển khai như tool variant của HTTP Request; vì vậy việc thấy upstream có version mới **không phải lý do tự động sửa blueprint**.

```text
upstream version newer
!= current blueprint broken
!= safe to auto-upgrade
```

Node migration, credential behavior và tool wrapping phải được kiểm bằng import + execution smoke trước khi đổi `typeVersion` trong repo.

## Bản ghi admission cho release

`tested_n8n_version` cố ý để **UNVERIFIED** cho tới khi maintainer chạy smoke test dưới đây trên đúng engine version và ghi version, ngày, kết quả import cùng execution IDs vào PR/release evidence. Không được biến “JSON parse được” thành “n8n hỗ trợ workflow này”.

```text
tested_n8n_version: UNVERIFIED
tested_node_versions: declared above; engine support unverified
upgrade_review_cadence: trước mỗi lần nâng n8n engine và ít nhất mỗi quý
```

## Smoke test bắt buộc

1. Import cả hai blueprint vào một n8n instance sạch/local; kiểm tra node/type version nào bị unknown hoặc tự migrate trước khi lưu.
2. M06: chạy hai lần với public fixture/source được phép. Xác nhận GET-only, `NEW → UNCHANGED`, `observation_id`, correlation và canonical-history handoff. Đổi nội dung để nhận `CHANGED`; chỉ đổi thứ tự key của cùng JSON object phải vẫn là `UNCHANGED`.
3. M07: dùng least-privilege read-only credential (credential quyền tối thiểu, chỉ đọc) hoặc mock. Xác nhận evidence output bình thường, write/prompt-injection request bị chặn và output boundary vẫn ở `HUMAN_REVIEW`.
4. Khi upgrade engine, lặp lại import và execution smoke. Bất kỳ thay đổi hành vi nào ở HTTP Request Tool, AI Agent, credential scope, static-data behavior hoặc node migration đều chặn activation cho tới khi ledger này được review lại.

Kết quả smoke chỉ là **integration evidence (bằng chứng tích hợp)**, tự nó không phải Reality/Operated evidence. Không commit production credential, write scope hay secret vào blueprint hoặc repo này.
