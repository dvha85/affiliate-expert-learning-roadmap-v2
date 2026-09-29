# Sổ tương thích n8n — n8n compatibility ledger

Blueprint JSON hợp lệ chỉ chứng minh cấu trúc file. Khả năng import và giữ đúng semantics khi chạy là thuộc tính của **engine n8n cụ thể**, nên phải được chứng minh trên instance thật.

## Baseline node đã khai báo

| Blueprint | Node | `typeVersion` đang khai báo |
|---|---|---|
| M06 | Schedule Trigger (bộ kích hoạt lịch) / Set (gán dữ liệu) / HTTP Request (yêu cầu HTTP) / Code (mã) | 1.2 / 3.4 / 4.2 / 2 |
| M07 | AI Agent (Agent AI) / OpenAI Chat Model (mô hình chat OpenAI) / HTTP Request Tool (tool yêu cầu HTTP) / Code (mã) | 3.1 / 1.2 / 1.2 / 2 |

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

`tested_n8n_version` là **release admission**, nên vẫn để UNVERIFIED cho tới khi
có đủ smoke của cả hai blueprint theo ma trận dưới đây trên topology được chọn.
Không được biến “JSON parse được” hay một local run thành “n8n hỗ trợ workflow
này” một cách tổng quát.

```text
tested_n8n_version: UNVERIFIED
tested_node_versions: declared above; only the documented synthetic 2.38.1 fixture paths are verified
upgrade_review_cadence: trước mỗi lần nâng n8n engine và ít nhất mỗi quý
```

### Evidence local có phạm vi hẹp — 2026-09-09

Đây **không** thay đổi release admission ở trên. M07 blueprint hiện hành (SHA-256
`80f9247c586eab1075aa347c57b7ca151eb45a4c5f110320ab461b54de62a97d`) đã chạy
qua Manual Trigger trên n8n `2.38.1` local với credential Cockpit chỉ đọc. Lần
chạy `execution_id=12` hoàn tất toàn bộ chuỗi node: nhận/đăng ký tool, dựng
context canonical, Agent chỉ đọc, kiểm grounding, lưu proposal và báo ACK.
Grounding là `VALID/HUMAN_REVIEW`, proposal canonical
`sha256:afdfbb9116ea609f8032443021956d03a27c8ae8da40199e3a1ea22191c9b7de`
ACK với `execution_permitted=false`. Sau restart n8n, health, canonical history
replay và cùng proposal ID vẫn resolve được.

Scope là fixture synthetic/read-only; không có campaign, provider diversity,
received-redirect transport, production deployment, business outcome hoặc
MACHINE_EXECUTION. PR #96/#97 bổ sung CI n8n `2.38.1` cho M06 và M07 negative
paths, nhưng không đưa credential/model-success vào CI.

M06 hiện cũng có regression disposable trên n8n `2.38.1`: một workflow copy
active giữ node Schedule Trigger thật ở cadence một giây, append rồi
`EXACT_DUPLICATE` trước/sau restart n8n và adapter, replay canonical history qua
adapter loopback, và fail closed khi adapter không khả dụng. Điều này chỉ xác
minh đường fixture synthetic/read-only; không nâng release admission, không
chứng minh selected source hoặc deployment topology.

### Re-run engine cục bộ — 2026-09-14

`scripts/run_n8n_engine_regression.py` đã PASS lại trên macOS arm64 với n8n
`2.38.1`, Node `24.21.0`, SQLite runtime disposable và loopback adapters/model
stub của chính repo. Node 22 bị n8n từ chối (`>=24.0.0`); khi native
`isolated-vm` từng được build bằng Node 22 nhưng n8n chạy dưới Node 24, n8n
không khởi tạo expression engine. Rebuild dependency bằng cùng Node 24 đã khôi
phục engine. Vì vậy khi tái lập local, dùng cùng major Node với lúc `npm install`
hoặc rebuild native dependencies trước import. Đây là reproduction cục bộ của CI
fixture path, không thay đổi `tested_n8n_version: UNVERIFIED` và không là
bằng chứng từ nhà cung cấp, nguồn đã chọn vận hành hay triển khai.

