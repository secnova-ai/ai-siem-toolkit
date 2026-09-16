# Testing and troubleshooting

Each `tests/*.yaml` file is one test. It selects a tool, supplies params/creds, lists HTTP expectations in order, and asserts the complete decoded output. An empty expectation means JSON null, not 'skip comparison'. Missing or extra requests fail the test. Header expectations compare the named headers; body and URL compare exactly. Mock status/body/headers supply the response. Tests never fall back to real networking.

```yaml
name: Read health
tool: my-api.health
params: {}
creds:
  base_url: https://api.example.com
  api_key: example-not-a-real-secret
http:
  - request:
      method: GET
      url: https://api.example.com/health
      headers: {X-API-Key: example-not-a-real-secret}
    response:
      status: 200
      body: '{"status":"ok"}'
expect: {status: ok}
```

Use `expect_error: substring` to assert an execution/input validation error instead of output. Unexpected requests still fail and are not excused by expect_error. Test at least success, empty results, invalid input, authentication rejection and malformed/error responses. The toolkit validates input and output schemas; output validation is an authoring check, not a claim that every platform invocation enforces it.

```sh
tcpkg test my-api
tcpkg test my-api --tool my-api.health
tcpkg test my-api --live --credentials my-api/.local/credentials.json
```

Live mode replaces fake creds and mock responses with real requests, retaining input and output assertions. Only run tests whose actual effects are intended. Filter to a read-only tool first. OAuth renewal is not performed. Redirects are rejected; TLS is verified; response bodies are capped at 1 MiB. The test runner suppresses actual result/credential values on mismatch. Guest stderr is retained privately during execution and not printed by the runner.

| Failure | Check |
|---|---|
| CEL undeclared function | Use supported functions or select WASM |
| No allowed hostname | Declare service address as a URL credential or configure allowed_domains |
| Request mismatch | Method, complete URL, exact serialized body, expected headers |
| Output mismatch | Correct response decoding, success/error schema and expected types |
| Missing WASM | Run build, check src/NAME and artifact_ref |
| WASM memory trap | Reduce dependencies/buffers; compilation alone does not prove runtime fitness |
| Invalid stdout | Output exactly one JSON value, log only to stderr |
| MCP test unavailable | Use validate/pack, then discover and invoke in SIEM |

Real deployment verification additionally covers tenant access, permission/approval policy, credential activation/default selection, OAuth expiration and target network reachability.
