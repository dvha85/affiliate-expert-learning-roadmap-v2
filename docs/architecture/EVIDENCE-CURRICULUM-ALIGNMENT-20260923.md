# Evidence đồng bộ chương trình học — 23/09/2026

## 1. Phạm vi và trạng thái

- Kế hoạch: [CURRICULUM-ALIGNMENT-PLAN-20260923.md](../plans/CURRICULUM-ALIGNMENT-PLAN-20260923.md)
- Mã: `CA-20260923`.
- Trạng thái nội dung: `DOCS_ALIGNED`.
- Trạng thái nghiệm thu runtime: `PARTIAL_EVIDENCE_GAPS`; full offline sau review PASS 38/38 trên clean snapshot ở mục 12, engine/Schedule Trigger fixture PASS riêng ở mục 11. Learner self-walkthrough và external evidence còn mở.
- Readiness của repo vẫn là `NOT_READY_FOR_PRODUCTION`.
- `PROGRESS.md`, Mission gates, authority boundaries và learner credit không bị thay đổi.

HEAD gốc của đợt triển khai là `277f1322830cf004e1597cfe97876eb5076ffccc`; tại thời điểm các snapshot dưới đây, thay đổi chưa commit/PR. Snapshot tạm `1531b8f188b067bacac29e0b2736684c48cf969e` đã chạy offline 38/38 **trước lượt review bổ sung**. Snapshot `cd18432f32bdfed523917868382481a5b6a3c664` sau đó chỉ cập nhật plan/evidence và chạy lại audit/static. Cả hai là commit tạm, không phải commit trên main; chúng không bao gồm các sửa lỗi review ở mục 10 và không chứng minh full offline runner trên diff hiện tại. Engine/Schedule Trigger rerun cục bộ trên diff sau review ngày 29/09 được ghi tại mục 11; không phải clean-tree hay hosted CI evidence. Phạm vi bàn giao Git được duyệt sau đó ở mục 13.

Full offline cho toàn bộ diff sau review đã được chạy mới ngày 29/09 trên clean snapshot `5972c2b647556454c495d75387ac248d4826fcce`: PASS 38/38. Đây là commit trong clone tạm, không phải commit trên main; provenance, lỗi tìm được và ranh giới nghiệm thu ở mục 12.

## 2. Baseline và toolchain

| Hạng mục | Quan sát |
|---|---|
| Baseline review | `main` tại `277f1322830cf004e1597cfe97876eb5076ffccc` |
| Mốc tiến độ learner | `1faf0c1`, M00.1 PASS ngày 04/09/2026; không cập nhật trong task này |
| OS/shell | Windows, PowerShell, repo root `D:\Project\Bot` |
| Python | `3.13.13` |
| Go | `go1.27.1 windows/386`, executable `C:\Program Files (x86)\Go\bin\go.exe`; test chạy với `GOARCH=amd64` và PATH tạm theo command |
| Node của lượt fixture trước | `v24.21.0`; portable runtime tạm đã được dọn sau kiểm tra |
| n8n CLI của lượt fixture trước | `2.38.1`, chạy bằng Node 24 trong home tạm; không coi runtime đó còn được cài/khả dụng sau cleanup |
| C compiler | LLVM-MinGW/Clang `22.1.8`, target `x86_64-w64-windows-gnu`; `gcc` frontend khả dụng |

Go hiện resolve được trong PATH tại `C:\Program Files (x86)\Go\bin\go.exe`; lượt kiểm auth/core bổ sung dùng `GOARCH=amd64`. Bảng mục 4 giữ kết quả lịch sử; Go test/vet/race đã rerun trong full offline ngày 29/09 ở mục 12, n8n engine được ghi riêng tại mục 11. Readiness audit mới vẫn kết luận `NOT_READY_FOR_PRODUCTION` với 154 claims; không nâng production readiness từ fixture runtime.

## 3. Thay đổi đã thực hiện

### Hợp đồng M06/M07 và regression

- Thêm [validate_learner_walkthroughs.py](../../scripts/validate_learner_walkthroughs.py) và [test_learner_walkthroughs.py](../../scripts/tests/test_learner_walkthroughs.py).
- Nối learner walkthrough validator vào [run_offline_checks.py](../../scripts/run_offline_checks.py) và [curriculum-ci.yml](../../.github/workflows/curriculum-ci.yml).
- Thống nhất `CANONICAL_ADAPTER_TOKEN` trong lesson/starter M06/M07; không thêm alias cho token cũ.
- Tách `NEW/UNCHANGED/CHANGED` (change detection) khỏi `APPENDED/EXACT_DUPLICATE` (persistence), và ghi `HANDOFF_ERROR` cho identity conflict; cập nhật compatibility ledger để không nhầm operated source lane với learner synthetic blueprint.
- Bổ sung negative tests cho token cũ, status layer sai, bridge/boundary bị thiếu và claim rehearsal vượt Mission/production.

### Đường thực hành M00–M05

