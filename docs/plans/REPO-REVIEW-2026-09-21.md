# Review toàn repo và kế hoạch chỉnh sửa — 21/09/2026

## 1. Kết luận

Repo có nền tảng tốt về phân quyền, schema, liên kết artifact, kiểm tra lỗi và khôi phục. Các bộ test hiện hữu đã chạy trong phiên review đều PASS. Tuy nhiên, **chưa nên coi các boundary M10/M11 đã hoàn tất chỉ dựa trên CI xanh**: review mới tái hiện được lỗi integrity, thời gian cấp quyền và sticky STOP mà test hiện tại chưa chặn.

Có **13 phát hiện cần xử lý: 5 P1 và 8 P2**, cộng một số cải tiến cấu trúc được tách riêng. Nên sửa theo các PR nhỏ, bổ sung regression trước khi refactor lớn. Không có cơ sở để viết lại toàn hệ thống hoặc thêm framework mới.

Giữ trạng thái `NOT_READY_FOR_PRODUCTION`. Sửa hết các lỗi dưới đây vẫn không thay thế bằng chứng selected source, provider, learner pilot, deployment và business outcome độc lập.

Đây là **báo cáo review và kế hoạch tại mốc baseline**; các kết luận
remediation khi đó chưa hoàn thành. Phần cập nhật triển khai ở §9 ghi riêng
những thay đổi đang có trong working tree sau review, không thay thế bằng chứng
vận hành độc lập, không tự commit/PR và không kích hoạt dịch vụ thật.

## 2. Mốc đánh giá và phạm vi

- Baseline: nhánh `main`, commit `707ac76aba0c1aac705ec3e52bcbd45e71edfb85`.
- Môi trường kiểm chứng: macOS `darwin/arm64`, Go `1.27.0`, Python hệ thống.
- Inventory tracked: 600 file; 195 file Go, trong đó 87 file `_test.go`; 257 Markdown; 55 Python.
- Đã rà theo các lớp: `contracts/`, `core/`, learner CLI/store trong `lab/affiliate-bot/`, conformance harness `lab/mission-runtime/`, ba blueprint `lab/n8n/`, scripts/CI, curriculum/missions/evals/starter-kits/examples, kiến trúc/governance/plans.
- Đối chiếu mã hiện hành và chạy probe độc lập, không sao chép kết luận từ các review cũ. Các probe dùng Go overlay hoặc bản copy trong thư mục tạm; không chèn test vào source checkout.
- File không tracked `.codex-checkpoint.md` đã có trước phiên review và được giữ nguyên.

Mức ưu tiên:

- **P1:** sửa trước khi tiếp tục dựa vào kết quả authority/integrity/conformance tương ứng, và trước mọi mở rộng live capability.
- **P2:** lỗi chức năng, độ bền, admission, CI hoặc tài liệu cần sửa ở đợt kế tiếp.
- Mức ưu tiên không phải khẳng định có sự cố production hay khai thác từ xa. Learner vẫn proposal/fixture-only; executor trong harness chỉ là local sandbox.

## 3. Các phát hiện

### F01 — P1: M11 chain audit làm tròn số và chấp nhận intent đã bị sửa

**Vị trí:** `lab/mission-runtime/cmd/demo/m11_chain.go:121–136`.

Helper `toCore` marshal rồi `json.Unmarshal` vào `corem08.Intent`. Trường `Parameters` chứa `map[string]any`, nên số bị chuyển thành `float64`, mất tính chính xác trước bước kiểm hash.

**Tái hiện:** tạo bundle `closed_cycle`, seal intent với `product_id=9007199254740992`; chỉ đổi giá trị thành `9007199254740993`, không đổi hash hoặc các liên kết. `CheckM11Chain` vẫn trả `VALID / CONSISTENT_UNVERIFIED`. Ngược lại, bundle được seal đúng với số thứ hai bị trả `TAMPERED_INTENT`.

**Tác động:** kết quả audit integrity sai ở cả hai chiều. Không phải bằng chứng đã thực thi hành động bên ngoài.

**Hướng sửa/nghiệm thu:** dùng chuyển đổi typed hoặc decoder giữ `json.Number`; rà các helper tương tự. Test cả input hợp lệ và tampered-neighbor tại `2^53−1`, `2^53`, `2^53+1`, số lồng trong object/array; kiểm cả API và CLI. Không sửa bằng cách tính lại hash cho dữ liệu đã bị làm tròn.

