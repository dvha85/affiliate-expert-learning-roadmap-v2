# Bộ khởi đầu — M06

Bài kiểm [m06-check offline](../../docs/architecture/M06-JSON-BOUNDARY.md) xuất Observation synthetic/test đã kiểm schema, không fetch/persist. Bài này không thay workflow thật và các bước Mission bên dưới.

1. Học `curriculum/M06/`.
2. Đọc `CHECKPOINTS.md` và dùng `M06-OPERATED-EVIDENCE-TEMPLATE.md`.
3. Chạy `cd lab/mission-runtime && go test ./...` và `go run ./cmd/demo M06`.
4. Import `lab/n8n/M06-readonly-watcher.blueprint.json`.
5. Theo [bài tập M06.3](../../curriculum/M06/M06.3-n8n-readonly-workflow.md):
   fixture đầu `APPENDED`, retry nguyên event `EXACT_DUPLICATE`; sửa body giữ
   identity cũ phải `HANDOFF_ERROR`, history bytes không đổi. Event mới có
   correlation/time hợp lệ mới được append record mới. Không đổi URL sang
   source public; selected-source là profile và operated evidence riêng.
6. Phân biệt change detection với persistence. `NEW/UNCHANGED/CHANGED` là state
   so sánh nội dung/identity; adapter persistence trả `APPENDED` hoặc
   `EXACT_DUPLICATE`, còn conflict/reject trả lỗi như `HANDOFF_ERROR`. Sửa body
   mà giữ identity cũ không phải `CHANGED`; history bytes phải không đổi.
7. Chạy riêng core change-detection cases có previous_hash để thấy
   `NEW/UNCHANGED/CHANGED`; không mong các state này từ report của blueprint.
   Output workflow là persistence report; resolve Observation từ record_id/history. Khi chạy
   adapter HTTP, cấp `CANONICAL_ADAPTER_TOKEN` qua môi trường và gửi Bearer
   header; adapter và n8n phải dùng cùng token. Không commit token hay đưa token
   vào workflow JSON.
8. Setup shell/cwd/history, token ngẫu nhiên dùng chung và n8n home cô lập nằm ở
   [M06.3](../../curriculum/M06/M06.3-n8n-readonly-workflow.md). Lưu evidence thật
   dưới `learner/M06/`; ghi fixture rehearsal riêng.

Không dùng credential có write scope. Content hash/change state không phải business truth.
