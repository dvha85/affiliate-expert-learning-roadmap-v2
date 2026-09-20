# Đối chiếu và xử lý phần còn lại của full review — 21/09/2026

Status: APPROVED — người dùng yêu cầu tự triển khai tuần tự, tạo, review và merge PR.
Baseline đã fetch: `ada7827a51dd44e9b670d24f69f79d01a449ea43`, clean `main`.
Authority: [review gốc](FULL-REPO-REVIEW-2026-09-19.md). Không mở rộng authority production.

## Đối chiếu đủ các finding

| Finding | Source/regression hiện có | Phần phải hoàn tất trong đợt này |
|---|---|---|
| RV-01 / F-01 | `mission_command.go`: immutable registry + mutable authority chặn rebind; `TestMissionIntentRebindAfterConsumptionHistoryFailsClosed` và restore/expiry tests | Giữ chính sách fail-closed, không đổi sang multi-intent ledger; chạy lại hosted learner/backup tests |
| RV-02 / F-02 | `n8n_runtime_env.py`, cả hai engine runner dùng chung builder; `test_n8n_runtime_env.py` | Bổ sung kiểm chứng environment tại ranh giới child process của cả hai runner; chạy n8n pin |
| RV-03 / F-03 | `core/m11/historical_chain.go` kiểm cost/authorization ID và hash trước phân profile | Test mutate riêng cost ID, cost hash, authorization ID/hash ở cả hai profile, rehash đúng và baseline non-authorizing |
| RV-04 / F-04 | `m11_chain.go` dùng `coreIntent`; large-integer regression | Bổ sung decimal/exponent và nested numeric values, so sánh hash/dữ liệu qua adapter |
| RV-05 / F-03 | Historical audit và artifact graph giới hạn authorization bằng health TTL | Test đúng biên, vượt 1 ns, execution tại/sau expiry, overflow; giữ historical clock độc lập wall clock |
| RV-06 / F-05 | `watcher.go` limit+1; hai bộ oversized/no-mutation tests cho M06/M07/history | Chạy lại learner tests, không thay body-limit contract |
| RV-07 / F-06 | `latest_execution_id` được gọi sau `stop_n8n` | SQLite regression: tick dư trước shutdown không được tính là tick sau restart; kiểm thứ tự gọi |
| RV-08 / F-07 | `n8n_change_scope.py` và table tests | Đảm bảo helper lỗi không bị coi là docs-only skip; bổ sung dependency tests |
| RV-09 / F-07 | Hai aggregate gate phụ thuộc child jobs và fail-closed | Thêm regression chống lệch workflow/governance. GitHub API hiện xác nhận main không protected, rules áp dụng rỗng: owner/admin phải cấu hình riêng |
| RV-10 / F-09 | `run_offline_checks.py`, README và capability docs | Nghiệm thu entrypoint đầy đủ trong hosted/toolchain-complete environment; cập nhật tracker theo kết quả thật |
| RV-11 / F-06 | Parser phân biệt reference/literal; ba unit tests | Bổ sung shared references và đối chiếu runtime flatted, chạy n8n thật |
| G-01 / F-08 | `missionNowUTC` cho admission; `TestMissionM11ExpiryRejectsAuthorityWritesWithoutMutation` có fresh process/restore | Không bỏ clock guard; sửa smoke fixture nếu lỗi chronology, không nới production invariant |

## Danh sách thực hiện tuần tự

1. **S-01 — baseline/reproduce:** Python baseline 174 tests PASS; audit PASS, 154 scoped claims, vẫn `NOT_READY_FOR_PRODUCTION`. Remote baseline mới nhất có Curriculum CI FAILURE tại BR-16a reservation race (run `35529283970`, job `106126893616`); Mission CI PASS. Điều tra lỗi này trước khi nói F-10 hoàn tất.
2. **S-02 — correctness/regression:** tái hiện lỗi fixture chronology, bổ sung các test nghiệm thu còn thiếu nêu trên, sửa đúng lỗi được chứng minh. Test bảo vệ logic cũ chạy trước/sau thay đổi.
3. **S-03 — CI contract:** fail-closed scope errors; kiểm gate dependencies và governance; nghiệm thu full offline runner trên hosted nếu local thiếu Go.
4. **S-04 — PR lifecycle:** self-review diff, Python/validators/audit; tạo PR, ghi review không giả human approval; đợi tất cả required gate và n8n executed trên exact head, sửa mọi failure trước merge.
5. **S-05 — post-merge:** fetch main, rebind baseline/plan/matrix/graph khi cần, kiểm post-merge CI; ghi rõ local/hosted/external evidence và đóng chỉ phạm vi đã chứng minh.

## Việc không thể tự đóng bằng thay đổi repo

- **EXT-01 / RV-09:** owner/admin bật PR requirement và required `curriculum-gate`, `mission-gate`, chặn force-push/deletion; xác minh negative-check/merge policy. GET branch protection trả `404 Branch not protected`; GET rules/branches/main trả `[]` trong phiên này. Không tự thay settings.
- **EXT-02 / RP-10:** selected-source/provider operated artifacts và n8n/provider deployment thật với nguồn/credentials do owner cấp; fixture/loopback không thay thế.
- **EXT-03 / RP-10:** live executor và business outcome có human authority, scope/limit/reconciliation và evidence độc lập; không bật live writes từ task sửa repo.
- **EXT-04 / RP-10:** target-host deployment/backup/restore drill và clean-machine beginner pilot độc lập; cần host/người tham gia được chọn.
- **EXT-05:** distributed locking, power-loss và atomic multi-file durability cần quyết định storage/deployment cùng fault-injection evidence; process-kill/local-lock tests không chứng minh các mục này.
- **MAINT-01 (không chặn correctness):** refactor `mission_command.go`, `backup_command.go`, `m11_registry.go` và thu gọn tracker theo mục 7 của review: PR riêng, có baseline test; không trộn vào bản sửa này.
- **SEC-01 (liên quan baseline, ngoài 11 RV):** govulncheck non-blocking failure phải được theo dõi riêng, không tuyên bố security scan sạch chỉ vì workflow tổng xanh.

## Kết quả

Đang thực hiện. Chưa gán PASS cho code mới hoặc hosted checks chưa chạy.