### F02 — P1: Executor M11 bỏ lỡ sticky STOP khi policy cũng bị lỗi

**Vị trí:** `lab/mission-runtime/cmd/demo/m11.go:916–925`; đối chiếu `m11_gate.go:8–30,66–73`.

`EnforceProductionGate` kiểm trusted critical STOP trước lỗi eligibility thông thường. Executor lại gọi thẳng `AuthorizeProduction`, chỉ persist STOP nếu gate trả đúng `STOP`. Một lỗi policy có thể trả `DENY` trước khi kiểm tín hiệu compliance.

**Tái hiện:** cấp authorization hợp lệ; làm sai `Policy.Reason` đồng thời đặt health đã được bind/trust có `ComplianceAlertCount=1`; executor trả `DENY` nhưng không lưu STOP. Khôi phục policy và bỏ alert, cùng lease/authorization chạy được: `first=DENY`, `second=EXECUTED`.

**Tác động:** mất sự kiện phải thu hồi quyền một cách bền vững; lease cũ có thể hoạt động lại mà chưa qua recovery review. Đã tái hiện trong sandbox của harness.

**Hướng sửa/nghiệm thu:** chia sẻ critical-stop enforcement dưới lock executor đang giữ, tránh lấy lock lồng nhau. Test tổ hợp policy lỗi với compliance/kill switch/reconciliation, rồi restart; lease cũ vẫn phải `STOPPED` dù tín hiệu tạm thời đã biến mất.

### F03 — P1: M10 registry vẫn công nhận authorization nối vào gate DENY

**Vị trí:** `core/m10/artifact_registry.go:179–192`; đường sử dụng `lab/affiliate-bot/cmd/bot/mission_command.go:2654–2691`.

Graph validator kiểm reference, hash và derived ID nhưng thiếu điều kiện gate phải thực sự cho phép cấp authorization.

**Tái hiện:** tạo chuỗi hợp lệ, sửa gate đã đăng ký thành `Decision=DENY`, `Reason=INVALID_GATE`, tính lại envelope digest bằng helper chuẩn, giữ authorization cũ. Kết quả:

```text
loadM10ArtifactRegistry -> không lỗi
m10-reserve-authorization -> RESERVED
backup create -> BACKED_UP
```

**Tác động:** một graph có checksum đúng nhưng authority sai vẫn được dùng và sao lưu như hợp lệ. Đây là lỗi semantic integrity; không suy ra attacker từ xa có quyền sửa registry.

**Hướng sửa/nghiệm thu:** tái sử dụng invariant issuance ở graph validation: gate `ALLOW_CANARY`, không còn yêu cầu approval chưa giải quyết, executor/scope, thời gian và dependency bindings hợp lệ. Thêm checksum-valid mutation tests ở core, learner reserve và backup/restore; reject trước mutation. Không bắt historical restore phải thỏa wall-clock hiện tại: kiểm quan hệ thời gian tại thời điểm lịch sử tương ứng.

### F04 — P1: Trusted cost bound từ tương lai được sử dụng ngay

**Vị trí:** `core/m10/cost_bound.go:69–77`; các caller trong `lab/affiliate-bot/cmd/bot/mission_command.go:2534,2585,2663`.

`ValidFor` chỉ kiểm `ExpiresAt > now`, không kiểm `ObservedAt <= now`. Decoder chỉ bảo đảm expiry sau observation, chưa bảo đảm quote đã tồn tại lúc sử dụng.

**Tái hiện:** giữ clock tại T, quote có `ObservedAt=T+1 phút` và expiry muộn hơn. Learner trả `m10-cost-register=APPENDED`, `m10-gate(T)=ALLOW_CANARY`. Probe core còn tạo được `ExecutionAuthorized=true`.

**Hướng sửa/nghiệm thu:** kiểm `observed <= now < expires` tại helper dùng chung; xác minh các đường M10/M11 dùng helper này. Test observed T−1ns/T/T+1ns và expiry exact boundary. Recovery/historical audit phải truyền đúng thời điểm sự kiện, không dùng clock hiện tại một cách máy móc.

### F05 — P1: Reserve authorization trước thời điểm cấp làm runtime không backup được

**Vị trí:** `lab/affiliate-bot/cmd/bot/mission_command.go:2658–2660`; đối chiếu `backup_command.go:1063–1067`.

Đường reserve chỉ kiểm authorization chưa hết hạn, thiếu not-before check. Gate/authorize cũng nhận timestamp do caller cung cấp.