- Thêm [PRACTICE-M00-M05.md](../../curriculum/PRACTICE-M00-M05.md), nối từ curriculum và starter M01–M05.
- Bổ sung walkthrough M03–M05 trên cùng learner Bot, gồm action/outcome/evaluation/proposal/review, continuity, failure cases và evidence handoff. `scripts/smoke_br12d.py` chạy packet t1/t2, M01/M02, rồi nối M03–M05 trên cùng history trong một temp workspace.
- Thêm liên kết có guard từ M01.1/M01.4/M02.4, M03.3/M04.3/M05.2, starter checkpoint/evidence templates và Continuity Checkpoint.
- Thêm input/output/failure/evidence matrix, lệnh PowerShell chuyển envelope sang input JSON UTF-8 không BOM và boundary nhấn mạnh rehearsal không tạo learner credit.
- README/ROADMAP phân biệt chặng học, personal-local lab, rehearsal và Mission PASS.

### Các file tài liệu đã rà/sửa

- Entry points: `README.md`, `ROADMAP.md`, `curriculum/README.md`.
- Lesson: `curriculum/M01/M01.4-failure-first-operated-proof.md`, `curriculum/M02/M02.4-restart-query-operated-proof.md`, `curriculum/M06/M06.1-readonly-watcher-contract.md`, `curriculum/M06/M06.2-idempotency-correlation-observability.md`, `curriculum/M06/M06.3-n8n-readonly-workflow.md`, `curriculum/M07/M07.3-n8n-evidence-agent.md`.
- Compatibility/runbook boundary: `lab/n8n/COMPATIBILITY.md` keeps operated source lanes separate from the learner synthetic blueprint.
- Regression coverage: `scripts/tests/test_learner_walkthroughs.py`, Windows n8n teardown test trong `scripts/tests/test_n8n_schedule_regression.py` và assertion wiring trong `scripts/tests/test_run_offline_checks.py`.
- Readiness audit: chỉ nhận `ROADMAP.md`, `lab/n8n/COMPATIBILITY.md` và các test-harness script khi đúng file được khai báo tại package CA-01/05/07; regression vẫn từ chối source drift khác.
- Starter M01–M06: README/checkpoint/evidence template liên quan.
- CI/offline wiring: `.github/workflows/curriculum-ci.yml`, `scripts/run_offline_checks.py`.

## 4. Kết quả kiểm chứng trước review bổ sung (snapshot lịch sử)

Bảng này giữ kết quả đã chạy trước mục 10, không phải acceptance của diff mới nhất.

| Kiểm tra | Kết quả | Phạm vi bằng chứng |
|---|---|---|
| Python regression suite | `199 PASS` | Unit/regression hiện có, walker/href/boundary tests, audit plan allowlist và Windows n8n teardown |
| Learner walkthrough validator | `PASS` | Token, status layers, bridge, input/output/evidence và boundary |
| Negative tests | `PASS` | Token cũ và M06 status layer bị bắt đúng; các fixture drift liên quan bị từ chối |
| Repo/mission/artifact/continuity validators | `PASS` | Static contract của repo |
| M06/M07 validators và M06 case validator | `PASS` | Static blueprint/fixture contract, không phải engine execution |
| Language policy | `PASS` | Tài liệu thuộc scope |
| Documentation links | `101 relative links / 34 changed Markdown files PASS` | Kiểm tra đích của link tương đối; bỏ qua external links và fenced code |
| `run_offline_checks.py --list` | `PASS` | Validator mới xuất hiện trong runner |
| `git diff --check` | `PASS` | Không có whitespace error trong diff |
| Node ACK isolated check | `PASS` | `APPENDED`/`EXACT_DUPLICATE` được nhận; `NEW`/`UNCHANGED`/`CHANGED` bị từ chối |
| Go test/vet (GOARCH=amd64) | `PASS` | `contracts`, `core`, `lab/affiliate-bot` và `lab/mission-runtime`; không bao gồm race |
| Learner Bot race regression | `PASS` | `GOARCH=amd64`, `CGO_ENABLED=1`, `CC=x86_64-w64-mingw32-gcc`; bốn package trong `lab/affiliate-bot` PASS |
| Offline smoke suite BR-08…BR-18b | `PASS` | Synthetic/local fixture; chạy lại với Go amd64, không phải provider/live/deployment evidence |
| n8n engine regression | `PASS_FIXTURE_RUNTIME` | n8n `2.38.1` + Node `24.21.0`, M06/M07 disposable SQLite/loopback fixture; engine thật đã import/execute |
| n8n Schedule Trigger regression | `PASS_FIXTURE_RUNTIME` | `APPENDED`, `EXACT_DUPLICATE` trước/sau restart và fail-closed khi adapter unavailable; Windows teardown và startup-budget regression đã PASS |
| n8n M07 adversarial/output execution | `PASS_FIXTURE_RUNTIME` | Real CLI rejects forged/write/redirect/missing-grounding paths and accepts only grounded read-only output |
| Full offline runner | `PARTIAL_CLEAN_TREE` | Steps 1–36 PASS, bước 37 readiness audit dừng tại `product baseline audit requires a clean working tree`; không stage/commit để lách guard |
| Readiness audit trên clean checkout | `PASS_AUDIT_NOT_READY` | Checkout tạm sạch `HEAD=277f1322830cf004e1597cfe97876eb5076ffccc`; audit trả `NOT_READY_FOR_PRODUCTION`, baseline `8f1c8aaf4524`, 154 scoped claims, mở BR-13/14/15/16a/17/18b/19 |
| Full offline runner trên snapshot implementation sạch | `PASS 38/38` | Checkout tạm sạch `1531b8f188b067bacac29e0b2736684c48cf969e`; Go test/vet/race, 199 Python tests, 15 validators, 11 smoke, readiness audit và diff check đều pass. Audit status là `NOT_READY_FOR_PRODUCTION`, baseline `8f1c8aaf4524`, 154 claims |
| M00–M05 continuity rehearsal | `PASS_SYNTHETIC_RUNTIME` | t1 `GET_MORE_DATA`, t2 `RANK_SCENARIO`, cả hai capture vào một history, tiếp tục action/outcome/advisor/evaluation/proposal/review qua learner CLI và fresh processes; chỉ là fixture synthetic |
| PowerShell M00→M01 handoff | `PASS_SYNTHETIC_RUNTIME` | Windows PowerShell: envelope `VALID`, input JSON UTF-8 không BOM, M01 `RANK_SCENARIO`; workspace tạm duy nhất |

