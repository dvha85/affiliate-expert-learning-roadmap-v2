# Rà soát toàn bộ repository và kế hoạch chỉnh sửa — 2026-09-19

## 0. Trạng thái tài liệu

- Trạng thái: `IMPLEMENTED_OFFLINE_WITH_HOSTED_BLOCKERS`.
- Phạm vi: review từ đầu toàn bộ repository tại `main` / `2ec78f1a40c264d796ec5b7c459abf864cd9a97a`; triển khai offline trên nhánh `codex/full-repository-hardening-20260919`.
- Tài liệu này **không** tự cấp quyền triển khai code, thay branch protection, dùng credential/provider, chạy live executor hoặc nâng readiness.
- Trạng thái readiness giữ nguyên: `NOT_READY_FOR_PRODUCTION`.
- Đã triển khai offline: adapter size/auth guard, n8n token wiring, CI pin/cache/gates,
  security/hygiene metadata, runbook Go version và readiness evidence sync.
- Còn chờ hosted/admin: Go/Windows/n8n exact-head checks, branch protection, provider,
  deployment/target-host, pilot và business-outcome evidence.
- Nguồn trạng thái chi tiết hiện hành vẫn là:
  - [REVIEW-REMEDIATION-PLAN.md](REVIEW-REMEDIATION-PLAN.md)
  - [READINESS-MATRIX.json](READINESS-MATRIX.json)
  - [READINESS-EVIDENCE-GRAPH.json](READINESS-EVIDENCE-GRAPH.json)

## 1. Kết luận ngắn

Repo có nền tảng kiểm soát tốt hơn đáng kể so với một lab thông thường: authority boundary rõ, strict JSON/provenance, fail-closed path, backup/restore, mutation tests, Windows/race jobs và readiness audit đều đã có. Không phát hiện secret rõ ràng trong source/config được track; Git/JSON integrity và toàn bộ lớp Python có thể chạy trên máy review đều đạt.

Tuy nhiên vẫn cần chỉnh sửa trước khi mở rộng operated/deployment:

1. CI có việc khẩn cấp do `actions/cache@v4` còn chạy Node 20, trong khi GitHub đã cảnh báo mốc loại bỏ Node 20 ngày 23/09/2026.
2. Required-check governance chưa bao phủ rõ toàn bộ 13 job quan trọng; cấu hình branch protection thực tế chưa được xác minh.
3. Ba HTTP input path dùng `io.LimitReader` nhưng không đọc thêm một byte để phát hiện payload vượt trần; JSON hợp lệ kèm padding vượt trần có thể bị cắt rồi vẫn được chấp nhận.
4. Canonical adapter chỉ dựa vào loopback, chưa có authentication giữa n8n/process và adapter; loopback không phải ranh giới trust khi host có nhiều process/user/container.
5. CI cache Go đang cấu hình thiếu dependency path ở hai job và đã phát warning trên run hiện hành.
6. PR path gate của n8n chưa bao phủ đầy đủ dependency làm thay đổi binary/contract.
7. Runbook deployment ghi Go `1.23+`, trái với `lab/affiliate-bot/go.mod` yêu cầu Go `1.27`.
8. `main_baseline` hiện có nghĩa chưa đủ rõ: ba readiness artifact cùng trỏ `ec2499a`, nhưng `HEAD` đã tiến thêm 15 commit audit/docs. Audit chỉ kiểm ba giá trị bằng nhau, chưa kiểm quan hệ Git hoặc loại thay đổi kể từ baseline.
9. Một số file đã quá lớn và trạng thái được lặp thủ công giữa Markdown/JSON/Python, làm tăng chi phí review và rủi ro drift.
10. Repo công khai chưa có lớp hygiene/security maintenance tối thiểu như ignore Python cache, dependency update config, security policy và quyết định license.

## 2. Snapshot và phạm vi đã kiểm

### 2.1 Repository snapshot

