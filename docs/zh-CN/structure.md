# 工具定义结构速查

这页给熟悉开发的作者和 AI 核对结构。下面是两份匹配的完整文件，接口为教学示例。复制后要统一修改 Provider ID、接口契约，并增加测试；其它页面的局部片段不能单独当成完整工具。

<!-- tutorial-file: _provider.yaml -->
```yaml
provider_id: sample-api
version: "1.0.0"
vendor: Example
name: sample-api
description: Read service health through a documented HTTP API.
category: utility
supported_locations: [saas]
credential_schema:
  base_url:
    type: url
    required: true
    description: Service origin, e.g. https://api.example.com, without a trailing slash.
  api_key:
    type: secret
    required: true
    description: API key with permission to read service health.
credential_test:
  type: http_probe
  method: GET
  url: "{{creds.base_url}}/health"
  headers:
    X-API-Key: "{{creds.api_key}}"
```

<!-- tutorial-file: tools/health.yaml -->
```yaml
tool_id: sample-api.health
name: Read service health
description: Read the service health without changing configuration. Returns the service status.
input_schema:
  type: object
  properties: {}
  additionalProperties: false
output_schema:
  type: object
  properties:
    status:
      type: string
      description: Service-reported health status, for example ok.
    error:
      type: string
      description: Sanitized explanation when the HTTP request failed.
    http_status:
      type: integer
      description: Non-success HTTP status reported by the service.
  oneOf:
    - required: [status]
    - required: [error, http_status]
risk_level: low
blast_radius: none
runtime_type: cel
runtime_config:
  program: >-
    [get_h(creds.base_url + "/health", {"X-API-Key": creds.api_key})]
      .map(r, r.ok ? {"status": r.body.decode_json().status} :
        {"error": "Health request failed", "http_status": r.status})[0]
```

## 结构不变量

- 根目录放 `_provider.yaml`；工具放 `tools/*.yaml`；WASM 放 `wasm/NAME.wasm`；图标放 `_assets/NAME`。压缩包不能额外套一层目录。
- Provider 必须有 provider_id、version、vendor、name、category。Tool 必须有 tool_id、name、risk_level、runtime_type 及对应必填运行配置。tool_id 以 provider_id 加点开头。
- credential_schema 是字段映射，使用字段内的 `required: true`，不是 JSON Schema 的 properties 格式。
- 输入输出是 JSON Schema：对象使用 properties 和 `required: [字段名]`，数组使用 items，整数类型为 integer，不能使用 secret/url。
- 嵌套 required 只约束当前对象；可选不等于可空。参数 default 不自动填入 params，缺省行为需要实现。
- CEL 必须有 program，变量为 params/creds。Provider 的双花括号模板不是 CEL，_h 函数需显式认证头。
- WASM 必须有纯文件名 artifact_ref，Go 源码位于 src/NAME，使用宿主 HTTP SDK，不填写 CEL program。
- 静态 MCP 使用 tool_name，可选 endpoint；动态 MCP 在 Provider 上设置 mcp_discovery: true，不混入静态工具定义。工具包不会启动 MCP Server。
- 测试选择工具的键是 tool，不是 tool_id；creds 是假凭证值，不是 credential_schema；http 是按顺序的请求预期。
- 单文件单文档，无别名和重复键。合并片段后必须校验整个项目，再测试、打包、验包。离线检查不能证明真实凭证或 MCP 可用。