Node ACK check chỉ chạy function tương đương đoạn code trong blueprint với dữ liệu synthetic cô lập. Nó không chạy n8n engine, adapter server, canonical history hoặc restart flow.

## 5. Mapping CA-T01…CA-T18 tại snapshot trước review

| Test | Trạng thái hiện tại | Bằng chứng / giới hạn |
|---|---|---|
| CA-T01 | `PASS_STATIC` | Negative test đổi token trong bản sao bị guard bắt; token runtime hiện hành thống nhất |
| CA-T02 | `VERIFIED_FIXTURE_RUNTIME` | Go auth/admission, n8n engine và negative paths đã chạy trên synthetic/loopback; không phải external authority evidence |
| CA-T03 | `VERIFIED_FIXTURE_RUNTIME` | Engine/Schedule Trigger fixture đã cho `APPENDED` rồi `EXACT_DUPLICATE`; learner self-walkthrough vẫn chưa được ghi |
| CA-T04 | `VERIFIED_FIXTURE_RUNTIME` | Go tests/smoke xác minh conflict và history không mutation; chưa phải learner credit |
| CA-T05 | `VERIFIED_FIXTURE_RUNTIME` | Go tests/BR-09 xác minh correlation/event identity, restart và replay trên fixture |
| CA-T06 | `VERIFIED_FIXTURE_RUNTIME` | Core Go tests, M06 case validator và learner guard giữ riêng change-detection với persistence |
| CA-T07 | `VERIFIED_FIXTURE_RUNTIME` | n8n engine và Schedule Trigger negative dừng trước success report khi adapter unavailable |
| CA-T08 | `VERIFIED_FIXTURE_RUNTIME` | Engine/Schedule Trigger xác minh restart adapter/n8n, exact record và replay |
| CA-T09 | `VERIFIED_SYNTHETIC_RUNTIME` | BR-09 packet → Bot → history/replay PASS; chưa phải learner-operated walkthrough |
| CA-T10 | `VERIFIED_SYNTHETIC_RUNTIME` | Go baseline và BR-08/09 xác minh field/ID/projection/null-vs-zero boundaries |
| CA-T11 | `VERIFIED_SYNTHETIC_RUNTIME` | BR-09 capture/duplicate/conflict/restart/replay PASS |
| CA-T12 | `VERIFIED_SYNTHETIC_RUNTIME` | BR-10a/b/c/d PASS cho action/outcome windows, pending/zero/late và fail-closed cases |
| CA-T13 | `VERIFIED_SYNTHETIC_RUNTIME` | BR-11a mock grounded/stale/future/orphan PASS; không có provider write |
| CA-T14 | `VERIFIED_SYNTHETIC_RUNTIME` | BR-12d synthetic evaluation/proposal/review continuity PASS; không auto-apply/business success |
| CA-T15 | `PASS_STATIC` | 101 relative links trên 34 Markdown files; 21 walkthrough entrypoints resolve đúng href; negative test href hỏng bị bắt |
| CA-T16 | `PASS_STATIC` | Boundary trong walkthrough/README/ROADMAP; không đổi `PROGRESS.md` hoặc Mission gate |
| CA-T17 | `PASS_RUNTIME_PARTIAL_AUDIT` | Full offline runner 38/38 trên clean snapshot; 199 Python tests, Go test/vet/race, validators và smokes PASS; audit còn `NOT_READY_FOR_PRODUCTION` với 154 claims |
| CA-T18 | `NOT_RUN_LEARNER` | Chưa có assisted/self-operated walkthrough thực tế của learner; không tự tạo credit |

## 6. Những kiểm tra còn bị chặn hoặc chưa chạy

