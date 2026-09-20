# Review toàn repo và kế hoạch chỉnh sửa — 19/09/2026

## 1. Kết luận

**Đối chiếu mới nhất (21/09/2026):** xem [checklist follow-up từng RV và phần còn phải làm](FULL-REPO-REVIEW-FOLLOWUP-20260921.md).
Đã sửa thêm lỗi fixture chronology và n8n scope fail-open, bổ sung nghiệm thu
còn thiếu, self-review và merge PR #491 tại `b447c04`. Hosted exact head PASS
Go/Windows/race, full offline runner và real n8n/flatted/Schedule Trigger.
RV-09 còn admin settings; RP-10 và các external gaps vẫn mở. Baseline `ada7827`
failure và các kết quả PR #489 bên dưới được giữ như lịch sử, không thay evidence mới.

Repo có nền tảng kiểm thử khá đầy đủ và đường chạy offline liên tục đã hoạt động. Trong phiên review này, cả bốn Go module qua test/vet; learner Bot qua race detector; 133 test Python, 14 validator không cần execution artifact, 11 smoke offline, quickstart clone/cache rỗng và HTTPS fixture smoke đều đạt.

Tuy nhiên, đã xác nhận **11 nhóm vấn đề: 2 P1, 8 P2 và 1 P3**. Hai việc cần xử lý trước là:

1. `mission bind` có thể xóa lịch sử tiêu thụ hạn mức M10, cho dùng lại cùng grant/authorization và làm hỏng liên kết backup.
2. Hai n8n regression runner kế thừa cấu hình database từ shell; runtime mang tên “disposable” vẫn có thể nhắm vào database có sẵn.

Các việc tiếp theo gồm độ chính xác và tính đầy đủ của M11 historical audit, giới hạn HTTP input, bằng chứng restart của n8n, phạm vi CI và tài liệu hướng dẫn hiện hành. Bộ test đang xanh chưa bao phủ các trường hợp mới được tái hiện dưới đây.

Trạng thái tổng thể vẫn là **`NOT_READY_FOR_PRODUCTION`**. Báo cáo này là kết quả review và kế hoạch thực hiện; các đầu việc chưa được sửa chỉ vì đã được ghi lại ở đây.

## 2. Baseline, phạm vi và cách kiểm

- Ngày review: **19/09/2026**, múi giờ Asia/Ho_Chi_Minh.
- Branch: `main`.
- HEAD: `707ac76aba0c1aac705ec3e52bcbd45e71edfb85` — `ci: run learner bot runtime tests on Windows (#415)`.
- Môi trường: macOS arm64, Go `1.27.0`, Python `3.9.6`; CI khai báo Python `3.12`.
- Inventory đầu phiên: **600 file tracked**, gồm 195 file Go, 55 file Python, 81 file JSON, 257 file Markdown, 4 Go module và 2 workflow CI. Mã Go có khoảng 22.997 dòng implementation và 21.221 dòng test.
- Review dựa trên working tree hiện hành, bao gồm các thay đổi đã có trước phiên; không gọi working tree này là một checkout sạch của HEAD.

Những file đã có thay đổi trước khi review và được giữ nguyên:

```text
M  docs/plans/READINESS-EVIDENCE-GRAPH.json
M  docs/plans/READINESS-MATRIX.json
M  docs/plans/REVIEW-REMEDIATION-PLAN.md
M  scripts/tests/test_audit_readiness.py
?? docs/architecture/EVIDENCE-PR415-POST-MERGE-20260918.md
```

| Phần repo | Nội dung đã rà | Bằng chứng kiểm tra trong phiên |
|---|---|---|
| `contracts/`, `core/` | Schema/decoder, identity/hash, M00–M11 semantics, approval, historical chain, registry graph | Test/vet; đầu dò M11 trên module sao chép |
| `lab/affiliate-bot/` | CLI, bind/approval/grant/reservation, M11 lifecycle, HTTP adapter, provider/fetch boundary, store/backup/recovery, primitive filesystem | Full test/vet/race; smoke; đầu dò M10 trên Go overlay; HTTP server loopback thật |
| `lab/mission-runtime/`, `evals/` | Vai trò oracle, chuyển đổi dữ liệu, chain audit, fixture và semantic checks | Full test/vet; bốn đầu dò M11 có baseline đối chứng |
| `lab/n8n/`, `scripts/` | Blueprint, environment isolation, restart proof, validators, readiness audit | Python tests/validators; mock child environment, SQLite tạm, thử trực tiếp regex CI |
| `.github/`, governance | Phân chia job, path filter, required-check hướng dẫn | Đối chiếu các job hiện có với tài liệu; không truy cập cấu hình branch protection thực tế |
| README, curriculum, missions, starter-kits, product/architecture/plans | Thứ tự học, authority, onboarding, mô tả capability và hướng dẫn chạy | Structural/language/continuity validators; quickstart; quét link Markdown nội bộ không thấy đích thiếu |

Đây là review theo toàn bộ subsystem, kết hợp đọc mã và chạy kiểm tra; không phải chứng minh mọi nhánh hoặc mọi tổ hợp input. Các số dòng bên dưới gắn với baseline này.

