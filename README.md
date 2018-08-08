# k8s-mattermost

### Infos
**k8s-mattermost** is a bot in Golang that connect on a channel's websocket and watch for commands. 

By default, you can trigger the bot with `!k <namespace> <verb> <ressource> -- !k default get pod `

The configuration is present in the `config.toml.dist` file (rename to `config.toml`) : 

```
[general]
bot_name = "k8s-bot"
kubectl_path = "/usr/local/bin/kubectl"

[mattermost]
host = "mattermost.org"
channel_name = "kubernetes"
team_name = "your-team"
user_login = "bot@email.org"
user_password = "averystr0ngpassw0rd"
```

You can load a different config file using the `-c ` flag. 

### Screenshot
![](https://i.imgur.com/6eFvItT.png)