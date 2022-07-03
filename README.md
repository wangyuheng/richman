# richman

基于飞书多维表格(bitable)实现的记账机器人。

## 飞书机器人

- 视频教程: https://www.bilibili.com/video/BV1AY411K7rn
- 文字教程: https://www.bilibili.com/read/cv17318657

### 使用方式

设置环境变量。

- LARK_APP_ID: 对应飞书开放平台 -> 开发者后台 -> 应用凭证 -> APP ID
- LARK_APP_SECRET: 对应飞书开放平台 -> 开发者后台 -> 应用凭证 -> App Secret
- LARK_APP_TOKEN: 新建多维表格后，通过浏览器url获取。
- LARK_APP_VERIFICATION_TOKEN: 对应飞书开放平台 -> 开发者后台 -> 事件订阅 -> Verification Token

比如

```shell
LARK_APP_ID=cli_a232fc4bceb8100b
LARK_APP_SECRET=AWkBwpc15kgsCOWf7Y7KQcCJyAdM1Clx
LARK_APP_TOKEN=bascnZkP4JxAWoFuO8R6LUJABme
LARK_APP_VERIFICATION_TOKEN=lqfwcLQ2msJvhDZD8Z5JibXB7fq8tTaD
```

如果是测试环境，可以直接在`env.go`文件中修改，生产环境建议通过系统环境变量进行设置。

## 微信公众号

Coming Soon