### Kết quả chạy lại

| Kiểm tra | Kết quả |
|---|---|
| `go test -count=1 ./...` và `go vet ./...` trong `contracts`, `core`, `lab/mission-runtime`, `lab/affiliate-bot` | PASS cả 4 module |
| `go test -race -count=1 ./...` trong `lab/affiliate-bot` | PASS, khoảng 97 giây |
| `python3 -m unittest discover -s scripts/tests -v` | PASS, 133 tests |
| 14 `validate*.py` không cần execution artifact | PASS |
| `python3 scripts/audit_readiness.py` | Exit 0; `NOT_READY_FOR_PRODUCTION`; resolve 100 scoped claims |
| BR-08, BR-09, BR-10a/b/c/d, BR-11a, BR-12d, BR-13b, BR-16a, BR-18b | PASS, 11 smoke offline |
| `python3 scripts/smoke_quickstart.py` | PASS: clone/cache rỗng, run/test, intentional FAIL, sửa lại, O00 |
| `python3 scripts/smoke_br13c.py` | PASS: pinned HTTPS fixture |
| `git diff --check` | PASS |
| Đầu dò M11 mới | Baseline PASS; 4 assertion mong muốn FAIL, xác nhận các thiếu sót RV-03/04/05 |
| Đầu dò M10 bind và HTTP malformed input | Tái hiện thành công RV-01/RV-06 |

Không chạy lại n8n engine thật vì `node`/`n8n` không có trên PATH; không chạy native Windows/Linux, live provider/executor, target deployment, kiểm thử mất điện hoặc toàn bộ mutation scripts. Hai validator `*_operated_execution.py` cần execution JSON/history riêng, không thể chạy không tham số như validator tĩnh. Không coi log CI cũ là lần chạy CI mới của phiên review này.

Log và đầu dò của phiên nằm ngoài repo tại `/tmp/bot-review-20260919/` và các thư mục tạm liên quan. Các thư mục này có thể bị hệ điều hành dọn; phần mô tả tái hiện và tiêu chí nghiệm thu trong báo cáo là thông tin cần giữ lâu dài. Không thêm test tạm hay thay đổi implementation vào repo trong phiên review.

## 3. Danh sách ưu tiên

P1: sửa trước khi tiếp tục sử dụng đường chức năng liên quan. P2: sửa trong đợt hardening tiếp theo vì ảnh hưởng correctness hoặc độ tin cậy kiểm chứng. P3: lỗi phạm vi hẹp, có thể xử lý cùng phần liên quan.

| ID | Mức | Vấn đề | Gói cũ liên quan |
|---|---|---|---|
| RV-01 | P1 | Rebind intent làm mất budget/reservation M10 và tạo execution mồ côi | RP-03, RP-06 |
| RV-02 | P1 | n8n regression kế thừa cấu hình database/queue bên ngoài | RP-08 |
| RV-03 | P2 | Historical audit M11 thiếu exact intent binding của cost/authorization | RP-07 |
| RV-04 | P2 | Adapter M11 đổi `json.Number` thành `float64`, làm sai hash | RP-02, RP-07 |
| RV-05 | P2 | Historical audit M11 không ràng buộc authorization expiry với health TTL | RP-07 |
| RV-06 | P2 | HTTP input bị cắt tại limit rồi vẫn được chấp nhận | RP-01, RP-04, RP-05 |
| RV-07 | P2 | Schedule restart regression có thể dùng execution trước restart | RP-08 |
| RV-08 | P2 | PR path filter bỏ qua dependency trực tiếp của n8n integration | RP-08 |
| RV-09 | P2 | Hướng dẫn required checks thiếu các job hiện đang giữ test quan trọng | RP-08, RP-09 |
| RV-10 | P2 | Runbook kiểm tra không chạy trọn; mô tả capability còn ở baseline cũ | RP-09 |
| RV-11 | P3 | Decoder flatted giải mã sai string chỉ chứa chữ số | RP-08 |

Các ID này bổ sung bằng chứng cho tracker RP hiện hữu, không tự đóng hoặc thay trạng thái các RP/BR. Ở thời điểm review ban đầu tất cả là **TODO**; trạng thái cập nhật bên dưới phân biệt implementation/regression offline với nghiệm thu hosted và evidence ngoài repo.

### RV-01 — `bind` xóa consumption history của M10

**Vị trí:** `lab/affiliate-bot/cmd/bot/mission_command.go:2365–2369`, `2437–2452`.

Khi chuyển sang intent khác, `bind` đặt `Approval`, `Canary`, `Lease`, `Reservations` về `nil`. Import lại grant chỉ lấy bộ đếm từ `s.Canary` đang tồn tại; sau rebind, cùng grant được khởi tạo lại với usage bằng 0.

**Tái hiện đã chạy:** tạo grant `max_executions_total=1`; authorize, reserve và ghi một execution `FAILED/NOT_PERFORMED`. Chạy `bind A → B → A`, nhập lại đúng approval/grant cũ, rồi reserve cùng authorization với reservation ID mới. Kết quả:

