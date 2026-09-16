"""Refresh maintained copies after editing their canonical sources."""
from pathlib import Path
from check_docs import REFERENCES, ROOT

for name in REFERENCES:
    src = ROOT / "docs/en" / f"{name}.md"
    dst = ROOT / "skills/create-custom-tool/references" / src.name
    dst.parent.mkdir(parents=True, exist_ok=True)
    dst.write_bytes(src.read_bytes())
for name in ("http_wasm.go", "http_stub.go"):
    src = ROOT / "sdk/go" / name
    (ROOT / "templates/wasm-go/sdk" / (name + ".txt")).write_bytes(src.read_bytes())
print("Updated skill references and embedded SDK source.")

for directory in [ROOT / "templates" / n for n in ("cel", "wasm-go", "mcp")] + [ROOT / "skills/create-custom-tool"] + [p for p in (ROOT / "examples").iterdir() if p.is_dir() and (p / "_provider.yaml").exists()]:
    (directory / "LICENSE").write_bytes((ROOT / "LICENSE").read_bytes())
print("Updated license copies.")
