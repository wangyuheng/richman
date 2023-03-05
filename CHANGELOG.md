# 2023-02-11

Feature

- 提供生成账本功能 `Make Command` 
- 增加获取源码Cmd

Refactor

- 使用`wire`优化依赖逻辑
- 使用飞书官方SDK(github.com/larksuite/oapi-sdk-go)优化OpenAPI调用

# 2022-07-31

- 增加Dream管理
- 允许自定义记账人字段名
- 兼容单选分类格式

# 2022-07-23

- 支持微信公众号
  - 绑定并获取事件回调URL
  - 记账

# 2022-07-10

- 支持绑定多个app应用
  - 提供Register接口进行注册
  - Webhook增加appId标识
- 增加`账单`指令, 查看当前账本

# 2022-07-09

- 支持绑定自定义多维表格账本
- 引入[feishu-bitable-db](https://github.com/geeklubcn/feishu-bitable-db)处理repo操作

# 2022-07-03

- 飞书记账机器人