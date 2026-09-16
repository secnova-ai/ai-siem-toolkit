# Nested JSON input example

This is an illustrative API, not an integration for a real vendor. It demonstrates nested array/object parameter descriptions, bounds, explicit limits, JSON body construction, read-only POST risk classification and HTTP error handling.

```sh
tcpkg validate examples/cel-search
tcpkg test examples/cel-search
tcpkg pack examples/cel-search
```

See `tools/search.yaml` and `tests/search.yaml`. Replace the API contract using your vendor's documentation before live testing.
