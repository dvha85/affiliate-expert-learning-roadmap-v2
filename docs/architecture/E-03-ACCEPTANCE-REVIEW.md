# E-03 — Review nghiệm thu BR-03

E-02b và review E-03: [PR #46](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/46), commit triển khai `94d5c1c`. Tests/vet ba module, 8 validators và 10 Python regressions local PASS. Chưa merge; quyết định đóng item cha chờ review PR.

Baseline `2cf3bc1` sau review/merge #45; review này gồm thay đổi E-02b cùng PR, chưa merge. Codex thực hiện review kỹ thuật; chủ repo review PR trước khi đổi item cha sang DONE.

## Quyết định đề xuất

**READY_FOR_REVIEW trong phạm vi lab/schema**, không phải Operated hoặc production-ready. E-01 đã merge #44 `1c68ad2`; E-02 đã merge #45 `2cf3bc1`; E-02b nối export boundary và regression ở PR này. Chỉ đề nghị đóng BR-03 sau review/merge E-02b và chấp thuận phạm vi bên dưới; hiện vẫn IN_PROGRESS.

| Tiêu chí BR-03 | Evidence và kết luận |
|---|---|
| Raw required/enum/null/array/unknown/duplicate/time | Ma trận 29 schema trong BR-03e; boundary tests M03–M11; contracts engine pin. Không suy mọi internal struct là canonical |
| Output thực kiểm schema tương ứng | M02 history/adapter; CLI checks M03–M08; M09 persistence; E-02 output matrix cho executor/gate/authorization M09–M11; E-02b kiểm byte snapshot tại mission-demo M10/M11 |
| Không gắn nhãn diagnostic thành canonical | TestGovernedDiagnosticExport: empty, thiếu cost/hash, invalid time, nil state và missing activation; result chỉ chứa diagnostic, execution_permitted=false; bỏ cả authorization/authority khi gate sai |
| Canonical output và fail closed | TestGovernedCanonicalExport: decode từng raw artifact đúng profile; sibling authorization null gây lỗi trước có bytes; không sửa schema hoặc điền fake IDs |
| Precision | #44 timestamp_precision_test: nanosecond expiry và window qua persist/load; không đổi file cũ ngầm |
| History compatibility | BR-03b history/decision adapter tests tiếp tục PASS; ngoại lệ ranked:null legacy giữ nguyên |
| Schema không thay trust/semantics | Các chain summary unverified; export không cấp quyền hoặc gọi executor; H-01–H-04 tiếp tục mở |

## Phạm vi chấp thuận cần rõ

1. E-02 kiểm các output shape qua đường runtime và builder chung; không phải branch coverage 100% hoặc fault-injection mọi Write/Sync/Close. Các nhánh lỗi dùng builder chung được truy vết trong ma trận; test crash/syscall thuộc H-03.
2. E-02b bảo vệ serializer mission-demo M10/M11 hiện hữu. Typed return của Evaluate/Enforce vẫn có thể là diagnostic; caller mới phải dùng export boundary, không marshal gate trực tiếp rồi tuyên bố canonical. Không thay mọi API typed bằng schema engine và không tạo endpoint live mới.
3. Gate đạt schema không xác thực grant/lease/health hay liên kết với sibling: export là conformance, không chain authorization. Gate không đạt schema xuất thông tin chẩn đoán; reason/decision là dữ liệu báo lỗi, không quyền thực thi.
4. Canonical cycle là artifact caller cung cấp được kiểm bởi chain/validator, chưa có issuer production độc lập. Không thêm issuer chỉ để tuyên bố coverage.
5. Output khác zero không phải lúc nào là canonical; zero sentinel/diagnostic không nằm trong lời hứa artifact schema. E-02b không biến lỗi input thành ID giả.

## Backlog không đóng cùng BR-03

H-01 nguồn trust/clock; H-02 rollback/migration; H-03 crash/concurrency; H-04 source/store và operated evidence giữ mở như BR-03e. BR-06b vẫn chưa có chương trình/kênh. Không có chứng nhận chạy affiliate thật từ tests sandbox.

## Quy trình review cuối

Chạy tests/vet ba module, 8 validators, 10 Python regressions và CI của head PR. Review diff E-02b cùng ma trận; nếu đồng ý phạm vi thì merge và cập nhật BR-03 DONE với link commit/evidence trong bước quản lý tiếp. Nếu yêu cầu bảo vệ mọi caller typed hoặc fault injection toàn syscall thì tạo scope riêng, không ghi nhận sai là đã có bằng chứng.
