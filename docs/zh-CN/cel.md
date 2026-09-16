# CEL 工具开发

`runtime_type: cel` 必须提供 `runtime_config.program`。表达式使用 `params`（调用参数）和 `creds`（凭证值）。`allowed_domains` 填不带协议的主机名；未填写时，从 `credential_schema` 中 `type: url` 的字段获取主机。`timeout_seconds` 控制平台 HTTP 超时，本地测试另有 30 秒总时限。

| 函数 | 返回 |
|---|---|
| `get(url)`、`delete(url)` | HTTP 响应对象 |
| `post(url, body)`、`put(url, body)` | HTTP 响应对象，默认 JSON Content-Type |
| `get_h(url, headers)`、`delete_h(url, headers)` | 使用显式请求头 |
| `post_h(url, body, headers)`、`put_h(url, body, headers)` | 使用显式请求头 |
| `value.encode_json()` | JSON 字符串 |
| `decode_json(text)` 或 `text.decode_json()` | 解码后的值 |
| `base64_encode(text)` | 标准 Base64 字符串 |

响应对象包含 `body` 字符串、`status` 整数、`ok`（是否 2xx）。HTTP 非 2xx 不自动变成 CEL 执行错误，必须明确处理，不能把接口错误响应当成业务成功返回。可用 CEL 原生比较、条件表达式和有界集合变换；不能编造 `patch`、HMAC 或 JavaScript 函数。

只发出一次请求，再根据结果分支：

```yaml
runtime_config:
  program: >-
    [get_h(creds.base_url + "/health", {"X-API-Key": creds.api_key})]
      .map(r, r.ok ? {"status": r.body.decode_json().status} :
        {"error": "Health request failed", "http_status": r.status})[0]
```

在输出 Schema 中描述成功和错误两种对象。JSON 请求体使用 `.encode_json()` 构造。不要直接把任意用户输入拼进 URL；当前表达式缺少所需编码器时用 WASM 实现。

普通 HTTP 函数在存在 `creds.access_token` 时注入 Bearer。`_h` 函数使用显式请求头，**不会**自动补 Bearer。API Key、Basic 等需自行引用凭证，例如 `{"Authorization": "Basic " + base64_encode(creds.username + ":" + creds.password)}`。

调用之间无共享登录状态，也没有写回凭证的 API。`openapi_request()` 依赖平台生成的请求计划，不在 v0.1 手写工具协议中。
