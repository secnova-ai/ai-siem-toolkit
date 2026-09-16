# {{ID}}

Starter health tool: replace `/health`, authentication, schemas and descriptions using your API documentation.

1. `tcpkg validate .`
3. `tcpkg test .` (mocked HTTP; no real requests)
4. `tcpkg pack .`
5. Upload the `.tcpkg` in SIEM → Tools → Add Tool → Upload Package, then configure credentials.

Local secrets belong in `.local/credentials.json`, never in the provider YAML or tests.
For real requests explicitly use `tcpkg test . --live --credentials .local/credentials.json`.
Tests assert exact outputs, so adapt expected values for the target environment.