- `python scripts/run_offline_checks.py`: trên snapshot tạm sạch `1531b8f188b067bacac29e0b2736684c48cf969e`, với `GOARCH=amd64`, `CGO_ENABLED=1` và LLVM-MinGW GCC, cả 38/38 bước PASS. Readiness audit là thành công về kiểm tra artifact, với kết luận `NOT_READY_FOR_PRODUCTION`.
- Tại lượt audit mục 12, working tree chính còn thay đổi chưa commit; clean-snapshot cũ không chứa sửa lỗi mục 10. Full offline sau review đã PASS 38/38 trên snapshot mới ngày 29/09; engine/Schedule Trigger fixture PASS riêng tại mục 11. Git handoff tiếp theo ở mục 13 phải kiểm đúng commit thực tế; không cập nhật product baseline `8f1c8aaf4524` hoặc coi snapshot là exact-head hosted CI/merged-main acceptance.
- Go unit/integration/vet và race trong `contracts`, `core`, `lab/affiliate-bot` và `lab/mission-runtime`: PASS với executable direct path, `GOARCH=amd64`, `CGO_ENABLED=1` và LLVM-MinGW gcc frontend.
- n8n engine regression, Schedule Trigger regression và M07 adversarial/output execution: PASS trên disposable local n8n `2.38.1`/Node `24.21.0`; lượt engine/Schedule Trigger sau review ngày 29/09 ghi tại mục 11. Đây là fixture runtime, chưa thay `tested_n8n_version: UNVERIFIED` hoặc operated deployment evidence.
- Lần Schedule Trigger đầu tiên hết timeout 35 giây; rerun với pool SQLite 1 và budget 90 giây PASS. Chưa có phép thử tách từng biến để kết luận nguyên nhân duy nhất hay startup chính xác 39–40 giây. Lượt review mới bổ sung phân biệt crash exit 1 với terminate chủ động.
- Fresh-machine/independent beginner pilot, provider/live source, deployment, multi-host, power-loss và production evidence: chưa chạy; nằm ngoài nghiệm thu docs-only này.
- Assisted learner walkthrough thực tế của chủ repo: chưa ghi nhận; không tự tạo learner credit hoặc cập nhật `PROGRESS.md`.

## 7. Trạng thái CA-00…CA-07

| Gói | Trạng thái | Ghi chú |
|---|---|---|
| CA-00 | `DONE_DOCS_STATIC` | Baseline, scope, working tree và toolchain blocker đã ghi |
| CA-01 | `DONE_DOCS_STATIC` | Guard, negative tests, runner và CI wiring đã có |
| CA-02 | `DONE_DOCS_STATIC` | Token/setup M06/M07 đã khớp implementation |
| CA-03 | `DONE_DOCS_STATIC` | Ba lớp semantics và failure expectations đã tách |
| CA-04 | `DONE_DOCS_STATIC` | Bridge M00–M02 đã nối vào lesson/starter; BR-09 t1/t2 import→M01→history/replay PASS |
| CA-05 | `DONE_DOCS_STATIC` | M03–M05 starter/template liên kết bridge; smoke chạy chuỗi learner CLI trên cùng history M00–M02 |
| CA-06 | `DONE_DOCS_STATIC` | Chặng học/personal-local boundary đã rõ |
| CA-07 | `PARTIAL_EVIDENCE_GAPS` | Full offline/clean-snapshot acceptance sau review PASS 38/38; Engine/Schedule Trigger fixture PASS riêng; còn learner walkthrough CA-T18 |

## 8. Kết luận và bước tiếp theo

Review bổ sung phát hiện checklist xanh trước đó chưa bao phủ mọi yêu cầu CA-01…CA-05; chi tiết finding và bản sửa nằm ở mục 10. Full offline 38/38 và n8n pin trước đó giữ nguyên như lịch sử; engine/Schedule Trigger fixture sau review PASS ở mục 11 và full offline trên clean snapshot chứa toàn bộ diff PASS 38/38 ở mục 12. CA-T17 đã có bằng chứng offline mới; CA-T18 chưa có learner tự walkthrough. Các gap external/operated/deployment vẫn mở.

Bước tiếp theo để nghiệm thu trải nghiệm học là chủ repo tự chạy walkthrough và tự ghi evidence vào nơi được quy định; smoke synthetic không làm thay bước đó. Muốn nâng readiness còn cần external/production evidence riêng; release admission vẫn `UNVERIFIED` theo ledger. Không cần học lại BOOT/M00 và không tự reset tiến độ learner.

## 9. Rerun theo release pin — 23/09/2026