| Thuộc tính | Kết quả |
|---|---|
| Branch / HEAD | `main` / `2ec78f1a40c264d796ec5b7c459abf864cd9a97a` |
| Working tree trước review | sạch |
| File được track | 657 |
| Thành phần chính | 203 Go, 66 Python, 81 JSON, 295 Markdown, 2 workflow YAML |
| Go modules | `contracts`, `core`, `lab/affiliate-bot`, `lab/mission-runtime` |
| Readiness | `NOT_READY_FOR_PRODUCTION` |
| Readiness package | `RP-01…RP-09=PARTIAL`, `RP-10=OPEN` |
| Structured criteria | `BR-13/14/15/16a/17/18b/19=PARTIAL` |

### 2.2 Hạng mục đã rà

- Entry points, curriculum, governance, architecture và runbook.
- Bốn Go module, source/test layout, file/network/path/durability boundaries.
- Python validators, smoke/mutation runners và readiness audit.
- Contracts/schema/eval/mission/starter-kit/n8n blueprint.
- Hai GitHub Actions workflows, path gating, cache, runner/toolchain và required-check documentation.
- Readiness plan/matrix/evidence graph và bằng chứng hậu merge.
- Git hygiene, generated files, file size/complexity, secret-pattern scan và public issue backlog.
- Live CI của commit hiện hành ở mức trang GitHub công khai; không có quyền admin để đọc branch protection thực tế.

### 2.3 Kết quả kiểm chứng trong phiên review

| Kiểm tra | Kết quả |
|---|---|
| `python scripts/audit_readiness.py .` | PASS; `NOT_READY_FOR_PRODUCTION`; 150 scoped claims |
| `python -m unittest discover -s scripts/tests -v` | PASS; 157 tests |
| Foundation validators | PASS: missions, repo, artifact spine, continuity, language |
| Semantic/static validators | PASS: agent semantics, semantic contracts, M11, n8n M06/M07 static |
| n8n M07 adversarial/output validators | không chạy hết vì máy review không có Go |
| Go test/vet/race/smoke local | không chạy vì máy review không có Go |
| JSON parse | PASS toàn bộ file JSON |
| `git diff --check` | PASS trước khi tạo tài liệu này |
| `git fsck --strict` | không có lỗi object; chỉ có hai dangling tree không được ref |
| Secret-pattern scan | không phát hiện credential/private key rõ ràng trong file source/config được rà |
| GitHub Actions trên `2ec78f1` | hai workflow thành công; 13 job được liệt kê, nhưng còn warning cache/runner |

Giới hạn: kết quả local không thay thế Go/Windows/race/n8n engine CI; CI thành công không thay thế selected-source, provider, pilot, target host, power-loss hoặc business evidence.

## 3. Phát hiện cần xử lý

### FR-01 — P0 — GitHub Actions cache còn phụ thuộc Node 20

**Bằng chứng**

- [mission-agent-path-ci.yml](../../.github/workflows/mission-agent-path-ci.yml) dùng `actions/cache@v4` cho n8n.
- Run hiện hành báo action này target Node 20 và đang bị ép chạy Node 24.
- GitHub công bố Node 20 sẽ bị loại khỏi runner ngày 23/09/2026; `actions/cache@v5` dùng Node 24.

**Rủi ro**

- CI có thể hỏng hoặc tiếp tục phát warning ngay sát mốc loại bỏ runtime.
- n8n engine là acceptance path quan trọng; không nên chờ đến khi runner thay đổi mới sửa.

**Chỉnh sửa đề xuất**

- Đổi sang `actions/cache@v5`; nếu dùng self-hosted runner phải xác nhận runner `>=2.327.1`.
- Pin action theo full commit SHA sau khi chọn version đã review; giữ comment version dễ tra.
- Chạy lại main-push n8n engine path và xác nhận warning Node 20 biến mất.

### FR-02 — P0 — Required checks chưa tạo một gate ổn định bao phủ toàn bộ suite

**Bằng chứng**

