package tcpkg

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/secnova-ai/ai-siem-toolkit/pkg/cel"
	"github.com/secnova-ai/ai-siem-toolkit/pkg/credential"
	"github.com/secnova-ai/ai-siem-toolkit/pkg/schema"
	"github.com/tetratelabs/wazero"
	"gopkg.in/yaml.v3"
)

type Provider struct {
	ProviderID         string                    `yaml:"provider_id"`
	Version            string                    `yaml:"version"`
	Vendor             string                    `yaml:"vendor"`
	Name               string                    `yaml:"name"`
	Description        string                    `yaml:"description"`
	Category           string                    `yaml:"category"`
	SupportedLocations []string                  `yaml:"supported_locations"`
	CredentialSchema   map[string]map[string]any `yaml:"credential_schema"`
	AuthStrategy       map[string]any            `yaml:"auth_strategy"`
	CredentialTest     map[string]any            `yaml:"credential_test"`
	API                map[string]any            `yaml:"api"`
	RateLimit          map[string]any            `yaml:"rate_limit"`
	MCPDiscovery       bool                      `yaml:"mcp_discovery"`
	MCPEndpoint        string                    `yaml:"mcp_endpoint"`
	Icon               string                    `yaml:"icon"`
	IconDark           string                    `yaml:"icon_dark"`
}
type Config struct {
	Endpoint         string   `yaml:"endpoint"`
	Program          string   `yaml:"program"`
	ArtifactRef      string   `yaml:"artifact_ref"`
	ArtifactChecksum string   `yaml:"artifact_checksum"`
	AllowedDomains   []string `yaml:"allowed_domains"`
	TimeoutSeconds   int      `yaml:"timeout_seconds"`
	ToolName         string   `yaml:"tool_name"`
}
type Tool struct {
	ToolID              string   `yaml:"tool_id"`
	ProviderID          string   `yaml:"provider_id"`
	Name                string   `yaml:"name"`
	Description         string   `yaml:"description"`
	HumanDescription    string   `yaml:"human_description"`
	InputSchema         any      `yaml:"input_schema"`
	OutputSchema        any      `yaml:"output_schema"`
	RiskLevel           string   `yaml:"risk_level"`
	RuntimeType         string   `yaml:"runtime_type"`
	RuntimeConfig       Config   `yaml:"runtime_config"`
	RequiredPermissions []string `yaml:"required_permissions"`
	ForceApproval       bool     `yaml:"force_approval"`
	SensitiveParams     []string `yaml:"sensitive_params"`
	Irreversible        bool     `yaml:"irreversible"`
	BlastRadius         string   `yaml:"blast_radius"`
	AllowedCallerTypes  []string `yaml:"allowed_caller_types"`
}
type Project struct {
	Provider Provider
	Tools    map[string]Tool
	Files    map[string][]byte
	Warnings []string
}

var idPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
var versionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

