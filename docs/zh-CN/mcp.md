# 接入已有 MCP 服务

## 第 1 步：取得服务信息

向服务提供方取得完整 MCP HTTP 地址、认证方式和可用工具清单。普通 REST 地址不是 MCP 地址。本教程使用 Bearer Token；OAuth 服务必须走实际授权流程，不能把 client secret 当作 Bearer Token。

## 第 2 步：初始化并填写 Provider

```sh
tcpkg init my-mcp --runtime mcp
```

将根目录的 `_provider.yaml` 替换为下面的完整内容。动态发现不需要 `tools/*.yaml`。真实 Token 在平台凭证中填写，不写入包。

<!-- tutorial-file: _provider.yaml -->
```yaml
provider_id: my-mcp
version: "1.0.0"
vendor: Example
name: my-mcp
description: Discover and call tools from an existing MCP service.
category: utility
supported_locations: [saas]
mcp_discovery: true
credential_schema:
  mcp_endpoint:
    type: url
    required: true
    description: Complete MCP HTTP endpoint, including the service path.
  api_key:
    type: secret
    required: true
    description: Bearer token accepted by the MCP service.
auth_strategy:
  headers:
    Authorization: "Bearer {{creds.api_key}}"
credential_test:
  type: skip
```

- `mcp_discovery: true`：从服务器发现工具及参数定义。
- `mcp_endpoint`：完整服务地址，包括服务要求的路径。
- `api_key`：本例使用的 Bearer Token 凭证字段。
- `auth_strategy.headers`：引用凭证字段，生成实际请求头。
- `credential_test.type: skip`：跳过独立凭证探测，不表示认证成功。

服务使用 X-API-Key 时，按服务约定修改请求头；服务无认证时，删除 `api_key` 字段和 `auth_strategy`。不能只把字段改为非必填，却保留对不存在值的引用。

## 第 3 步：校验、打包和上传

```sh
tcpkg validate my-mcp
tcpkg pack my-mcp -o my-mcp-1.0.0.tcpkg
tcpkg verify my-mcp-1.0.0.tcpkg
```

上传包，在凭证中填写地址和 Token，设定默认凭证，刷新 Available Tools。核对服务端开放的工具清单，再调用一个只读工具。动态模式没有本地工具测试用例，不能用本地 `test` 代替这一步。

## 第 4 步：排查连接问题

| 现象 | 先检查 |
|---|---|
| 401 / 403 | Token 是否有效、请求头格式、账号权限 |
| 404 / initialize 失败 | 完整地址及路径、服务是否支持 HTTP MCP 协议 |
| 工具数量为 0 | 服务端给该账号开放的工具、同步返回信息 |
| 本地能访问，平台连接失败 | 执行端的网络；localhost 指向执行端而非你的电脑 |

## 可选：维护固定工具集

从 `_provider.yaml` 删除 `mcp_discovery: true`，创建 `tools/get_device.yaml`。下面是完整工具示例，远端工具名与参数必须按真实服务契约替换。

```yaml
tool_id: my-mcp.get_device
name: Get device
description: Read one device by its immutable inventory ID without modifying it.
input_schema:
  type: object
  additionalProperties: false
  required: [device_id]
  properties:
    device_id:
      type: string
      minLength: 1
      description: Exact immutable ID from the server inventory, not a hostname or IP address.
risk_level: low
blast_radius: none
runtime_type: mcp
runtime_config:
  tool_name: get_device
```

`tool_id` 是本地标识，`runtime_config.tool_name` 是远端工具名。重新校验、打包，并在 SIEM 验证调用。

## 接入边界

MCP 模板使用 `mcp_discovery: true`，不带工具定义。凭证提供完整 HTTP 地址 `mcp_endpoint` 和 `api_key`，`auth_strategy.headers` 将密钥作为 Bearer 请求头发送。服务采用其他认证方式时修改定义。打包过程不连接服务。

上传后在 SIEM 中配置凭证，需要 OAuth 时完成授权，然后刷新 Available Tools。核对工具名称、参数描述、数量，并执行至少一个只读工具。包校验成功、跳过凭证探测，都不能证明认证和发现成功。

静态工具模式：移除 `mcp_discovery: true`，添加 `tools/NAME.yaml`，设置 `runtime_type: mcp`，将 `runtime_config.tool_name` 填为服务端真实工具名，按服务端定义填写输入输出 Schema。可选固定地址 `runtime_config.endpoint`；否则从凭证取地址。本工具集不允许静态定义与动态发现混用。

MCP 工具包不包含或启动 MCP Server。`tcpkg test` 不模拟 MCP 会话、工具发现和 OAuth，这些流程在 SIEM 中对真实服务验证。支持 HTTP 接入，本地 stdio 服务需要另行部署合适的 HTTP 桥接或服务。
