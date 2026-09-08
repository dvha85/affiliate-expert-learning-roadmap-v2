# Affiliate Bot — deterministic learner baseline và continuity anchor

Workspace này là reference baseline (mốc tham chiếu) cho **M01 — Smallest Deterministic Bot v0.1** và **M02 — Trustworthy History + Replay v0.2**.

Sau M02, learner **không bỏ Bot này để chuyển sang một demo khác**. M03–M11 được tích hợp dần vào cùng learner Bot/workspace; `lab/mission-runtime/` chỉ là conformance oracle/harness để đối chiếu behavior/failure semantics, không phải Bot thứ hai.

Continuity contract: `docs/architecture/LEARNER-BOT-CONTINUITY.md`  
Continuity checklist: `starter-kits/CONTINUITY-CHECKPOINT.md`

## M01 baseline

```bash
cd lab/affiliate-bot
go run ./cmd/bot
go test ./...
```

Fixture mặc định `data/sample-observations.json` là canonical-shaped synthetic fixture (mẫu mô phỏng có shape chuẩn); nó chỉ chứng minh behavior kỹ thuật M01.

```text
known observations
→ deterministic formula + stable tie-break
→ RANK_SCENARIO | GET_MORE_DATA | HUMAN_REVIEW
→ NO external action
```

Baseline cố ý yếu: `price × commission_rate`.

```text
real evidence != RECOMMEND
RANK_SCENARIO != Approval != Execution
```

## M02 history + replay

M02 capture yêu cầu canonical identity/time: mỗi observation có `observation_id`, `subject_id`, provenance và `observed_at`. Fixture synthetic riêng là `data/m02-sample-observations.json`.

Capture một immutable decision snapshot:

```bash
go run ./cmd/bot history capture data/history.jsonl data/m02-sample-observations.json demo-1 2026-09-01T01:00:00Z 2026-09-03T10:00:00Z
```

Dừng process, chạy lại rồi query:

```bash
go run ./cmd/bot history list data/history.jsonl
```

Replay bằng `formula_version` đã record:

```bash
go run ./cmd/bot history replay data/history.jsonl
```

Replay states:

```text
MATCH        — same version + same input integrity + same result
DRIFT        — same supported version nhưng result khác record
UNREPLAYABLE — formula version không được registry hiện tại hỗ trợ
```

Nếu `input_hash` sai, evidence linkage sai, timestamp/identity invalid hoặc JSONL corrupt, record fail closed với integrity error **trước khi** được coi là replay hợp lệ.

`input_hash` bảo vệ integrity của canonical input snapshot; nó không chứng minh input là market truth.

```text
replay MATCH != business truth
replay MATCH != RECOMMEND
history exists != execution permission
```

M02 vẫn A0 deterministic, local-only: không scrape/login/publish/message/spend và không tạo Approval/Execution permission.

## Từ M03 trở đi — mở rộng cùng learner Bot

Repo không copy reference implementation M03–M11 vào workspace này vì việc đó sẽ tạo hai implementation phải giữ parity. Thay vào đó, với mỗi Mission learner phải:

1. đọc contract + lesson + starter của Mission;
2. chạy `lab/mission-runtime` như test oracle;
3. thêm capability/adapter vào **cùng learner Bot/workspace**;
4. bind artifact mới tới exact artifact IDs từ Mission trước;
5. lưu Integration Evidence + Reality/Operated evidence.

Ví dụ M03 bắt đầu từ `DecisionPacket.decision_id` hiện có rồi thêm human-only `ActionRecord`/`OutcomeRecord`; M06 thêm watcher adapter nhưng canonical Observation/History vẫn thuộc Deterministic Core; M11 chỉ mở production sau E5 promotion + finite ProductionLease.

## BR-03b: DecisionPacket từ history