- [REPOSITORY-GOVERNANCE.md](../governance/REPOSITORY-GOVERNANCE.md) chỉ khuyến nghị bốn check.
- Hai workflow hiện có tổng cộng 13 job, gồm Windows, shard, race, quickstart, backup/mutation và n8n engine.
- Branch protection thực tế không đọc được từ public/API không đăng nhập; chính governance cũng ghi là chưa xác nhận.
- GitHub issue [#2](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/issues/2) vẫn mở và còn nêu các tên check cũ đã bị governance cấm dùng.

**Rủi ro**

- PR có thể đủ bốn check khuyến nghị nhưng một job consequential khác fail mà vẫn đủ điều kiện merge.
- Tên job tăng/đổi theo thời gian làm branch protection drift.

**Chỉnh sửa đề xuất**

- Thêm một final gate ổn định cho mỗi workflow (`curriculum-required-gate`, `mission-required-gate`) với `needs` toàn bộ job và `if: always()`; gate chỉ pass khi mọi job bắt buộc đạt hoặc skip đúng policy được kiểm.
- Required checks trên `main` trỏ tới hai final gate ổn định, kèm policy PR/review/conversation/force-push/delete.
- Xác minh cấu hình bằng quyền admin và lưu evidence; không suy từ file workflow.
- Đóng hoặc cập nhật issue #2 vì M01/M02 đã tồn tại và tên check đã lỗi thời.

### FR-03 — P1 — Cache Go không tìm được dependency file ở hai job

**Bằng chứng**

- Live CI báo `Restore cache failed: Dependencies file is not found ... Supported file pattern: go.mod` tại `mission-runtime` và `n8n-engine-regression`.
- Hai bước `actions/setup-go@v7` trong [mission-agent-path-ci.yml](../../.github/workflows/mission-agent-path-ci.yml) chưa khai báo `cache-dependency-path`.

**Rủi ro**

- Cache không hoạt động, tốn thời gian/băng thông và che warning thật trong CI.

**Chỉnh sửa đề xuất**

- `mission-runtime`: khai báo `core/go.sum`, `contracts/go.sum`, `lab/mission-runtime/go.sum`.
- `n8n-engine-regression`: khai báo `lab/affiliate-bot/go.sum`, `core/go.sum`, `contracts/go.sum` và chỉ thêm module khác nếu binary thật sự phụ thuộc.
- Thêm assertion/log ngắn cho cache key, nhưng cache miss không được làm thay đổi correctness.

### FR-04 — P1 — PR path gate của n8n có false-negative dependency

**Bằng chứng**

- Regex `Select n8n engine coverage` chỉ liệt kê một số file adapter và không bao phủ `go.mod/go.sum`, toàn bộ `core/`, `contracts/`, hoặc mọi helper có thể thay binary.
- Main push luôn chạy full engine, nhưng PR không liên quan theo regex sẽ báo skip thành công.

**Rủi ro**

- Thay dependency/contract có thể merge mà engine regression chỉ phát hiện sau merge trên `main`.

**Chỉnh sửa đề xuất**

- Mở rộng gate tối thiểu tới `lab/affiliate-bot/**`, `core/**`, `contracts/**`, `lab/n8n/**`, hai engine runner và workflow này; sau khi đo chi phí có thể thu hẹp bằng dependency map được test.
- Thêm unit test cho path selector với case `go.mod`, `go.sum`, schema, shared core và helper mới.
- Final gate phải phân biệt `skip hợp lệ` với `runner không được kích hoạt do regex lỗi`.

### FR-05 — P1 — HTTP request-size ceiling có thể bị bypass bằng truncation

**Bằng chứng**

- [watcher.go](../../lab/affiliate-bot/cmd/bot/watcher.go) tại `decodeM06AdapterRequest`, `decodeM07AdapterRequest` và `historyHandoffHTTPHandler` dùng `io.ReadAll(io.LimitReader(body, limit))`.
- Code không đọc `limit+1`, không so độ dài và không dùng `http.MaxBytesReader`.
- Nếu JSON hợp lệ kết thúc trước/trùng biên và phần vượt trần là whitespace/padding, parser chỉ thấy phần bị cắt và có thể chấp nhận request lớn hơn contract.

**Rủi ro**

- Size guard không đúng nghĩa; local process có thể tiêu tốn network/read budget hoặc đưa payload vượt policy qua adapter.
- Ba endpoint có thể khác hành vi với `m07ToolResponseBody`, nơi đã dùng đúng `limit+1` và kiểm `len`.

**Chỉnh sửa đề xuất**

- Viết regression trước: payload JSON hợp lệ + padding làm tổng kích thước `limit+1` phải trả `413`/status riêng và không mutate history/artifact.
- Tạo shared helper đọc body `limit+1`, reject trước decode; hoặc dùng `http.MaxBytesReader` và map `MaxBytesError` ổn định.
- Áp dụng cùng contract cho M06, M07 và history handoff; giữ duplicate-key/trailing-value strict decode hiện có.
- Kiểm Content-Type nếu endpoint chỉ nhận JSON; không log raw body.

### FR-06 — P1 trước deployment — Canonical adapter chưa có process-to-process authentication

**Bằng chứng**

- `watcher serve` buộc bind `127.0.0.1`, nhưng các endpoint đọc/append/register không kiểm credential.
- Loopback ngăn truy cập trực tiếp từ mạng ngoài, không ngăn process/user/container khác trên cùng host.

**Rủi ro**

- Local process có thể đọc canonical history hoặc gửi request làm thay đổi artifact store.
- Khi đặt n8n và adapter trên cùng host/container network, “loopback” dễ bị hiểu nhầm là authorization.

**Chỉnh sửa đề xuất**

- Chọn một profile cross-platform rõ ràng: bearer secret ngẫu nhiên qua environment/secret store, hoặc local IPC có ACL; không commit secret vào blueprint.
- So sánh token constant-time, không log token, hỗ trợ rotation và fail startup nếu profile operated thiếu secret.
- Có thể để `/healthz` không auth nhưng tuyệt đối không trả data nhạy cảm; mọi endpoint đọc/ghi phải auth.
- Regression: thiếu/sai token trả 401/403, byte-identical state; token đúng giữ behavior cũ; n8n workflow lấy token từ credential store.

### FR-07 — P1 — Runbook deployment khai báo sai Go minimum

**Bằng chứng**

- [BR-18A-DEPLOYMENT-RECOVERY-RUNBOOK.md](../architecture/BR-18A-DEPLOYMENT-RECOVERY-RUNBOOK.md) ghi Go `1.23+`.
- [lab/affiliate-bot/go.mod](../../lab/affiliate-bot/go.mod) khai báo `go 1.27`; [QUICKSTART.md](../../curriculum/BOOT/QUICKSTART.md) đã dùng 1.27.

**Rủi ro**

- Clean host làm theo runbook có thể dừng ngay ở build hoặc tự hạ `go.mod` sai cách.

**Chỉnh sửa đề xuất**

- Runbook phải nói “dùng version theo `lab/affiliate-bot/go.mod` (hiện 1.27)”.
- Thêm validator so các hướng dẫn vận hành với directive thật để không drift lần nữa.
- Giữ rõ các module `core/contracts/mission-runtime` có minimum riêng; không suy 1.23 là đủ để build learner Bot.

### FR-08 — P1 theo lịch — `ubuntu-latest` sắp đổi sang Ubuntu 26

**Bằng chứng**

- Live CI của commit hiện hành báo `ubuntu-latest` bắt đầu migrate sang Ubuntu 26 từ 19/10/2026.
- Repo có native dependency n8n/`isolated-vm`, POSIX openat/process-kill và race tests nhạy với runner image.

**Chỉnh sửa đề xuất**

- Pin acceptance baseline vào `ubuntu-24.04` trước mốc migration.
- Tạo compatibility job tạm thời cho Ubuntu 26 khi image sẵn sàng; chỉ chuyển baseline sau khi Go/race/n8n/mutation đều pass.
- Ghi rõ image vào evidence; không dùng “ubuntu-latest PASS” như bằng chứng ổn định qua migration.

### FR-09 — P1/P2 — `main_baseline` chưa có semantics và invariant Git đủ mạnh

**Bằng chứng**

- Plan/matrix/graph cùng trỏ `ec2499a51dee…`; `HEAD` là `2ec78f1a40c…`.
- Diff từ baseline tới HEAD chỉ là docs/audit/tests, không đổi learner runtime, nhưng tên field là `main_baseline` và prose có chỗ gọi đây là current main.
- [audit_readiness.py](../../scripts/audit_readiness.py) hiện chỉ kiểm format và equality giữa ba artifact; không kiểm commit tồn tại, là ancestor, hoặc diff thuộc allowlist.

**Rủi ro**

- Người đọc không biết baseline là product code, audit snapshot hay exact HEAD.
- Ba file có thể cùng trỏ một SHA hợp lệ về format nhưng không liên quan checkout.

**Chỉnh sửa đề xuất**

- Quyết định một trong hai contract:
  - `snapshot_head`: luôn bằng exact reviewed HEAD; hoặc
  - `product_baseline`: có thể cũ hơn HEAD nhưng audit bắt buộc `is-ancestor` và diff chỉ thuộc nhóm docs/audit/test đã khai báo.
- Đổi tên field/prose theo contract được chọn; migration cả plan/matrix/graph/audit/test trong cùng PR.
- Negative tests: missing commit, non-ancestor, product file đổi sau baseline nhưng không bump, shallow checkout không đủ history phải fail với thông báo rõ.

### FR-10 — P2 — File quá lớn và trạng thái lặp thủ công

**Bằng chứng**

- `REVIEW-REMEDIATION-PLAN.md`: 3.718 dòng / khoảng 285 KB.
- `READINESS-MATRIX.json`: khoảng 227 KB; `READINESS-EVIDENCE-GRAPH.json`: khoảng 153 KB.
- `audit_readiness.py`: 2.622 dòng; `mission_command.go`: 3.242 dòng/99 hàm; `backup_command.go`: 1.894 dòng.
- Kết luận/current status/history được lặp giữa Markdown, hai JSON, evidence files và assertions Python.

**Rủi ro**

- Review khó, merge conflict cao, thêm claim dễ quên fixture hoặc cập nhật một trong ba nguồn.
- Một file command lớn làm ownership/fault-boundary khó quan sát dù test hiện mạnh.

**Chỉnh sửa đề xuất**

- Readiness: tách “current snapshot” nhỏ khỏi append-only history; sinh summary Markdown từ data canonical; dùng JSON Schema + canonical formatter.
- Audit: tách package theo snapshot, runtime boundary, evidence graph, CI wiring; giữ CLI/output tương thích.
- Go: refactor theo bounded component (`m10 journal`, `m11 journal`, `artifact IO`, `command dispatch`) bằng move-only PR nhỏ; test baseline trước/sau, không đổi schema/status/output.
- Không làm refactor này cùng PR với security/behavior fix FR-05/FR-06.

### FR-11 — P2 — Repo hygiene và security maintenance còn thiếu

**Bằng chứng**

- Chạy Python tests tạo `scripts/**/__pycache__`; `.gitignore` chưa ignore `__pycache__/`, `*.py[cod]`, `.pytest_cache/`.
- Không có `dependabot.yml`, `SECURITY.md`, CodeQL/govulncheck job, `.editorconfig` hoặc `.gitattributes`.
- Actions dùng major tags thay vì full SHA.

**Chỉnh sửa đề xuất**

- Bổ sung ignore Python cache và test sạch working tree sau validators.
- Thêm Dependabot cho GitHub Actions và bốn Go module; giới hạn lịch/PR concurrency.
- Thêm `govulncheck` và CodeQL ở scheduled/non-blocking trước; sau giai đoạn ổn định mới cân nhắc required.
- Thêm `SECURITY.md`; pin actions theo full SHA và có quy trình update.
- `.editorconfig`/`.gitattributes` khóa UTF-8/LF cho source/docs, tránh diff do Windows newline.

### FR-12 — P3 — Metadata cộng tác/phân phối chưa quyết định

**Bằng chứng**

- Repo public nhưng chưa có `LICENSE`, `CONTRIBUTING.md`, `CODEOWNERS` hoặc code-of-conduct.

**Chỉnh sửa đề xuất**

- Chủ repo quyết định license trước khi khuyến khích reuse/contribution.
- Nếu vẫn là repo cá nhân: tối thiểu ghi rõ contribution policy và owner/reviewer của runtime, contracts, readiness evidence.
- Đây là quyết định quản trị, không tự chọn license thay chủ repo.

## 4. Kế hoạch triển khai theo PR và dependency

### PR-A — CI runtime khẩn cấp

Phạm vi: FR-01, FR-03, FR-04, FR-08 phần pin runner, timeout/concurrency cơ bản.

Thứ tự:

1. Chụp baseline 13 job trên exact HEAD.
2. Update `actions/cache`, khai báo Go cache paths.
3. Pin `ubuntu-24.04`; thêm kế hoạch/job compatibility Ubuntu 26.
4. Mở rộng n8n path gate và thêm test selector.
5. Thêm `timeout-minutes` hữu hạn cho mọi job; thêm concurrency cancel cho PR superseded.

Nghiệm thu:

- Hai workflow pass toàn bộ.
- Không còn warning Node 20 và “Dependencies file is not found”.
- PR chỉ đổi `contracts/go.mod`, `core/**`, `lab/affiliate-bot/go.sum` hoặc helper adapter phải kích hoạt n8n engine.
- Main push vẫn luôn chạy n8n engine.

### PR-B — Required gate và governance

Phạm vi: FR-02, phần repo của FR-11.

Thứ tự:

1. Thêm final gate cho từng workflow.
2. Update governance bằng exact stable check names.
3. Dùng quyền admin bật/verify branch protection; lưu evidence riêng.
4. Triage issue #2: đóng hoặc thay bằng issue hiện hành có link plan này.

Nghiệm thu:

- Mutation tạm làm một job fail thì final gate fail.
- Job skip hợp lệ vẫn có kết quả xác định; job bị cancel/fail không được coi pass.
- Evidence admin chứng minh PR required, status required, force-push/delete blocked theo policy được duyệt.

### PR-C — HTTP adapter size boundary

Phạm vi: chỉ FR-05; không gộp authentication để diff nhỏ và regression rõ.

Thứ tự test-first:

1. Chạy baseline watcher/history tests hiện có.
2. Thêm ba regression fail trên code cũ: M06, M07, history handoff nhận valid JSON + padding vượt một byte.
3. Implement shared limited-body reader.
4. Rerun tests, n8n static/engine, history replay và byte-identical no-mutation cases.

Nghiệm thu:

- Exact limit hợp lệ được xử lý theo contract.
- `limit+1` bị reject trước parse/persist.
- Duplicate key, trailing JSON value, large-number precision và ACK semantics không đổi.

### PR-D — Authentication canonical adapter

Phạm vi: FR-06.

Điều kiện trước: chủ repo chọn profile token/IPC và nơi giữ secret; không dùng credential production trong test.

Nghiệm thu:

- Missing/wrong/rotated credential fail closed và không mutate/read data.
- Blueprint dùng n8n credential store; export workflow không chứa secret.
- Loopback vẫn bắt buộc ở profile local.
- Engine regression có unauthorized/authorized case qua adapter thật.

### PR-E — Toolchain/runbook/hygiene

Phạm vi: FR-07, phần nhẹ của FR-11, FR-12 nếu chủ repo đã quyết định.

Nghiệm thu:

- Clean clone dùng Go version theo `lab/affiliate-bot/go.mod` build được runbook path.
- Validator bắt lỗi nếu runbook/quickstart quay lại minimum thấp hơn module.
- Chạy Python suite không làm working tree xuất hiện `__pycache__`.

### PR-F — Readiness metadata contract

Phạm vi: FR-09.

Dependency: sau PR-A/B để snapshot CI ổn định.

Nghiệm thu:

- Baseline semantics có tên và định nghĩa duy nhất.
- Audit bắt non-existent/non-ancestor/unauthorized-diff cases.
- Plan/matrix/graph/test fixture được migrate đồng bộ.
- Readiness vẫn `NOT_READY_FOR_PRODUCTION` trừ khi có evidence mới độc lập.

### PR-G — Maintainability split

Phạm vi: FR-10, phần còn lại FR-11.

Nguyên tắc:

- Một PR chỉ move/split một boundary; không đổi output/status/schema.
- Trước mỗi functional refactor phải thêm/chạy baseline regression, rồi chạy lại sau sửa.
- Không xóa historical evidence; chuyển sang archive/index có mapping và audit migration.

## 5. Lệnh nghiệm thu chuẩn

Chạy từ repo root; CI/Linux dùng shell phù hợp. Máy review Windows hiện không có Go nên các lệnh Go phải chạy ở CI hoặc môi trường có đúng toolchain.

```text
python -m unittest discover -s scripts/tests -v
python scripts/validate_missions.py
python scripts/validate_repo.py
python scripts/validate_artifact_spine.py
python scripts/validate_continuity.py
python scripts/validate_language_policy.py
python scripts/validate_agent_semantics.py
python scripts/validate_semantic_contracts.py
python scripts/validate_m11.py
python scripts/audit_readiness.py .
git diff --check
```

Với Go:

```text
contracts: go test ./... && go vet ./...
core: GOWORK=off go test ./... && GOWORK=off go vet ./...
lab/mission-runtime: go test ./... && go vet ./...
lab/affiliate-bot: go test ./... && go vet ./...
lab/affiliate-bot: go test -race ./...
```

Với n8n/engine:

```text
python scripts/run_n8n_engine_regression.py --n8n-cli <pinned-n8n-cli>
python scripts/run_n8n_m06_schedule_regression.py --n8n-cli <pinned-n8n-cli>
```

## 6. Các blocker không được “sửa bằng tài liệu”

Các mục sau tiếp tục nằm ngoài bằng chứng local/offline và giữ `RP-10=OPEN`:

- selected-source/provider operated run với account/credential được cấp quyền;
- live executor và business outcome thật;
- clean-machine beginner pilot có người ghi nhận độc lập;
- target-host deployment/restart/backup/recovery drill;
- power-loss/filesystem crash tại seam nhiều file;
- distributed/multi-host locking và transaction;
- reviewer độc lập xác nhận evidence và promotion/recovery authority.

Không đổi chúng thành DONE chỉ vì CI xanh, có fixture, có checklist hoặc có file evidence mới.

## 7. Thứ tự ưu tiên đề nghị

```text
Ngay lập tức: PR-A
→ tạo merge gate: PR-B
→ sửa input boundary: PR-C
→ chọn và triển khai local authentication: PR-D
→ sửa runbook/hygiene: PR-E
→ làm rõ readiness baseline: PR-F
→ refactor maintainability: PR-G
→ chỉ sau đó mới chạy RP-10 operated/deployment khi có authority và môi trường
```

## 8. Quyết định cần chủ repo chốt trước khi triển khai

1. Canonical adapter dùng bearer token cross-platform hay IPC/ACL theo OS.
2. Branch protection require hai final gate hay require từng job riêng; đề nghị final gate để tên ổn định.
3. `main_baseline` sẽ là exact snapshot HEAD hay product baseline có allowlist.
4. Có công bố license/contribution hay tiếp tục repo cá nhân không nhận contribution.
5. Có sẵn target host/provider/pilot cho RP-10 hay chỉ triển khai các PR offline A–G.

## 9. Kết quả thực hiện tự động — 2026-09-19

| Hạng mục | Kết quả | Ghi chú |
|---|---|---|
| FR-01/03/04/08 | `DONE_OFFLINE` | CI action SHA, Ubuntu 24.04, cache paths, timeout/concurrency, n8n path gate và cache v5. |
| FR-02 | `DONE_REPO` | Thêm `curriculum-gate` và `mission-gate`; bật branch protection thật vẫn cần admin. |
| FR-05/06 | `DONE_OFFLINE` | Body `limit+1`/413, bearer token bắt buộc, test missing/wrong/correct và caller/blueprint wiring. |
| FR-07/11 | `DONE_OFFLINE` | Runbook Go 1.27, hygiene files, Dependabot, CodeQL/govulncheck workflow, ignore/generated-file policy. |
| FR-09 | `DONE_REPO` | Chọn `product` baseline; audit kiểm commit tồn tại, ancestor, working tree sạch và docs-only drift. |
| FR-10 | `DEFERRED` | Move-only refactor lớn chưa gộp cùng security fix; cần PR riêng có Go baseline và review ownership. |
| FR-12 | `BLOCKED_EXTERNAL` | Chưa tự chọn license/legal policy thay chủ repo. |

### Kiểm chứng cuối nhánh

- `python -m unittest discover -s scripts/tests -q`: PASS (157 tests).
- `python scripts/audit_readiness.py .`: PASS, 151 scoped claims,
  `NOT_READY_FOR_PRODUCTION`.
- JSON parse, Python compile, static n8n validators và `git diff --check`: PASS.
- Go và real n8n: chưa chạy vì host không có executable; hosted CI là gate tiếp theo.
