package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/secnova-ai/ai-siem-toolkit/pkg/cel"
	"github.com/secnova-ai/ai-siem-toolkit/pkg/credential"
	"github.com/secnova-ai/ai-siem-toolkit/pkg/schema"
	"github.com/secnova-ai/ai-siem-toolkit/pkg/tcpkg"
	"github.com/secnova-ai/ai-siem-toolkit/pkg/wasm"
)

type Mock struct {
	Request  cel.Request  `yaml:"request"`
	Response cel.Response `yaml:"response"`
}
type Case struct {
	Name        string         `yaml:"name"`
	Tool        string         `yaml:"tool"`
	Params      map[string]any `yaml:"params"`
	Creds       map[string]any `yaml:"creds"`
	HTTP        []Mock         `yaml:"http"`
	Expect      any            `yaml:"expect"`
	ExpectError string         `yaml:"expect_error"`
}
type Options struct {
	Live           bool
	CredentialFile string
	Tool           string
}

func Run(root string, opt Options, w io.Writer) error {
	p, err := tcpkg.Load(root, true)
	if err != nil {
		return err
	}
	liveCreds := map[string]any{}
	if opt.Live {
		if opt.CredentialFile == "" {
			return fmt.Errorf("--live requires --credentials <local.json>")
		}
		b, err := os.ReadFile(opt.CredentialFile)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(b, &liveCreds); err != nil {
			return err
		}
	} else if opt.CredentialFile != "" {
		return fmt.Errorf("--credentials requires --live")
	}
	names, err := filepath.Glob(filepath.Join(root, "tests", "*.yaml"))
	if err != nil {
		return err
	}
	count := 0
	for _, name := range names {
		b, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		var c Case
		if err = tcpkg.Decode(b, &c); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if opt.Tool != "" && c.Tool != opt.Tool {
			continue
		}
		t, ok := p.Tools[c.Tool]
		if !ok {
			return fmt.Errorf("%s: unknown tool %s", name, c.Tool)
		}
		if t.RuntimeType == "mcp" {
			return fmt.Errorf("MCP execution belongs to the remote server; validate its package here and test discovery/calls in SIEM")
		}
		count++
		creds := c.Creds
		if opt.Live {
			creds = liveCreds
		}
		if creds == nil {
			creds = map[string]any{}
		}
		if c.Params == nil {
			c.Params = map[string]any{}
		}
		credSchema, _ := json.Marshal(p.Provider.CredentialSchema)
		creds = credential.ApplySchemaDefaults(credSchema, creds)
		secretJSON, _ := json.Marshal(creds)
		if e := credential.ValidateSchema(credSchema, secretJSON); e != nil {
			return fmt.Errorf("%s: credential validation: %w", name, e)
		}
		cursor := 0
		var transportErr error
		transport := func(ctx context.Context, req cel.Request) (cel.Response, error) {
			if !allowed(p, t, creds, req.URL) {
				transportErr = fmt.Errorf("request hostname is not in allowed_domains or a URL credential")
				return cel.Response{}, transportErr
			}
			if opt.Live {
				return live(ctx, req)
			}
			if cursor >= len(c.HTTP) {
				transportErr = fmt.Errorf("unexpected HTTP request #%d", cursor+1)
				return cel.Response{}, transportErr
			}
			m := c.HTTP[cursor]
			cursor++
			if req.Method != m.Request.Method || req.URL != m.Request.URL || req.Body != m.Request.Body {
				transportErr = fmt.Errorf("HTTP request #%d method/URL/body mismatch", cursor)
				return cel.Response{}, transportErr
			}
			for k, v := range m.Request.Headers {
				actual := ""
				for rk, rv := range req.Headers {
					if strings.EqualFold(k, rk) {
						actual = rv
					}
				}
				if actual != v {
					transportErr = fmt.Errorf("HTTP request #%d header %s mismatch", cursor, k)
					return cel.Response{}, transportErr
				}
			}
			return m.Response, nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		input, _ := json.Marshal(c.Params)
		sch, _ := json.Marshal(t.InputSchema)
		err = (&schema.Validator{}).Validate(t.ToolID, sch, input)
		var result json.RawMessage
		if err == nil {
			switch t.RuntimeType {
			case "cel":
				result, err = cel.Execute(ctx, t.RuntimeConfig.Program, c.Params, creds, transport)
			case "wasm":
				result, err = wasm.Execute(ctx, p.Files["wasm/"+t.RuntimeConfig.ArtifactRef], c.Params, creds, transport)
			}
		}
		cancel()
		if transportErr != nil {
			return fmt.Errorf("%s: %w", name, transportErr)
		}
		if !opt.Live && cursor != len(c.HTTP) {
			return fmt.Errorf("%s: %d unused HTTP expectations", name, len(c.HTTP)-cursor)
		}
		if c.ExpectError != "" {
			if err == nil || !strings.Contains(err.Error(), c.ExpectError) {
				return fmt.Errorf("%s: expected error %q was not observed", name, c.ExpectError)
			}
		} else {
			if err != nil {
				if opt.Live {
					return fmt.Errorf("%s: live execution failed (details suppressed to protect credentials)", name)
				}
				return fmt.Errorf("%s: %w", name, err)
			}
			outSchema, _ := json.Marshal(t.OutputSchema)
			if e := (&schema.Validator{}).Validate(t.ToolID, outSchema, result); e != nil {
				return fmt.Errorf("%s: output does not match output_schema", name)
			}
			var actual any
			if err = json.Unmarshal(result, &actual); err != nil {
				return err
			}
			expectedBytes, _ := json.Marshal(c.Expect)
			var expected any
			json.Unmarshal(expectedBytes, &expected)
			if !reflect.DeepEqual(actual, expected) {
				return fmt.Errorf("%s: output differs from expect (values suppressed)", name)
			}
		}
		fmt.Fprintf(w, "PASS %s\n", filepath.Base(name))
	}
	if count == 0 {
		return fmt.Errorf("no matching tests/*.yaml found; MCP discovery is tested after installation in SIEM")
	}
	fmt.Fprintf(w, "%d tests passed\n", count)
	return nil
}
func allowed(p *tcpkg.Project, t tcpkg.Tool, creds map[string]any, raw string) bool {
	u, e := url.Parse(raw)
	if e != nil || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil {
		return false
	}
	hosts := map[string]bool{}
	for _, h := range t.RuntimeConfig.AllowedDomains {
		hosts[strings.ToLower(h)] = true
	}
	for k, f := range p.Provider.CredentialSchema {
		if f["type"] == "url" {
			if s, ok := creds[k].(string); ok {
				v, e := url.Parse(s)
				if e == nil {
					hosts[strings.ToLower(v.Hostname())] = true
				}
			}
		}
	}
	return hosts[strings.ToLower(u.Hostname())]
}
func live(ctx context.Context, r cel.Request) (cel.Response, error) {
	req, err := http.NewRequestWithContext(ctx, r.Method, r.URL, strings.NewReader(r.Body))
	if err != nil {
		return cel.Response{}, fmt.Errorf("invalid request")
	}
	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return fmt.Errorf("redirects are disabled during live testing")
	}}
	res, err := client.Do(req)
	if err != nil {
		return cel.Response{}, fmt.Errorf("live HTTP transport failed")
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, (1<<20)+1))
	if err != nil || len(b) > 1<<20 {
		return cel.Response{}, fmt.Errorf("response unreadable or exceeds 1 MiB")
	}
	return cel.Response{Status: res.StatusCode, Body: string(b), Headers: res.Header}, nil
}