- Runtime xác minh: n8n `2.38.1`, Node `24.21.0`, SQLite disposable, loopback adapter/model stub; compiler `x86_64-w64-mingw32-gcc` và `CGO_ENABLED=1` vẫn được giữ cho các bước Go/race.
- `scripts/run_n8n_engine_regression.py`: `PASS`; M06/M07 import/execute thật, retry/replay/exact-number và fail-closed/grounded output đều đạt.
- `scripts/run_n8n_m06_schedule_regression.py`: `PASS`; output quan sát được là `APPENDED` execution 1, `EXACT_DUPLICATE` execution 2 và execution 3 sau restart, cùng rejection trước ACK/report khi adapter down.
- Regression unit sau cập nhật harness và guard: `199 PASS`; riêng schedule test `8 PASS`, learner walkthrough tests `10 PASS`.
- `scripts/smoke_br12d.py`: `PASS`; hai M00 snapshots t1/t2 được import/M01/capture vào cùng history, M03–M05 tiếp tục qua action/outcome/advisor/evaluation/proposal/review bằng cùng Bot workspace. Rollback FAIL/PASS/rollback/restore cũng được kiểm trong workspace riêng.
- PowerShell command path cho M00→M01: `PASS`; artifact được đổi từ envelope thành input UTF-8 không BOM, M01 in `RANK_SCENARIO`.
- `run_offline_checks.py`: `PASS 38/38` trên temporary clean snapshot; Readiness audit in `NOT_READY_FOR_PRODUCTION`, 154 claims, còn mở BR-13/14/15/16a/17/18b/19.
- Không ghi nhận credential, secret, learner credit hay thay đổi `PROGRESS.md`. Phạm vi vẫn là synthetic/read-only/loopback; không đóng release admission, provider/live, deployment, pilot hoặc production gap.

## 10. Review lại diff theo plan — 23/09/2026

Người review/sửa: Codex (self-review theo yêu cầu chủ repo, không phải reviewer độc lập).
Phạm vi: toàn bộ diff tracked và 5 file untracked của CA-20260923 trên HEAD
`277f132`; bảo toàn các thay đổi đang có, không tạo commit/PR trong lượt review.

| Finding | Yêu cầu plan chưa đạt | Sửa và bằng chứng |
|---|---|---|
| CR-01 / P1 | CA-02: M06 tạo token hai lần, thiếu parent history/cwd PowerShell, chưa hướng dẫn n8n kế thừa cùng token và home cô lập; M07 vẫn bắt đầu bằng credential thật | M06.3 có một token mỗi shell, temp store, parent-child process, isolated_n8n_environment và env-access chỉ trong child; M07/starter/runbooks dẫn fixture/model stub trước provider. PowerShell parser và hai Python launcher snippets được kiểm với subprocess mock, chứng minh bỏ inherited external DB/home |
| CR-02 / P1 | CA-03: starter vẫn yêu cầu blueprint trả NEW/UNCHANGED/CHANGED; checkpoint cấm POST nội bộ; correlation mới cho retry bị hiểu sai | Starter/lesson/ledger thống nhất bảng APPENDED/EXACT_DUPLICATE/HANDOFF_ERROR/new_event; core previous_hash kiểm riêng; template thêm count/hash/ACK/replay/auth. Core M06 và learner auth/retry tests PASS |
| CR-03 / P1 | CA-04/05: walkthrough đổi cwd lặp, placeholders chưa giải quyết, capture nhầm as_of/ingested_at, thiếu mẫu input dùng chung; smoke xanh chưa kiểm lệnh Markdown | Viết lại setup một lần, 6 mẫu JSON nối d12→a12→o12→e12→p12→r12, shell commands chạy từ repo root, Go/JSON đúng chỗ và bài diagnostic FAIL/PASS/rollback do learner tự viết. smoke_br12d đọc mẫu và chạy chính các command blocks từ Markdown trên Windows PowerShell; PASS |
| CR-04 / P2 | CA-01: token trong notes có thể che thiếu Authorization; guard chỉ có keyword, từ chối historical quote/negated warning và khóa nguyên heading tiếng Anh | Parse node/header/POST endpoint/ACK/report edges; test negative auth/endpoint/bypass/ACK, xóa boundary, lịch sử và câu phủ định, heading tương đương. Static checks không được diễn giải thành execution proof |
| CR-05 / P2 | CA-07: Windows teardown nhận mọi exit 1, kể cả process đã crash | Chỉ nhận exit 1 khi đã gọi terminate; test process đã exit 1 trước cleanup phải raise; không sửa workflow semantics |
| CR-06 / P2 | CA-01/07: đọc section CA-07 kéo dài tới cuối file, path ở mục ngoài cũng được chấp nhận là scope | Giới hạn tới heading cấp 1–3 tiếp theo. Regression chứng minh path ở mục sau vẫn bị audit từ chối; giữ nguyên product baseline/claim count |
| CR-07 / P2 | CA-07: evidence gọi snapshot cũ là diff hiện tại, runtime đã xóa vẫn ghi như khả dụng; startup 39–40s chưa đủ bằng chứng nguyên nhân | Mục 4/5/9 giữ snapshot lịch sử; mục này ghi kiểm chứng mới và các giới hạn. Chưa coi full offline/engine cũ là PASS cho thay đổi sau review |
| CR-08 / P2 | CA-06/07: entrypoint chưa nói trực tiếp NOT_READY không chặn M00–M02; checklist reviewer/PR vô điều kiện mâu thuẫn phạm vi personal-local và mục 9 | README/ROADMAP/curriculum dẫn PROGRESS và giải thích missing account/outcome giữ gate; checklist tách self-review, learner walkthrough và Git workflow có điều kiện |

