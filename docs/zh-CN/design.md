# 设计有意义的工具

先明确用户要完成的事情，再确定工具清单，不要把每个厂商接口机械地变成工具。查询设备、隔离指定设备等可以独立使用的动作有价值；仅输出登录 token 的中间步骤通常不能满足用户需求。目标选择、权限、行为或影响范围存在歧义时先澄清。创建前确认工具清单及副作用；用户已明确批准同一清单时无需重复确认。

Provider ID 使用简短稳定名称，例如 `firewall`，Tool ID 使用 `firewall.block_ip`。描述必须说明适用情形、实际操作、前置条件和返回含义。界面短说明可以精简，面向 AI 的说明必须保留正确调用所需细节。

每个参数写明格式、单位、合法值、值从哪里获取，以及省略、null、空值是否不同。接口有枚举或边界时写进 Schema。不要把嵌套 JSON 隐藏成未说明的字符串。

```yaml
input_schema:
  type: object
  required: [targets]
  properties:
    targets:
      type: array
      minItems: 1
      description: 从资产查询结果中选择的待隔离设备，每个元素影响一台设备。
      items:
        type: object
        additionalProperties: false
        required: [device_id]
        properties:
          device_id:
            type: string
            description: 资产接口返回的精确设备 ID，不是主机名或 IP。
          reason:
            type: string
            description: 可选审计原因，省略时由服务使用默认原因，不得包含秘密。
      examples:
        - [{device_id: "device-42", reason: "Confirmed incident"}]
```

风险按实际效果确定。只读 GET 通常低，配置写入通常中，删除、隔离通常高；POST 搜索可以低，具有破坏性的 GET 必须提高风险。根据实际语义设置 `force_approval`、`irreversible`、`blast_radius`、`sensitive_params`，复用平台凭证与权限体系。
