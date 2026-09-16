# Go WASM 工具开发

```sh
tcpkg init my-wasm --runtime wasm --language go
tcpkg build my-wasm
tcpkg test my-wasm
tcpkg pack my-wasm
```

模板包含独立 Go 模块、`sdk/` 源码、`src/health/main.go`、工具定义和 HTTP 模拟测试，无需公司私有依赖。`artifact_ref: health.wasm` 对应 `src/health/`。多个独立入口使用不同产物名称及源码目录。

编译目标是 `GOOS=wasip1 GOARCH=wasm`，不是浏览器使用的 `GOOS=js`。平台提供 WASI 和 `toolcenter.http_call` 宿主函数。程序从 stdin 读取 `{"params": {...}, "creds": {...}}`，向 stdout 输出一个 JSON 值，日志写 stderr。执行失败以非零状态退出。不要把日志混入 stdout，也不要输出凭证。

HTTP 使用 `sdk.Get`、`sdk.Post`、`sdk.Put`、`sdk.Delete` 或 `sdk.Do(method, url, headers, body)`，宿主检查出站主机白名单。响应提供 Status、Body、OK、Headers、Header(name)。SDK 不自动重试，避免写操作重复执行。先检查状态码，再解析响应。

平台限制为 16 MiB 线性内存、30 秒执行、1 MiB stdout、64 KiB stderr；SDK 响应缓冲区为 4 MiB。本地 WASM 测试应用相同的内存、输出、时间限制及 HTTP ABI。即使编译成功，大型依赖和临时缓冲也可能超限。

沙箱不挂载文件系统，不保留调用状态，不运行本机进程，也不提供任意 socket。使用 SDK 宿主 HTTP，不直接使用 Go 网络客户端。单次调用应是有界工作，持久会话、后台任务应使用外部服务或 MCP。

发布后也可引用 `github.com/secnova-ai/ai-siem-toolkit/sdk/go`。模板直接附带 SDK 源码，因此初始化和示例编译不需要下载私有依赖；升级时保持 SDK 与平台 ABI 一致。
