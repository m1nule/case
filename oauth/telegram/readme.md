# Telegram Oath Demo

[widget-configuration]<https://core.telegram.org/widgets/login#widget-configuration>

```go
// 机器人token
token       = os.Getenv("bot_token")
// 机器人名,不带@
botName     = os.Getenv("bot_name")
// 回调地址,BotFather中设置的doamin且必须要是https
callbackURL = os.Getenv("bot_callback_url")
```
