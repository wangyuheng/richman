# Richman

<p align="center">
    <img src=".design/logo.png">
</p>

基于飞书多维表格(bitable)实现的记账机器人。

![architecture](http://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/geeklubcn/richman/master/.design/architecture.puml)

## 飞书机器人

- 视频教程: https://www.bilibili.com/video/BV1AY411K7rn
- 文字教程: https://www.bilibili.com/read/cv17318657

![flow](http://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/geeklubcn/richman/master/.design/flow.puml)

## 微信公众号

搜索并关注 `莫比乌斯的code`

![flow](https://raw.githubusercontent.com/geeklubcn/richman/master/.design/qrcode.jpg)

### 获取事件回调地址

向公众号发送消息，获取回调地址URL
```json
{
    "id":"cli_a2513784b578530c",
    "secret":"DEieGDn72PusZ7lSGT7Gec3XOtgyRQhg",
    "token":"iejPgIsMMneCnyOBhbZ6hhCTs05YJDhb"
}
```

### 绑定账本

- 向飞书机器人发送`wx`获取`bindToken`
- 将`bindToken`发给公众号，完成绑定

### 开发

设置环境变量。

- LARK_APP_ID: 对应飞书开放平台 -> 开发者后台 -> 应用凭证 -> APP ID
- LARK_APP_SECRET: 对应飞书开放平台 -> 开发者后台 -> 应用凭证 -> App Secret
- SEVER_URL: 服务公网域名，用于生成回调地址
比如

```shell
LARK_APP_ID=cli_a232fc4bceb8100b
LARK_APP_SECRET=AWkBwpc15kgsCOWf7Y7KQcCJyAdM1Clx
SEVER_URL=https://richman.geeklub.cn
```

如果是测试环境，可以直接在`env.go`文件中修改，生产环境建议通过系统环境变量进行设置。

### 开发测试

1. 在`env.go` 设置环境变量
2. 发送POST请求，信息参考 ./dev/feishu.http

```shell
curl --location --request POST 'localhost:8080/feishu/webhook' \
--header 'Content-Type: application/json' \
--data-raw '
{
    "schema": "2.0",
    "header": {
        "event_id": "5e3702a84e847582be8db7fb73283c02",
        "event_type": "im.message.receive_v1",
        "create_time": "1608725989000",
        "token": "{{verification_token}}",
        "app_id": "{{app_id}}",
        "tenant_key": "2ca1d211f64f6424"
    },
    "event": {
        "sender": {
            "sender_id": {
                "union_id": "on_8ed6aa67826108097d9ee143816345",
                "user_id": "e33ggbyz",
                "open_id": "ou_84aad35d084aa403a838cf73ee18467"
            },
            "sender_type": "user",
            "tenant_key": "736588c9260f175e"
        },
        "message": {
            "message_id": "om_c7f35970552ecb3a0dab7dc698796121",
            "root_id": "om_5ce6d572455d361153b7cb5xxfsdfsdfdaf",
            "parent_id": "om_5ce6d572455d361153b7cb5xxfsdfsdddsf",
            "create_time": "1609073151345",
            "chat_id": "oc_5ce6d572455d361153b7xx51da1c3945",
            "chat_type": "p2p",
            "message_type": "text",
            "content": "{\"text\":\"辣条 20\"}"
        }
    }
}'
```

## Docker

使用
```shell
docker run -e "LARK_APP_ID=cli_a232fc4bceb8100b" \
-e "LARK_APP_SECRET=AWkBwpc15kgsCOWf7Y7KQcCJyAdM1Clx" \
-e "LARK_APP_TOKEN=bascnZkP4JxAWoFuO8R6LUJABme" \
geeklubcn/richman:v1 
```

打包

```shell
docker build -t geeklubcn/richman:v1 .
```

PUSH

```shell
docker push geeklubcn/richman:v1
```