### Kiểm chứng của lượt review

- Baseline trước sửa: 18 Python tests liên quan và BR-12d cũ PASS. Năm test mới
  tái hiện CR-02/03/04/05 (hai test cho CR-04) đều RED trước sửa;
  test section leak CR-06 và test trình bày status tương đương cũng RED.
- BR-12d sau sửa: `DOCUMENTED WALKTHROUGH ps PASS` (lệnh/mẫu đọc trực tiếp từ
  Markdown), chain M00–M05 và rollback exit 1→0→1→0 PASS; các process dùng
  workspace tạm, history hai record và downstream stores mỗi loại một record.
- Go: 20 tests/subtests auth/admission/retry M06/M07 trong learner Bot PASS;
  7 tests core M06 PASS với GOARCH=amd64. Không đổi Go production source/schema.
- PowerShell parser và Bash `-n` PASS cho các shell block. Bash syntax check trên
  Git Bash Windows không phải bằng chứng runtime macOS; POSIX command path được
  wiring vào BR-12d trên CI/Linux nhưng chưa có hosted result mới cho diff này.
- Full Python regression: 211/211 PASS sau các sửa review; riêng walkthrough
  20 tests, schedule 9 tests PASS. Shell command smoke và focused Go checks ở
  trên độc lập với static validator; không chỉ dựa vào keyword/marker.
- Repo/mission/artifact/continuity/language/learner/n8n static validators và
  git diff --check PASS; 113 link tương đối trên 37 Markdown files thay đổi PASS.
  Không đổi PROGRESS.md,
  CURRICULUM.md, schema/core hoặc production Go source.
- Chưa rerun n8n engine 2.38.1/Node24 sau thay đổi teardown. Manual editor/startup
  chỉ được kiểm parser/env contract; chưa coi là operated run của learner.
- Clean-tree audit trực tiếp cho diff mới còn chờ revision được commit theo
  workflow được phép; unit fixtures của audit kiểm code drift và section scope.
  Không đóng CA-T18, independent review hay release/provider/deployment gap.
- Selector n8n_change_scope trả `run` với danh sách diff hiện tại; CI engine
  không được claim PASS/skipped trước khi có hosted result trên revision đó.

### Đính chính sự cố cô lập ở lượt chẩn đoán trước

Log công cụ trước review có một lệnh PowerShell truyền environment sai do quoting;
n8n tiếp tục chạy với settings mặc định, bind `::` cổng 5678 và thực hiện migrations.
Log ghi `Recorded version change: 2.35.7 -> 2.38.1` và dependency index 0 workflows.
Process đó đã được dừng; đây không phải disposable-home run. Chưa kiểm chứng tác
động/khả năng rollback database mặc định, không tự phục hồi hoặc ghi đè dữ liệu.
Vì vậy không khẳng định toàn bộ lượt chẩn đoán trước không chạm state n8n cá nhân.
Hai regression harness có home cô lập vẫn có output PASS riêng như mục lịch sử.
Lượt review này không khởi động n8n trên home mặc định.

## 11. Rerun n8n sau review — 29/09/2026

- Runtime thực tế: n8n CLI `2.38.1` chạy bằng Node `24.21.0` trên Windows; n8n `N8N_USER_FOLDER`, SQLite database và Go cache đều nằm trong thư mục tạm. Go adapter build với `GOARCH=amd64`, `CGO_ENABLED=1`, `CC=x86_64-w64-mingw32-gcc`.
- `scripts/run_n8n_engine_regression.py`: `PASS` trên working tree hiện tại sau CR-01…CR-08. M06 kiểm tra `APPENDED`/`EXACT_DUPLICATE`, reorder JSON, identity conflict/source rejection không mutate history, event mới, selected-source sanitized capture/retry/rejection, replay và adapter-unavailable fail-closed. M07 dùng adapter loopback và OpenAI-compatible model stub chạy trong process; POST/redirect/adapter-down, forged commission, malformed output bị reject; grounded proposal, selected-source `null` claim, exact number `9007199254740993`, restart/replay/revalidation PASS.
- `scripts/run_n8n_m06_schedule_regression.py`: `PASS`. Schedule Trigger thật ghi execution 1 `APPENDED`, execution 2 `EXACT_DUPLICATE`, execution 3 vẫn `EXACT_DUPLICATE` sau restart cả n8n và adapter, cùng record ID; adapter unavailable dừng trước ACK/report. Log có một SQLite ping timeout rồi recovered; runner vẫn hoàn tất toàn bộ assertion.
- Adapter/model là fixture loopback/local; không dùng provider credential hay gọi affiliate endpoint. npm 11 chặn install scripts mặc định nên chỉ `isolated-vm@7.0.1` và `sqlite3@5.1.7` được approve trong prefix tạm để khởi tạo expression engine/SQLite native module. Runtime và log cài đặt nằm ngoài repo.
- Đây là fixture evidence trên working tree chưa commit, không phải full offline runner trên clean checkout, learner-operated walkthrough, Reality/Operated, release admission hay production deployment evidence. Giữ `tested_n8n_version: UNVERIFIED`, `NOT_READY_FOR_PRODUCTION`, `PROGRESS.md` và Mission gates không đổi.

