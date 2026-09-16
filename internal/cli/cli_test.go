package cli

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/secnova-ai/ai-siem-toolkit/pkg/tcpkg"
)

func command(t *testing.T, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	if err := Run(args, &out, "test"); err != nil {
		t.Fatalf("%v: %v\n%s", args, err, out.String())
	}
	return out.String()
}
func TestCELWorkflow(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "health-api")
	command(t, "init", dir, "--runtime", "cel")
	command(t, "validate", dir)
	command(t, "test", dir)
	os.Mkdir(filepath.Join(dir, ".local"), 0700)
	os.WriteFile(filepath.Join(dir, ".local", "credentials.json"), []byte(`{"api_key":"must-not-ship"}`), 0600)
	pkg := filepath.Join(t.TempDir(), "health.tcpkg")
	command(t, "pack", dir, "-o", pkg)
	command(t, "verify", pkg)
	z, err := zip.OpenReader(pkg)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	if len(z.File) != 2 {
		t.Fatalf("unexpected files: %v", z.File)
	}
	var out bytes.Buffer
	if err := Run([]string{"pack", dir, "-o", pkg}, &out, "test"); err == nil {
		t.Fatal("overwrote existing archive")
	}
	// An unexpected request must fail even when an expected tool error is declared.
	path := filepath.Join(dir, "tests", "health.yaml")
	b, _ := os.ReadFile(path)
	b = bytes.ReplaceAll(b, []byte("method: GET"), []byte("method: POST"))
	os.WriteFile(path, b, 0644)
	if err := Run([]string{"test", dir}, &out, "test"); err == nil || !strings.Contains(err.Error(), "mismatch") {
		t.Fatalf("unexpected request passed: %v", err)
	}
}
func TestMCPDiscovery(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "mcp-service")
	command(t, "init", dir, "--runtime", "mcp")
	command(t, "validate", dir)
	pkg := filepath.Join(t.TempDir(), "mcp.tcpkg")
	command(t, "pack", dir, "-o", pkg)
	p, err := tcpkg.Verify(pkg)
	if err != nil || !p.Provider.MCPDiscovery || len(p.Tools) != 0 {
		t.Fatalf("%+v %v", p, err)
	}
}
func TestWASMWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("requires local Go compiler")
	}
	dir := filepath.Join(t.TempDir(), "wasm-api")
	command(t, "init", dir, "--runtime", "wasm", "--language", "go")
	command(t, "validate", dir)
	command(t, "build", dir)
	command(t, "test", dir)
	pkg := filepath.Join(t.TempDir(), "wasm.tcpkg")
	command(t, "pack", dir, "-o", pkg)
	command(t, "verify", pkg)
}
func TestBadDefinitions(t *testing.T) {
	for _, tc := range []struct{ name, old, new string }{
		{"unknown-field", "risk_level: low", "risk_levle: low"},
		{"cel-function", "get_h(", "imaginary_http("},
		{"schema", "type: object", "type: int"},
		{"provider-mismatch", "name: Read service health", "provider_id: other\nname: Read service health"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "bad-api")
			command(t, "init", dir)
			file := filepath.Join(dir, "tools", "health.yaml")
			b, _ := os.ReadFile(file)
			os.WriteFile(file, bytes.ReplaceAll(b, []byte(tc.old), []byte(tc.new)), 0644)
			var out bytes.Buffer
			if err := Run([]string{"validate", dir}, &out, "test"); err == nil {
				t.Fatal("invalid definition accepted")
			}
		})
	}
}
func TestInitDoesNotOverwrite(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "existing")
	os.Mkdir(dir, 0755)
	var out bytes.Buffer
	if err := Run([]string{"init", dir}, &out, "test"); err == nil {
		t.Fatal("existing directory accepted")
	}
}
