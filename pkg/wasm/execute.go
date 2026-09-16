// Package wasm provides the local WASI test host for Tool Center's HTTP ABI.
package wasm

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/secnova-ai/ai-siem-toolkit/pkg/cel"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type capped struct {
	bytes.Buffer
	limit int
}

func (w *capped) Write(p []byte) (int, error) {
	if w.Len()+len(p) > w.limit {
		return 0, fmt.Errorf("WASM output limit exceeded")
	}
	return w.Buffer.Write(p)
}
func Execute(ctx context.Context, code []byte, params, creds map[string]any, transport cel.Transport) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithMemoryLimitPages(256).WithCloseOnContextDone(true))
	defer r.Close(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return nil, err
	}
	var hostErr error
	_, err := r.NewHostModuleBuilder("toolcenter").NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, p, n, o, cap uint32) int32 {
		b, ok := m.Memory().Read(p, n)
		if !ok || n > 1<<20 {
			hostErr = fmt.Errorf("invalid WASM HTTP request memory")
			return -1
		}
		var req cel.Request
		if e := json.Unmarshal(b, &req); e != nil {
			hostErr = e
			return -1
		}
		res, e := transport(ctx, req)
		if e != nil {
			hostErr = e
			return -1
		}
		raw, e := json.Marshal(map[string]any{"status": res.Status, "body": res.Body, "ok": res.Status >= 200 && res.Status < 300, "headers": res.Headers})
		if e != nil {
			hostErr = e
			return -1
		}
		if len(raw) > int(cap) {
			hostErr = fmt.Errorf("HTTP response exceeds guest buffer")
			return -2
		}
		if !m.Memory().Write(o, raw) {
			return -1
		}
		return int32(len(raw))
	}).Export("http_call").Instantiate(ctx)
	if err != nil {
		return nil, err
	}
	input, err := json.Marshal(map[string]any{"params": params, "creds": creds})
	if err != nil {
		return nil, err
	}
	out := &capped{limit: 1 << 20}
	stderr := &capped{limit: 64 << 10}
	_, err = r.InstantiateWithConfig(ctx, code, wazero.NewModuleConfig().WithStdin(bytes.NewReader(input)).WithStdout(out).WithStderr(stderr).WithSysWalltime().WithSysNanotime().WithRandSource(rand.Reader))
	if hostErr != nil {
		return nil, hostErr
	}
	if err != nil {
		return nil, fmt.Errorf("WASM execution: %w", err)
	}
	// Logs are deliberately not printed; guest code could include credentials.
	raw, err := io.ReadAll(out)
	if err != nil {
		return nil, err
	}
	if !json.Valid(raw) {
		return nil, fmt.Errorf("WASM stdout must contain exactly one JSON value; write logs to stderr")
	}
	return raw, nil
}
