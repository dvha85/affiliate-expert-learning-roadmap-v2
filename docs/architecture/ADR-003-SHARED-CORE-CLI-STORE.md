# ADR-003 — Shared core, CLI và quyền sở hữu store

Status: Proposed — BR-08a, chờ review; chưa triển khai refactor. Baseline khảo sát: main `8c7c86d` sau nghiệm thu BR-03 lab/schema. Không thay ADR-001/002 hoặc authority của learner Bot.

## Bối cảnh và bằng chứng

| Hiện trạng | Source / hệ quả |
|---|---|
| Ba module riêng | contracts/go.mod (Go 1.23), lab/mission-runtime/go.mod (`missionlab`, Go 1.23), lab/affiliate-bot/go.mod (Go 1.27). Hai lab require contracts v0.0.0 + replace ../../contracts; chưa có shared business package |
| Learner baseline M01/M02 | [main.go](../../lab/affiliate-bot/cmd/bot/main.go) chứa evaluate, types và CLI; [history.go](../../lab/affiliate-bot/cmd/bot/history.go) chứa NewHistoryRecord/LoadHistory/AppendHistory/Replay; [history_schema.go](../../lab/affiliate-bot/cmd/bot/history_schema.go) kiểm boundary |
| Capability M03–M11 chưa import được trực tiếp | lab/mission-runtime/cmd/demo là package main. [m03_m05.go](../../lab/mission-runtime/cmd/demo/m03_m05.go) trộn type/semantic M03 và M05; [m03_boundary.go](../../lab/mission-runtime/cmd/demo/m03_boundary.go) ghép decode/semantic/file handler |
| CLI hiện tại khác vai trò | bot nhận observations file hoặc history capture/list/replay/decision; cmd/demo nhận Mxx hoặc m03-check…m11-chain-check. Demo M09–M11 có effect trong thư mục sandbox tạm, không phải tất cả đều read-only |
| History M02 là JSONL | AppendHistory kiểm toàn bộ existing records/ID conflict rồi O_APPEND; chưa có transaction/CAS cho nhiều writer. Không tự gọi đây là store đa tiến trình an toàn |
| State M09–M11 thuộc sandbox/harness | m09_persistence.go, m10_persistence.go, m11_persistence.go có boundary riêng và marker/ledger; không thay thế canonical history của learner |
| CI đang kiểm từng lab | .github/workflows/curriculum-ci.yml và mission-agent-path-ci.yml chạy tests, eval packs, demo smoke; thay đường source có thể ảnh hưởng cả scripts static validation |

Theo [continuity](LEARNER-BOT-CONTINUITY.md), learner phải mở rộng một Bot, harness không thành Bot thứ hai; n8n cache không là history.

## Lựa chọn

| Phương án | Đánh giá |
|---|---|
| Copy M03–M11 sang affiliate-bot | Loại: hai implementation dễ lệch semantics, tests parity không chữa nguồn trùng |
| Gộp repo thành một module ngay | Hoãn: đổi build/CI/tutorial/toolchain rộng, rủi ro vượt lát cắt M03 |
| Public package trong missionlab | Khả thi nhưng buộc learner phụ thuộc module tên/role harness; dễ kéo executor/demo state vào dependency |
| Module core riêng, extract từng capability | Chọn đề xuất: ranh giới import rõ, giữ hai entrypoint mỏng và tests conformance độc lập; đánh đổi thêm một go.mod/replace cần CI kiểm |

## Quyết định đề xuất

Thêm module `core/`, module path `github.com/dvha85/affiliate-expert-learning-roadmap-v2/core`, Go 1.23 cho lát cắt đầu (không dùng API mới hơn). Đây là đường dẫn **đề xuất, chưa tồn tại**. Không đổi directive 1.27 của affiliate-bot trong ADR này. Không cần go.work để build; local replace giống contracts, CI phải chạy được từng module độc lập từ clean checkout.

Import graph mục tiêu (mũi tên là import):

```text
affiliate-bot/cmd/bot → affiliate-bot/internal/app → core/m03 → contracts
mission-runtime/cmd/demo ────────────────────────→ core/m03
affiliate-bot/internal/store → contracts (+ core types khi cần)
```

Core không import cmd, app/store, harness, n8n, HTTP client hoặc provider SDK; contracts không import core. Core nhận bytes/context/instants tường minh, không đọc file, getenv, clock hay tự ghi dữ liệu. File/stdio/status mapping thuộc adapter CLI. Không tạo internal package trong module A rồi import nó từ module B.

### Lát cắt M03 trước

Extract types và pure validation/decode cần cho ActionRecord + OutcomeRecord + EffectRef, CheckM03Pair; chuyển ownership implementation, không copy. Tách phụ thuộc M05 trong m03_m05.go bằng graph symbol thực tế trước commit. Harness có thể giữ alias/delegating wrapper tạm để tránh sửa mọi test cùng lúc; wrapper không chứa bản sao logic. Alias ActionID nội bộ M11 phải được bảo toàn tương thích riêng, canonical raw boundary vẫn từ chối alias.

Giữ eval JSON và expected statuses trong harness; tests không tạo expected bằng chính hàm đang test. Core có unit tests độc lập, harness chạy fixture golden/negative từ evals, learner có CLI integration tests. Nếu type alias kéo cả M05/executor vào core, dừng lát cắt và tách type tối thiểu; không đưa toàn cmd/demo sang core.

