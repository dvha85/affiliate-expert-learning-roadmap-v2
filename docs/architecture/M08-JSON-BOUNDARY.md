# BR-03c.5 — ActionIntent/PolicyDecision M08

Conformance offline, không executor/approval/store. Intent qua canonical schema + strict decode trước EvaluateShadowPolicy; PolicyDecision serialize được kiểm canonical schema trước xuất. Không nhận alias shadow_only/dry_run/approval_required nội bộ từ JSON. Không gọi SealShadowActionIntent để sửa hash/quyền của input.

Từ repo root:

```text
cd lab/mission-runtime
go run ./cmd/demo m08-check testdata/m08-intent.json testdata/m08-context.json
go test ./cmd/demo -run TestM08 -count=1 -v
```

Fixture synthetic đã seal sẵn, clock cố định 03/09/2026 không phải thời gian vận hành hiện tại. Output intent giữ PROPOSAL_ONLY/false; policy ALLOW vẫn NON_AUTHORIZING/false. Parameters giữ json.Number tránh làm tròn số lớn; hash vẫn theo hàm hiện hữu, không tự chuyển hash cũ.

Context là shape local, không có canonical schema: policy_version/now và toàn bộ arrays/maps phải có mặt, không null. Lists từ chối ID rỗng/whitespace/trùng; maps chỉ nhận string, action_risk thuộc RISK0/1/2; key lạ/trùng bị chặn. Context do người cung cấp, chưa phải trusted policy/clock/store.

Schema chặn required/null/enum/const, hash sai hình dạng, evidence IDs trùng/rỗng và agent thiếu proposal_ref. Sau đó semantic runtime giữ kiểm hash thật, decision/evidence/proposal links, thời gian, host, idempotency. Decoder không tự cấp compatibility aliases.

CLI xuất envelope intent/policy cho mọi policy result hợp lệ, kể cả DENY/WAIT/GET_MORE_DATA/HUMAN_REVIEW. Exit 0 chỉ nói phép kiểm hoàn tất, không phải ALLOW hay approval. Schema/context/I/O/output lỗi trả exit 1, không có success envelope; writer error truyền về. Schema áp dụng từng artifact, không phải cả envelope.

## Bài tập

Copy hai fixture vào thư mục bài tập và thay path lệnh. Đổi target sau seal → DENY/TAMPERED_INTENT, không tự sửa hash. Đổi execution_authorized=true → INVALID_SCHEMA. Bỏ syn-d khỏi known_decision_ids → GET_MORE_DATA/MISSING_DECISION_LINK. Đổi context.now bằng expires_at → DENY/EXPIRED_INTENT. Khôi phục từng sửa đổi rồi kiểm lại. Không dùng ALLOW để gọi executor hoặc cập nhật PROGRESS.

## Evidence và giới hạn

Tests kiểm required/null intent/policy, duplicate/unknown/authority, agent conditional proposal_ref, policy review flag, số lớn parameters, hash tamper, risk levels, missing links, expiry, idempotency và CLI/writer. Typed eval M08 cũ giữ nguyên và vẫn seal fixture; không coi bước seal đó là kiểm JSON gốc.

Không resolve IDs với store/payload, xác thực người đề xuất/policy owner, persist seen_idempotency hoặc enforce network allowlist. Hash là consistency, không phải chữ ký/quyền. Không có file size quota cho endpoint công khai; chỉ file người vận hành chọn. BR-17 và M09–M11 còn mở, không thay Reality/Operated proof.
