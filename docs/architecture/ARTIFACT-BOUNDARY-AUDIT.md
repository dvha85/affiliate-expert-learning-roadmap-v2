# BR-03 — Đối chiếu artifact và boundary

## Phạm vi đã triển khai

BR-03a kiểm `AdvisorOutput` M04 tại JSON boundary và trước SUPPORTED. Đây **không phải** tuyên bố toàn bộ 29 schema đã được cưỡng chế ở runtime. BR-03b/c bên dưới còn mở; bảng này là bản đồ để review từng phần, không phải báo cáo conformance đầy đủ.

Chọn standard library Go cho contract nhỏ M04: kiểm JSON gốc, required/unknown/duplicate keys, null/array/item types, enum, constant false, unique/nonempty evidence IDs và reason. Không thêm dependency. `reason`/ID chỉ chứa khoảng trắng cũng bị từ chối (chặt hơn minLength của schema). Test khóa cấu trúc schema M04 để buộc review validator khi contract đổi. Không dùng giải pháp thủ công này như một JSON Schema engine tổng quát cho `$ref`, `allOf` hoặc các contract phức tạp.

Schema không thay semantic checks: ID phải resolve, nguồn không rỗng, context không có ID trùng ghi đè nhau, timestamp đúng RFC3339 và evidence được dùng không stale/future. `SUPPORTED` chỉ phản ánh các check hiện có, không xác minh nội dung ngôn ngữ có được evidence hỗ trợ. Không có quyền write hoặc quyết định kinh doanh tự động.

## Bản đồ Go ↔ schema ↔ boundary

Schema bên dưới nằm trong `contracts/`; type nằm trong module tương ứng. Trừ dòng M04, trạng thái là **cần audit/kiểm serialized output**, không suy từ tên type giống schema.

| Artifact / Go type | Schema | Input/output hiện tại và phần cần kiểm |
|---|---|---|
| `affiliate-bot.Observation` | `observation.schema.json` | M02 capture nhận file observations; phần domain/history dùng schema kết hợp |
| `affiliate-bot.HistoryRecord` | `history-record.schema.json` | capture/JSONL/list/replay; cần test raw input và serialized output theo schema |
| `affiliate-bot.Result` | projection trong `history-record.schema.json` | `recorded_result` không phải DecisionPacket; giữ dữ liệu replay hiện tại |
| Chưa có type DecisionPacket canonical ở bot | `decision-packet.schema.json` | M00 template; cần adapter được review, không ép cast Result |
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

Giữ `HistoryRecord.recorded_result` là projection của deterministic ranking (`Result`), giữ `formula_version`, input hash và thuật toán replay hiện hành. BR-03a không sửa schema/history bytes; test/replay cũ vẫn phải PASS. Đây là quyết định bảo toàn, **chưa hoàn thành adapter**.

BR-03b cần adapter tường minh từ Result + context do người cung cấp sang DecisionPacket: giữ exact `decision_id/evidence_ids`, map state theo enum canonical; câu hỏi, supported facts, assumptions, unknowns, next measurement phải có nguồn/căn cứ, không tự bịa để điền required field. `action` luôn null. Phải review mapping `reasons` → `reason`, arrays nil/rỗng và field domain `ranked` không được lọt vào canonical packet. Packet mới là artifact riêng, không ghi đè recorded_result; kiểm schema và replay fixtures trước/sau.

## Phần còn lại của BR-03

- BR-03a: Advisor boundary, eval và lệnh kiểm file — chờ review PR.
- BR-03b: M02 adapter + schema conformance tests cho capture/serialized history/replay — TODO.
- BR-03c: lựa chọn/pin bộ kiểm JSON Schema đầy đủ nếu cần; kiểm output thực và raw boundaries của các artifact còn lại ở bảng; không nới schema hoặc xóa semantic guards để lấy PASS — TODO.

Chỉ đánh dấu BR-03 DONE khi BR-03b/c có implementation/evidence và được review; bảng mapping hoặc unit test M04 không thay được các phần đó.
