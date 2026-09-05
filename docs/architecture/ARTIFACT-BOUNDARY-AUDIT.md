# BR-03 — Đối chiếu artifact và boundary

## Phạm vi đã triển khai

BR-03a kiểm `AdvisorOutput` M04 tại JSON boundary và trước SUPPORTED. BR-03b bổ sung schema package, M02 capture/load/serialized history và adapter xuất DecisionPacket (đã review/merge PR #24 tại `9102f00`). Đây **không phải** tuyên bố toàn bộ 29 schema đã được cưỡng chế ở runtime. BR-03c còn mở; bảng này là bản đồ để review từng phần, không phải báo cáo conformance đầy đủ.

Chọn standard library Go cho contract nhỏ M04: kiểm JSON gốc, required/unknown/duplicate keys, null/array/item types, enum, constant false, unique/nonempty evidence IDs và reason. Không thêm dependency. `reason`/ID chỉ chứa khoảng trắng cũng bị từ chối (chặt hơn minLength của schema). Test khóa cấu trúc schema M04 để buộc review validator khi contract đổi. Không dùng giải pháp thủ công này như một JSON Schema engine tổng quát cho `$ref`, `allOf` hoặc các contract phức tạp.

Schema không thay semantic checks: ID phải resolve, nguồn không rỗng, context không có ID trùng ghi đè nhau, timestamp đúng RFC3339 và evidence được dùng không stale/future. `SUPPORTED` chỉ phản ánh các check hiện có, không xác minh nội dung ngôn ngữ có được evidence hỗ trợ. Không có quyền write hoặc quyết định kinh doanh tự động.

## Bản đồ Go ↔ schema ↔ boundary

Schema bên dưới nằm trong `contracts/`; type nằm trong module tương ứng. Trừ dòng M04, trạng thái là **cần audit/kiểm serialized output**, không suy từ tên type giống schema.

| Artifact / Go type | Schema | Input/output hiện tại và phần cần kiểm |
|---|---|---|
| `affiliate-bot.Observation` | `observation.schema.json` | BR-03b capture kiểm raw observations theo history subschema có `$ref`/`allOf`, strict decode trước hash |
| `affiliate-bot.HistoryRecord` | `history-record.schema.json` | BR-03b kiểm capture/JSONL/load/append + serialized output; ngoại lệ legacy ranked:null được ghi riêng |
| `affiliate-bot.Result` | projection trong `history-record.schema.json` | `recorded_result` không phải DecisionPacket; giữ dữ liệu replay hiện tại |
| `affiliate-bot.DecisionPacket` | `decision-packet.schema.json` | BR-03b: history decision xuất packet riêng đã kiểm schema từ replay=MATCH + context tường minh; không ép cast Result |
| `HumanActionRecord` | `action-record.schema.json` | M03 eval nhận record; demo chỉ trả trạng thái validation |
| `EffectRef` | `effect-ref.schema.json` | Nested trong Outcome/Evaluation; raw decode cần kiểm kind/id/field lạ |
| `OutcomeRecord` | `outcome-record.schema.json` | M03/M05 eval và M10/M11 runtime; còn internal ActionID alias cần audit |
| `AdvisorOutput` | `advisor-output.schema.json` | BR-03a: eval raw JSON và `advisor-check OUTPUT.json EVIDENCE.json AS_OF MAX_AGE_HOURS`; output envelope chứa artifact đã kiểm |
| `AdvisorEvidence` | Không có schema riêng | Context rút gọn eID/time/source, không phải Observation canonical |
| `EvaluationRecord` | `evaluation-record.schema.json` | M05 eval; required arrays/serialized output còn cần kiểm |
| `ImprovementProposal`, `ReviewRecord` | `improvement-proposal.schema.json`, `review-record.schema.json` | M05 eval/demo; raw boundary còn mở |
| `CanonicalObservation` | `observation.schema.json` | M06 normalizer; output thêm domain/correlation fields cần đối chiếu |
| `ToolSpec` | `tool-registry.schema.json` | Phần tử registry M07, không phải toàn registry envelope |
| `AgentProposal` | Không có schema canonical riêng | M07 internal proposal/tool-call model; không gọi nó là AdvisorOutput |
| `ShadowActionIntent`, `ShadowPolicyDecision` | `action-intent.schema.json`, `policy-decision.schema.json` | M08 eval/demo; mode/authority đã có semantic guard, schema conformance còn cần chứng minh |
| `ApprovalRecord`, `ExecutionAuthorization`, `ExecutionRecord` | `approval-record.schema.json`, `execution-authorization.schema.json`, `execution-record.schema.json` | M09 eval/state/demo; required fields/unknown properties tại raw boundary còn mở |
| `CanaryGrant`, `CanaryCostBound`, `CanaryLedger`, `CanaryGateDecision` | `canary-grant.schema.json`, `trusted-cost-bound.schema.json`, `canary-ledger.schema.json`, `canary-gate-decision.schema.json` | M10; alias nội bộ/canonical mapping cần test cụ thể |
| `CanaryExecutionAuthorization`, `CanaryExecutionRecord` | execution authorization/record schemas | M10 output chuyên biệt; chưa claim khớp hoàn toàn |
| `ProductionLease`, `ProductionLeaseApproval`, `ProductionHealthSnapshot` | production lease/approval/health schemas | M11 typed input/state; kiểm serialized fields là phần còn lại |
| `ProductionLedger`, `ProductionGateDecision`, `ProductionCycleRecord` | production ledger/gate/cycle schemas | M11 output/state; chưa claim khớp hoàn toàn |
| `ProductionExecutionAuthorization`, `ProductionExecutionRecord` | execution authorization/record schemas | M11 output chuyên biệt; cần kiểm mapping |
| `ProductionReconciliationResolution`, `productionActivationRecord` | production reconciliation resolution/activation schemas | M11 recovery/activation; cần audit raw boundary |
| Chưa xác minh type cho grant approval | `canary-grant-approval.schema.json` | Cần xác định boundary thực dùng trong BR-03c |

## M02 — Quyết định bảo toàn và đề xuất adapter

Giữ `HistoryRecord.recorded_result` là projection của deterministic ranking (`Result`), giữ `formula_version` và input hash. BR-03b thêm adapter riêng và kiểm raw schema, không viết lại history cũ. Ranking rỗng ở record mới xuất []; riêng ranked:null cũ được xem như array rỗng khi kiểm/replay. Các kiểm ID/hash/formula khác không bị bỏ. Xem [mapping và lệnh adapter](M02-DECISION-ADAPTER.md), [ngoại lệ tương thích](../../contracts/README.md).

Adapter BR-03b từ Result + context do người cung cấp sang DecisionPacket giữ exact `decision_id/evidence_ids`, giữ state; context không override action hoặc ID. `action` luôn null; reasons nối newline, các array copy độc lập và không null trong packet. Field domain ranked không lọt vào canonical packet. Tests kiểm output serialize, mutation input, legacy replay và file không đổi; nội dung supported_facts vẫn cần người review.

## Phần còn lại của BR-03

BR-03c.3 đã review/merge #30 `891562a`, CI 4/4 PASS. BR-03c.4 M07: [boundary](M07-JSON-BOUNDARY.md), nhánh `codex/br-03c-m07-boundary`, Codex thực hiện/chủ repo chưa review; canonical registry toàn mảng, proposal shape local, không model/network/store.

BR-03c.2 đã review/merge #29 tại `7a0f6c4`, CI 4/4 PASS. BR-03c.3 M06 đang triển khai trên nhánh `codex/br-03c-m06-boundary` (Codex; chủ repo chưa review): [output schema/profile](M06-JSON-BOUNDARY.md), provenance synthetic của harness và ID binding; không nối source parser/store/n8n.

BR-03c.1 đã review/merge #28 tại `d89a04c`, sau #27 `09a2f50`. BR-03c.2 (Codex, reviewer chủ repo chưa review) đang triển khai [M05 boundary](M05-JSON-BOUNDARY.md) trên nhánh `codex/br-03c-m05-boundary`: schema-first Evaluation/Proposal/Review và cặp M03 qua m05-check, strict IDs/timeline, output schema, không apply. Notes tùy chọn được giữ; risks nil/rỗng bỏ khi serialize thay vì null. Không claim M06–M11 đã có raw conformance.

BR-03c.1 đang triển khai trên nhánh `codex/br-03c-m03-boundary` (Codex; reviewer chủ repo, chưa review). [Boundary M03](M03-JSON-BOUNDARY.md) nối ActionRecord/OutcomeRecord raw schema + strict decode vào lệnh m03-check và raw eval; output serialize được kiểm lại. Alias M11 không nhận ở boundary này. Chưa đóng các dòng M05–M11; typed validators cũ không trở thành raw schema boundary chỉ nhờ thêm lệnh mới.

- BR-03a: Advisor boundary, eval và lệnh kiểm file — PR #22 đã merge tại `487ed73`.
- BR-03b: M02 adapter + schema conformance tests cho capture/serialized history/replay — đã merge #24. Schema package pin jsonschema/v6 v6.0.2, embed canonical files và chặn loader ngoài, bật format assertions; CI có tests và smoke export.
- BR-03c: lựa chọn/pin bộ kiểm JSON Schema đầy đủ nếu cần; kiểm output thực và raw boundaries của các artifact còn lại ở bảng; không nới schema hoặc xóa semantic guards để lấy PASS — TODO.

Chỉ đánh dấu BR-03 DONE khi BR-03b/c có implementation/evidence và được review; bảng mapping hoặc unit test M04 không thay được các phần đó.
