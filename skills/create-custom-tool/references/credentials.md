# Credentials and authentication

`credential_schema` is a field map, not a JSON Schema object. It declares what users enter after installation. Supported field types are string, text, secret, url, boolean, number, integer and select. `required: true` rejects missing/empty values. Optional defaults must match the field type. URL fields also provide hostnames for the execution allowlist.

```yaml
credential_schema:
  base_url:
    type: url
    required: true
    description: Service origin without a trailing slash.
  api_key:
    type: secret
    required: true
    description: Read-only API key generated in the service administration page.
```

The package contains the schema, never a customer's credential values. Configure multiple credentials for separate accounts/endpoints through existing SIEM credential management. A package belongs to the uploading tenant; do not encode tenant IDs into its provider ID.

| Runtime | Authentication behavior |
|---|---|
| CEL | Read values from `creds`; ordinary HTTP helpers inject an available access_token as Bearer; `_h` helpers use explicit headers |
| WASM | stdin contains creds; the guest explicitly constructs its HTTP headers |
| MCP | Platform MCP client applies the provider authentication configuration and manages supported OAuth flows |

For API Key/Basic, explicitly reference the right credential fields. A CEL provider's template header declaration is not a substitute for its `_h` header expression. WASM also does not inherit headers automatically.

Platform CEL supports its existing client-credentials exchange when the required token URL/client credential fields are provided. Authorization-code OAuth credentials are configured/authorized in SIEM according to the deployment's supported credential workflow. CLI tests never exchange, refresh or persist tokens; supply a fake access_token in mock tests or an existing token for live tests. Verify renewal and callback behavior in SIEM, not by assuming a mocked test covers it.

`credential_test.type` is `http_probe`, `oauth2_client_credentials`, or `skip`. A probe uses a documented harmless endpoint. It is a connectivity check, not a persistent login script. Never run a destructive action to test credentials.

Local live test secrets belong in `.local/credentials.json`; this directory is ignored and excluded from packing. Live testing verifies TLS certificates and refuses redirects, even if the platform credential supports other policies. Do not publish real tokens in examples, command lines, screenshots or logs.