Sau capture/replay, chạy `go run ./cmd/bot history decision <history.jsonl> <record_id> <context.json>` để xuất packet mới đã kiểm schema. Mẫu context: `data/m02-decision-context.json`; [hướng dẫn và mapping](../../docs/architecture/M02-DECISION-ADAPTER.md). Lệnh chỉ đọc, không thay history; không redirect stdout vào file history đầu vào.
# M03: kiểm action/outcome từ file

[BR-11a mock advisor](../../docs/architecture/BR-11A-MOCK-ADVISOR.md): `bot advisor mock HISTORY ACTIONS OUTCOMES CONFIG.json` chỉ đọc context và trả HUMAN_REVIEW, không gọi provider hay ghi store.

Các guard freshness có trạng thái `ABSTAIN_STALE` và `ABSTAIN_FUTURE`; yêu cầu write tool bị `REJECT_WRITE_REQUEST`.

[BR-10c: audit toàn chain từ M00](../../docs/architecture/BR-10C-ACCEPTANCE.md): chạy `python3 scripts/smoke_br10c.py` từ root để kiểm workspace lab độc lập; không phải live proof.

## M06–M11 learner entrypoints

Trạng thái sau review `ece6a32`: **PARTIAL**. Các lệnh dưới đây đã tồn tại nhưng
chưa nghiệm thu tích hợp: còn lỗi grounding/ID, budget/expiry/concurrency,
output-path safety và backup/restore. Chưa dùng các guard này cho live executor.
Xem [kế hoạch sửa R01–R16](../../docs/plans/REVIEW-REMEDIATION-PLAN.md) và
[baseline review](../../docs/plans/evidence/REVIEW-ECE6A32.md); việc lưu kế hoạch
không có nghĩa các lỗi đã được sửa.

Watcher fixture và watcher n8n dùng chung canonical adapter. Chạy adapter local
trước khi import blueprint M06:

```bash
go run ./cmd/bot watcher serve /tmp/affiliate-runtime/history.jsonl 127.0.0.1:8787
go run ./cmd/bot watcher history-handoff HISTORY.jsonl HISTORY-RECORD.json
```

M07 không tự gắn IDs từ context vào câu trả lời. Lệnh `context` xuất payload đã
resolve từ history; blueprint n8n cũng GET lại cùng `record_id` từ canonical
adapter trước khi gọi Agent. `validate` chỉ nhận output model có `A2-RO`,
`write_permission=false` và mỗi claim có `field_or_claim`/`value` khớp exact
với evidence IDs đã resolve:

```bash
go run ./cmd/bot m07 context HISTORY.jsonl DECISION_ID
go run ./cmd/bot m07 validate HISTORY.jsonl DECISION_ID MODEL-OUTPUT.json REGISTRY.json
```

Các entrypoint learner proposal-only cho M08–M11 là `mission m08-intent`,
`m08-policy`, `m09-approval`, `m10-canary`, `m10-cost-register`, `m10-gate`,
`m10-authorize`, `m10-reserve-authorization`, `m10-cancel`, `m10-reserve`,
`m10-record-failed`, `m10-outcome`, `m11-register`, `m11-resolve`, `m11-activate`, `m11-stop` và
`status`. Chúng ghi `mission-state.json`, kiểm hash/link trước khi ACK, giữ
budget usage sau restart, từ chối risk cần review nếu chưa có approval hợp lệ,
và STOP không bị `init` ghi đè. `m10-cancel` chỉ phát hành execution record
`CANCELLED`/`NOT_PERFORMED` được bind với authorization; nó không gọi executor,
không reserve thêm budget và không tạo side effect. Đây vẫn không phải live
executor.

Đường governed mới là `m10-reserve-authorization`: chỉ nhận authorization đã
được registry cấp, resolve lại trusted cost bound và charge đúng số đã được
ủy quyền. Reservation lưu `authorization_id`; `m10-cancel` chỉ tạo record sau
khi reservation đó tồn tại, rồi ghi ngược `execution_id` vào state. Điều này
ngăn một authorization bị reserve/record nhiều lần. `m10-reserve` cũ vẫn có
để tương thích các lab trước, nhưng không có authorization/execution binding
và không dùng cho chuỗi governed mới.

