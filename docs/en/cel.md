# CEL tools

`runtime_type: cel` requires `runtime_config.program`. The expression receives `params` (validated invocation input) and `creds` (resolved credential values). `allowed_domains` contains hostnames without a scheme; when absent, hosts derive from credential fields declared `type: url`. `timeout_seconds` controls the platform HTTP timeout; local tests impose an overall 30-second budget.

| Function | Result |
|---|---|
| `get(url)`, `delete(url)` | HTTP response map |
| `post(url, body)`, `put(url, body)` | HTTP response map; JSON Content-Type |
| `get_h(url, headers)`, `delete_h(url, headers)` | HTTP response map with explicit headers |
| `post_h(url, body, headers)`, `put_h(url, body, headers)` | HTTP response map with explicit headers |
| `value.encode_json()` | JSON string |
| `decode_json(text)` or `text.decode_json()` | Decoded value |
| `base64_encode(text)` | Standard Base64 string |

The response map contains `body` (string), `status` (integer), `ok` (HTTP 2xx). A non-2xx response is not automatically a CEL execution error. Handle it deliberately; do not silently return an error body as a successful business result. Native CEL expressions support comparisons, ternaries and bounded collection transformations. `patch`, HMAC and arbitrary JavaScript functions are not defined.

Evaluate a request once and branch on its result:

```yaml
runtime_config:
  program: >-
    [get_h(creds.base_url + "/health", {"X-API-Key": creds.api_key})]
      .map(r, r.ok ? {"status": r.body.decode_json().status} :
        {"error": "Health request failed", "http_status": r.status})[0]
```

Describe both success and error objects in `output_schema`. Do not log authentication material. Build JSON bodies using `.encode_json()` rather than string interpolation. Do not concatenate unescaped arbitrary user input into a URL; select WASM for requests requiring encoders unavailable in this environment.

Ordinary HTTP functions inject `Authorization: Bearer` when `creds.access_token` is available. The `_h` functions use your explicit headers and do **not** inject that header. API Key or Basic authentication must therefore be assembled from `creds`, for example `{"Authorization": "Basic " + base64_encode(creds.username + ":" + creds.password)}`.

Every invocation is independent. There is no shared login session or credential mutation API. `openapi_request()` requires a platform-generated request plan and is not part of the hand-authored v0.1 toolkit contract.
