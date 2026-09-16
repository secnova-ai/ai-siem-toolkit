"""Cross-build standalone CLI archives and SHA256SUMS; does not publish."""
import argparse
import hashlib
import os
from pathlib import Path
import re
import subprocess
import tarfile
import tempfile
import zipfile

parser = argparse.ArgumentParser()
parser.add_argument("--version", required=True)
args = parser.parse_args()
if not re.fullmatch(r"[A-Za-z0-9_.-]+", args.version):
    parser.error("version must use letters, digits, dots, underscores or hyphens")
root = Path(__file__).resolve().parents[1]
out = root / "dist"
out.mkdir(exist_ok=True)
checksums = []
for target in ("linux", "darwin", "windows"):
    for arch in ("amd64", "arm64"):
        filename = "tcpkg.exe" if target == "windows" else "tcpkg"
        with tempfile.TemporaryDirectory() as tmp:
            binary = Path(tmp) / filename
            env = dict(os.environ, GOOS=target, GOARCH=arch, CGO_ENABLED="0")
            subprocess.run(["go", "build", "-trimpath", "-ldflags", f"-s -w -X main.version={args.version}", "-o", str(binary), "./cmd/tcpkg"], cwd=root, env=env, check=True)
            ext = "zip" if target == "windows" else "tar.gz"
            archive = out / f"tcpkg-{args.version}-{target}-{arch}.{ext}"
            if target == "windows":
                with zipfile.ZipFile(archive, "w", zipfile.ZIP_DEFLATED) as z:
                    z.write(binary, filename)
                    z.write(root / "README.md", "README.md")
            else:
                with tarfile.open(archive, "w:gz") as z:
                    info = z.gettarinfo(str(binary), arcname=filename)
                    info.mode = 0o755
                    with binary.open("rb") as source:
                        z.addfile(info, source)
                    z.add(root / "README.md", arcname="README.md")
            checksums.append(f"{hashlib.sha256(archive.read_bytes()).hexdigest()}  {archive.name}")
            print(archive.name, flush=True)
(out / "SHA256SUMS").write_text("\n".join(checksums) + "\n", encoding="utf-8")
