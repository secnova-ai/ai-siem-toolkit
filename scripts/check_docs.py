"""Check local Markdown links and synchronized SDK/skill reference copies."""
from pathlib import Path
import re
import sys

ROOT = Path(__file__).resolve().parents[1]
REFERENCES = ("design", "cel", "wasm", "mcp", "credentials", "reference", "testing", "cli", "compatibility", "structure")


def main():
    failures = []
    license_text = (ROOT / "LICENSE").read_bytes()
    for directory in [ROOT / "templates" / n for n in ("cel", "wasm-go", "mcp")] + [ROOT / "skills/create-custom-tool"] + [p for p in (ROOT / "examples").iterdir() if p.is_dir() and (p / "_provider.yaml").exists()]:
        license_file = directory / "LICENSE"
        if not license_file.exists() or license_file.read_bytes() != license_text:
            failures.append(f"license copy missing or out of sync: {directory.relative_to(ROOT)}")
    for file in ROOT.rglob("*.md"):
        if ".git" in file.parts or "dist" in file.parts:
            continue
        text = file.read_text(encoding="utf-8")
        for target in re.findall(r"\[[^\]]*\]\(([^)]+)\)", text):
            if "://" in target or target.startswith("#"):
                continue
            target = target.split("#", 1)[0]
            if not (file.parent / target).exists():
                failures.append(f"{file.relative_to(ROOT)}: missing {target}")
    for name in REFERENCES:
        source = ROOT / "docs/en" / f"{name}.md"
        target = ROOT / "skills/create-custom-tool/references" / source.name
        if not target.exists() or source.read_bytes() != target.read_bytes():
            failures.append(f"skill reference out of sync: {name}")
    for name in ("http_wasm.go", "http_stub.go"):
        source = ROOT / "sdk/go" / name
        target = ROOT / "templates/wasm-go/sdk" / (name + ".txt")
        if source.read_bytes() != target.read_bytes():
            failures.append(f"WASM template SDK out of sync: {name}")
    if failures:
        print("\n".join(failures))
        return 1
    print("Documentation links, skill references and SDK template copies are valid.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
