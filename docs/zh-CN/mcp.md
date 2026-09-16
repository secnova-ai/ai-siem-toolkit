# 接入已有 MCP 服务

MCP 模板使用 `mcp_discovery: true`，不带工具定义。凭证提供完整 HTTP 地址 `mcp_endpoint` 和 `api_key`，`auth_strategy.headers` 将密钥作为 Bearer 请求头发送。服务采用其他认证方式时修改定义。打包过程不连接服务。

上传后在 SIEM 中配置凭证，需要 OAuth 时完成授权，然后刷新 Available Tools。核对工具名称、参数描述、数量，并执行至少一个只读工具。包校验成功、跳过凭证探测，都不能证明认证和发现成功。

静态工具模式：移除 `mcp_discovery: true`，添加 `tools/NAME.yaml`，设置 `runtime_type: mcp`，将 `runtime_config.tool_name` 填为服务端真实工具名，按服务端定义填写输入输出 Schema。可选固定地址 `runtime_config.endpoint`；否则从凭证取地址。本工具集不允许静态定义与动态发现混用。

MCP 工具包不包含或启动 MCP Server。`tcpkg test` 不模拟 MCP 会话、工具发现和 OAuth，这些流程在 SIEM 中对真实服务验证。支持 HTTP 接入，本地 stdio 服务需要另行部署合适的 HTTP 桥接或服务。