Cùng runtime đó, `scripts/run_n8n_m06_schedule_regression.py` PASS: Schedule
Trigger thực append một record, trả `EXACT_DUPLICATE` trước/sau restart n8n và
adapter, và dừng trước ACK/report khi adapter loopback không khả dụng. Đây vẫn
là fixture read-only, không phải vận hành Schedule Trigger trên deployment thật.

### Kiểm tra lại local Windows — 2026-09-23

Trên Windows PowerShell, Go `go1.27.1 windows/386` được gọi bằng executable
`C:\Program Files (x86)\Go\bin\go.exe` với `GOARCH=amd64`; Node là `22.23.2` và
n8n là `2.35.7` từ npx cache. `scripts/run_n8n_engine_regression.py` và
`scripts/run_n8n_m06_schedule_regression.py` đều PASS trên SQLite disposable,
loopback adapter và fixture synthetic của repo. M07 adversarial/output execution
cũng PASS qua implementation thật. Đây là local fixture evidence có phạm vi hẹp;
không đổi `tested_n8n_version: UNVERIFIED`, vì runtime này khác n8n `2.38.1` +
Node 24 đang được dùng làm release/CI fixture.

Host hiện đã có LLVM-MinGW/Clang `22.1.8`; với `GOARCH=amd64`,
`CGO_ENABLED=1` và `CC=x86_64-w64-mingw32-gcc`, `go test -race` PASS.
Lúc dừng disposable n8n trên Windows, `Popen.terminate()` trả exit code `1`;
runner đã có test riêng để chấp nhận đúng mã kết thúc platform-specific này sau
khi workflow đã được quan sát chạy, không biến lỗi khởi động thành PASS.

### Release-pin rerun trên Windows — 2026-09-23

Đã cài và xác minh runtime portable n8n `2.38.1` chạy bằng Node `24.21.0`.
`scripts/run_n8n_engine_regression.py` PASS với engine thật trên SQLite disposable,
M06/M07 loopback adapter và model stub; các retry/replay, exact-number và
fail-closed/grounded-output checks đều đạt.

`scripts/run_n8n_m06_schedule_regression.py` cũng PASS: execution đầu là
`APPENDED`, execution retry là `EXACT_DUPLICATE`, execution sau restart n8n/adapter
vẫn `EXACT_DUPLICATE`, và adapter unavailable dừng trước ACK/report.

Lần chạy đầu với pin này hết timeout 35 giây; đồng thời `DB_SQLITE_POOL_SIZE=0` bị n8n
2.38.1 fallback vì pool phải từ 1 trở lên. Disposable harness đã được cập nhật
thành pool `1` và execution wait budget `90s`, có regression test; rerun sau cập
nhật PASS. Đây là điều chỉnh harness tương thích runtime, không phải thay đổi
semantics workflow. Chưa có phép thử tách biến để khẳng định thời gian startup
39–40 giây hoặc nguyên nhân duy nhất. Review sau đó siết teardown Windows:
exit code 1 chỉ được chấp nhận nếu harness đã yêu cầu terminate; server tự crash
vẫn làm test fail. Thay đổi teardown mới đã có unit regression, chưa rerun engine
2.38.1 sau review này vì runtime portable trước đã được dọn.

Evidence này vẫn chỉ là synthetic/read-only/loopback fixture. Giữ
`tested_n8n_version: UNVERIFIED` cho tới khi import/execution smoke được lặp lại
trên topology release được chọn và có operated/provider/deployment evidence tương
ứng; không suy ra production readiness từ local PASS.

## Smoke test bắt buộc

1. Import cả hai blueprint vào một n8n instance sạch/local; kiểm tra node/type version nào bị unknown hoặc tự migrate trước khi lưu.
2. M06 blueprint synthetic `br13-offer-fixture/v1`: lần đầu `APPENDED`, retry cùng event `EXACT_DUPLICATE`, giữ record ID/count/hash. Sửa body giữ identity cũ phải reject; event mới hợp lệ mới append record mới. Kiểm restart/replay và ACK/persisted flags theo [case M06.3](../../curriculum/M06/M06.3-n8n-readonly-workflow.md). Không đổi URL sang public/selected source trong starter này.
   Chạy **riêng** core normalizer với previous_hash để xác minh `NEW/UNCHANGED/CHANGED` và đổi thứ tự key không đổi fingerprint. Blueprint không trả ba state đó; không dùng persistence report thay change-detection evidence.