### Store ownership và tương thích

- Learner app là owner duy nhất của canonical workspace; adapter store thực hiện I/O theo request đã validate. Core chỉ trả artifact/validation, không ghi store. Caller truyền path tường minh, không tự chọn home/cwd để ghi.
- BR-08b/c không đổi JSONL M02, record IDs, input hash, formula version, decision adapter, thứ tự records hoặc ngoại lệ ranked:null legacy. Không reseal/rewrite history cũ; replay fixture phải giữ MATCH. Duplicate/conflict semantics phải giữ.
- Chưa tạo action/outcome database ở BR-08a–c. BR-08d định nghĩa interface/store adapter nhỏ cho M02 và hướng dẫn ownership; BR-10 mới thiết kế append Action/Outcome và linkage với decision. Không nhét record loại mới vào JSONL history M02 khi chưa có contract/version migration.
- M09–M11 ledger/activation/marker vẫn là state sandbox tương ứng; không rename/migrate/delete chúng khi extract M03. n8n static data chỉ cache.
- Một writer theo quy trình lab là giới hạn, không bảo đảm locking đa tiến trình. Nếu có nhu cầu đa writer hoặc thay storage, cần ADR bổ sung + migration/backup/recovery; H-02/H-03 không bị coi đã giải quyết.

## CLI mục tiêu và ranh giới side effect

Các lệnh mới dưới đây **chưa tồn tại**, không đưa vào quickstart như lệnh chạy được trước PR implementation.

| Lệnh | Hành vi dự kiến | Thời điểm |
|---|---|---|
| `bot action validate ACTION.json OUTCOME.json` | M03 pair, read-only; file người học, không fixture hard-code | BR-08c |
| `bot history capture/list/replay/decision …` | Giữ lệnh/đối số cũ; capture là ghi, các lệnh khác theo hành vi hiện hữu | Tương thích xuyên migration |
| `bot action record …`, `bot outcome import …` | Persist record thủ công đã validate + resolve exact IDs, không publish | BR-10, không implement ở BR-08c |
| `bot demo …` | Fixture/sandbox được ghi nhãn; không dùng canonical workspace | Chỉ thêm sau khi spec riêng được review; lệnh cmd/demo cũ vẫn giữ |
| `bot execute …` | Không cung cấp đường consequential mới trong BR-08 | BR-17 và authority prerequisites |

CLI mới: stdout chỉ JSON một envelope có command/status, artifact canonical đã validate hoặc diagnostic riêng; stderr cho lỗi thao tác. Exit 0 = validation đạt (không là business truth/approval); 2 = usage; 1 = invalid input, semantic rejection hoặc I/O, có status phân biệt. Không bắt CLI cũ đổi exit/output ngay; compatibility tests khóa hiện trạng. Ghi file chỉ qua subcommand ghi rõ, path tường minh; validate không tạo thư mục hoặc mutate input.

CheckM03Pair không tự resolve DecisionPacket từ store; `VALID` phải nói rõ local schema/linkage giữa pair, không claim continuity đã operated. BR-09/10 thêm nối artifact thật. E-02b diagnostic/canonical export phải được giữ nếu di chuyển code M10/M11 về sau.

## Kế hoạch PR và nghiệm thu

| Bước | Phạm vi | Điều kiện đạt / rollback |
|---|---|---|
| BR-08a (PR này) | ADR + khảo sát + kế hoạch | Review lựa chọn module/import/store/CLI; không đổi Go/data |
| BR-08b | Tạo core/m03, extract tối thiểu, wrappers harness, CI module core | Eval expected độc lập không đổi; raw mutations/precision/aliases PASS; build từng module không go.work; revert PR không đụng store |
| BR-08c | Entrypoint learner action validate, adapter app, output contract/test | File input thật, invalid không có success artifact, stdout JSON/stderr/exit đúng; input bytes không đổi; old commands PASS |
| BR-08d | Store ownership seam M02 + hướng dẫn mở rộng workspace | Load/append/replay/decision giữ ID/hash/duplicate/conflict/legacy; lỗi store không reset; không migration dữ liệu; rollback code không sửa history |
| BR-08e | Review tích hợp/nghiệm thu BR-08 | Clean checkout smoke, fixture M03 qua bot và harness với expected độc lập, full regression/core CI, cập nhật curriculum/README; owner/path/limits rõ |

Mỗi PR kiểm tests/vet tất cả module bị ảnh hưởng và regression ba module hiện tại; chạy 8 validators + 10 Python regressions, CLI smoke cũ và link docs. Khi di chuyển source phải cập nhật static checks theo invariant, không bỏ guard để lấy PASS. Go minimum/core và toolchain CI phải được kiểm thực tế trong BR-08b, ADR không claim đã build core.

BR-08 chỉ DONE khi capability M03 nhận file qua learner entrypoint, core không duplicate, store owner/compatibility có tests và hướng dẫn đã đồng bộ. BR-09/10 mới hoàn thiện decision→action→outcome continuity; không chờ chương trình affiliate cho BR-08a nhưng không tự đóng BR-06b.

## Không thuộc ADR này

Không đổi schema, đưa SQLite/Postgres/provider SDK vào repo, migrate ledger/history, gọi affiliate API, bật Agent execution, gộp M09–M11 vào learner, thay PROGRESS hay đóng H-01–H-04. ADR chỉ Proposed tới khi được review/merge.