`m10-record-failed` là fixture stub local-only: nó chỉ có thể ghi
`FAILED / NOT_PERFORMED` sau reservation governed, không gọi executor hay
network. `m10-outcome` chỉ nhận outcome fixture `CANCELLED`, metrics rỗng, và
`EffectRef` `MACHINE_EXECUTION` resolve đúng execution record no-side-effect;
outcome được lưu trong `m10-outcomes.jsonl`. Nó không chứng minh outcome kinh
doanh hay mở đường live executor.

Các artifact M10 của learner được sao chép canonical vào registry bất biến
`m10-artifacts.jsonl` trong state directory. Grant, trusted cost bound, gate,
authorization và execution record chỉ được ACK sau khi registry nhận chúng;
authorization/execution resolve lại input từ registry và từ chối artifact hợp
schema nhưng không phải bản đã đăng ký. `m10-outcomes.jsonl` giữ fixture
outcome riêng; khi restore, Bot resolve lại execution record, reservation và
EffectRef trước khi chấp nhận graph. Registry/outcome fixture vẫn không phải
executor hay bằng chứng outcome kinh doanh.

Có thể chỉ đọc một artifact đã được registry sở hữu theo loại/ID (và tùy chọn
content hash); lệnh không tạo quyền thực thi:

```bash
go run ./cmd/bot mission m10-resolve /tmp/affiliate-runtime EXECUTION_RECORD canary-exec-IDENTIFIER
```

Mỗi lần reserve mới phải có `RESERVATION_ID` ổn định. Retry cùng ID, cost và
binding trả `EXACT_DUPLICATE` thay vì charge lần hai; tái dùng ID với cost hay
binding khác bị từ chối. Dạng cũ không có ID chỉ còn tương thích tạm thời và
không phù hợp cho attempt mới.

```bash
go run ./cmd/bot mission m10-reserve /tmp/affiliate-runtime 100 attempt-001
go run ./cmd/bot mission m10-reserve-authorization /tmp/affiliate-runtime authorization.json governed-attempt-001
go run ./cmd/bot mission m10-record-failed /tmp/affiliate-runtime authorization.json failed.json 2026-09-08T00:01:00Z fixture-dispatch-failed-before-executor
go run ./cmd/bot mission m10-outcome /tmp/affiliate-runtime cancelled-outcome.json
go run ./cmd/bot mission m10-cancel /tmp/affiliate-runtime authorization.json cancelled.json 2026-09-08T00:01:00Z learner-cancelled-before-executor
```

```bash
go run ./cmd/bot mission init /tmp/affiliate-runtime
go run ./cmd/bot mission status /tmp/affiliate-runtime
go run ./cmd/bot backup create /tmp/affiliate-runtime /tmp/affiliate-backup
go run ./cmd/bot backup restore /tmp/affiliate-backup /tmp/affiliate-restored
```

Backup profile hiện là `affiliate-bot-backup/v2`. Với runtime đã khởi tạo
canary, manifest bắt buộc có `m10-artifacts.jsonl`; nếu có execution `FAILED`
thì bắt buộc có thêm `m10-outcomes.jsonl`. Restore kiểm checksum, inventory bắt
buộc, replay history, mission state, rồi kiểm link reservation → execution và
FAILED execution → fixture outcome/`MACHINE_EXECUTION` EffectRef. Một backup
checksum hợp lệ nhưng orphan outcome vẫn bị từ chối với `GRAPH_FAILED`. Phạm vi
này chỉ là flat M10 learner graph: chưa là snapshot transaction, inventory
recursive M00–M10 hay lifecycle M11.