**Tái hiện:** clock T → tạo gate tại T+1 phút → authorize tại T+1 phút → reserve ngay tại T. Ba bước lần lượt trả `ALLOW_CANARY`, `AUTHORIZED`, `RESERVED`; backup sau đó trả:

```text
INPUT_ERROR: governed reservation falls outside authorization lifetime
```

**Tác động:** chính lệnh được ACK tạo state vi phạm điều kiện của backup; budget đã bị charge trước khi authorization có hiệu lực.

**Hướng sửa/nghiệm thu:** chụp clock một lần cho transition; kiểm `AuthorizedAt <= now < ExpiresAt` trước ghi counters/state; làm rõ timestamp do runtime sở hữu so với thời gian replay fixture. Future-issued request bị từ chối và toàn bộ canonical state không đổi; state được ACK phải backup/restore được.

### F06 — P2: Authorization M10/M11 còn hạn bị từ chối nếu execute muộn hơn lúc cấp

**Vị trí:** `lab/mission-runtime/cmd/demo/m10.go:573–575,629–635`; `m11.go:916–926`.

Executor tái tạo expected authorization với `AuthorizedAt=ctx.Now` rồi so toàn bộ struct với artifact đã cấp. Vì vậy kiểm expiry đúng ở phía sau không giải quyết được sai lệch timestamp này.

**Tái hiện:** authorize lúc `08:00:00Z`, execute lúc `08:00:00.000000001Z`. Cả M10 và M11 trả `DENY_AUTHORIZATION`, dù expiry còn ở `09:00:00Z` / `08:04:00Z` trong fixture.

**Hướng sửa/nghiệm thu:** tách kiểm binding bất biến của authorization khỏi tái đánh giá điều kiện hiện tại. Không bỏ revalidation policy/health/budget. Test execute trễ hợp lệ, future-issued, hết hạn, dependency bị sửa và retry idempotent. Kiểm tương tác với F02 khi sửa cùng executor.

### F07 — P2: Một số authorization bị cắt phần giây lẻ của expiry

**Vị trí:** `lab/mission-runtime/cmd/demo/m09.go:193`; cùng pattern ở `lab/affiliate-bot/cmd/bot/m11_registry.go:653–667`.

Code tính expiry bằng `time.Time` nhưng ghi lại với `Format(time.RFC3339)`, làm mất nanosecond.

**Tái hiện M09:** approval hết hạn `08:00:00.999999999Z`, now `08:00:00.5Z`; authorizer trả `AUTHORIZED` với expiry `08:00:00Z`; execute ngay trả `DENY_EXPIRED_AUTHORIZATION`. Learner M11 có cùng phép format gây mất precision, xác nhận bằng đọc mã; chưa có probe riêng cho nhánh này.

**Hướng sửa/nghiệm thu:** dùng formatter bảo toàn precision như `RFC3339Nano` nhất quán; bổ sung M09 và learner M11 vào precision matrix. Test expiry −1ns/exact/+1ns, round-trip persist/load và không kéo dài quyền sau expiry.

### F08 — P2: Tạo output parent trên POSIX làm rò file descriptor

**Vị trí:** `lab/affiliate-bot/cmd/bot/output_parent_posix.go:29–46`.

`defer unix.Close(directory)` giữ giá trị descriptor ban đầu. Vòng lặp đóng descriptor đó và gán descriptor mới, nên defer không còn đóng đúng descriptor đang sở hữu ở cuối traversal.

**Tái hiện trên macOS:** gọi helper 12 lần để tạo 12 parent một cấp; số descriptor trong `/dev/fd` tăng từ 5 lên 17. Probe regression thất bại với `leaked 12 file descriptors`.

**Tác động:** leak khi tái sử dụng trong cùng process. Còn có rủi ro đóng nhầm descriptor nếu số cũ được tái sử dụng; phần đóng nhầm chưa được tái hiện. CLI ngắn sống được OS thu hồi khi exit, nhưng helper vẫn sai ownership.

**Hướng sửa/nghiệm thu:** quản lý descriptor hiện hành nhất quán, đóng đúng một lần trên mọi đường success/error. Test số cấp 0/1/2/3, lỗi giữa traversal, gọi lặp nhiều lần; chạy Linux và macOS. Không làm yếu `openat/O_NOFOLLOW`.

### F09 — P2: HTTP adapter local thiếu kiểm soát caller trước khi ghi canonical state

