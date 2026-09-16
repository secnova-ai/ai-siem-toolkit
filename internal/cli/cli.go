package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/secnova-ai/ai-siem-toolkit/internal/runner"
	"github.com/secnova-ai/ai-siem-toolkit/pkg/tcpkg"
	"github.com/secnova-ai/ai-siem-toolkit/templates"
)

const help = `tcpkg — SecNova AI-SIEM custom tool development

  tcpkg init <directory> --runtime cel|wasm|mcp [--language go]
  tcpkg validate <directory>
  tcpkg build <directory>
  tcpkg test <directory> [--tool <id>] [--live --credentials <local.json>]
  tcpkg pack <directory> [-o <output.tcpkg>]
  tcpkg verify <archive.tcpkg> [archive.tcpkg...]
  tcpkg version

validate/test/verify never upload or install tools. test uses mocked HTTP by
default. --live sends real requests, including writes defined by your tools.
`

func Run(args []string, w io.Writer, version string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		fmt.Fprint(w, help)
		return nil
	}
	if args[0] == "version" {
		fmt.Fprintf(w, "tcpkg %s (Tool Center dev/3.0.7 authoring contract)\n", version)
		return nil
	}
	if len(args) < 2 {
		return fmt.Errorf("directory/archive is required\n%s", help)
	}
	cmd, dir := args[0], args[1]
	opts := map[string]string{}
	allowed := map[string]map[string]bool{"init": {"--runtime": true, "--language": true}, "pack": {"-o": true}, "test": {"--tool": true, "--live": true, "--credentials": true}, "validate": {}, "build": {}, "verify": {}}
	permitted, ok := allowed[cmd]
	if !ok {
		return fmt.Errorf("unknown command %q", cmd)
	}
	if cmd != "verify" {
		for i := 2; i < len(args); i++ {
			k := args[i]
			if !permitted[k] {
				return fmt.Errorf("unknown option %q for %s", k, cmd)
			}
			if _, exists := opts[k]; exists {
				return fmt.Errorf("duplicate option %s", k)
			}
			if k == "--live" {
				opts[k] = "true"
				continue
			}
			i++
			if i >= len(args) {
				return fmt.Errorf("%s requires a value", k)
			}
			opts[k] = args[i]
		}
	}
	switch cmd {
	case "init":
		runtime := opts["--runtime"]
		if runtime == "" {
			runtime = "cel"
		}
		if lang := opts["--language"]; lang != "" && (runtime != "wasm" || lang != "go") {
			return fmt.Errorf("only --runtime wasm --language go is supported")
		}
		if err := templates.Init(dir, runtime); err != nil {
			return err
		}
		fmt.Fprintf(w, "created %s (%s)\n", dir, runtime)
	case "validate":
		p, err := tcpkg.Load(dir, false)
		if err != nil {
			return err
		}
		for _, v := range p.Warnings {
			fmt.Fprintln(w, "warning:", v)
		}
		fmt.Fprintf(w, "valid %s: %d static tools, discovery=%t\n", p.Provider.ProviderID, len(p.Tools), p.Provider.MCPDiscovery)
	case "pack":
		out := opts["-o"]
		if err := tcpkg.Pack(dir, out); err != nil {
			return err
		}
		fmt.Fprintln(w, "package written and verified")
	case "verify":
		for _, f := range args[1:] {
			p, err := tcpkg.Verify(f)
			if err != nil {
				return fmt.Errorf("%s: %w", f, err)
			}
			fmt.Fprintf(w, "verified %s: %s %s\n", f, p.Provider.ProviderID, p.Provider.Version)
		}
	case "test":
		return runner.Run(dir, runner.Options{Live: opts["--live"] == "true", CredentialFile: opts["--credentials"], Tool: opts["--tool"]}, w)
	case "build":
		return build(dir, w)
	}
	return nil
}
func build(dir string, w io.Writer) error {
	p, err := tcpkg.Load(dir, false)
	if err != nil {
		return err
	}
	if _, err = exec.LookPath("go"); err != nil {
		return fmt.Errorf("Go 1.25.7 or newer is required to build WASM")
	}
	root, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, t := range p.Tools {
		if t.RuntimeType != "wasm" {
			continue
		}
		ref := t.RuntimeConfig.ArtifactRef
		if seen[ref] {
			continue
		}
		seen[ref] = true
		src := filepath.Join(root, "src", strings.TrimSuffix(ref, ".wasm"))
		if s, e := os.Stat(src); e != nil || !s.IsDir() {
			return fmt.Errorf("expected Go source directory src/%s", strings.TrimSuffix(ref, ".wasm"))
		}
		outdir := filepath.Join(root, "wasm")
		if err = os.MkdirAll(outdir, 0755); err != nil {
			return err
		}
		tmp, err := os.CreateTemp(outdir, ".build-*")
		if err != nil {
			return err
		}
		tmp.Close()
		path := tmp.Name()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		c := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", path, ".")
		c.Dir = src
		c.Env = []string{}
		for _, v := range os.Environ() {
			key := strings.ToUpper(strings.SplitN(v, "=", 2)[0])
			if key != "GOOS" && key != "GOARCH" && key != "CGO_ENABLED" && key != "GOWORK" {
				c.Env = append(c.Env, v)
			}
		}
		c.Env = append(c.Env, "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0", "GOWORK=off")
		c.Stdout = w
		c.Stderr = w
		err = c.Run()
		cancel()
		if err != nil {
			os.Remove(path)
			return fmt.Errorf("build %s: %w", ref, err)
		}
		err = os.Rename(path, filepath.Join(outdir, ref))
		if err != nil {
			os.Remove(path)
			return err
		}
		fmt.Fprintln(w, "built wasm/"+ref)
	}
	if len(seen) == 0 {
		return fmt.Errorf("no WASM tools; CEL and MCP do not need compilation")
	}
	_, err = tcpkg.Load(root, true)
	return err
}