3. M07: dùng least-privilege read-only credential (credential quyền tối thiểu, chỉ đọc) hoặc mock. Xác nhận evidence output bình thường, write/prompt-injection request bị chặn và output boundary vẫn ở `HUMAN_REVIEW`.
4. Khi upgrade engine, lặp lại import và execution smoke. Bất kỳ thay đổi hành vi nào ở HTTP Request Tool, AI Agent, credential scope, static-data behavior hoặc node migration đều chặn activation cho tới khi ledger này được review lại.

Kết quả smoke chỉ là **integration evidence (bằng chứng tích hợp)**, tự nó không phải Reality/Operated evidence. Không commit production credential, write scope hay secret vào blueprint hoặc repo này.

## Ranh giới truyền JSON nguyên dạng của M07 — 2026-09-17

Runner disposable của M07 gửi kết quả tool do adapter sở hữu qua
`tool_result_text`, thay vì parse thành object JavaScript. Blueprint trong repo
truyền output của model qua node chỉ nhận chuỗi `Require Raw Model JSON Text`,
sau đó gửi `model_output_text` tới adapter dùng chung. Adapter của learner từ
chối request đồng thời cung cấp cả raw text và object.

Fixture regression có số `9007199254740993` và kiểm tra sidecar tool, output
model, bytes của proposal, restart và revalidation; `9007199254740992` được xem
là dấu hiệu làm tròn và phải làm test fail. Local Windows re-check đã có Node và
n8n executable, nhưng release/CI vẫn phải cài n8n 2.38.1 cùng Node 24 trước khi
chấp nhận đường chạy này. Ranh giới
này chỉ là bằng chứng synthetic/read-only/loopback; không chứng minh provider,
selected-source, business outcome, live executor, pilot hoặc deployment
readiness.

### Post-review rerun trên Windows — 2026-09-29

Sau CR-01…CR-08, hai regression runner chạy lại trên working tree hiện tại
(chưa commit) bằng n8n CLI `2.38.1` + Node `24.21.0`. Mỗi runner tạo home/SQLite
tạm riêng; adapter dùng loopback và model M07 là OpenAI-compatible stub chạy
local trong process. Không dùng provider credential hay gọi affiliate endpoint.

`scripts/run_n8n_engine_regression.py` PASS: M06 kiểm tra lưu trữ/retry,
đổi thứ tự key JSON, từ chối identity/source sai mà không đổi history, event mới,
capture đã làm sạch từ nguồn được chọn, replay và dừng an toàn khi adapter down.
M07 kiểm tra từ chối theo policy, proposal bám evidence, claim `null` của nguồn
được chọn và giữ chính xác số `9007199254740993`; restart, replay và
revalidation đều PASS.

`scripts/run_n8n_m06_schedule_regression.py` PASS với Schedule Trigger thật:
execution 1 `APPENDED`, execution 2 `EXACT_DUPLICATE`, execution 3 vẫn
`EXACT_DUPLICATE` sau restart n8n/adapter và giữ nguyên record ID; adapter down
dừng trước ACK/report. Log có một SQLite ping timeout rồi recovered; toàn bộ
assertion vẫn PASS.

Windows npm 11 mặc định chặn install scripts; trong prefix tạm chỉ approve
`isolated-vm@7.0.1` và `sqlite3@5.1.7` để có expression engine và SQLite native
module tương thích Node 24. Đây chỉ là local synthetic/read-only fixture trên
working tree chưa commit: không phải clean-tree/full offline, learner Operated,
provider/live hoặc deployment proof. Giữ `tested_n8n_version: UNVERIFIED` và
`NOT_READY_FOR_PRODUCTION`.