```text
first execution -> APPENDED
bind -> BOUND
bind -> BOUND
m09-approval -> ACK
m10-canary -> ACK
m10-reserve-authorization -> RESERVED
backup graph: M10 execution record is orphaned from restored reservation
```

State chỉ giữ reservation thứ hai và usage `1`, dù đã có hai lần reserve. Execution cũ còn trong immutable registry nhưng mất cạnh nối tới reservation. Nếu reservation đầu chưa có execution, kiểm graph hiện tại còn chấp nhận việc mất lịch sử đó.

**Tác động:** tái sử dụng hạn mức đã tiêu thụ và làm hỏng graph bằng chính chuỗi lệnh CLI hợp lệ. Đây là lỗi runtime offline có bằng chứng; không cần giả định có live executor. Yêu cầu giữ consumed history qua `bind` đã có trong RP-03a nhưng chưa được thực thi đủ.

**Hướng sửa:** chặn đổi binding khi đã có authority/consumption chưa được quản lý an toàn; sau đó tách ledger theo grant/authorization khỏi con trỏ current intent. Kiểm identity trên lịch sử đã lưu, không chỉ trên intent đang active. Với dữ liệu đã mất reservation, phải báo tình trạng cần đối soát; không tự dựng lại usage bằng 0.

**Nghiệm thu:** A→B→A cho pending reservation và FAILED execution không cấp thêm hạn mức; cùng authorization không tạo reservation thứ hai ngoài chính sách; restart/backup/restore giữ nguyên history và graph; rejection không sửa file trạng thái. Test phải chạy qua learner command thật, không chỉ hàm tính budget.

### RV-02 — Runtime n8n tạm chưa cô lập cấu hình database

**Vị trí:** `scripts/run_n8n_engine_regression.py:440–452`; `scripts/run_n8n_m06_schedule_regression.py:215–216`.