`m11-register STATE_DIR KIND ARTIFACT.json` và `m11-resolve STATE_DIR KIND ID`
là artifact spine cho M11: chúng dùng `core/m11` để strict-decode và lưu
canonical record vào `m11-artifacts.jsonl`; registry chỉ nhận link lifecycle
đã resolve (lease, approval, health, cost, ledger, gate, authorization,
execution, activation, reconciliation hoặc cycle). Backup v2 gồm registry này
nếu runtime đã có nó và kiểm lại graph trước `RESTORED`. Đây chưa phải lệnh
lease activation, production gate, executor, reconciliation/recovery hay M11
lifecycle đầy đủ.

`m11-activate STATE_DIR LEASE_ID ACTIVATED_AT` chỉ tạo `ProductionActivation`
khi lease và lease approval khớp exact đã có trong registry, thời điểm còn nằm
trong lease và runtime chưa STOP. Retry cùng activation trả `EXACT_DUPLICATE`.
`m11-ledger-init STATE_DIR LEASE_ID INITIALIZED_AT` chỉ tạo durable ledger
rỗng sau activation đã resolve, trong thời hạn lease và khi chưa STOP; retry
exact cũng không reset usage. Các lệnh này chưa thay thế health gate hay cấp
phép/executor production.

`m11-gate STATE_DIR LEASE_ID HEALTH_ID COST_BOUND_ID LEDGER_ID EVALUATED_AT`
resolve toàn bộ artifact trên từ registry, kiểm lease/activation/policy/scope,
budget, freshness và health, rồi persist `ProductionGateDecision`. Nó trả
`ALLOW_PRODUCTION`, `DEGRADE`, `STOP`, `REQUIRE_APPROVAL` hoặc `DENY`, nhưng
luôn giữ `execution_authorized=false`. Bất kỳ ID/link không resolve được đều bị
reject.

`m11-authorize STATE_DIR LEASE_ID GATE_ID EXECUTOR_ID AUTHORIZED_AT` chỉ tạo
`GOVERNED_PRODUCTION` authorization từ gate `ALLOW_PRODUCTION` đã persist,
executor được lease allow và health/cost link còn exact. Hạn dùng là mốc sớm
nhất của lease, intent, human approval và cost bound. Authorization không gọi
executor; envelope command vẫn `execution_permitted=false`.

[BR-10b: nhập và đọc OutcomeRecord](../../docs/architecture/BR-10B-OUTCOME-STORE.md): `bot outcome import HISTORY ACTIONS OUTCOMES INPUT`, `bot outcome list HISTORY ACTIONS OUTCOMES`; nối action đã lưu, store riêng, không execution.

[BR-10a: ghi nhận ActionRecord thủ công](../../docs/architecture/BR-10A-ACTION-STORE.md): `bot action record HISTORY.jsonl ACTIONS.jsonl ACTION.json`, đọc lại bằng `bot action list HISTORY.jsonl ACTIONS.jsonl`. Store action riêng, không đăng bài/thực thi; decision phải tồn tại và replay MATCH. Lệnh validate dưới đây vẫn chỉ đọc.

Đường nhập trước M03: [BR-09 — packet M00 → input M01 → history M02 → DecisionPacket](../../examples/m00-import/README.md), lệnh `bot evidence import PACKET.json` read-only; capture vẫn là bước ghi tường minh.

[BR-08e: kiểm tích hợp learner/harness và giới hạn nghiệm thu](../../docs/architecture/BR-08E-ACCEPTANCE.md) có smoke script dùng workspace tạm, không sửa dữ liệu cá nhân.

Lệnh mới `bot action validate ACTION.json OUTCOME.json` dùng shared core M03, chỉ đọc file. Xem [hướng dẫn, exit code và giới hạn](../../docs/architecture/BR-08C-ACTION-CLI.md). VALID không là approval hoặc bằng chứng vận hành; chưa ghi action/outcome vào store.
