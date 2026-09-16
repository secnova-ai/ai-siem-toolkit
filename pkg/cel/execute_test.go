package cel

import (
	"context"
	"testing"
)

func TestHeaderSemantics(t *testing.T) {
	for _, tc := range []struct{ program, auth string }{{`get(creds.base_url).body`, "Bearer token"}, {`get_h(creds.base_url, {"X-API-Key":"key"}).body`, ""}, {`post(creds.base_url, {"ids":["a"]}.encode_json()).status`, "Bearer token"}} {
		_, err := Execute(context.Background(), tc.program, map[string]any{}, map[string]any{"base_url": "https://example.com", "access_token": "token"}, func(_ context.Context, r Request) (Response, error) {
			if r.Headers["Authorization"] != tc.auth {
				t.Errorf("auth=%q want=%q", r.Headers["Authorization"], tc.auth)
			}
			return Response{Status: 200, Body: `{"ok":true}`}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
func TestUnsupportedFunctions(t *testing.T) {
	for _, p := range []string{`fetch("url")`, `patch("url", "body")`, `openapi_request()`} {
		if Validate(p) == nil {
			t.Errorf("accepted %s", p)
		}
	}
}
