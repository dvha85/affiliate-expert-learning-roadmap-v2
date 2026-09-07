# BR-12a — M05 dùng chung trước khi nối store

Baseline `9862d9a`. BR-11 đã nghiệm thu lab/fixture sau #68. Lát cắt này chuyển types, semantic validation và raw chain boundary từ harness sang `core/m05`; không sao chép thêm một bộ quy tắc vào learner.

Đã review/merge #69 tại `1cad062`, head `db477a0`: core/harness/learner tests/race/vet, 8 validators và CI 4/4 PASS. Codex thực hiện/review theo quyền chủ repo; không có lỗi chặn trong phạm vi di chuyển implementation. Validator semantic được cập nhật đọc core/m05 thay vì chỉ tìm marker ở file cũ.

- `EvaluationRecord`, `ImprovementProposal`, `ReviewRecord` và các validator thuộc core; `CheckM05Chain` kiểm schema tất cả input trước semantic, link/time/duplicate và luôn execution_authorized=false.
- Demo giữ type aliases/wrappers để M05 và M11 cũ tiếp tục dùng đúng implementation. File I/O, CLI và output serialization vẫn ở demo. Core không đọc store, mạng, key hay clock.
- Contracts vẫn là nguồn schema; không thay required fields, trạng thái hoặc quyền. `ActionID` nội bộ `json:"-"` giữ tương thích M11, không trở thành trường JSON canonical.
- Tests trực tiếp qua package public có expected literal, kiểm INCONCLUSIVE/PENDING, approval không execution, orphan, thời gian, auto_apply, machine review và unknown/duplicate fields. Toàn bộ regression harness cũ tiếp tục chạy qua wrappers.

Giới hạn quan trọng: raw chain checker chưa resolve decision với history thật của bot; typed validators không thay raw schema boundary. Validator hiện kiểm cấu trúc/liên kết, không tự suy ra kết luận từ metrics hoặc bảo đảm người nhập INCONCLUSIVE khi thiếu dữ liệu. Không tuyên bố semantic hardening mới trong PR di chuyển code.

## Các bước BR-12 tiếp theo

1. BR-12b: evaluation CLI/store riêng trong learner, load/replay history và exact action/outcome, bảo vệ orphan/time/duplicate/conflict/restart; tạo INCONCLUSIVE hoặc NEEDS_MORE_DATA khi dữ liệu ít/pending, không tự coi zero là hiệu quả. Không coi kết quả test là business outcome.
2. BR-12c: proposal có version/benefit/risk/rollback gắn evaluation đã lưu; import human review với reason/time/link, auto_apply=false; approval chỉ để người sửa thủ công.
3. BR-12d: walkthrough M00 → M05 cùng workspace, bài thay đổi nhỏ có FAIL/PASS regression, diff/review/rollback và nghiệm thu riêng. Giữ kết quả đo tách khỏi kết quả test.

BR-12 cha IN_PROGRESS; BR-12a chưa tạo learner evaluation store/CLI và chưa hoàn tất checklist M05. Không gọi API hoặc thay ledger DeepSeek.
