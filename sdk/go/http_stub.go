//go:build !wasip1

// Package sdk provides helper utilities for tool-center WASM guest programs.
// This file is the non-WASM stub so the package compiles cleanly in the host
// binary (where it is never called at runtime).
package sdk

import (
	"errors"
	"strings"
)

// ErrNotWASI is returned by all SDK functions when not running under WASI.
var ErrNotWASI = errors.New("sdk: not running under WASI")

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

func Do(method, rawURL string, headers map[string]string, body string) (*HTTPResponse, error) {
	return nil, ErrNotWASI
}

func Get(rawURL string, headers map[string]string) (*HTTPResponse, error) {
	return nil, ErrNotWASI
}

func Post(rawURL string, headers map[string]string, body string) (*HTTPResponse, error) {
	return nil, ErrNotWASI
}

func Put(rawURL string, headers map[string]string, body string) (*HTTPResponse, error) {
	return nil, ErrNotWASI
}

func Delete(rawURL string, headers map[string]string) (*HTTPResponse, error) {
	return nil, ErrNotWASI
}
