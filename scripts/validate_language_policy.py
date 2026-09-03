from pathlib import Path
import re
import sys

ROOT = Path(__file__).resolve().parents[1]

EXACT_FILES = [
    ROOT / "README.md",
    ROOT / "CURRICULUM.md",
    ROOT / "ROADMAP.md",
    ROOT / "PROGRESS.md",
    ROOT / "lab/n8n/COMPATIBILITY.md",
]
GLOB_ROOTS = [
    ROOT / "curriculum",
    ROOT / "missions",
    ROOT / "starter-kits",
    ROOT / "templates",
]

VIETNAMESE_MARKERS = set("ăâđêôơưáàảãạấầẩẫậắằẳẵặéèẻẽẹếềểễệíìỉĩịóòỏõọốồổỗộớờởỡợúùủũụứừửữựýỳỷỹỵ")
VIETNAMESE_WORDS = {
    "la", "va", "khong", "co", "cua", "cho", "trong", "khi", "neu", "voi", "de",
    "mot", "nguoi", "hoc", "bang", "giu", "theo", "hoac", "duoc", "tu", "den",
    "thanh", "qua", "phai", "dung", "nay", "do", "vi", "sau", "truoc", "tren",
    "duoi", "giua", "them", "chi", "ro", "that", "gia", "dinh", "chay", "kiem",
}

ALLOW_LINE_RE = [
    re.compile(r"^https?://"),
    re.compile(r"^[A-Z0-9_./:+|<>=!`* -]+$"),
    re.compile(r"^[-*+]\s+`[^`]+`(?:\s*[—:-].*)?$")
]


def strip_markdown(line: str) -> str:
    line = re.sub(r"`[^`]+`", " ", line)
    line = re.sub(r"https?://\S+", " ", line)
    line = re.sub(r"[#>*_|()\[\]{}:;,.!?=+/-]", " ", line)
    return re.sub(r"\s+", " ", line).strip()


def has_vietnamese_signal(text: str) -> bool:
    lower = text.lower()
    if any(ch in VIETNAMESE_MARKERS for ch in lower):
        return True
    words = set(re.findall(r"[a-zA-Z]+", lower))
    return bool(words & VIETNAMESE_WORDS)


def obvious_english_only(line: str) -> bool:
    stripped = line.strip()
    if not stripped:
        return False
    if any(pattern.match(stripped) for pattern in ALLOW_LINE_RE):
        return False
    prose = strip_markdown(stripped)
    words = re.findall(r"[A-Za-z]+", prose)
    if len(words) < 4:
        return False
    if has_vietnamese_signal(prose):
        return False
    technical_tokens = [w for w in words if w.isupper() or re.search(r"\d", w)]
    if len(technical_tokens) >= max(1, len(words) - 1):
        return False
    return True


def iter_files():
    seen = set()
    for path in EXACT_FILES:
        if path.exists():
            seen.add(path)
            yield path
    for root in GLOB_ROOTS:
        if not root.exists():
            continue
        for path in sorted(root.rglob("*.md")):
            if path not in seen:
                seen.add(path)
                yield path


errors = []
for path in iter_files():
    in_fence = False
    in_frontmatter = False
    for line_no, raw in enumerate(path.read_text(encoding="utf-8").splitlines(), start=1):
        stripped = raw.strip()
        if line_no == 1 and stripped == "---":
            in_frontmatter = True
            continue
        if in_frontmatter:
            if stripped == "---":
                in_frontmatter = False
            continue
        if stripped.startswith("```"):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        if obvious_english_only(raw):
            errors.append(f"{path.relative_to(ROOT)}:{line_no}: English-only learner prose: {stripped}")

if errors:
    print("LANGUAGE POLICY VALIDATION FAILED")
    for error in errors:
        print(f"- {error}")
    sys.exit(1)

print("LANGUAGE POLICY PASS: learner-facing prose is Vietnamese-first")
