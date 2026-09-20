# Đối chiếu và xử lý phần còn lại của full review — 21/09/2026

Status: REPO_VERIFIED / EXTERNAL_OPEN — đã triển khai, self-review và merge PR #491; không phải independent human approval.
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

PR [#491](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/491)
đã merge vào main tại `b447c04262bafd94125b05c2d6df4ecaa8edbeac`, từ exact head
`294e35722cf861fe91337a1a96c9647562367e00`. Self-review được ghi trên PR;
không giả independent human approval, không bypass check hoặc đổi settings.

| Bước | Kết quả đã xác minh |
|---|---|
| S-01 | Baseline Python 174 PASS; xác nhận Curriculum failure mới nhất trên ada7827 và main chưa protected |
| S-02 | Reproduce timestamp rollover và scope-helper fail-open trước sửa; sau sửa targeted 20 và full Python 185 PASS; các regression M11 mới PASS hosted |
| S-03 | Hai gate thực thi fail-closed; parity flatted 3 ca từ dependency n8n pin PASS; full offline runner chạy hết 37 bước PASS hosted |
| S-04 | PR #491 đã self-review, CI exact head PASS và normal merge; không squash mất ancestry của product baseline |
| S-05 | Clean merged-main audit PASS; baseline được rebind tới merge b447c04 trong follow-up docs. Post-merge CI và docs PR được kiểm trước bàn giao; không coi run trước merge là run sau merge |

Hosted exact-head evidence:

- [Curriculum CI 35531456259](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35531456259): `curriculum-gate`, Windows, learner shards/race, quickstart, BR-16a, backup/mutation và `offline-review-contract` PASS.
- [Full offline runner job 106132677723](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35531456259/job/106132677723): bốn Go module test/vet, learner race, Python 185, 14 validators, 11 smoke, readiness audit, diff check; kết thúc `OFFLINE CHECKS PASS`.
- [Mission CI 35531456254](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35531456254): core/mission và `mission-gate` PASS; n8n job **executed**, không skip.
- [n8n job 106132677716](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35531456254/job/106132677716): Node 24/n8n 2.38.1, flatted parity, engine M06/M07 và Schedule Trigger/restart PASS. Chỉ synthetic/sanitized fixture, canonical loopback và model stub.
- [Security run 35531456192](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/actions/runs/35531456192): CodeQL PASS; `govulncheck` panic `unexpected expr: *ast.KeyValueExpr`, non-blocking theo policy hiện có. **Không phải vulnerability scan sạch.**

Local: full Python 185 PASS hai lần sau sửa; compileall, structure, language,
continuity, artifact-spine và clean-worktree readiness audit PASS. Không chạy
Go/n8n local; full runner local exit 2 trước mọi check vì thiếu `go`.

RV-01…RV-08, RV-10, RV-11 và G-01 đã đạt phạm vi repository/offline/fixture
của đợt review; RV-09 đạt workflow/governance regression nhưng **ADMIN_OPEN**.
Không còn correction RV đã xác định nào chờ sửa source. Overall giữ
`NOT_READY_FOR_PRODUCTION`, 154 scoped claims; RP-10 và các BR partial không tự đóng.

### Phần còn phải làm ngoài nghiệm thu RV

Đây là backlog kế thừa từ mục 4/7 của review và `missing_evidence` trong matrix,
không phải tuyên bố toàn bộ đã được xử lý bởi PR #491:

| Phạm vi | Việc còn thiếu để đóng hoàn toàn | Điều kiện tiếp tục |
|---|---|---|
| RV-09 / EXT-01 | Bật branch protection/ruleset, require hai aggregate gate, PR, chặn force push/deletion; thử failure/cancel để kiểm policy thật | Owner/admin chọn và cho phép cấu hình; phiên này chỉ đọc settings |
| BR-13/14 / EXT-02 | Nghiệm thu source/profile drift và live capture, selected deployment/import run độc lập, canonical ACK/replay và reject/no-mutation | Nguồn/profile, target n8n và operator được cấp quyền; không suy business truth từ metadata |
| BR-15 / EXT-02 | Provider diversity và operated selected-source M07 grounding/tool enforcement | Provider/credentials và nguồn hợp lệ, independent evidence; không dùng model stub thay provider |
| BR-16a / EXT-04 | Beginner pilot trên máy sạch; broader malformed-graph, native Windows ancestor/path parity và non-cooperating-writer coverage | Người tham gia/máy mục tiêu; xác định thêm fault matrix ngoài các RV đã đóng |
| BR-17 / EXT-03 | Business outcome ingestion/evaluation và authorized live executor/lease operation | Thiết kế business contract, human execution authority, giới hạn/stop/reconciliation trước live writes |
| BR-17/18b / EXT-05 | Broader multi-file crash/power-loss, filesystem-crash, semantic-orphan combinations và distributed locking | Chốt storage/deployment contract, multi-host/fault-injection environment; local process-kill không đủ |
| BR-18b / EXT-04 | Target-host deployment recovery drill, governed external-source snapshot | Host/operator/backup policy thực tế và independent restore evidence |
| BR-19 | Independent review các claim còn mở; tiếp tục gắn CI với exact head sau mỗi thay đổi | Owner/reviewer; auditor cấu trúc không tự chứng minh remote/business readiness |
| MAINT-01 | Refactor các file lớn theo use case và tách nhật ký tracker khỏi trạng thái hiện hành | PR bảo trì riêng, giữ regression và marker audit; không phải correction bắt buộc của 11 RV |
| SEC-01 | Sửa toolchain compatibility của govulncheck rồi chạy lại đủ bốn module | PR security riêng; vẫn giữ rõ scan chưa hoàn tất, không xóa/skip scanner để làm xanh |
