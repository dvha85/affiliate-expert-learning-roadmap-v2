# Repository Governance — repo học chính v2

## 1. Authority

`CURRICULUM.md` là authority duy nhất cho sequence/evidence/authority/PASS. Repo lịch sử không được dùng để ghi đè behavior hiện hành.

## 2. Change path

```text
issue/spec
→ branch
→ Pull Request
→ required CI
→ human review
→ merge/reject
```

Development Agent có thể sửa code/docs/tests và mở PR nhưng không tự cấp merge authority hoặc production activation.

## 3. Required checks

Khuyến nghị bật hai aggregate required checks trên `main`. Chúng chỉ xanh khi
tất cả job con hiện hành thành công; `skipped`, `cancelled` hoặc thiếu job đều
là thất bại:

| Workflow | Tên check chính xác trên GitHub |
|---|---|
| `curriculum-ci.yml` — Curriculum CI | `curriculum-gate` |
| `mission-agent-path-ci.yml` — Mission Agent Path CI | `mission-gate` |

`curriculum-gate` bao phủ `structure-language-and-foundations`, `deterministic-runtime`,
Windows, hai learner-test shard, race, quickstart và ba nhóm smoke/mutation.
`mission-gate` bao phủ semantics/blueprints, mission-runtime và n8n engine.
Tên check là tên job, không phải tên step. Khi cấu hình, chọn đúng check do
GitHub Actions phát hành trên commit hiện hành. Không dùng các tên cũ
`curriculum`, `evidence-and-safety`, `python-regression` hoặc tự thêm tiền tố
workflow vào tên check.

Phần trên là contract của các gate; trạng thái cấu hình quản trị thực tế được
ghi riêng bên dưới. Sửa tài liệu/workflow không tự sửa branch protection; việc
bật hoặc thay required checks vẫn cần thao tác quản trị riêng của chủ repo.

Đồng thời bật:

- Require a pull request before merging;
- Require status checks to pass;
- Block force pushes;
- Block branch deletion;
- Require conversation resolution nếu không gây ma sát không cần thiết.

### Current stable gate contract

The active `main` ruleset requires `curriculum-gate` from `curriculum-ci.yml`
and `mission-gate` from `mission-agent-path-ci.yml`. Each final gate waits for every
job in its workflow and fails closed when a job fails or is skipped. The child
jobs remain visible for diagnosis, but do not need to be listed individually;
this keeps the required-check contract stable as shards and smoke coverage grow.

### Verified GitHub configuration — 2026-09-21

Ruleset **Protect main with required CI gates** (`23753234`) is active for
`main`. The verified settings are:

- pull request required before merge;
- branch must be up to date before merge;
- required checks: `mission-gate`, `curriculum-gate`, `codeql-go`;
- force-push and branch deletion blocked;
- no bypass actors configured;
- no required approval count configured by this ruleset.

This records the repository setting observed after the ruleset was created; it
does not claim an independent human review or production authorization.

## 4. No legacy compatibility layer

Không thêm lại:

- numeric lesson map làm reading order;
- migration redirects;
- duplicate Mission cho spine cũ;
- validator chỉ để giữ compatibility với repo lịch sử;
- tool-specific authority assumptions.

Nếu cần xem provenance/history, dùng repo `dvha85/affiliate-expert-learning-roadmap`.

## 5. Technology changes

Technology update phải ghi rõ:

```text
observed bottleneck
→ candidate
→ adoption gate
→ rollback/fallback
→ Mission authority unchanged
```

Không đổi curriculum chỉ vì framework mới có feature mới.
