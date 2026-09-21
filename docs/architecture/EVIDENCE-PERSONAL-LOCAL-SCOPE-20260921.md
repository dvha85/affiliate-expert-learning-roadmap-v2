# Personal-local lab scope decision — 2026-09-21

Trạng thái: **PERSONAL_LOCAL_SCOPE_ACCEPTED**. Đây là quyết định phạm vi vận
hành cho chủ repo, không phải production acceptance và không thay đổi
`NOT_READY_FOR_PRODUCTION` của readiness audit.

## Quyết định

Repo này được dùng bởi một maintainer/operator duy nhất. Vì vậy, phạm vi chấp
nhận hiện tại là một local lab tự vận hành trên macOS native với:

- dữ liệu synthetic/read-only và model stub;
- adapter và n8n chỉ bind loopback mặc định;
- Go 1.27, Node 24 và n8n 2.38.1;
- runtime, backup và restore nằm ở các đường dẫn tách biệt dưới
  `$HOME/affiliate-lab`;
- không provider credential, public ingress, live executor, affiliate write
  hoặc business-outcome claim.

Không yêu cầu người thứ hai chạy lại runbook, user/máy sạch hoặc independent
beginner pilot cho phạm vi personal-local này. Khoảng trống đó được ghi là
**out of scope**, không được biến thành bằng chứng learner-operable hay
production.

## Bằng chứng được chấp nhận

Run `macos-br18a-20260921T155228Z` trong
[macOS local lab evidence](EVIDENCE-BR18A-MACOS-LOCAL-LAB-20260921.md) là
operated evidence đủ cho personal-local bounded lifecycle: health/authenticated
history read, restart, replay, backup/restore vào thư mục trống, durable STOP,
và fail-closed activation/gate sau restore. Các harness n8n M06/M07 và Schedule
Trigger vẫn chỉ là synthetic/read-only loopback evidence.

Kết quả này không đóng target-host/provider/public-network, Windows/Ubuntu,
power-loss hoặc atomic multi-file durability, distributed locking, live
executor, business outcome hay production readiness.

## Nhịp kiểm tra vận hành

Maintainer tự ghi một run record mới:

1. trước mỗi thay đổi lớn về code, schema, toolchain hoặc runtime;
2. sau restore hoặc thay đổi đường dẫn; và
3. định kỳ mỗi tháng khi local lab còn được sử dụng.

Mỗi run phải ghi `tested_at_utc`, host/OS, exact Go/Node/n8n versions, bind
addresses, runtime/backup/restore paths và kết quả health/restart/replay.
Backup phải được tạo vào target mới, restore vào thư mục trống khác, replay
được history, xác nhận `mission status` vẫn giữ `stop: true`, và xác nhận
activation/gate mới vẫn bị từ chối sau durable STOP. Token chỉ nằm trong
process environment; không ghi token vào log, URL, fixture hay backup.

Các lệnh chuẩn nằm trong
[BR-18a deployment/recovery runbook](BR-18A-DEPLOYMENT-RECOVERY-RUNBOOK.md).

## Ranh giới tuyên bố

`PERSONAL_LOCAL_SCOPE_ACCEPTED` chỉ có nghĩa là maintainer đã chấp nhận quy
trình local bounded cho mục đích cá nhân. Nó không nâng `BR-16a`, `BR-18b` hay
`BR-19` lên `IMPLEMENTED`, không bỏ qua các kiểm soát fail-closed và không cho
phép mở live write. Nếu sau này cần production hoặc public deployment, phải
mở lại các evidence gap tương ứng và chạy target-host/recovery drill riêng.
