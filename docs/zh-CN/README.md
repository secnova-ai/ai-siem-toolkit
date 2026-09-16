# 工具创作流程

1. [快速开始](quickstart.md)：创建并安装第一个工具。
2. [选择类型](runtimes.md)：判断使用 CEL、Go WASM 还是已有 MCP 服务。
3. [设计工具](design.md)：确定有意义的操作、参数描述和风险。
4. 阅读 [CEL 开发](cel.md)、[WASM 开发](wasm.md) 或 [MCP 接入](mcp.md)。
5. 定义 [凭证与认证](credentials.md)。
6. [测试排错](testing.md)，再 [打包安装](packaging.md)。
7. 按需查询 [字段参考](reference.md)、[CLI 命令](cli.md) 和 [兼容范围](compatibility.md)。

命令中的 `<directory>` 指包含 `_provider.yaml` 的源项目目录，不是 `.tcpkg` 文件。
