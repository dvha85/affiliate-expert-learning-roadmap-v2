# Bộ khởi đầu — M07

Tham khảo [m07-check offline](../../docs/architecture/M07-JSON-BOUNDARY.md) để kiểm ba file trước integration; không thay model/tool trace thật.

1. Học `curriculum/M07/`.
2. Đọc `CHECKPOINTS.md` và dùng `M07-OPERATED-EVIDENCE-TEMPLATE.md`.
3. Chạy `cd lab/mission-runtime && go test ./...` và `go run ./cmd/demo M07`.
4. Import `lab/n8n/M07-readonly-evidence-agent.blueprint.json` và cấu hình model credential cục bộ trong n8n.
5. Giữ Tool Registry read-only/GET-only + host allowlist; không nối write tool.
6. Chạy normal case, unknown/write case, hallucinated evidence và prompt-injection/tool-output case.
7. Đối chiếu Agent output với offline evaluator; output vẫn A2-RO/HUMAN_REVIEW.
8. Lưu operated evidence dưới `learner/M07/`.

Không commit secret/API key. Agent proposal không phải authorized ActionIntent.

Ghi chú: blueprint chỉ mở quyền đọc; mọi quyền hành động có hậu quả vẫn thuộc các Mission sau.