Hai runner dùng `dict(os.environ)` rồi thay `N8N_USER_FOLDER`, nhưng giữ `DB_TYPE`, `DB_POSTGRESDB_*`, biến dạng `_FILE` và cấu hình queue từ shell. Trong môi trường đã cấu hình PostgreSQL, lệnh import workflow/credential có thể nhắm vào database có sẵn trước khi runner kiểm SQLite tạm. `N8N_USER_FOLDER` không thay thế lựa chọn database qua `DB_TYPE`. [Nguồn: tài liệu cấu hình database chính thức của n8n](https://github.com/n8n-io/n8n-docs/blob/main/docs/deploy/host-n8n/configure-n8n/basic-configuration/use-environment-variables/database.md).

**Tái hiện an toàn:** chặn `run()` trước subprocess, truyền các giá trị giả `DB_TYPE=postgresdb`, `DB_POSTGRESDB_HOST=existing.example.invalid`, `EXECUTIONS_MODE=queue`; cả `main()` và `run_case()` giữ nguyên chúng trong child environment. Không kết nối database ngoài. Việc có thể ghi vào DB ngoài là hệ quả của environment này và cách n8n chọn backend, chưa phải một thử nghiệm ghi vào DB thật.

**Hướng sửa:** dùng chung hàm tạo environment cô lập; lọc cấu hình n8n/database/queue kế thừa, kể cả `_FILE`; pin SQLite, regular execution mode và kiểm đường database nằm dưới temp root trước import. Chỉ giữ các biến hệ thống/build thực sự cần.

**Nghiệm thu:** fake CLI ghi lại environment trong các ca parent env chứa database/queue/config-file giả; không giá trị bên ngoài nào chi phối backend. Sau đó chạy cả hai runner trên n8n đã pin và chỉ thấy file/state trong runtime tạm.

### RV-03 — Historical chain M11 thiếu hai cạnh exact binding

**Vị trí:** `core/m11/historical_chain.go:216–240`.

Hai đầu dò đã được chạy:

- Đổi `cost.intent_id` và `cost.intent_hash` sang intent khác, tính lại cost hash và cập nhật các hash tham chiếu: audit vẫn trả `VALID / CONSISTENT_UNVERIFIED`.
- Với profile `resolved_stop`, đổi riêng `authorization.intent_hash`: audit vẫn trả `VALID`.

Nhánh chung kiểm correlation và cost artifact refs, nhưng thiếu so sánh `Cost.IntentID/IntentHash` với intent; có so sánh `Authorization.IntentID` nhưng thiếu `Authorization.IntentHash`. Nhánh closed-cycle có thêm kiểm tra riêng nên che khuất trường hợp resolved-stop.

**Tác động:** kết luận tính nhất quán của audit lịch sử bị sai. Learner registry/gate có một số kiểm tra binding riêng, vì vậy không suy finding này thành bypass cấp quyền thực thi.

**Hướng sửa và nghiệm thu:** đưa các so sánh exact ID/hash vào nhánh chung trước phân profile. Mutate từng trường riêng, rehash hợp lệ, chạy cả `closed_cycle` và `resolved_stop`; tất cả mismatch phải bị từ chối, baseline phải giữ `CONSISTENT_UNVERIFIED` và `execution_permitted=false`.

### RV-04 — M11 chain adapter làm tròn số hợp lệ

**Vị trí:** `lab/mission-runtime/cmd/demo/m11_chain.go:121–136`.

M08 decoder đã giữ số dưới dạng `json.Number`, nhưng helper `toCore` Marshal rồi Unmarshal vào struct có `map[string]any`, đổi số thành `float64`. Với `parameters.large_integer=9007199254740993`, giá trị thành `9007199254740992` và audit trả `TAMPERED_INTENT` dù input/hash ban đầu hợp lệ.

**Đối chứng đã chạy:** chỉ thay phép chuyển intent bằng helper có sẵn `coreIntent(i)` trên bản sao, cùng fixture chuyển từ FAIL sang `VALID`. Không sửa repo để tạo đối chứng.

**Hướng sửa và nghiệm thu:** chuyển kiểu trực tiếp hoặc decoder giữ `UseNumber`; tránh vòng JSON trung gian cho các kiểu chứa `any`. Thêm fixture số trên `2^53`, decimal và exponent theo canonical contract đang dùng; giữ nguyên dữ liệu và hash qua decode → adapter → shared audit.

### RV-05 — Audit M11 bỏ sót giới hạn tuổi health khi kiểm authorization

**Vị trí:** `core/m11/historical_chain.go:288–294`; đối chiếu issuer tại `lab/mission-runtime/cmd/demo/m11.go:699–705`.

Audit kiểm health còn mới tại gate và authorization expiry không vượt lease/intent/cost, nhưng thiếu giới hạn theo health snapshot. Đầu dò dùng health lúc `07:59`, TTL 300 giây, authorization expiry `09:00`, execution `08:07`; cập nhật `post_ledger.last_execution_at` tương ứng. Audit vẫn trả `VALID`, dù health chỉ bảo đảm tới `08:04` và issuer hiện tại sẽ giới hạn authorization theo mốc đó.

**Hướng sửa:** đối chiếu authorization expiry với `health.observed_at + TTL`, xử lý phép tính thời gian an toàn. Bổ sung regression cho registry graph; vị trí `core/m11/artifact_registry.go:280–285` cũng cần kiểm tương ứng, nhưng chưa có đầu dò độc lập cho graph nên không tính là finding thứ hai.

**Nghiệm thu:** test đúng biên, vượt biên 1 ns, execution sau health expiry; historical audit giữ cùng invariant với issuer, không phụ thuộc đồng hồ hiện tại để đánh giá hồ sơ quá khứ.

### RV-06 — HTTP decoder chấp nhận phần đầu của request quá lớn

**Vị trí:** `lab/affiliate-bot/cmd/bot/watcher.go:184`, `368`, `787`.

Ba đường đọc body dùng `io.LimitReader` với đúng limit rồi decode; không đọc thêm một byte để phát hiện vượt limit. JSON hợp lệ cộng whitespace cho đủ limit sẽ được decode thành công, còn trailing data sau limit bị bỏ qua.

**Đã tái hiện trên server loopback thật:** body M06 gồm JSON fixture hợp lệ, padding tới 65.536 byte, rồi thêm `{"unexpected_second_json":true}`. Tổng 65.567 byte; server trả:

```json
{"http":200,"status":"APPENDED","canonical_history_ack":true,"history_exists":true}
```

**Tác động:** nhận request malformed/oversized và ghi canonical history, trái ranh giới strict input. Chưa cần giả định input này cấp được execution permission.

**Hướng sửa:** helper đọc body giới hạn dùng `http.MaxBytesReader` hoặc đọc `limit+1` và reject trước decode/append. Áp dụng nhất quán cho M06, M07 và history handoff.

**Nghiệm thu:** limit−1/limit/limit+1, `Content-Length` và chunked body, JSON hợp lệ + padding + JSON thứ hai; reject không tạo history, sidecar hoặc ACK. Input hợp lệ trong hạn vẫn giữ nguyên hành vi.

### RV-07 — Schedule restart regression có thể dùng tick cũ

**Vị trí:** `scripts/run_n8n_m06_schedule_regression.py:241–247`; query tại dòng 132.

Sau khi lấy `duplicate_id`, workflow có thể tạo thêm tick trước khi process dừng. Sau restart, query chỉ yêu cầu `id > duplicate_id`; vì vậy tick trước shutdown được coi là bằng chứng hoạt động sau restart.

**Tái hiện:** SQLite tạm có sẵn execution success ID 3; gọi `wait_for_execution(..., after_id=2)` với server giả còn sống nhưng không tạo execution mới, hàm trả ngay ID 3. Đây là đầu dò selection logic, không phải một lần chạy n8n engine mới.

**Hướng sửa và nghiệm thu:** sau khi process cũ dừng hoàn toàn, lấy maximum execution ID làm mốc; process mới phải tạo ID lớn hơn mốc này. Ca có tick dư trước shutdown nhưng không có tick mới phải timeout; ca restart thật phải sinh execution mới và giữ đúng canonical record.

### RV-08 — PR filter thiếu dependency của n8n integration

**Vị trí:** `.github/workflows/mission-agent-path-ci.yml:100`.

Chạy chính regex hiện tại với các path sau đều cho `SKIP`:

```text
core/m07/m07.go
contracts/action-intent.schema.json
lab/affiliate-bot/go.mod
scripts/n8n_cli_preflight.py
scripts/validate_n8n_m06_operated_execution.py
```

Trong khi đó `lab/affiliate-bot/cmd/bot/watcher.go` cho `RUN`. Core/contracts và các helper này là dependency trực tiếp của adapter/runner, nên thay chúng có thể làm hỏng engine integration mà PR không chạy kiểm tra tương ứng. Full check trên main push chỉ phát hiện sau merge.

**Hướng sửa và nghiệm thu:** mở rộng filter theo dependency thực tế hoặc chạy engine cho mọi PR thay code liên quan. Tách logic chọn path thành phần test được; table-test core, contracts, mọi module manifest/checksum cần thiết, runner/helper/validator, blueprint, adapter và docs-only. Bảo đảm required job vẫn xuất hiện với kết quả hợp lệ khi chỉ đổi docs.

### RV-09 — Required-check hướng dẫn đã thiếu các job quan trọng

**Vị trí:** `docs/governance/REPOSITORY-GOVERNANCE.md:24–29`; đối chiếu cả hai workflow.

Tài liệu chỉ khuyến nghị bốn check. `deterministic-runtime` hiện chỉ chạy learner internal tests cùng contracts/vet và một số smoke; full `cmd/bot` tests đã chuyển sang hai shard, race job, Windows job và các smoke job riêng. Nếu cấu hình branch protection theo bảng này, lỗi ở các job mới không nhất thiết chặn merge.

**Giới hạn:** chưa đọc branch protection thực tế, nên finding là hướng dẫn quản trị không đầy đủ; không khẳng định repo hiện cho phép merge khi test đỏ.

**Hướng sửa:** liệt kê đầy đủ check cần bắt buộc hoặc tạo một check tổng hợp phụ thuộc tất cả job cần thiết và fail khi dependency failure/cancelled. Sau khi có thiết kế cụ thể mới cập nhật cấu hình GitHub bởi người có quyền.

**Nghiệm thu:** failure của learner shard, race, Windows, backup/mutation hoặc n8n được phản ánh ở check bắt buộc; docs-only skip có chủ đích không gây pending vô hạn. Tài liệu dùng đúng tên job thực tế, có kiểm tra chống lệch giữa workflow và danh sách.

### RV-10 — Hướng dẫn kiểm repo không chạy trọn, mô tả capability chưa được cập nhật

**Vị trí chính:** `docs/plans/REVIEW-REMEDIATION-PLAN.md:2986–2988`; `lab/affiliate-bot/README.md:76–105`; `docs/architecture/LEARNER-BOT-CONTINUITY.md:25`; `docs/product/MVP-SPEC.md:23`, `50`.

Vòng `for validator in scripts/validate*.py` trong runbook gọi cả operated validators không có `execution_json`/`--history`. Đã chạy thử M06 operated validator không tham số: exit 2, báo thiếu arguments; vòng hướng dẫn dừng trước các bước audit/smoke phía sau.

Một số entrypoint vẫn hướng người học theo mô tả baseline M01–M02 hoặc nói CLI/store BR-08–BR-12 là việc tương lai, trong khi shared core và các command tích hợp đã tồn tại. README root chỉ dẫn chủ yếu tới kế hoạch/review ngày 05/09, còn tracker hiện hành có hơn 3.000 dòng trộn cập nhật và lịch sử. Việc này làm khó phân biệt capability đã chạy được với operated proof còn mở.

**Hướng sửa:** đưa ra một lệnh kiểm offline tái lập được, liệt kê rõ operated checks có arguments và engine checks cần Node/n8n. Cập nhật mục lục/tóm tắt capability hiện tại; giữ tài liệu cũ như snapshot có nhãn ngày/baseline và link tới trạng thái mới. Không đánh đồng cập nhật tài liệu với đóng live/provider/pilot gaps.

**Nghiệm thu:** người bảo trì copy lệnh từ repo root chạy hết đúng tập offline; tài liệu chỉ rõ dependency còn thiếu. Người học tìm được entrypoint hiện có và hiểu phần nào chỉ là fixture. Structural/language/link validators vẫn đạt.

### RV-11 — Decoder flatted xử lý sai string số

**Vị trí:** `scripts/run_n8n_m06_schedule_regression.py:149–180`.

Hàm resolve tiếp một string literal đã lấy ra từ string table như thể đó là index mới. Đầu vào flatted `[ {"value":"1"}, "1" ]` cần cho `{"value":"1"}`, nhưng kết quả thực tế là `{"value":{"$flatted_ref":1}}`.

**Tác động:** dữ liệu execution có numeric-string có thể bị thay nghĩa hoặc sinh cycle marker giả khi làm bằng chứng. Fixture hiện có chưa chứng minh lỗi này làm hỏng một run engine thật, nên ưu tiên P3.

**Hướng sửa và nghiệm thu:** phân biệt reference trong container với literal tại string table, hoặc dùng parser chuẩn của runtime được pin. Kiểm string số, string thường, array/object lồng nhau và shared references; đối chiếu với output của package flatted trong môi trường n8n.

## 4. Khoảng trống đã có và quyết định thiết kế còn phải chốt

### G-01 — Phân biệt thời gian mô phỏng M11 với đồng hồ cấp quyền

`lab/affiliate-bot/cmd/bot/m11_registry.go:605–609`, `684–695` dùng timestamp CLI làm `now`; activation/gate có cùng kiểu xử lý. Đầu dò đặt đồng hồ process ở `01:02`, lease hết hạn `00:02`, nhưng truyền timestamp cũ `00:00:10/20/30`; vẫn tạo gate `ALLOW_PRODUCTION`, authorization và reservation `APPENDED`.

Đây là hành vi đã tái hiện. Tuy nhiên đường hiện tại là offline fixture và tài liệu đã để trusted external time ngoài phạm vi, nên không xếp nó như bằng chứng khai thác live executor. Cần chốt trước khi mở rộng runtime:

- Admission thực dùng clock do process sở hữu; timestamp đầu vào chỉ làm provenance.
- Historical/simulation giữ thời gian giả lập nhưng phải được đặt tên và ghi phạm vi rõ; không dùng regression đó để claim chống backdating theo wall clock.

Nghiệm thu thiết kế đã chọn bằng test clock hết hạn + timestamp cũ, bao gồm restart/restore. Không thay historical audit bằng `time.Now()` vì sẽ làm hồ sơ quá khứ hợp lệ trở thành không hợp lệ theo ngày chạy.

### Các blocker vận hành vẫn còn

- Selected-source/provider operated evidence, live executor, business outcome và pilot người mới độc lập vẫn phải có bằng chứng riêng theo RP-10.
- Native Windows đã có job trong repo, nhưng phiên này chỉ chạy macOS; không tái xác nhận mọi platform hoặc ancestor-race parity.
- Các test process-kill hiện hữu không chứng minh power-loss, transaction đa file hoặc distributed locking.
- Readiness audit kiểm tính nhất quán cấu trúc của plan/matrix/graph và command refs; không chứng minh code không còn lỗi, remote CI đang xanh hay business readiness.

Giữ các giới hạn này trong tracker, nhưng không dùng chúng để trì hoãn các sửa offline RV-01…RV-11 đã xác định.

## 5. Kế hoạch thực hiện

Chia thành PR nhỏ có test bắt lỗi trước/sau. Quy mô S/M/L dưới đây là độ lớn tương đối, không phải cam kết thời gian. Người thực hiện là người nhận gói; review/merge theo governance hiện hành.

| Gói | Nội dung | Phụ thuộc | Quy mô | Trạng thái |
|---|---|---|---|---|
| F-01 | Bảo toàn consumption M10 qua bind; RV-01 | Không | M | VERIFIED_REPO |
| F-02 | Cô lập environment n8n trước mọi import; RV-02 | Không; làm song song F-01 | S/M | VERIFIED_REPO |
| F-03 | Sửa exact binding và health TTL audit M11; RV-03/RV-05 | Không; cần đối chiếu graph/issuer | M | VERIFIED_REPO |
| F-04 | Bỏ chuyển số qua float64 trong M11 adapter; RV-04 | Không; có thể song song F-03 | S | VERIFIED_REPO |
| F-05 | Reject HTTP body quá lớn trước mutation; RV-06 | Không | S | VERIFIED_REPO |
| F-06 | Chứng minh restart bằng tick mới; sửa decoder; RV-07/RV-11 | F-02 trước khi chạy engine thật | S/M | VERIFIED_REPO |
| F-07 | Đóng khoảng trống CI filter và merge checks; RV-08/RV-09 | F-02, F-06 để engine evidence đáng tin | M | REPO_VERIFIED_ADMIN_OPEN |
| F-08 | Chốt semantics thời gian M11; G-01 | Độc lập về thiết kế; trước live admission | M | VERIFIED_REPO |
| F-09 | Đồng bộ runbook/capability và tracker; RV-10 | Cập nhật sơ bộ ngay; chốt sau F-01…F-08 | S/M | VERIFIED_REPO |
| F-10 | Nghiệm thu lại offline rồi tiếp tục RP-10 | F-01…F-09 | M | OFFLINE_VERIFIED_EXTERNAL_OPEN |

Trạng thái bảng được cập nhật theo PR #491 (21/09/2026). `VERIFIED_REPO` chỉ
phạm vi source/regression/hosted offline, không phải production readiness;
chi tiết CI và checklist còn mở nằm trong follow-up link ở đầu tài liệu.

### Cập nhật implementation trên current product main — 20/09/2026

Sau khi lập báo cáo, working tree đã có các bản sửa cục bộ cho F-01…F-07:

- `bind` từ chối đổi intent khi runtime đã có approval, grant, lease hoặc reservation; regression chạy qua learner Bot thật cho reservation đang chờ và execution FAILED, đồng thời kiểm tra backup graph không bị orphan.
- Hai n8n runner dùng chung environment builder: xoá cấu hình database/queue/config-file kế thừa, ép SQLite và regular execution trong thư mục tạm; có Python tests chống biến môi trường bên ngoài.
- M11 historical/registry audit kiểm exact intent ID/hash của cost và authorization, ràng buộc authorization expiry với health TTL an toàn; M11 adapter giữ `json.Number` qua projection; các mutation/large-integer regressions đã được thêm.
- Ba HTTP endpoint dùng đọc `limit+1` để reject body vượt giới hạn trước khi decode hoặc append; có test M06, M07 và history handoff chứng minh không mutation.
- Schedule Trigger regression lấy execution watermark sau shutdown trước khi chờ tick sau restart; decoder flatted giữ string chỉ gồm chữ số là literal và có unit tests cho nested reference/cycle.
- PR path filter đã bao gồm contracts/core dependencies, module manifests, runner/helper/validator và toàn bộ learner adapter dependencies; quyết định scope được tách thành `scripts/n8n_change_scope.py` với table-test cho dependency và docs-only; governance table đã liệt kê các job CI hiện hành.
- Backup/restore smoke đã tách budget drill sang runtime mới với immutable M11 registry, để không dựa vào hành vi rebind đã bị chặn; artifact-graph gate cũng kiểm health TTL tại thời điểm `ALLOW_PRODUCTION`.
- Runbook baseline đã bỏ glob gọi operated validators thiếu artifact và dùng danh sách static validators có exit-code guard; README, learner README và MVP spec trỏ tới trạng thái capability hiện hành.

Đây là trạng thái implementation đã có trên product main `3e12ffe`; riêng
runner/tài liệu F-09 đang được hoàn tất trên PR follow-up. Đây chưa phải
evidence của một commit/CI run mới cho follow-up. F-07 vẫn `PARTIAL_LOCAL` vì
branch protection thực tế và exact-head hosted acceptance chưa được xác nhận;
F-10 chỉ đạt `PARTIAL_LOCAL` vì host thiếu Go và chưa có n8n/provider/deployment
evidence. Overall readiness không đổi: `NOT_READY_FOR_PRODUCTION`.

### Cập nhật F-09/F-10 và đối chiếu RV — 20/09/2026

- RV-01…RV-08 và RV-11 đã có implementation/regression tương ứng trên current
  product main, trong phạm vi offline/fixture/read-only; RV-08 đã được nối vào
  helper dùng chung và merge ở PR #486. Những kết quả này không thay thế
  hosted exact-head, native Windows hoặc real n8n evidence.
- RV-09: strategy trong governance đã dùng `curriculum-gate` và `mission-gate`
  fail-closed, bao phủ shard/race/Windows/smoke/n8n; branch protection thật
  vẫn là việc quản trị chưa được xác nhận.
- RV-10: runbook glob gây gọi operated validator đã được thay bằng
  `scripts/run_offline_checks.py`; README/capability docs và exclusion boundary
  đã được cập nhật.
- F-09 targeted tests đạt 3/3, Python regression suite đạt 174 tests,
  compile/JSON/`--list` đạt. F-10 full runner dừng trước bước đầu vì thiếu
  `go`; cần chạy lại trên hosted/toolchain-complete environment rồi ghi riêng
  local và remote results.
- Các blocker còn mở: Go/Windows hosted run trên exact follow-up head, real
  n8n/Schedule Trigger, operated execution artifact, provider/live executor,
  target deployment drill, clean-machine beginner pilot, business outcome,
  distributed locking, power-loss/atomic multi-file proof và branch protection
  thực tế.

### Cập nhật exact-head PR #489 — 21/09/2026

- RV-08/F-07: `scripts/n8n_change_scope.py` nay bắt cả offline runner và các
  test parser/scope n8n; table-test 7 ca và Python suite 174 test đạt.
- RV-11/F-06: decoder giữ numeric-string là literal; hosted exact-head PR #489
  đã chạy thành công n8n 2.38.1 trên Node 24, M06/M07 engine và Schedule
  Trigger thật. Đây vẫn là fixture/loopback evidence, không phải provider/live.
- F-10: Go/core/mission, Windows, race, validators, smoke, mutation và các
  required aggregate gate đều PASS trên hosted exact head; full offline runner
  local vẫn chưa chạy vì host thiếu Go. `govulncheck` vẫn non-blocking failure.

### F-01 — Chặn mất lịch sử trước, hoàn thiện ledger sau

- [x] Đưa đầu dò A→B→A thành regression thất bại trên implementation hiện tại.
- [x] Chặn transition làm mất approval/grant/reservation/execution history; exact retry không reset counters.
- [x] Xác định ledger/consumption owner độc lập current intent; giữ invariant identity qua lịch sử.
- [x] Test pending, FAILED/NOT_PERFORMED, restart và restore; backup graph luôn hợp lệ hoặc báo cần đối soát trước mutation.
- [x] Ghi rõ cách xử lý runtime cũ đã thiếu reservation; không tự xóa execution hoặc reset hạn mức.

### F-02…F-06 — Sửa boundary và bằng chứng kiểm thử

- [x] Environment builder chung và test chống cấu hình database/queue kế thừa.
- [x] Negative tests M11 cho cost intent, authorization hash ở cả hai profile, health expiry.
- [x] Positive tests số chính xác qua adapter; kiểm cùng dữ liệu trước/sau để loại trường hợp test fixture sai.
- [x] HTTP limit test kiểm byte-level no-mutation cho cả ba endpoint family.
- [x] Restart watermark lấy sau shutdown; test không có tick mới phải fail.
- [x] Decoder flatted giữ literal string; đối chiếu parser runtime trong môi trường n8n pin trên exact-head PR #489.

### F-07…F-09 — Biến test và tài liệu thành gate dùng được

- [x] Path selection có table-test cho dependency trực tiếp và docs-only.
- [x] Required-check strategy phản ánh shard/race/Windows/smoke/n8n; branch protection thực tế vẫn cần owner/admin xác nhận.
- [x] Thời gian mô phỏng và admission được tách rõ trong API/CLI, docs và test.
- [x] Một entrypoint kiểm offline; operated validators nhận artifact cụ thể, không lẫn vào vòng glob không arguments.
- [x] README dẫn tới trạng thái hiện tại; tài liệu lịch sử gắn nhãn baseline; map RV mới vào RP tương ứng sau khi fix có bằng chứng.

### F-10 — Điều kiện đóng đợt review

- [x] Mỗi RV có test hoặc kiểm chứng phù hợp trong phạm vi offline; không đánh dấu DONE chỉ vì đã thêm test hoặc cập nhật prose.
- [x] Bốn Go module qua test/vet; learner race, Python regression, validator và smoke liên quan đều đạt trên cùng hosted exact head PR #489; local full runner vẫn dừng vì host thiếu Go.
- [x] n8n engine và Schedule Trigger chạy thật bằng environment cô lập trên phiên bản pin của repo trên exact-head PR #489.
- [x] CI trên head cần merge có đầy đủ check; kết quả local và remote ghi riêng.
- [x] Baseline/plan/matrix/evidence graph nhất quán; dữ liệu đang thay đổi trước phiên review đã được xử lý bởi chủ sở hữu, không ghi đè.
- [x] Overall vẫn `NOT_READY_FOR_PRODUCTION` cho tới khi các điều kiện operated/deployment/business tương ứng được chứng minh.

## 6. Lệnh baseline có thể chạy lại

Chạy từ repo root bằng entrypoint đã version-control; runner chủ ý loại trừ
operated validators và n8n engine khi chưa có artifact/runtime phù hợp.

```bash
python3 scripts/run_offline_checks.py --list
python3 scripts/run_offline_checks.py
```

Quickstart cache rỗng và HTTPS fixture cần mạng:

```bash
python3 scripts/smoke_quickstart.py
python3 scripts/smoke_br13c.py
```

Các runner n8n cần Node/n8n đã pin và phải hoàn tất F-02 trước khi dùng trên shell có cấu hình n8n riêng. Operated validators cần execution artifact/history được tạo từ run có ghi nhận. Các yêu cầu này là dependency của kiểm chứng, không phải lý do để sửa kỳ vọng test hoặc tự nâng readiness.

## 7. Bảo trì sau đợt sửa

Không cần đổi framework hoặc viết lại toàn repo để xử lý các finding này. Sau khi có regression bảo vệ, có thể tách `mission_command.go` (3.235 dòng), `backup_command.go` (1.869 dòng) và `m11_registry.go` (1.386 dòng) theo use case; ưu tiên các kiểu dữ liệu chung và adapter không chuyển đổi JSON vòng lại. Đây là việc giảm khó khăn bảo trì, không phải điều kiện để trì hoãn F-01/F-02.

Nên giữ phần trạng thái hiện tại của tracker ngắn và chuyển nhật ký cũ thành các evidence record có baseline rõ. Bảo toàn các marker/reference đang được `audit_readiness.py` kiểm khi tổ chức lại tài liệu. Mọi refactor nên là PR riêng sau sửa correctness, để reviewer thấy rõ thay đổi hành vi và thay đổi cấu trúc.
## 7. Post-review implementation update — F-08 (2026-09-20)

F-08/G-01 is implemented on learner head `35394a3`. M11 authority uses the
process-owned `missionNowUTC()` clock for lease validity, activation, ledger
initialization, gate, authorization, reservation and FAILED/UNKNOWN execution.
Input timestamps remain historical provenance for chronology and deterministic
artifact identity. The real learner Bot expiry regression restores checkpoints
and rejects both exact-expiry and backdated authority writes before mutation.

The result is bounded offline/fixture/read-only evidence. It does not close the
trusted-time, live-executor, provider, deployment, pilot, business-outcome or
production-readiness gaps listed above.
