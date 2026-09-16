//go:build wasip1

// Package sdk provides helper utilities for tool-center WASM guest programs.
// It is compiled only when targeting GOOS=wasip1 and wraps the host functions
// exported by the "toolcenter" module.
package sdk

import (
	"encoding/json"
	"fmt"
	"strings"
	"unsafe"
)

// maxResponseBuf is the initial guest-side buffer for HTTP responses.
// Responses exceeding this size are rejected; requests are never retried implicitly.
const maxResponseBuf = 4 * 1024 * 1024 // 4 MiB

// httpCallRaw is provided by the toolcenter host module registered in the
// WASM runtime. It performs a controlled outbound HTTP request on the host and
// writes the JSON response into the caller-supplied buffer.
//
// Parameters: reqPtr, reqLen, outPtr, outCap (all uint32 mapped from i32)
// Returns: bytes written (>= 0) or error code (-1 = blocked/error, -2 = buf too small)
//
//go:wasmimport toolcenter http_call
//go:noescape
func httpCallRaw(reqPtr unsafe.Pointer, reqLen uint32, outPtr unsafe.Pointer, outCap uint32) int32

// HTTPResponse is the parsed result of a host HTTP call.
type HTTPResponse struct {
	Status  int
	Body    []byte
	OK      bool
	Headers map[string][]string
}

// Header returns the first value for the given header name, matching
// http.Header.Get's case-insensitive lookup semantics.
func (r *HTTPResponse) Header(name string) string {
	for k, v := range r.Headers {
		if len(v) > 0 && strings.EqualFold(k, name) {
			return v[0]
		}
	}
	return ""
}

// httpRequest mirrors the JSON structure expected by the host function.
type httpRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// httpResponseRaw mirrors the JSON structure written by the host function.
type httpResponseRaw struct {
	Status  int                 `json:"status"`
	Body    string              `json:"body"`
	OK      bool                `json:"ok"`
	Headers map[string][]string `json:"headers"`
}

// Do performs an HTTP request through the toolcenter host function.
// headers may be nil. body may be empty.
func Do(method, rawURL string, headers map[string]string, body string) (*HTTPResponse, error) {
	if headers == nil {
		headers = map[string]string{}
	}
	reqData, err := json.Marshal(httpRequest{
		Method:  method,
		URL:     rawURL,
		Headers: headers,
		Body:    body,
	})
	if err != nil {
		return nil, fmt.Errorf("sdk.Do: marshal request: %w", err)
	}

	outBuf := make([]byte, maxResponseBuf)

	n := httpCallRaw(
		unsafe.Pointer(&reqData[0]), uint32(len(reqData)),
		unsafe.Pointer(&outBuf[0]), uint32(len(outBuf)),
	)
	if n == -2 {
		return nil, fmt.Errorf("sdk.Do: response too large for buffer (%d bytes)", maxResponseBuf)
	}
	if n < 0 {
		return nil, fmt.Errorf("sdk.Do: host blocked or failed (code %d) for %s %s", n, method, rawURL)
	}

	var raw httpResponseRaw
	if err := json.Unmarshal(outBuf[:n], &raw); err != nil {
		return nil, fmt.Errorf("sdk.Do: parse response: %w", err)
	}
	return &HTTPResponse{
		Status:  raw.Status,
		Body:    []byte(raw.Body),
		OK:      raw.OK,
		Headers: raw.Headers,
	}, nil
}

// Get is a convenience wrapper for HTTP GET.
func Get(rawURL string, headers map[string]string) (*HTTPResponse, error) {
	return Do("GET", rawURL, headers, "")
}

// Post is a convenience wrapper for HTTP POST with a JSON body.
func Post(rawURL string, headers map[string]string, body string) (*HTTPResponse, error) {
	h := map[string]string{"Content-Type": "application/json"}
	for k, v := range headers {
		h[k] = v
	}
	return Do("POST", rawURL, h, body)
}

// Put is a convenience wrapper for HTTP PUT with a JSON body.
func Put(rawURL string, headers map[string]string, body string) (*HTTPResponse, error) {
	h := map[string]string{"Content-Type": "application/json"}
	for k, v := range headers {
		h[k] = v
	}
	return Do("PUT", rawURL, h, body)
}

// Delete is a convenience wrapper for HTTP DELETE.
func Delete(rawURL string, headers map[string]string) (*HTTPResponse, error) {
	return Do("DELETE", rawURL, headers, "")
}
