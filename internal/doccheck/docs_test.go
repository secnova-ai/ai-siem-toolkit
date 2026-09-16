// These tests execute the files readers copy from the documentation, so example
// regressions cannot be hidden by testing only separately maintained fixtures.
package doccheck

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/secnova-ai/ai-siem-toolkit/internal/cli"
	"github.com/secnova-ai/ai-siem-toolkit/pkg/cel"
	"github.com/secnova-ai/ai-siem-toolkit/pkg/credential"
	"github.com/secnova-ai/ai-siem-toolkit/pkg/schema"
	"github.com/secnova-ai/ai-siem-toolkit/pkg/tcpkg"
)

var yamlBlock = regexp.MustCompile("(?ms)^```yaml[^\\n]*\\n(.*?)^```[ \\t]*$")
var projectFile = regexp.MustCompile("(?ms)<!-- tutorial-file: ([^\\r\\n]+) -->\\s*\\n```yaml\\s*\\n(.*?)^```[ \\t]*$")

func TestDocumentationYAML(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, dir := range []string{"docs", "skills/create-custom-tool/references"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			text := strings.ReplaceAll(string(b), "\r\n", "\n")
			t.Run(filepath.ToSlash(path), func(t *testing.T) {
				for i, match := range yamlBlock.FindAllStringSubmatch(text, -1) {
					var value map[string]any
					if err := tcpkg.Decode([]byte(match[1]), &value); err != nil {
						t.Fatalf("YAML block %d: %v", i+1, err)
					}
					for _, key := range []string{"input_schema", "output_schema"} {
						if s, ok := value[key]; ok {
							raw, err := json.Marshal(s)
							if err != nil {
								t.Fatal(err)
							}
							if err = schema.ValidateDefinition(raw); err != nil {
								t.Fatalf("block %d %s: %v", i+1, key, err)
							}
						}
					}
					if s, ok := value["credential_schema"]; ok {
						raw, err := json.Marshal(s)
						if err != nil {
							t.Fatal(err)
						}
						if err = credential.ValidateSchemaDefinition(raw); err != nil {
							t.Fatalf("block %d credential_schema: %v", i+1, err)
						}
					}
					if cfg, ok := value["runtime_config"].(map[string]any); ok {
						if p, ok := cfg["program"].(string); ok {
							if err := cel.Validate(p); err != nil {
								t.Fatalf("block %d CEL: %v", i+1, err)
							}
						}
					}
				}
				files := projectFile.FindAllStringSubmatch(text, -1)
				if len(files) == 0 {
					return
				}
				project := filepath.Join(t.TempDir(), "project")
				seen := map[string]bool{}
				hasTests := false
				for _, m := range files {
					name := m[1]
					if !fs.ValidPath(name) || strings.Contains(name, "\\") {
						t.Fatalf("invalid tutorial path %s", name)
					}
					if seen[name] {
						t.Fatalf("duplicate tutorial file %s", name)
					}
					seen[name] = true
					out := filepath.Join(project, filepath.FromSlash(name))
					if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(out, []byte(m[2]), 0644); err != nil {
						t.Fatal(err)
					}
					hasTests = hasTests || strings.HasPrefix(name, "tests/")
				}
				commands := [][]string{{"validate", project}}
				if hasTests {
					commands = append(commands, []string{"test", project})
				}
				archive := filepath.Join(t.TempDir(), "tutorial.tcpkg")
				commands = append(commands, []string{"pack", project, "-o", archive}, []string{"verify", archive})
				for _, args := range commands {
					var output bytes.Buffer
					if err := cli.Run(args, &output, "test"); err != nil {
						t.Fatalf("%s: %v\n%s", args[0], err, output.String())
					}
				}
			})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
