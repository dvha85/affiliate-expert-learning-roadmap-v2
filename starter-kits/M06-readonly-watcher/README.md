# Bộ khởi đầu — M06

Bài kiểm [m06-check offline](../../docs/architecture/M06-JSON-BOUNDARY.md) xuất Observation synthetic/test đã kiểm schema, không fetch/persist. Bài này không thay workflow thật và các bước Mission bên dưới.

1. Học `curriculum/M06/`.
2. Đọc `CHECKPOINTS.md` và dùng `M06-OPERATED-EVIDENCE-TEMPLATE.md`.
3. Chạy `cd lab/mission-runtime && go test ./...` và `go run ./cmd/demo M06`.
4. Import `lab/n8n/M06-readonly-watcher.blueprint.json`.
5. Giữ nguyên synthetic fixture profile `br13-offer-fixture/v1` và chạy ít nhất
   ba lần: fixture mới → `NEW`, cùng bytes → `UNCHANGED`, rồi đổi một field
   được phép → `CHANGED`. Không đổi URL sang source public trong starter này;
   selected-source metadata là profile riêng, cần adapter/policy/operated
   evidence tương ứng.
6. Kiểm output là canonical Observation và có correlation/history. Khi chạy
   adapter HTTP, cấp `AFFILIATE_ADAPTER_TOKEN` qua môi trường và gửi Bearer
   header; không commit token hay đưa token vào workflow JSON.
7. Lưu operated evidence dưới `learner/M06/`.

Không dùng credential có write scope. Content hash/change state không phải business truth.
