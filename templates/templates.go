// Package templates embeds complete starter projects; init needs no network.
package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/secnova-ai/ai-siem-toolkit/pkg/tcpkg"
)

//go:embed all:cel all:wasm-go all:mcp
var content embed.FS

func Init(dir, runtime string) error {
	mode := runtime
	if mode == "wasm" {
		mode = "wasm-go"
	}
	if mode != "cel" && mode != "wasm-go" && mode != "mcp" {
		return fmt.Errorf("runtime must be cel, wasm or mcp")
	}
	id := filepath.Base(filepath.Clean(dir))
	if !tcpkg.ValidID(id) {
		return fmt.Errorf("directory name must start with a lowercase letter, then lowercase letters, digits or hyphens, at most 64 characters")
	}
	if _, err := os.Lstat(dir); err == nil {
		return fmt.Errorf("destination already exists: %s", dir)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Mkdir(dir, 0755); err != nil {
		return err
	}
	return fs.WalkDir(content, mode, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(path, mode+"/")
		if rel == "gitignore.txt" {
			rel = ".gitignore"
		} else {
			rel = strings.TrimSuffix(rel, ".txt")
		}
		out := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
			return err
		}
		b, err := content.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(out, []byte(strings.ReplaceAll(string(b), "{{ID}}", id)), 0644)
	})
}
