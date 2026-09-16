# 从一个接口请求，逐步写出可安装的工具

完成本教程后，你将拥有一个 `device-api.search` 工具：按结构化条件查询设备，返回设备 ID，能本地测试并打包上传。所有地址和密钥均为教学数据；`example.com` 不是可联调的设备接口。无需真实设备或凭证即可完成模拟测试。

先按仓库 README 安装 CLI，用 `tcpkg version` 确认命令可用。不熟悉缩进、数组和多行字符串时，先阅读 [YAML 基础](yaml-basics.md)。下面命令均在项目目录的上一级执行。

## 第 1 步：读懂接口，不急着写 YAML

假设设备文档提供了以下请求（仅用于理解，不需要执行）：

```sh
curl -X POST 'https://inventory.example.com/devices/search' -H 'X-API-Key: example-only' -H 'Content-Type: application/json' -d '{"filters":[{"field":"hostname","value":"web-01"}],"limit":10}'
```

成功响应：

```json
{"devices":[{"id":"device-42"}]}
```

没有结果时返回 `{"devices":[]}`，认证失败返回 HTTP 401。这里的 POST 是只读搜索，不修改设备。

先把接口拆成这些信息：

| 接口信息 | 放在哪里 | 原因 |
|---|---|---|
| 服务根地址 | 凭证 `base_url` | 不同客户的地址不同 |
| X-API-Key 的值 | 凭证 `api_key` | 是账户秘密，不应每次由 AI 填入 |
| filters、limit | 工具的 input_schema | 每次调用由用户或 AI 决定 |
| POST 和 /devices/search | CEL program | 是此动作固定的请求逻辑 |
| devices 及设备 id | output_schema | 让调用方知道如何理解结果 |
| 查询是只读 | risk_level: low | 风险以实际效果为准 |

`provider_id` 选 `device-api`，动作名选 `search`，完整工具 ID 为 `device-api.search`。一个 Provider 下可有多个工具；这次只实现一个明确有用的动作。

## 第 2 步：初始化项目，清理教学占位文件

```sh
tcpkg init device-api --runtime cel
```

初始化生成健康检查示例。用编辑器移除新建目录中的 `tools/health.yaml` 及 `tests/` 下原有的健康检查用例，接下来替换成下面的搜索工具。只删除新项目的样例文件，不要删除你已有项目的工具。保留 `.gitignore`、README、LICENSE。 旧版 v0.1.0 CLI 若未生成 LICENSE，从 Release 的 LICENSE 附件复制到项目根目录。

最终目录应为：

```text
device-api/
  _provider.yaml
  tools/search.yaml
  tests/search.yaml
  tests/empty.yaml
  tests/unauthorized.yaml
  .gitignore
  README.md
  LICENSE
```

文件名帮助你组织源码；真正决定工具 ID 的是文件内的 `tool_id`。

## 第 3 步：完整替换 `_provider.yaml`

<!-- tutorial-file: _provider.yaml -->
```yaml
provider_id: device-api
version: "1.0.0"
vendor: Example
name: Device inventory example
description: Illustrative inventory API; adapt endpoints to your vendor documentation.
category: utility
supported_locations: [saas]
credential_schema:
  base_url:
    type: url
    required: true
    description: API origin without a trailing slash.
  api_key:
    type: secret
    required: true
    description: API key with read-only inventory access.
credential_test:
  type: skip
```

这是一个完整文件，可以直接保存。逐项理解：

- `provider_id` 是稳定技术标识；`name` 是界面显示名称，可以修改显示名称而不改变 ID。
- `version` 是工具包自己的版本，不是 CLI 或设备版本。使用带引号的 `"1.0.0"`。
- `vendor` 是服务厂商，`category` 是分类；真实项目应换成实际值。
- `supported_locations: [saas]` 表示平台侧执行。本地设备必须从实际执行环境可达，浏览器能访问不代表平台能访问。
- `credential_schema` 定义需要填写哪些字段，不放真实字段值。
- `base_url.type: url` 既帮助校验 URL，也用于运行时提取允许访问的主机。
- `api_key.type: secret` 声明敏感值；`required: true` 是凭证字段的必填标志。
- `credential_test.type: skip` 仅代表这个例子未配置独立探测接口，不证明凭证有效。真实产品有安全探测接口时按 [认证配方](auth-recipes.md) 改成 http_probe。

## 第 4 步：创建完整的 `tools/search.yaml`

