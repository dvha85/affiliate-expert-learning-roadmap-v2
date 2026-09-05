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

Khuyến nghị bật required checks trên `main`:

| Workflow | Tên check chính xác trên GitHub |
|---|---|
| `curriculum-ci.yml` — Curriculum CI | `structure-language-and-foundations` |
| `curriculum-ci.yml` — Curriculum CI | `deterministic-runtime` |
| `mission-agent-path-ci.yml` — Mission Agent Path CI | `mission-semantics-and-blueprints` |
| `mission-agent-path-ci.yml` — Mission Agent Path CI | `mission-runtime` |

Tên check là tên job, không phải tên step; regression tests Python thuộc job `structure-language-and-foundations`. Khi cấu hình, chọn đúng check do GitHub Actions phát hành trên commit hiện hành. Không dùng các tên cũ `curriculum`, `evidence-and-safety`, `python-regression` hoặc tự thêm tiền tố workflow vào tên check.

Đây là cấu hình **khuyến nghị**, không xác nhận branch protection hiện đã bật. Sửa tài liệu/workflow không tự sửa branch protection; việc bật hoặc thay required checks cần thao tác quản trị riêng của chủ repo.

Đồng thời bật:

- Require a pull request before merging;
- Require status checks to pass;
- Block force pushes;
- Block branch deletion;
- Require conversation resolution nếu không gây ma sát không cần thiết.

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
