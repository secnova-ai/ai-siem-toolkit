# 工具创作流程

## 第一次编写工具：按这个顺序阅读

1. [YAML 基础](yaml-basics.md)：缩进、类型、多行字符串，以及常见错误。
2. [从接口到工具的完整教程](tutorial.md)：逐步创建 5 个完整文件，运行成功、空结果、认证失败测试，并打包安装。
3. [参数 Schema 教程](schemas.md)：对象、数组、嵌套必填、枚举、可空与默认行为。
4. [认证配方](auth-recipes.md)：API Key、Bearer、Basic、OAuth、MCP 和凭证探测的正确放置位置。

已经了解流程后，再查下面的专项指南和 [结构速查](structure.md)。CI 检查文档中的 YAML，并将带 `tutorial-file` 标记的示例组装成项目校验、打包；主教程还会实际运行测试。


1. [快速开始](quickstart.md)：创建并安装第一个工具。
2. [选择类型](runtimes.md)：判断使用 CEL、Go WASM 还是已有 MCP 服务。
3. [设计工具](design.md)：确定有意义的操作、参数描述和风险。
4. 阅读 [CEL 开发](cel.md)、[WASM 开发](wasm.md) 或 [MCP 接入](mcp.md)。
5. 定义 [凭证与认证](credentials.md)。
6. [测试排错](testing.md)，再 [打包安装](packaging.md)。
7. 按需查询 [字段参考](reference.md)、[CLI 命令](cli.md) 和 [兼容范围](compatibility.md)。

命令中的 `<directory>` 指包含 `_provider.yaml` 的源项目目录，不是 `.tcpkg` 文件。