<!-- tutorial-file: tools/search.yaml -->
```yaml
tool_id: device-api.search
name: Search devices
description: Search device inventory using exact-match filters. Read-only despite using POST. Returns up to limit devices; an empty devices array means no match. Does not isolate or modify devices.
risk_level: low
blast_radius: none
runtime_type: cel
input_schema:
  type: object
  additionalProperties: false
  required: [filters, limit]
  properties:
    filters:
      type: array
      minItems: 1
      maxItems: 10
      description: Exact-match filters combined with AND. Each entry supplies a supported field and a nonempty value.
      items:
        type: object
        additionalProperties: false
        required: [field, value]
        properties:
          field:
            type: string
            enum: [hostname, ip]
            description: Inventory field to match; hostname is the exact registered name, ip is the device address.
          value:
            type: string
            minLength: 1
            description: Exact value for the selected field, without wildcard characters; for example web-01.
      examples:
        - [{field: hostname, value: web-01}]
    limit:
      type: integer
      minimum: 1
      maximum: 100
      description: Maximum returned device count, 1 through 100. Required explicitly; this tool does not fetch additional pages.
output_schema:
  type: object
  properties:
    devices:
      type: array
      description: Matched device records; empty when nothing matches.
      items:
        type: object
        required: [id]
        properties:
          id:
            type: string
            description: Immutable device identifier.
    error:
      type: string
      description: Sanitized API error explanation.
    http_status:
      type: integer
      description: HTTP error status.
  oneOf:
    - required: [devices]
    - required: [error, http_status]
runtime_config:
  program: >-
    [post_h(creds.base_url + "/devices/search",
      {"filters": params.filters, "limit": params.limit}.encode_json(),
      {"X-API-Key": creds.api_key})]
      .map(r, r.ok ? {"devices": r.body.decode_json().devices} :
        {"error": "Inventory request failed", "http_status": r.status})[0]
```

这个文件稍长，因为把用户和 AI 正确调用所需的信息写完整了。分成四部分读：

1. 开头的 ID、名称、描述和风险说明“这是什么动作”。描述特别说明只读、空结果和分页边界。
2. `input_schema` 描述调用参数。`filters` 是数组，数组里的每项是对象；`items.properties` 定义 `field` 和 `value`。`limit` 是整数，范围 1～100。详细规则见 [参数 Schema 教程](schemas.md)。
3. `output_schema` 描述返回值，`oneOf` 表示成功与错误结构恰好匹配其中一种。非 2xx 返回显式错误对象；此结构不意味着平台自动把调用标为执行失败，使用结果的流程需检查 `error`。
4. `runtime_config.program` 是 CEL，不是 YAML 字段模板。

表达式逐段解释：

| 片段 | 作用 |
|---|---|
| `creds.base_url + "/devices/search"` | 从所选凭证取得地址，加上固定路径 |
| `params.filters`、`params.limit` | 取得本次调用的业务参数 |
| `{...}.encode_json()` | 将对象编码成 JSON 请求体，不手工拼接 JSON 字符串 |
| `{"X-API-Key": creds.api_key}` | 显式将秘密放入认证请求头 |
| `[post_h(...)]` | 执行一次请求，把响应放进单元素列表 |
| `.map(r, ...)[0]` | 使用响应 r 构造最终对象，取出唯一结果；不会再请求一次 |
| `r.ok ? ... : ...` | 2xx 解析 devices，非 2xx 返回可识别错误对象 |

`post_h` 已默认使用 JSON Content-Type。`_h` 函数不会额外注入 Bearer，认证请求头由你明确提供。不要把 `{{creds.api_key}}` 写进 CEL：那是 Provider 模板语法，CEL 应使用 `creds.api_key`。

## 第 5 步：先校验结构

```sh
tcpkg validate device-api
```

应看到 `valid device-api: 1 static tools, discovery=false`。如果提示 unknown field、CEL compile 或 Schema 错误，先按文件和字段修复，不要进入打包。校验不会发出 HTTP 请求，也不会验证真实 API Key。

## 第 6 步：用可控数据测试成功路径

创建 `tests/search.yaml`：

<!-- tutorial-file: tests/search.yaml -->
```yaml
name: Nested filters produce the correct request
tool: device-api.search
params:
  filters:
    - field: hostname
      value: web-01
  limit: 10
creds:
  base_url: https://inventory.example.com
  api_key: example-only
http:
  - request:
      method: POST
      url: https://inventory.example.com/devices/search
      headers:
        X-API-Key: example-only
        Content-Type: application/json
      body: '{"filters":[{"field":"hostname","value":"web-01"}],"limit":10}'
    response:
      status: 200
      body: '{"devices":[{"id":"device-42"}]}'
expect:
  devices:
    - id: device-42
```

