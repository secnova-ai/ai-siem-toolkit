# 从认证文档写出凭证配置

先确认接口文档说的是哪一种认证，再选择下面配方。片段必须合并到已有 `_provider.yaml` 或工具 YAML 中；不是完整独立文件。相同顶层字段只保留一份，不能在文件末尾重复追加第二个 credential_schema。

## 第 1 步：区分“凭证定义”和“使用凭证”

Provider 的 credential_schema 告诉界面让用户填什么；工具的 CEL/WASM 告诉运行时把值放到请求哪里。仅声明一个名叫 api_key 的字段，不会自动把它放到任意厂商需要的请求头。

凭证不应该作为普通工具输入让 AI 每次填写，也不应写进代码、模拟用例之外的公开真实值。服务地址通常也属于凭证，从而复用平台多凭证能力。

## 第 2 步：API Key 请求头

接口要求 `X-API-Key: ...` 时，在 Provider 中定义：

```yaml
credential_schema:
  base_url:
    type: url
    required: true
    description: API 根地址，不带结尾斜杠。
  api_key:
    type: secret
    required: true
    description: 从目标服务管理页面获取的只读 API Key。
```

工具中使用值：

```yaml
runtime_config:
  program: 'get_h(creds.base_url + "/health", {"X-API-Key": creds.api_key}).body.decode_json()'
```

这里的返回表达式仅演示认证，正式工具还应按 [CEL 指南](cel.md) 处理非 2xx。请求头名字必须与目标文档一致，不能统一猜成 Authorization。

## 第 3 步：固定 Bearer Token

将上面的 api_key 字段换为 token，仍使用 type: secret，再在工具中显式设置：

```yaml
runtime_config:
  program: 'get_h(creds.base_url + "/health", {"Authorization": "Bearer " + creds.token}).body.decode_json()'
```

用户填写的是 token 本身，不要再填一遍 `Bearer ` 前缀。固定 token 不代表 OAuth，平台不会因为字段叫 token 就知道怎样续期。

## 第 4 步：Basic 用户名和密码/API Token

在 Provider 中把秘密拆成清晰字段：

```yaml
credential_schema:
  base_url:
    type: url
    required: true
  username:
    type: string
    required: true
    description: 目标服务的用户名或邮箱，按服务要求填写。
  password:
    type: secret
    required: true
    description: 密码或该服务要求的 API Token，不是已经编码的 Basic 字符串。
```

对应工具片段：

```yaml
runtime_config:
  program: >-
    get_h(creds.base_url + "/health",
      {"Authorization": "Basic " + base64_encode(creds.username + ":" + creds.password)}).body.decode_json()
```

不要求客户手工 Base64 编码。这里的 Base64 是编码，不提供额外加密，仍使用正确的服务传输协议。

## 第 5 步：OAuth client_credentials

只有厂商明确支持该 grant 时才采用。下面的 token 路径也是教学路径，必须替换为真实文档值：

```yaml
credential_schema:
  base_url:
    type: url
    required: true
  client_id:
    type: string
    required: true
  client_secret:
    type: secret
    required: true
auth_strategy:
  type: oauth2_client_credentials
  token_url: "{{creds.base_url}}/oauth/token"
credential_test:
  type: oauth2_client_credentials
```

平台使用现有 client-credentials 流程获取 access_token。工具可用普通 `get(url)` 自动携带已有 access_token；若用 `_h`，必须显式写 `Authorization: "Bearer " + creds.access_token`。CLI 本地不换 token：模拟测试还需给 creds 一个假的 access_token，以及 schema 中要求的假 client_id/client_secret。

不能把浏览器授权码/PKCE 流程写成 client_credentials。如果厂商要求用户打开授权页，使用 SIEM 已支持的 OAuth 凭证配置及授权流程，在平台验证 callback、scope、token 续期。此处不承诺仅添加一个 YAML 字段就能建立浏览器授权。

## 第 6 步：MCP 的认证写在 Provider 模板中

```yaml
credential_schema:
  mcp_endpoint:
    type: url
    required: true
  token:
    type: secret
    required: true
auth_strategy:
  headers:
    Authorization: "Bearer {{creds.token}}"
credential_test:
  type: skip
```

这里的双花括号由平台 MCP 认证模板替换，不是 CEL 表达式。创建 OAuth MCP 凭证时按平台 UI 的实际流程授权。`skip` 不验证发现或调用，上传后仍要 Refresh tools 并调用只读工具。

## 第 7 步：配置无副作用的凭证探测

只有确知 `/health` 可用于认证验证时才使用此例：

```yaml
credential_test:
  type: http_probe
  method: GET
  url: "{{creds.base_url}}/health"
  expected_status: 200
  headers:
    X-API-Key: "{{creds.api_key}}"
```

expected_status 填整数。认证接口必须真正检查凭证；一个任何人都返回 200 的公开健康页不能证明 API Key 有效。不要用登录探测期待保存 cookie，也不要用封禁、删除等接口探测。

## 第 8 步：逐层排查

1. validate：字段结构和类型是否正确？本地 CLI 能解析不等于授权成功。
2. mock test：请求头实际是否正确？测试 creds 必须是值，别复制 credential_schema 进去。
3. SIEM 凭证：真实值、默认凭证和 active 状态是否正确？
4. 实际调用：目标地址、执行位置、接口权限、scope、证书是否匹配？OAuth 还要验证过期后续期。

WASM 从 stdin 获得 creds，再通过 SDK 显式构造同样的请求头；不会自动继承这些 CEL 表达式。需要复杂签名时在 WASM 中实现，不编造 CEL 函数。
