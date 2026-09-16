# Compatibility contract

Initial baseline: Tool Center dev/3.0.7. The toolkit retains the existing ZIP-based `.tcpkg` layout. It targets customer-authored CEL, Go WASM and MCP providers, not exports of every internal platform feature. Toolkit version and provider version are independent.

The package parser, credential type/default validation, JSON Schema validator, CEL function declarations and WASM HTTP SDK originate from the matching Tool Center implementation. Public packages have no database or private service dependency. CLI validation adds stricter authoring checks, so it can reject a legacy package previously accepted by a server. It does not bypass server validation.

Public library entry points include `pkg/tcpkg.ParsePackage`, `Load`, `ValidateFiles`, `Pack`, `Verify`; `pkg/schema.ValidateDefinition`; `pkg/credential.ValidateSchemaDefinition`; and `pkg/cel.Validate`. Server consumers should pin a reviewed release and keep their tenant-specific checks. This repository does not by itself update an existing server dependency.

Local mock tests execute real CEL expressions and real compiled WASM, with injected HTTP responses. They do not exercise OAuth renewal, RBAC, approvals, tenant scoping, platform credential activation, SSRF deployment policy or MCP protocol sessions. Live tests use explicit credentials, strict TLS and no redirects. These differences are deliberate and must not be presented as production parity.

Before release: run all tests, cross-compile CLI binaries, validate docs and SDK template synchronization, and validate generated CEL/WASM/MCP packages using the target Tool Center parser/loader. A deployment upload and actual invocation remain separate acceptance checks.

Unknown new platform fields should first be reviewed against their server implementation, then added with tests and docs. Do not weaken all unknown-field checks to accommodate one field.
