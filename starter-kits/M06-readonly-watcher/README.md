# Bộ khởi đầu — M06

1. Học `curriculum/M06/`.
2. Đọc `CHECKPOINTS.md` và dùng `M06-OPERATED-EVIDENCE-TEMPLATE.md`.
3. Chạy `cd lab/mission-runtime && go test ./...` và `go run ./cmd/demo M06`.
4. Import `lab/n8n/M06-readonly-watcher.blueprint.json`.
5. Thay source bằng public/allowlisted source; giữ GET-only; chạy ít nhất hai lần để chứng minh `NEW/UNCHANGED/CHANGED`.
6. Kiểm output là canonical Observation và có correlation/history.
7. Lưu operated evidence dưới `learner/M06/`.

Không dùng credential có write scope. Content hash/change state không phải business truth.
