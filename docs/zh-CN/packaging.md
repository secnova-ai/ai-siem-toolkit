# 打包、安装与更新

源项目包含 `_provider.yaml`、`tools/*.yaml`、可选图标 `_assets/` 和产物 `wasm/`，也可包含源码、测试、README。包根目录保留平台约定，不能额外包一层项目目录。

```text
_provider.yaml
tools/health.yaml
wasm/health.wasm       # 工具引用时包含
_assets/icon.svg       # Provider 引用时包含
```

`pack` 只收集定义及其引用的 WASM、图标，不包含源码、测试和 `.local` 凭证。先校验并验证 ZIP 完整性，再生成目标文件，不覆盖已有包。多个工具引用相同文件时不会重复写入 ZIP。

解析器限制压缩包及展开总大小为 256 MiB，单文件 32 MiB；具体部署的 UI、反向代理可能限制更低，需以部署为准。verify 检查 CRC、路径、重复条目和定义。校验成功不代表目标服务可达或凭证可用。

在 SIEM → Tools → Add Tool → Upload Package 上传，配置凭证后验证执行。发布变更时递增语义版本，保留旧源码和包，以便按平台支持的更新方式恢复。CLI 不提供自动回滚承诺。

工具使用者只需 `.tcpkg`，开发者可另外获取源码。分发工具包不会授予目标服务权限，仍由租户配置适当凭证。