`tool` 选择要测的 ID；`params` 是调用实例，不是参数定义；`creds` 是假凭证值，不是 credential_schema；`http` 按调用顺序列出预期请求和模拟响应；`expect` 是完整预期结果。

请求体和 URL 精确比较，JSON 对象键顺序也会影响这里的字符串比较。此例 encode_json 会生成示例中的稳定键序。mock 只检查列出的请求头，可省略不关心的头。

```sh
tcpkg test device-api
```

现在应看到 `PASS search.yaml` 和 `1 tests passed`。测试实际执行 CEL，只把 HTTP 传输替换为模拟数据，不会访问 inventory.example.com。

## 第 7 步：补上空结果和认证失败

创建 `tests/empty.yaml`：

<!-- tutorial-file: tests/empty.yaml -->
```yaml
name: No matching device is a valid result
tool: device-api.search
params:
  filters:
    - field: hostname
      value: web-01
  limit: 10
creds:
  base_url: https://inventory.example.com
  api_key: example-only
http:
  - request:
      method: POST
      url: https://inventory.example.com/devices/search
      headers:
        X-API-Key: example-only
        Content-Type: application/json
      body: '{"filters":[{"field":"hostname","value":"web-01"}],"limit":10}'
    response:
      status: 200
      body: '{"devices":[]}'
expect:
  devices: []
```

创建 `tests/unauthorized.yaml`：

<!-- tutorial-file: tests/unauthorized.yaml -->
```yaml
name: Authentication failure is explicit
tool: device-api.search
params:
  filters:
    - field: hostname
      value: web-01
  limit: 10
creds:
  base_url: https://inventory.example.com
  api_key: example-only
http:
  - request:
      method: POST
      url: https://inventory.example.com/devices/search
      headers:
        X-API-Key: example-only
        Content-Type: application/json
      body: '{"filters":[{"field":"hostname","value":"web-01"}],"limit":10}'
    response:
      status: 401
      body: '{"devices":[{"id":"device-42"}]}'
expect:
  error: Inventory request failed
  http_status: 401
```

再次运行 `tcpkg test device-api`，应有 3 个测试通过。401 测试使用 `expect`，因为表达式返回了一个错误对象；只有真正的输入校验或 CEL/WASM 执行错误才使用 `expect_error`。

再做一个可撤销练习：把成功用例的 `limit` 改成 0，测试应在输入校验阶段失败，不发出 HTTP。恢复 10 后重新通过。不要为了通过测试放宽接口本身规定的范围。

## 第 8 步：打包和检查产物

```sh
tcpkg pack device-api -o device-api-1.0.0.tcpkg
tcpkg verify device-api-1.0.0.tcpkg
```

归档包含 `_provider.yaml` 与 `tools/search.yaml`；测试、凭证、README、LICENSE 不放进 .tcpkg。分发复制模板代码形成的包时，同时提供项目 LICENSE。已存在同名包时 pack 会拒绝覆盖；修改工具并增加版本后使用新文件名。

## 第 9 步：在 SIEM 中验证真实服务

将教学路径和响应契约改成真实接口后，再上传包。进入 Tools → Add Tool → Upload Package，打开安装后的 Credentials，填写真实 base_url 和 api_key，按需设为默认凭证。执行只读 search，传入与下面一致结构的参数：

```json
{"filters":[{"field":"hostname","value":"web-01"}],"limit":10}
```

核对返回的设备 ID 确实属于目标环境。遇到 error/http_status 时先检查认证和接口，不能把它当作“没有查询结果”。真实测试应另外验证权限、审批、网络可达性及凭证状态。

不要直接对教学域名使用 live。确需从开发机联调时，按 [测试指南](testing.md) 将凭证保存在 `.local/credentials.json`，显式执行 `--live --credentials`。

## 第 10 步：下一次如何扩展

增加一个操作：新增一个 tools/*.yaml，保持同一 provider_id 前缀，为它增加 tests/*.yaml，并递增包版本。不要把“先登录、以后复用会话”拆成假设有共享状态的工具。CEL 不支持的签名、编码或复杂流程按 [WASM 教程](wasm.md) 实现；持久状态、后台运行等考虑 [MCP](mcp.md)。