## 12. Full offline audit sau review trên clean snapshot — 29/09/2026

Thực hiện theo yêu cầu chủ repo: chốt full offline trên bản sao sạch chứa toàn bộ thay đổi hiện tại, không chỉ checkout HEAD cũ. Không stage/commit/push repo chính hoặc tạo PR.

### Snapshot và cách kiểm

- HEAD nguồn: `277f1322830cf004e1597cfe97876eb5076ffccc`; 44 file tracked thay đổi và 5 file mới chưa tracked.
- Snapshot PASS: `5972c2b647556454c495d75387ac248d4826fcce`; Git tree `b552cca0ad123ecafc15f2044be8e99d1386fe29`. Commit chỉ tồn tại trong clone tạm, giữ ancestry của product baseline.
- Chép toàn bộ file tracked hiện hữu và untracked không bị ignore; không chép `.env`, workspace/private evidence, runtime database hoặc cache bị ignore. Inventory và SHA-256 của **693/693 file** khớp working tree trước chạy; kiểm lại sau chạy vẫn khớp, Git snapshot sạch và HEAD/status repo chính không đổi.
- Chạy `python scripts/run_offline_checks.py` từ root snapshot trên Windows PowerShell, Python `3.13.13`, Go `go1.27.1 windows/386` với `GOOS=windows`, `GOARCH=amd64`, `CGO_ENABLED=1`, `CC=x86_64-w64-mingw32-gcc` (LLVM-MinGW/Clang `22.1.8`). `GOWORK=off`, `GOTOOLCHAIN=local`, `GOPROXY=off`, `GOSUMDB=off`; dùng toolchain/dependency đã có, không cài hoặc tải runtime mới.
- Lượt PASS chạy từ `10:38:49` đến `10:47:54` ngày 29/09/2026, UTC+07:00; exit `0`. Log ghi đủ 38 step markers và `OFFLINE CHECKS PASS: offline/read-only contract completed`.

### Finding trong lượt audit và kết quả sau sửa

Snapshot đầu `591651cff859967133df2a7cf05fd0aff845bb64` dừng đúng tại bước 15/38: `validate_language_policy.py` báo hai dòng English-only trong ghi chú n8n mới tại `lab/n8n/COMPATIBILITY.md`. Go test/vet/race và 211 Python tests trước đó đã PASS, nhưng **lượt này không được tính full PASS**.

Đã dịch đoạn mô tả M06/M07 sang tiếng Việt, không đổi runtime, blueprint, assertion, allowlist hoặc semantics. Validator language chuyển từ FAIL sang PASS; sau đó tạo snapshot mới nêu trên và chạy lại **từ bước 1**, không resume hoặc ghép kết quả hai lượt.

| Nhóm kiểm | Kết quả mới |
|---|---|
| Go test/vet | PASS cho `contracts`, `core`, `lab/affiliate-bot`, `lab/mission-runtime`; test chạy `-count=1` |
| Go race | PASS cho 4 package learner Bot với `-race -count=1` và cgo/gcc |
| Python regression | 211/211 PASS trong 166.502 giây |
| Offline validators | 15/15 PASS, gồm language policy, learner walkthrough và M06/M07 contract/CLI cases |
| Offline smoke | 11/11 PASS; BR-12d thực thi đúng command blocks PowerShell và 6 JSON inputs M00–M05; BR-16a/BR-18b kiểm runtime, restart/replay, STOP và backup/restore trên fixture |
| Readiness audit | PASS về nhất quán artifact: 154 scoped claims, product baseline `8f1c8aaf4524`; kết luận vẫn `NOT_READY_FOR_PRODUCTION` |
| Whitespace và tính toàn vẹn | PASS cho runner, diff từ HEAD nguồn tới snapshot và working tree; snapshot sạch trước/sau; 693 file không bị audit thay đổi |

### Bàn giao và giới hạn

- CA-T17: `PASS_OFFLINE_CLEAN_SNAPSHOT`. CA-07 vẫn `PARTIAL_EVIDENCE_GAPS` vì CA-T18 `NOT_RUN_LEARNER`; không cấp learner, Reality hoặc Operated PASS từ smoke synthetic.
- `PROGRESS.md` không đổi; SHA-256 trước/sau là `33a8ce1662a5003a9fb117494913bde71d6827173ebcb1a2bedd33fb03a02509`. Không sửa `CURRICULUM.md`, Mission gates, production source, readiness matrix/graph, product baseline hoặc claim count.
- Runner này không chạy real n8n engine/Schedule Trigger; bằng chứng riêng trên n8n `2.38.1` + Node `24.21.0` ở mục 11 vẫn giữ nguyên. Lượt audit chỉ sửa cách diễn đạt trong ledger, không thay code của engine/adapter/model.
- Đây là local Windows offline acceptance, không phải hosted CI, POSIX/macOS runtime, learner walkthrough, fresh-machine pilot, provider/live, deployment hoặc production evidence. Các mục BR-13/14/15/16a/17/18b/19 vẫn partial/open; release admission giữ `UNVERIFIED`.
- Log và manifest local được giữ ngoài repo trong `C:\Users\hadv\AppData\Local\Temp\bot-offline-audit-20260929-d7472d18`: lượt FAIL ở `full-offline-audit.log`/`snapshot-manifest.json`; lượt PASS ở `rerun/full-offline-audit.log`/`rerun/snapshot-manifest.json`. Đây là artifact tạm trên máy hiện tại, không phải log đã được lưu trên CI.
- Sau snapshot PASS, phần đồng bộ biên bản chỉ gồm file plan và evidence này. Kiểm bổ sung repo/learner/language/readiness trên clean snapshot của biên bản được lưu riêng tại `post-sync-checks.log` và `final-manifest.json` trong thư mục audit; không gắn nhãn một lượt full offline thứ ba cho kiểm tra tài liệu đó.