**Vị trí:** `lab/affiliate-bot/cmd/bot/watcher.go:292–320,734–767`.

Server giới hạn bind loopback nhưng gắn handler trực tiếp, không có authentication/Origin/Host admission. Decoder nhận JSON mà không kiểm media type.

**Tái hiện ở HTTP handler:** POST fixture hợp lệ với `Origin: https://untrusted.example`, `Content-Type: text/plain`, không credential → HTTP 200, `canonical_history_ack=true`, history tạm đã được ghi.

**Giới hạn kết luận:** đây là bằng chứng thiếu admission tại server, **không** phải chứng minh exploit qua browser hiện hành hay kết nối từ Internet. Runbook synthetic/loopback giảm phạm vi tác động nhưng không nêu chấp nhận mọi caller là trusted.

**Hướng sửa/nghiệm thu:** chốt threat model trước khi dùng persistent learner data; chọn token cục bộ hoặc IPC phù hợp làm boundary chính, bổ sung Host/Origin/media-type checks. Caller được cấu hình vẫn hoạt động; request thiếu/sai credential hoặc origin không cho phép bị reject trước mutation. Cập nhật cả blueprint và smoke khi thay giao thức; không commit token, không tự thêm write credential/provider.

### F10 — P2: PR thay dependency thật có thể bỏ qua n8n engine regression

**Vị trí:** `.github/workflows/mission-agent-path-ci.yml:99–105`.

Path gate chỉ liệt kê một số file adapter. PR chỉ đổi `core/m06/`, `core/m07/`, `contracts/`, Go mod/sum hoặc helper được import trực tiếp như `scripts/n8n_cli_preflight.py`, `scripts/validate_n8n_m06_operated_execution.py` không khớp regex. Các helper cuối được runner import tại `scripts/run_n8n_engine_regression.py:26–28`.

**Tác động:** kiểm tra integration bị skip trước merge khi dependency có thay đổi. Push vào `main` vẫn chạy engine, nên không khẳng định test này mất hoàn toàn.

**Hướng sửa/nghiệm thu:** bỏ path gate hoặc dùng dependency set bảo thủ, dễ duy trì. Thêm unit test bảng must-run/may-skip; mọi thay đổi dependency ở trên phải chạy engine trước merge. Khi không xác định diff/dependency được, fail closed hoặc chạy đầy đủ.

### F11 — P2: Hướng dẫn required checks thiếu job test learner hiện tại

**Vị trí:** `docs/governance/REPOSITORY-GOVERNANCE.md:24–29`; `.github/workflows/curriculum-ci.yml:56–58,123–165`.

Governance chỉ đề nghị bốn required checks. Job `deterministic-runtime` hiện test learner `./internal/...`; test `cmd/bot` đã chuyển sang hai shard/race và các job khác. Theo đúng bảng hiện tại vẫn có thể cho merge khi job learner trọng yếu fail.

**Giới hạn:** chưa truy vấn remote branch protection; đây là lỗi hướng dẫn/cấu trúc gate, không xác nhận repo GitHub đang cấu hình sai.

**Hướng sửa/nghiệm thu:** cập nhật required-check set hoặc tạo aggregate gate có semantics rõ cho failed/cancelled/skipped. Đồng bộ governance bằng consistency test. Việc thay branch protection thật cần chủ repo phê duyệt và thực hiện riêng; không coi sửa YAML/Markdown là đã bật protection.

### F12 — P2: Bài M07 hướng dẫn direct HTTP tool đã bị loại khỏi blueprint

**Vị trí:** `curriculum/M07/M07.3-n8n-evidence-agent.md:15–22`.

Bài yêu cầu giữ HTTP tool GET-only và mô tả node `HTTP Request Tool`/`Grounding Boundary`. Blueprint hiện gọi `Fetch and Register Tool Adapter` bằng POST tới loopback (`lab/n8n/M07-readonly-evidence-agent.blueprint.json:55–66`); validator còn cấm direct `httpRequestTool` (`scripts/validate_agent_semantics.py:101–105`).

**Tác động:** learner tìm node không tồn tại hoặc thêm lại capability trái boundary hiện hành. POST tới adapter local không đồng nghĩa Agent được phép POST ra nguồn bên ngoài.

**Hướng sửa/nghiệm thu:** viết lại lesson theo fetch/register → canonical context → Agent → validate/register proposal → ACK. Đồng bộ starter/checkpoint/runbook; walkthrough phải chạy đúng blueprint, không thêm direct tool. Có doc-contract test cho tên node và boundary quan trọng.

