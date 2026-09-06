# Shared core — BR-08 / BR-09

BR-08 đã nghiệm thu scoped sau #47–#51. BR-09 bổ sung `m00.Convert` và `m00.SourceFields`: chuyển packet JSON thành M02 input, kiểm lại projection/provenance, không I/O trong core. [Mapping và giới hạn](../examples/m00-import/README.md); BR-09 còn chờ review.

Module import được từ harness và learner, có `m00` và `m03`. Implementation M03 được chuyển từ cmd/demo, không có bản sao validation thứ hai. Package m03 nhận bytes/types, không đọc file, gọi clock, ghi store hoặc gọi executor. contracts vẫn là nguồn schema.

API: HumanActionRecord, EffectRef, OutcomeRecord; ValidateHumanActionRecord, ValidateEffectRef, ValidateOutcomeRecord, ValidateActionOutcomeLink; DecodeM03Action, DecodeM03Outcome, CheckM03Pair. Status giữ tương thích. VALID là schema/semantic của pair, không resolve decision store hoặc cấp quyền.

OutcomeRecord.ActionID là field tương thích typed `json:"-"` cho harness cũ, không là input canonical; DecodeM03Outcome vẫn từ chối action_id và effect_ref sai. Chưa đổi semantics typed legacy hoặc import M05/M11/executor vào core.

Từ thư mục core: `GOWORK=off go test ./...` và `GOWORK=off go vet ./...`. Module Go 1.23, replace contracts về ../contracts; không cần go.work. Test package m03_test có expected literal độc lập; fixture/eval harness cũ giữ nguyên, không tính expected bằng wrapper.

Learner CLI mới thuộc BR-08c, store seam thuộc BR-08d. Không đổi history M02, learner go.mod hoặc state sandbox. [ADR-003](../docs/architecture/ADR-003-SHARED-CORE-CLI-STORE.md).