## 13. Bàn giao CA-T18 và Git workflow — 29/09/2026

Chủ repo đã giao hoàn tất phần có thể thực hiện và commit/push/tạo PR trên repo gốc. Nhánh bàn giao: `codex/curriculum-alignment-20260929`, base `main`; không tự merge. Các kết quả exact-head local/hosted CI được ghi trên PR với SHA tương ứng. Kết quả mục 12 không tự bao phủ sửa bổ sung ở mục này.

### Rà phần còn lại trước PR

- Phần kỹ thuật CA-01…CA-06 và các fixture CA-07 đã có bằng chứng; learner walkthrough CA-T18 vẫn `NOT_RUN_LEARNER`. Không thể thay thao tác/explain-back của người học bằng smoke tự động.
- Self-review phát hiện walkthrough ghi `learner/M01…M05` là đã ignore, nhưng `git check-ignore` không nhận đường dẫn đó. Đã thêm `test_private_learner_evidence_path_is_actually_gitignored`, quan sát RED trước sửa, rồi đổi hướng dẫn sang `workspace/learner/M01…M05`; test kiểm cả đường dẫn trong tài liệu lẫn Git ignore thực tế cho M01, M05 và CA-T18. Không sửa `.gitignore`, không di chuyển hoặc đọc dữ liệu learner riêng tư.
- Các gap provider/live, deployment, độc lập/máy sạch và production vẫn ngoài phạm vi này; không tự mở credential, thu thập observations thay learner hoặc sửa dữ liệu n8n cá nhân từ sự cố lịch sử mục 10.

### Phiếu CA-T18 cho người học tự điền

Sao chép phần phiếu trống vào `workspace/learner/CA-T18.md` (đã ignore), không điền vào tài liệu public này. Đi theo [walkthrough](../../curriculum/PRACTICE-M00-M05.md) phần 2–5 cho M00–M02 trước; M03–M05 là rehearsal tùy chọn, không được dùng để vượt Mission gate. Ghi rõ `synthetic rehearsal` hoặc evidence thật; không trộn hai loại.

- Người thực hiện/ngày/OS/shell: **chưa ghi nhận**.
- Commit/workspace riêng dùng cho bài: **chưa ghi nhận**.
- Mức trợ giúp (tự làm/được hướng dẫn, ai giúp và giúp ở bước nào): **chưa ghi nhận**.
- Trạng thái nghiệm thu trải nghiệm: `NOT_RUN_LEARNER` cho tới khi có thao tác và review thực tế.

| Bước người học tự làm | Kết quả dự kiến với fixture | Kết quả thực tế/log đã bỏ secret | Vướng mắc/trợ giúp |
|---|---|---|---|
| Setup và import M00 packet | Có workspace mới; envelope `VALID`, tách đúng artifact làm input | Chưa ghi nhận | Chưa ghi nhận |
| Dự đoán rồi chạy M01 với t1/t2 | t1 `GET_MORE_DATA`, t2 `RANK_SCENARIO`; giải thích missing không phải zero | Chưa ghi nhận | Chưa ghi nhận |
| Capture hai snapshot vào cùng history M02 | `APPENDED`, retry `EXACT_DUPLICATE`, giữ hai record | Chưa ghi nhận | Chưa ghi nhận |
| Khởi động process mới, list/replay | IDs resolve; `replay=MATCH`, không duplicate/history rewrite | Chưa ghi nhận | Chưa ghi nhận |
| Ca lỗi và explain-back theo walkthrough | Input/identity sai bị từ chối; chỉ rõ limitation và chủ sở hữu canonical history | Chưa ghi nhận | Chưa ghi nhận |

Chỉ xem xét đóng CA-T18 khi có bản ghi người học thực sự tự làm, expected/observed, lỗi cần trợ giúp, explain-back và người review/kết luận thực tế. Đây là nghiệm thu khả năng làm theo hướng dẫn; vẫn không tự cấp M00 Reality hoặc Mission PASS. Bằng chứng E1 thật và các gate học tập tiếp tục theo `PROGRESS.md`, không đổi từ PR này.