### F13 — P2: Starter M06 yêu cầu đổi source trái với bài học/profile fixture

**Vị trí:** `starter-kits/M06-readonly-watcher/README.md:8–10`; đối chiếu `curriculum/M06/M06.3-n8n-readonly-workflow.md:11–15`.

Starter yêu cầu thay source thành public/allowlisted source sau khi import fixture blueprint. Bài học lại cấm thay URL theo cách đó; blueprint dùng `/v1/m06/fixture-import`, không phải generic public fetch.

**Hướng sửa/nghiệm thu:** tách rõ bài synthetic fixture và selected-source profile ACCESSTRADE-Shopee; nêu adapter prerequisite, output/ACK thực tế và giới hạn PASS. Learner mới phải làm theo được mà không tự sửa authority/source semantics để vượt bước.

## 4. Kế hoạch triển khai

Tại mốc review, tất cả đầu việc dưới đây đang **OPEN**. Quy trình mỗi PR:
regression đỏ trên baseline → sửa tối thiểu → regression xanh → chạy matrix liên
quan → human review. Trạng thái working-tree remediation hiện tại được ghi ở
§9; không tự merge hoặc nâng readiness.

| Đợt / PR đề xuất | Phạm vi và kết quả cần có | Phụ thuộc | Tiêu chí hoàn tất |
|---|---|---|---|
| A1 — Integrity số và precision | F01, F07; bỏ lossy conversion, giữ nanosecond; rà các điểm chuyển typed/raw liên quan | Không | Valid/tampered large-number phân biệt đúng; M09/learner M11 không truncate expiry; không thay hash/version ngoài chủ đích |
| A2 — M10 authority graph và timeline | F03, F04, F05; graph invariant dùng chung, quote/auth not-before, một clock snapshot mỗi transition | Không; kiểm tích hợp lại sau A1 | DENY gate không dùng được; future input reject không mutation; mọi state ACK được backup/restore; historical chain hợp lệ vẫn load được |
| A3 — Executor conformance M10/M11 | F02, F06; sticky STOP ưu tiên và revalidation không tái cấp authorization | Không; chạy toàn matrix sau A1/A2 | Delayed execution hợp lệ hoạt động; critical STOP bền qua restart; giữ expiry/budget/idempotency/tamper guards |
| B1 — Ownership descriptor | F08; sửa close/defer và test mọi nhánh traversal | Không | FD count không tăng; Linux/macOS pass; path-race guards không yếu đi |
| B2 — Admission HTTP local | F09; ghi rõ threat model, chọn caller authentication, đồng bộ adapter/blueprint/smoke | Chốt giao thức local trước sửa | Caller hợp lệ dùng được; request không được phép không tạo history/sidecar/network fetch; secret không xuất hiện trong log/repo |
| C1 — CI trước merge | F10, F11; path dependency tests, required-check gate và governance nhất quán | Có thể làm sớm song song A/B | Sửa dependency bắt buộc chạy n8n; fail một shard thì required gate fail; skip hợp lệ phân biệt với bỏ sót |
| C2 — Learner docs | F12, F13; cập nhật M06/M07, core API index, README chỉ dẫn trạng thái hiện hành | Chốt B2 để tránh viết lại giao thức | Fresh walkthrough khớp node/command/ACK; doc-contract checks pass; không đổi criteria/PASS của curriculum |
| D — Nghiệm thu và giảm nợ cấu trúc | Full regression, readiness/evidence đồng bộ; refactor nhỏ theo mục 5 | A/B/C hoàn tất | Bằng chứng có commit/OS/command/result/limitation; không nâng production readiness bằng fixture |

### Checklist chung khi nghiệm thu A/B

- [ ] Regression mới thất bại đúng vì lỗi trên baseline, không vì setup fixture hỏng.
- [ ] Lỗi bị từ chối trước side effect; so canonical bytes/counters trước và sau rejection.
- [ ] Artifact hợp lệ vẫn đi hết flow; tránh chỉ thêm negative test làm mọi request đều bị reject.
- [ ] Preserve exact identity/hash, số lớn và timestamp precision qua persist/load/backup.
- [ ] Kiểm giao điểm runtime ↔ shared core ↔ harness ↔ backup, không chỉ hàm riêng lẻ.
- [ ] Race/process-kill/idempotent-retry tests liên quan vẫn pass.
- [ ] Chỉ đóng finding khi có regression và kết quả trên commit sửa; ghi rõ giới hạn chưa kiểm chứng.

