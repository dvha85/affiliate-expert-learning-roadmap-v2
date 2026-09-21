# Shared core — BR-08 → M11

BR-13a bổ sung `m06`: normalizer/hash/identity và raw fixture boundary dùng chung với harness. Không fetch, không history persistence hoặc trusted allowlist. [Phạm vi và giới hạn trước network/handoff](../docs/architecture/BR-13A-SHARED-M06.md).

BR-12a bổ sung `m05`: EvaluationRecord/ImprovementProposal/ReviewRecord và raw `CheckM05Chain` dùng chung với harness qua aliases/wrappers. Không I/O, không resolve history store hoặc cấp execution permission. [Phạm vi và các bước tiếp theo](../docs/architecture/BR-12A-SHARED-M05.md).

BR-08 đã nghiệm thu scoped sau #47–#51. BR-09 bổ sung `m00.Convert` và `m00.SourceFields`: chuyển packet JSON thành M02 input, kiểm lại projection/provenance, không I/O trong core. [Mapping và giới hạn](../examples/m00-import/README.md); BR-09 đã nghiệm thu lab sau merge #52 `5fa86b9`, không là live proof.

Module hiện được import từ harness và learner cho `m00`, `m03`–`m11`. Các package là pure/shared validation: nhận bytes/types, không đọc file, gọi clock, ghi store hoặc gọi executor; contracts vẫn là nguồn schema. M10/M11 chỉ tạo/kiểm artifact authority và graph, không tự cấp quyền thực thi cho caller.

API chính theo module:

- `m00`: `Convert`, `SourceFields` cho projection/provenance.
- `m03`: `HumanActionRecord`, `EffectRef`, `OutcomeRecord`, các validator/decoder và `CheckM03Pair`.
- `m04`: `AdvisorOutput`, `AdvisorEvidence`, `EvaluateAdvisorOutput`, `DecodeAdvisorOutput`.
- `m05`: evaluation/proposal/review records, validators và `CheckM05Chain`.
- `m06`: watch/fixture normalizers, canonical content hashes và selected-source profile builders.
- `m07`: read-only registry/tool-result/proposal validation and grounding helpers.
- `m08`: `Intent`, `PolicyDecision`, `DecodeIntent`, `EvaluatePolicy`, `ComputeIntentHash`.
- `m09`: approval/authorization/execution decoders and historical-chain validation.
- `m10`: trusted cost bounds, canary gate/authorization/execution records and artifact-graph validation.
- `m11`: production lifecycle artifact decoding, exact graph validation and historical-chain checks.

`VALID`/`SUPPORTED` chỉ là schema/semantic kết quả tương ứng; chúng không resolve decision store, invoke a provider, or grant execution permission.

OutcomeRecord.ActionID là field tương thích typed `json:"-"` cho harness cũ, không là input canonical; DecodeM03Outcome vẫn từ chối action_id và effect_ref sai. Compatibility aliases không được xem là canonical JSON hoặc authority.

Từ thư mục core: `GOWORK=off go test ./...` và `GOWORK=off go vet ./...`. Module Go 1.23, replace contracts về ../contracts; không cần go.work. Test package m03_test có expected literal độc lập; fixture/eval harness cũ giữ nguyên, không tính expected bằng wrapper.

Learner CLI mới thuộc BR-08c, store seam thuộc BR-08d. Không đổi history M02, learner go.mod hoặc state sandbox. [ADR-003](../docs/architecture/ADR-003-SHARED-CORE-CLI-STORE.md).