func ValidID(id string) bool { return idPattern.MatchString(id) }
func Decode(data []byte, v any) error {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return err
	}
	var walk func(*yaml.Node, int) error
	count := 0
	walk = func(n *yaml.Node, d int) error {
		count++
		if d > 64 || count > 100000 {
			return fmt.Errorf("YAML too complex")
		}
		if n.Kind == yaml.AliasNode {
			return fmt.Errorf("YAML aliases are not supported")
		}
		for _, c := range n.Content {
			if err := walk(c, d+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(&node, 0); err != nil {
		return err
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(v); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("expected one YAML document")
	}
	return nil
}
func regular(root, name string) ([]byte, error) {
	// Reject symlinks in every component, including directories.
	cur := root
	for _, p := range strings.Split(filepath.ToSlash(name), "/") {
		cur = filepath.Join(cur, p)
		s, err := os.Lstat(cur)
		if err != nil {
			return nil, err
		}
		if s.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("symlink is not allowed: %s", name)
		}
	}
	st, err := os.Stat(cur)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() || st.Size() > maxPkgEntrySize {
		return nil, fmt.Errorf("not a regular file or exceeds 32 MiB: %s", name)
	}
	return os.ReadFile(cur)
}

// Load validates a source tree; requireArtifacts=false allows validating before WASM build.
func Load(root string, requireArtifacts bool) (*Project, error) {
	files := map[string][]byte{}
	b, err := regular(root, "_provider.yaml")
	if err != nil {
		return nil, err
	}
	files["_provider.yaml"] = b
	entries, err := os.ReadDir(filepath.Join(root, "tools"))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() && isYAMLFile(e.Name()) {
			name := "tools/" + e.Name()
			b, err = regular(root, name)
			if err != nil {
				return nil, err
			}
			files[name] = b
		}
	}
	p, err := ValidateFiles(files, false)
	if err != nil {
		return nil, err
	}
	refs := map[string]bool{}
	for _, t := range p.Tools {
		if t.RuntimeType == "wasm" {
			refs["wasm/"+t.RuntimeConfig.ArtifactRef] = true
		}
	}
	for _, s := range []string{p.Provider.Icon, p.Provider.IconDark} {
		if s != "" {
			refs["_assets/"+s] = true
		}
	}
	for name := range refs {
		b, err := regular(root, name)
		if err != nil {
			if !requireArtifacts && strings.HasPrefix(name, "wasm/") && os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("%s: %w (WASM: run tcpkg build first)", name, err)
		}
		files[name] = b
	}
	return ValidateFiles(files, requireArtifacts)
}
func plain(s string) bool { return s != "" && s != "." && s != ".." && !strings.ContainsAny(s, "/\\:") }
func ValidateFiles(files map[string][]byte, requireArtifacts bool) (*Project, error) {
	p := &Project{Files: files, Tools: map[string]Tool{}}
	if err := Decode(files["_provider.yaml"], &p.Provider); err != nil {
		return nil, fmt.Errorf("_provider.yaml: %w", err)
	}
	v := p.Provider
	credSchema, err := json.Marshal(v.CredentialSchema)
	if err != nil {
		return nil, err
	}
	if err = credential.ValidateSchemaDefinition(credSchema); err != nil {
		return nil, fmt.Errorf("_provider.yaml: credential_schema: %w", err)
	}
	if typ, ok := v.CredentialTest["type"]; ok && !contains([]string{"skip", "http_probe", "oauth2_client_credentials"}, fmt.Sprint(typ)) {
		return nil, fmt.Errorf("_provider.yaml: unsupported credential_test.type")
	}
	if !ValidID(v.ProviderID) || !versionPattern.MatchString(v.Version) || v.Vendor == "" || v.Name == "" || v.Category == "" {
		return nil, fmt.Errorf("_provider.yaml: valid provider_id, semver version, vendor, name and category are required")
	}
	for _, l := range v.SupportedLocations {
		if l != "saas" && l != "edge" {
			return nil, fmt.Errorf("unsupported location %q", l)
		}
	}
	for _, s := range []string{v.Icon, v.IconDark} {
		if s != "" && !plain(s) {
			return nil, fmt.Errorf("icon must be a bare filename")
		}
		if s != "" && requireArtifacts && files["_assets/"+s] == nil {
			return nil, fmt.Errorf("missing icon %s", s)
		}
	}
	names := []string{}
	for n := range files {
		if strings.HasPrefix(n, "tools/") {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		var t Tool
		if err := Decode(files[name], &t); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if !strings.HasPrefix(t.ToolID, v.ProviderID+".") || strings.TrimPrefix(t.ToolID, v.ProviderID+".") == "" || t.Name == "" {
			return nil, fmt.Errorf("%s: name and tool_id prefixed with %s. are required", name, v.ProviderID)
		}
		if t.ProviderID != "" && t.ProviderID != v.ProviderID {
			return nil, fmt.Errorf("%s: provider_id mismatch", name)
		}
		if _, ok := p.Tools[t.ToolID]; ok {
			return nil, fmt.Errorf("%s: duplicate tool_id %s", name, t.ToolID)
		}
		if !contains([]string{"low", "medium", "high", "critical"}, t.RiskLevel) {
			return nil, fmt.Errorf("%s: invalid risk_level", name)
		}
		if !contains([]string{"", "none", "single", "multi", "global"}, t.BlastRadius) {
			return nil, fmt.Errorf("%s: invalid blast_radius", name)
		}
		for _, caller := range t.AllowedCallerTypes {
			if !contains([]string{"chat_agent", "async_agent", "automation", "external_api"}, caller) {
				return nil, fmt.Errorf("%s: invalid allowed_caller_types entry %q", name, caller)
			}
		}
		for _, s := range []string{"input_schema", "output_schema"} {
			x := t.InputSchema
			if s == "output_schema" {
				x = t.OutputSchema
			}
			b, err := json.Marshal(x)
			if err != nil {
				return nil, err
			}
			if err = schema.ValidateDefinition(b); err != nil {
				return nil, fmt.Errorf("%s: %s: %w", name, s, err)
			}
		}
		switch t.RuntimeType {
		case "cel":
			if err := cel.Validate(t.RuntimeConfig.Program); err != nil {
				return nil, fmt.Errorf("%s: runtime_config.program: %w", name, err)
			}
		case "wasm":
			ref := t.RuntimeConfig.ArtifactRef
			if !plain(ref) || !strings.HasSuffix(ref, ".wasm") {
				return nil, fmt.Errorf("%s: artifact_ref must be a .wasm filename", name)
			}
			b := files["wasm/"+ref]
			if b == nil && requireArtifacts {
				return nil, fmt.Errorf("%s: missing wasm/%s; run tcpkg build", name, ref)
			}
			if b != nil {
				if checksum := t.RuntimeConfig.ArtifactChecksum; checksum != "" && checksum != fmt.Sprintf("sha256:%x", sha256.Sum256(b)) {
					return nil, fmt.Errorf("%s: artifact_checksum mismatch", name)
				}
				ctx := context.Background()
				r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithMemoryLimitPages(256))
				_, err := r.CompileModule(ctx, b)
				r.Close(ctx)
				if err != nil {
					return nil, fmt.Errorf("%s: WASM compile: %w", name, err)
				}
			}
		case "mcp":
			if t.RuntimeConfig.ToolName == "" {
				return nil, fmt.Errorf("%s: runtime_config.tool_name is required", name)
			}
		default:
			return nil, fmt.Errorf("%s: runtime_type must be cel, wasm or mcp", name)
		}
		if t.Description == "" {
			p.Warnings = append(p.Warnings, name+": description is missing")
		}
		describe(t.InputSchema, name+": input_schema", &p.Warnings)
		p.Tools[t.ToolID] = t
	}
	if len(p.Tools) == 0 && !v.MCPDiscovery {
		return nil, fmt.Errorf("tools/*.yaml required unless mcp_discovery is true")
	}
	if v.MCPDiscovery && len(p.Tools) > 0 {
		return nil, fmt.Errorf("dynamic MCP discovery must not mix static tool definitions")
	}
	return p, nil
}
func contains(a []string, s string) bool {
	for _, v := range a {
		if s == v {
			return true
		}
	}
	return false
}
func describe(x any, path string, w *[]string) {
	m, ok := x.(map[string]any)
	if !ok {
		return
	}
	if props, ok := m["properties"].(map[string]any); ok {
		keys := []string{}
		for k := range props {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := props[k]
			p := path + ".properties." + k
			if f, ok := v.(map[string]any); ok {
				if f["description"] == nil || f["description"] == "" {
					*w = append(*w, p+": description is missing")
				}
			}
			describe(v, p, w)
		}
	}
	if m["type"] == "object" && m["properties"] == nil {
		*w = append(*w, path+": describe nested properties or document the free-form object")
	}
	describe(m["items"], path+".items", w)
}