Thứ tự ưu tiên: **A1/A2/A3 trước**, B1 có thể song song; C1 nên làm sớm để bảo vệ các PR sau. B2/C2 cần thống nhất giao thức trước khi hoàn tất hướng dẫn. Không nên bắt đầu refactor lớn trước khi khóa các invariant mới bằng test.

## 5. Cải tiến cấu trúc sau khi sửa lỗi

Đây là khuyến nghị bảo trì, không cộng vào 13 lỗi và không phải điều kiện tự động nâng production readiness.

1. **Tách CLI theo use case, giữ shared invariant một nguồn.** `mission_command.go` 3.235 dòng, `backup_command.go` 1.869 dòng, `m11_registry.go` 1.386 dòng. Tách command parsing, domain validation và persistence; tránh copy công thức authority vào nhiều adapter. Dùng F01/F03/F04 làm parity tests trước khi di chuyển code. Không gộp harness vào learner theo cách mất tính kiểm chứng độc lập.
2. **Tách trạng thái hiện hành khỏi lịch sử remediation.** `REVIEW-REMEDIATION-PLAN.md` 3.040 dòng, `audit_readiness.py` 2.150 dòng. Giữ một index ngắn cho trạng thái/blocker/claim hiện tại; chuyển chronology sang evidence archive có link. Không xóa lịch sử hay thay metadata để làm readiness xanh.
3. **Ưu tiên behavioral tests thay cho exact source-string guards.** Guard văn bản hữu ích cho curriculum authority, nhưng không đủ chứng minh invariant runtime và làm refactor tốn công. Các finding mới phải có test chạy hành vi; doc-contract checks chỉ giữ phần hướng dẫn có thể thực thi.
4. **Coverage theo rủi ro, không theo tỷ lệ đơn lẻ.** Unit coverage của core M08/M10/M11 còn thấp hơn các module đầu, nhưng harness/learner cũng test xuyên module. Lập matrix invariant → test → consumer; bổ sung boundary-time, số lớn, lỗi đồng thời và graph mutation trước khi đặt mục tiêu phần trăm.
5. **Giữ local-first và scope nhỏ.** Chưa có bằng chứng cần thêm database, queue, agent framework hoặc hệ thống orchestration khác. Chỉ cân nhắc sau khi quan sát bottleneck thực tế và có migration/rollback/authority review riêng.

## 6. Kiểm chứng đã thực hiện

| Nhóm | Kết quả tại baseline |
|---|---|
| `contracts`: test, test `-count=1 -cover`, vet | PASS; statement coverage 79,7% |
| `core`: test, test `-count=1 -cover`, vet | PASS cho m00, m03–m11 |
| `lab/affiliate-bot`: `go test -count=1 -cover ./...`, vet | PASS; cmd/bot 75,1%, internal/app 80,8%, learning 100%, store 70,6% |
| `lab/affiliate-bot`: `go test -race -count=1 ./...` | PASS |
| `lab/mission-runtime`: test, vet, race | PASS |
| Python unittest discovery trong `scripts/tests` | 133 tests PASS |
| Standalone validators/auditor | 15 checks PASS, liệt kê bên dưới |
| `scripts/smoke_br16a_offline.py` | PASS: shared runtime M00→M11, reconciliation, backup/restore, replay, STOP |
| `scripts/smoke_br18b_backup_restore.py` | PASS: typed v3 inventory và các graph mutations hiện có |
| Inline relative file links của 257 tracked Markdown | Không phát hiện đích file không tồn tại trong kiểu link đã quét |
| Probe mới ngoài repo | Xác nhận F01–F06, M09 của F07, F08 và handler admission của F09; không phải bộ regression đã được commit |

Các standalone checks đã chạy:

```text
validate_missions.py
validate_repo.py
validate_artifact_spine.py
validate_continuity.py
validate_language_policy.py
validate_agent_semantics.py
validate_semantic_contracts.py
validate_m11.py
validate_n8n_m06.py
validate_n8n_m06_selected_source.py
validate_n8n_m06_cases.py
validate_n8n_m07.py
validate_n8n_m07_adversarial.py
validate_n8n_m07_output_cases.py
audit_readiness.py
```

