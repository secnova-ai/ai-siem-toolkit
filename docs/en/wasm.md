# Go WASM tools

```sh
tcpkg init my-wasm --runtime wasm --language go
tcpkg build my-wasm
tcpkg test my-wasm
tcpkg pack my-wasm
```

The starter includes a standalone Go module, SDK source in `sdk/`, `src/health/main.go`, a tool definition and a mocked HTTP test. No private Go modules are required. `artifact_ref: health.wasm` maps to `src/health/`. Add another tool with another artifact name and matching source directory for independent entry points.

Build target: `GOOS=wasip1 GOARCH=wasm`, not JavaScript's `GOOS=js`. Tool Center supplies WASI and a `toolcenter.http_call` import. The module reads one JSON object `{"params": {...}, "creds": {...}}` from stdin, writes one JSON value to stdout, and writes logs to stderr. Return a nonzero exit status on execution failure. Never print logs to stdout or credentials to either stream.

Use `sdk.Get`, `sdk.Post`, `sdk.Put`, `sdk.Delete`, or `sdk.Do(method, url, headers, body)` for HTTP. The host enforces allowed domains. `HTTPResponse` exposes Status, Body, OK, Headers and Header(name). Calls are not implicitly retried: retrying a write may duplicate an action. Check status before decoding JSON.

The runtime provides 16 MiB linear memory, a 30-second execution budget, 1 MiB stdout and 64 KiB stderr. The SDK response buffer is 4 MiB. Local tests apply the same WASM memory/output/time limits and the HTTP host ABI. Large dependency trees and large temporary buffers may exceed the budget even when compilation succeeds.

The sandbox has no mounted filesystem, persistent state, native process execution or general socket networking. Use host HTTP rather than Go's ordinary network client. Only bounded work belongs in a single invocation. Stateful or long-running workflows should use an external service/MCP.

The public SDK is also importable as `github.com/secnova-ai/ai-siem-toolkit/sdk/go` once a revision is published. Starter projects vendor its source so initialization and example builds work offline. Keep the SDK and platform ABI compatible when upgrading.
