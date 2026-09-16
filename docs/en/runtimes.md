# Choose a runtime

| Need | Runtime | Boundary |
|---|---|---|
| Simple HTTP request and response mapping | CEL | Only the documented expression environment and functions |
| Signatures, complex parsing, bounded multi-step calls | Go WASM | WASI sandbox, host HTTP ABI, 16 MiB memory, 30-second execution |
| Existing MCP server, long-lived service state or background processing | MCP | Service runs outside the tool package and owns its state |

CEL is an expression language, not JavaScript or Python. Do not invent functions for HMAC, arbitrary networking, persistent variables, loops or sessions. Select WASM when the required transformation cannot be expressed with supported functions. A WASM module cannot open arbitrary sockets or execute native programs; outbound HTTP must use the provided SDK. Requirements outside those constraints belong in an external service, possibly exposed through MCP.

Each CEL/WASM invocation receives parameters and credential values afresh. A login action cannot establish a session for a later action. Platform-managed OAuth is a separate credential capability; it is not a general session store for custom code.

Builtin and legacy Dify provider development, OpenAPI conversion, server deployment, automatic upload and marketplace publication are outside the v0.1 CLI scope.