Core unit coverage riêng: m00 83,1%; m03 75,8%; m04 76,2%; m05 80,5%; m06 82,1%; m07 74,1%; m08 29,2%; m09 75,0%; m10 63,2%; m11 59,2%. Không cộng các tỷ lệ này thành coverage end-to-end.

Readiness auditor kiểm 99 scoped claims và giữ `NOT_READY_FOR_PRODUCTION`; metadata readiness đang tham chiếu baseline `598bb21801d4`, không phải chứng nhận production cho HEAD mới.

### Lệnh regression nền để chạy lại sau mỗi đợt

Chạy từ root repo; mỗi Go module độc lập, không giả định có root `go.mod`:

```bash
(
  set -e
  for review_module in contracts core lab/affiliate-bot lab/mission-runtime; do
    (cd "$review_module" && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...)
  done
  (cd lab/affiliate-bot && go test -race -count=1 ./...)
  (cd lab/mission-runtime && go test -race -count=1 ./...)
  python3 -m unittest discover -s scripts/tests -v
  python3 scripts/smoke_br16a_offline.py
  python3 scripts/smoke_br18b_backup_restore.py
  python3 scripts/audit_readiness.py
)
```

Ngoài các lệnh nền, từng PR phải chạy standalone validators/mutation tests/engine tests đúng phạm vi và CI Linux/Windows; B1 thêm macOS. Không coi danh sách này thay thế toàn bộ required CI.

### Probe chẩn đoán còn giữ tạm trên máy review

Các đường dẫn sau chỉ giúp tái hiện ngay trong phiên/máy hiện tại, không phải artifact bền vững của repo. Đợt A/B cần đưa regression tối thiểu vào test suite chính thức.

```text
/tmp/mission-review.Ge2iP1/lab/mission-runtime/cmd/demo/review_regression_test.go
/tmp/bot-core-review.S5zi5j/probe_test.go
/tmp/bot-core-review.S5zi5j/overlay.json
/tmp/bot-review-20260921.nbODgr/review_probe_test.go
/tmp/bot-review-20260921.nbODgr/overlay.json
```

Probe runtime/numeric/STOP và FD/HTTP được viết theo kỳ vọng đúng, nên FAIL trên baseline. Ba probe M10 ghi nhận hành vi lỗi hiện tại để xác nhận, nên có thể PASS trong khi log chứa `RESERVED`/`BACKED_UP` không đúng invariant. Không đọc exit code của các probe chẩn đoán như bằng chứng đã sửa lỗi.

## 7. Giới hạn và điều kiện ngoài code

- Chưa chạy live n8n engine/provider, gọi ACCESSTRADE thật, tạo affiliate link, publish, gửi message hoặc tiêu tiền. Static blueprint/CLI validation không thay engine-operated proof.
- Chưa chạy lại cold-cache quickstart, toàn bộ mutation scripts, Windows/Linux runtime tại phiên này; kết quả local không thay required CI đa nền tảng.
- Chưa audit cấu hình GitHub branch protection thực, secrets/deployment của máy chủ hay vulnerability database dependency. Không suy ra chúng an toàn từ review local.
- Chưa có browser exploit test cho F09; chỉ xác nhận server-side admission bằng `httptest`.
- Không tiến hành power-loss/filesystem-crash hoặc distributed-lock audit mới. Giới hạn đã công khai trong readiness vẫn mở.
- Independent selected-source/provider operation, beginner pilot trên máy sạch, business outcome/evaluation, authorized live executor/production lease và recovery drill trên target host vẫn cần người có thẩm quyền và môi trường thật. Fixture/CI/docs cleanup không được dùng để đóng các mục này.
- Review bao phủ các khu vực của repo theo rủi ro và kiểm chứng chọn lọc; không phải chứng minh mọi dòng mã đã hết lỗi.

## 8. Điểm kết thúc của kế hoạch

Kế hoạch remediation chỉ hoàn tất khi F01–F13 được xử lý hoặc có quyết định chấp nhận rủi ro có lý do/phạm vi rõ ràng, test mới đã được commit, CI bắt buộc thực sự bảo vệ các đường quan trọng và tài liệu learner chạy đúng implementation. Mọi thay đổi authority/PASS vẫn do `CURRICULUM.md` chi phối; production readiness chỉ xét lại khi có bằng chứng vận hành riêng.

## 9. Cập nhật triển khai sau review — working tree 21/09/2026

Phần này là nhật ký remediation của working tree hiện tại, không phải một
claim production. F01–F13 đã được xử lý ở mức mã nguồn/test/tài liệu như sau:

