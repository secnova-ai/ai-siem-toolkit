package cel

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Request is the fully constructed outbound request, before transport.
type Request struct {
	Method  string            `json:"method" yaml:"method"`
	URL     string            `json:"url" yaml:"url"`
	Headers map[string]string `json:"headers" yaml:"headers"`
	Body    string            `json:"body" yaml:"body"`
}
type Response struct {
	Status  int                 `json:"status" yaml:"status"`
	Body    string              `json:"body" yaml:"body"`
	Headers map[string][]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}
type Transport func(context.Context, Request) (Response, error)
type httpExecutor struct {
	ctx       context.Context
	transport Transport
	token     string
}

func (e *httpExecutor) call(method, url, body string, h map[string]any, explicit bool) (map[string]any, error) {
	if e.transport == nil {
		return nil, fmt.Errorf("HTTP transport unavailable during validation")
	}
	headers := map[string]string{}
	if method == "POST" || method == "PUT" {
		headers["Content-Type"] = "application/json"
	}
	if !explicit && e.token != "" {
		headers["Authorization"] = "Bearer " + e.token
	}
	for k, v := range h {
		if s, ok := v.(string); ok {
			headers[k] = s
		}
	}
	r, err := e.transport(e.ctx, Request{method, url, headers, body})
	if err != nil {
		return nil, err
	}
	return map[string]any{"status": int64(r.Status), "body": r.Body, "ok": r.Status >= 200 && r.Status < 300}, nil
}
func (e *httpExecutor) get(u string) (map[string]any, error) { return e.call("GET", u, "", nil, false) }
func (e *httpExecutor) post(u, b string) (map[string]any, error) {
	return e.call("POST", u, b, nil, false)
}
func (e *httpExecutor) put(u, b string) (map[string]any, error) {
	return e.call("PUT", u, b, nil, false)
}
func (e *httpExecutor) delete(u string) (map[string]any, error) {
	return e.call("DELETE", u, "", nil, false)
}
func (e *httpExecutor) getH(u string, h map[string]any) (map[string]any, error) {
	return e.call("GET", u, "", h, true)
}
func (e *httpExecutor) postH(u, b string, h map[string]any) (map[string]any, error) {
	return e.call("POST", u, b, h, true)
}
func (e *httpExecutor) putH(u, b string, h map[string]any) (map[string]any, error) {
	return e.call("PUT", u, b, h, true)
}
func (e *httpExecutor) deleteH(u string, h map[string]any) (map[string]any, error) {
	return e.call("DELETE", u, "", h, true)
}

func Validate(program string) error {
	if strings.TrimSpace(program) == "" || len(program) > 65536 {
		return fmt.Errorf("program must contain 1..65536 bytes")
	}
	env, err := newCELEnv(&httpExecutor{})
	if err != nil {
		return err
	}
	_, issues := env.Compile(program)
	return issues.Err()
}

// Execute never exchanges OAuth tokens. Supply a test access_token or explicit headers.
func Execute(ctx context.Context, program string, params, creds map[string]any, t Transport) (json.RawMessage, error) {
	token, _ := creds["access_token"].(string)
	env, err := newCELEnv(&httpExecutor{ctx: ctx, transport: t, token: token})
	if err != nil {
		return nil, err
	}
	ast, issues := env.Compile(program)
	if issues.Err() != nil {
		return nil, issues.Err()
	}
	p, err := env.Program(ast)
	if err != nil {
		return nil, err
	}
	v, _, err := p.ContextEval(ctx, map[string]any{"params": params, "creds": creds})
	if err != nil {
		return nil, err
	}
	return valToJSON(v)
}
