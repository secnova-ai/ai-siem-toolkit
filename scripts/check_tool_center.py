"""Validate example packages with a local Tool Center checkout without changing it.

Run with the Go toolchain/environment appropriate for that checkout:
  python scripts/check_tool_center.py --tool-center ../tool-center
Uses a Go overlay; no database or live deployment is required.
"""
import argparse
import json
import os
from pathlib import Path
import subprocess
import tempfile

parser = argparse.ArgumentParser()
parser.add_argument("--tool-center", type=Path, required=True)
args = parser.parse_args()
root = Path(__file__).resolve().parents[1]
center = args.tool_center.resolve()
test_source = r'''package catalog
import (
 "encoding/json"
 "os"
 "path/filepath"
 "testing"
 "gopkg.in/yaml.v3"
 celruntime "tool-center/internal/runtime/cel"
 mcpRuntime "tool-center/internal/runtime/mcp"
 wasmRuntime "tool-center/internal/runtime/wasm"
)
func TestToolkitCompatibility(t *testing.T) {
 files,err:=filepath.Glob(filepath.Join(os.Getenv("TOOLKIT_PACKAGES"),"*.tcpkg"));if err!=nil||len(files)<3{t.Fatalf("missing packages: %v",err)}
 for _,name:=range files {t.Run(filepath.Base(name),func(t *testing.T){
  f,err:=os.Open(name);if err!=nil{t.Fatal(err)};defer f.Close();st,_:=f.Stat();pkg,err:=ParsePackage(f,st.Size());if err!=nil{t.Fatal(err)}
  var raw providerYAML;if err=yaml.Unmarshal(pkg.ProviderYAML,&raw);err!=nil{t.Fatal(err)}
  p,err:=buildProvider(&raw,"test",raw.Version,"test","uploaded");if err!=nil{t.Fatal(err)}
  for _,data:=range pkg.ToolYAMLs{tool,err:=buildTool(data,"toolkit",p);if err!=nil{t.Fatal(err)};switch string(tool.RuntimeType){
   case "cel":var c struct{Program string `json:"program"`};json.Unmarshal(tool.RuntimeConfig,&c);err=celruntime.ValidateProgram(c.Program)
   case "mcp":err=mcpRuntime.New().ValidateRuntimeConfig(tool.RuntimeConfig)
   case "wasm":err=wasmRuntime.New(nil).ValidateRuntimeConfig(tool.RuntimeConfig)
  };if err!=nil{t.Fatal(err)}}
 })}
}
'''
with tempfile.TemporaryDirectory(prefix="toolkit-compat-") as tmp:
    temp = Path(tmp)
    exe = temp / ("tcpkg.exe" if os.name == "nt" else "tcpkg")
    subprocess.run(["go", "build", "-o", str(exe), "./cmd/tcpkg"], cwd=root, check=True)
    for project in sorted((root / "examples").iterdir()):
        if not project.is_dir() or not (project / "_provider.yaml").exists():
            continue
        if (project / "src").exists():
            subprocess.run([str(exe), "build", str(project)], check=True)
        subprocess.run([str(exe), "pack", str(project), "-o", str(temp / (project.name + ".tcpkg"))], check=True)
    source = temp / "compat_test.go"
    source.write_text(test_source, encoding="utf-8")
    overlay = temp / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {str(center / "internal/catalog/zz_toolkit_compat_test.go"): str(source)}}), encoding="utf-8")
    subprocess.run(["go", "test", "-overlay", str(overlay), "./internal/catalog", "-run", "^TestToolkitCompatibility$", "-count=1", "-v"], cwd=center, env=dict(os.environ, TOOLKIT_PACKAGES=str(temp)), check=True)