| Finding | Thay đổi và regression chính | Trạng thái hiện tại |
|---|---|---|
| F01 | M11 chain dùng `coreIntent`/`json.Number`, không round-trip intent qua `float64`; có test large-number hợp lệ và tampered-neighbor. | IMPLEMENTED_OFFLINE |
| F02 | Executor M11 kiểm tra trusted critical STOP trước policy eligibility và persist sticky STOP; test policy lỗi + compliance alert và restart-equivalent. | IMPLEMENTED_OFFLINE |
| F03 | M10 registry graph tái kiểm gate `ALLOW_CANARY`/reason/risk, executor/correlation/cost scope và chronology; checksum-valid gate/authorization mutations bị reject. | IMPLEMENTED_OFFLINE |
| F04 | `core/m10.ValidFor` yêu cầu `ObservedAt <= now < ExpiresAt`; learner cost-register/gate/authorize/reserve dùng một runtime clock snapshot; có boundary tests và future-observation learner test. | IMPLEMENTED_OFFLINE |
| F05 | M10 governed reservation kiểm `AuthorizedAt <= now < ExpiresAt` trước mutation và giữ canonical state nguyên vẹn khi authorization ở tương lai. | IMPLEMENTED_OFFLINE |
| F06 | M10/M11 executor revalidate eligibility nhưng so authorization theo issuance time; delayed execution regressions giữ capability còn hạn hoạt động. | IMPLEMENTED_OFFLINE |
| F07 | M09 và learner M11 marshal expiry bằng `RFC3339Nano`; precision/round-trip và exact-expiry regressions được bổ sung. | IMPLEMENTED_OFFLINE |
| F08 | POSIX output-parent đóng descriptor hiện hành bằng deferred closure; repeated traversal FD regression xanh. | IMPLEMENTED_OFFLINE |
| F09 | HTTP adapter loopback yêu cầu Bearer token local, host/origin loopback và JSON media type trước handler; blueprints, smoke, runbook, lesson và cả Schedule Trigger runner truyền token qua môi trường. | IMPLEMENTED_OFFLINE |
| F10 | Bỏ path selector không đầy đủ, n8n engine regression chạy mọi PR/push; `mission-gate` fail closed khi job bị skip/cancel/fail. | IMPLEMENTED_OFFLINE |
| F11 | Thêm `curriculum-gate` bao phủ learner shards, Windows, race, quickstart, smoke/mutation và readiness jobs; governance chỉ rõ hai aggregate check, không giả định branch protection đã bật. | IMPLEMENTED_OFFLINE |
| F12 | M07 lesson/starter mô tả đúng chuỗi adapter POST → canonical context → Agent → validate/register proposal, không còn direct HTTP tool. | IMPLEMENTED_OFFLINE |
| F13 | M06 starter phân biệt synthetic fixture profile với selected-source profile và không yêu cầu đổi source trái lesson. | IMPLEMENTED_OFFLINE |

### Evidence đã chạy trên working tree

```text
contracts/core/lab/affiliate-bot/lab/mission-runtime: go test -count=1 ./...   PASS
contracts/core/lab/affiliate-bot/lab/mission-runtime: go vet ./...             PASS
lab/affiliate-bot: go test -race -count=1 ./...                                PASS
lab/mission-runtime: go test -race -count=1 ./...                               PASS
scripts/tests: 185 unittest tests                                             PASS
scripts/smoke_br16a_offline.py                                                 PASS
scripts/smoke_br18b_backup_restore.py                                          PASS
17 mutation guards (`scripts/mutate_*.py`)                                    PASS
scripts/smoke_quickstart.py (isolated clone at baseline HEAD)                   PASS
workflow YAML + n8n blueprint JSON parse                                        PASS
standalone validators (M00–M11, n8n M06/M07)                                   PASS
scripts/audit_readiness.py                                                      NOT_READY_FOR_PRODUCTION
```

Các kết quả trên là local working-tree evidence, chưa phải commit/PR evidence;
CI đa nền tảng phải chạy lại trên commit sửa. Máy review không có `node`/`n8n`,
nên chưa có n8n engine-operated rerun trong phiên này. F09 chưa có browser/live
network exploit test; F10/F11 chưa thay branch-protection settings thật. Các
giới hạn vận hành, provider, selected-source, learner pilot, deployment và
business outcome ở §7 vẫn mở. Vì vậy readiness vẫn giữ
`NOT_READY_FOR_PRODUCTION`.
