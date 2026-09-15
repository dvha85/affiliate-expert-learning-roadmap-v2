# Deterministic runtime re-run — 2026-09-16 — `main` `2e86636`

## Scope

Đây là maintainer-run local verification của đúng chuỗi `deterministic-runtime`
đang khai báo trong `.github/workflows/curriculum-ci.yml`. Run dùng learner Bot
được build một lần trong thư mục tạm, sau đó dùng cùng binary cho M01 baseline
và toàn bộ M02 capture/list/replay/decision. History nằm trong workspace tạm;
không thay đổi repository hoặc dữ liệu n8n đang chạy.

Run chỉ dùng fixture local và giữ boundary `synthetic`/`read-only`. Không dùng
credential ACCESSTRADE, không gọi provider, live executor, business outcome,
clean-machine pilot, target-host deployment, multi-host hay power-loss.

## Environment and result

- `run_id`: `deterministic-runtime-rerun-20260916-2e86636`
- `repo_commit`: `2e86636` (`main`)
- `host`: local macOS arm64
- `result`: PASS cho toàn bộ command trong chuỗi dưới đây
- `wall_time`: khoảng `0.8s` trong local workspace có Go build/test cache nóng

Wall-clock local này chỉ mô tả lần chạy đã ghi; không phải benchmark GitHub
Actions và không chứng minh cache hit trên runner khác.

## Commands and results

| Check | Result |
|---|---|
| `go test ./internal/...` trong `lab/affiliate-bot` | PASS |
| `go test ./...` và `go vet ./...` trong `contracts` | PASS |
| `go build -o <tmp>/learner-bot ./cmd/bot` | PASS |
| `<tmp>/learner-bot` (M01 baseline) | PASS; `RANK_SCENARIO` |
| `<tmp>/learner-bot history capture <tmp>/m02-history.jsonl data/m02-sample-observations.json ci-demo ...` | PASS; `APPENDED` |
| `<tmp>/learner-bot history list <tmp>/m02-history.jsonl` | PASS; thấy `ci-demo` |
| `<tmp>/learner-bot history replay <tmp>/m02-history.jsonl` | PASS; `replay=MATCH` |
| `<tmp>/learner-bot history decision <tmp>/m02-history.jsonl ci-demo data/m02-decision-context.json` | PASS; packet có `"action": null` |
| `python3 scripts/audit_readiness.py` | PASS; `NOT_READY_FOR_PRODUCTION` |

Các artifact tạm dùng cho capture/list/replay/decision đều được tạo trong
`mktemp` directory và không được ghi vào input path của repository.

## Interpretation

Đường single-build và M01/M02 CLI smoke hiện chạy được trên commit này. Audit
vẫn giữ `NOT_READY_FOR_PRODUCTION`, với các gap BR-13, BR-14, BR-15, BR-16a,
BR-17, BR-18b và BR-19 còn partial/open. Record này chỉ xác nhận local
fixture/offline path; không nâng readiness hoặc thay thế bằng chứng provider,
business outcome, live executor, pilot hay deployment.
