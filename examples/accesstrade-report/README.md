# ACCESSTRADE report fixture — sanitized only

`sanitized-orders.csv` mirrors the visible columns of the Publisher **Báo cáo
đơn hàng** screen on 08/09/2026. It is invented, contains no account, link,
customer, or real order data, and proves neither a conversion nor payment.

The adapter accepts only these observed columns: `Mã đơn`, `Trạng thái`, `Giá
trị đơn hàng`, and `Hoa hồng`. Other columns are retained in this fixture so a
learner can see their provenance context, but they do not become canonical
OutcomeRecord fields.

`manifest.json` is a separate, explicit mapping from a **sanitized** order key
to a previously recorded human `action_id`. Do not infer that link from UTM:
UTM helps analysis, but does not prove ACCESSTRADE attribution. Use a new
`snapshot_id`, `source_ref`, and `outcome_id` for a later export; do not reuse
or overwrite prior outcomes.

The current observed status mapping is intentionally narrow:

| ACCESSTRADE report status | Outcome status |
|---|---|
| `Chờ xử lý`, `Tạm duyệt` | `PENDING` |
| `Được duyệt` | `VALID` |
| `Từ chối` | `CANCELLED` |

An unknown status is rejected. In particular, report commission is stored as
`reported_commission_vnd`, never `commission_paid_vnd`; this adapter has not
verified a payment-status export and must not infer payment.
