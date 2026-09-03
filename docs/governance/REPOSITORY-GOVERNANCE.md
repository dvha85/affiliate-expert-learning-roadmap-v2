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

- `Curriculum CI / curriculum`
- `Curriculum CI / evidence-and-safety`
- `Curriculum CI / python-regression`
- `Curriculum CI / deterministic-runtime`

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
