# BR-05 — Evidence quickstart

- Người thực hiện: Codex; reviewer: chủ repo, chưa review. Tài liệu: [QUICKSTART](../../../curriculum/BOOT/QUICKSTART.md).
- [PR #25](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/25), commit tài liệu/harness `59fb37b`; [job CI theo commit](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/25/checks) kiểm lại trên Ubuntu. Phụ thuộc PR #24.
- Đối tượng smoke: code commit `4e62eb7cc70d936883994f7c93eabeb6a8ce1fe2` (BR-03b), clone mới từ checkout local; Go 1.27.0 darwin/arm64, Git 2.50.1 Apple Git-155; Python 3.9.6 để chạy harness.
- Lệnh: `python3 scripts/smoke_quickstart.py --go /usr/local/go/bin/go`. Người có go trong PATH không cần `--go`.
- Mỗi lần dùng thư mục clone, GOCACHE và GOMODCACHE mới, GOTOOLCHAIN=local; dependency pin tải lại. Không dùng lại build cache để tạo PASS giả. Clone local kiểm source/working tree, không giả làm kiểm network clone GitHub hoặc installer.

## Kết quả quan sát

| Bước | Kết quả |
|---|---|
| Clone / git status / go version | Clone sạch; Go 1.27.0 darwin/arm64 |
| go mod download | Exit 0 với module cache rỗng |
| go run ./cmd/bot | Synthetic; Product B 9.60, A/C 8.00; RANK_SCENARIO |
| go test ./... / go vet ./... | Exit 0 |
| Sửa ProductName < thành > trong clone tạm | `cmd/bot/main.go` dòng 150 tại commit này; test TestSameInputSameOutput exit 1, đúng assertion A-first tie break |
| Sửa lại và chạy full tests | Exit 0; tests contracts và mission-runtime cũng PASS |
| go run ./cmd/demo O00 | external_side_effects=false; final_state=DRY_RUN_ONLY |
| git status cuối | Không thay đổi trong clone tạm; clone nguồn không bị sửa |

Harness dùng timeout 300 giây mỗi lệnh, fail ngay khi exit code/marker sai. `finally` trả lại code trong clone tạm sau deliberate failure; thư mục tạm được dọn khi kết thúc. Không xóa/reset repo nguồn hoặc dữ liệu học viên.

## Phạm vi chưa chứng minh

- Không cài lại Git/Go/editor trên máy chưa có công cụ; không dùng smoke để claim installer GUI đã kiểm.
- Windows/PowerShell chưa được chạy; chỉ có lệnh tham khảo, không support claim.
- Linux chỉ có bằng chứng khi job CI `Beginner quickstart isolated-clone smoke` của PR/commit PASS; không suy macOS PASS thành Linux PASS.
- Chưa có pilot độc lập với một người mới; BR-16 vẫn cần kiểm người học không phải hỏi thêm lệnh.
- Không có market/affiliate/Reality/Operated evidence. BR-05 chỉ giải quyết đường vào môi trường kỹ thuật và tài liệu.

Nghiệm thu BR-05 giữ IN_REVIEW cho người review xác nhận phạm vi hỗ trợ; không đánh dấu DONE chỉ từ bảng này.
