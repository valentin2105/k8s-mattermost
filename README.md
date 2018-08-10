# k8s-mattermost
[![Go Report Card](https://goreportcard.com/badge/github.com/valentin2105/k8s-mattermost)](https://goreportcard.com/report/github.com/valentin2105/k8s-mattermost)
[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/dwyl/esta/issues)

### What is it ?
**k8s-mattermost** is a botthat connects to a Mattermost channel's websocket and watches for kubectl commands. 

By default, you can trigger the bot with `!k <namespace> <verb> <ressource>` :

```
!k - get cs  # You can use "-" if ressource doesn't get namespace

!k kube-system get deploy

!k all get pod # You can use "all" to show all namespaces
```

The configuration is located in the `config.toml.dist` file (you should rename it to `config.toml`) : 

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


### How can you run it ?

You can fetch the latest build for Linux with :
```
wget https://github.com/valentin2105/k8s-mattermost/releases/download/v0.1.0/k8s-mattermost
chmod +x k8s-mattermost 
./k8s-mattermost -c config.toml
```

Or build it from source : 

```
# Install dep
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# Clone repo
mkdir -p $GOPATH/src/github.com/valentin2105/ && cd $GOPATH/src/github.com/valentin2105/
git clone git@github.com:valentin2105/k8s-mattermost.git && cd k8s-mattermost 

# Install dependencies
dep ensure

# Build
go build
```

### Screenshot
![](https://i.imgur.com/6eFvItT.png)
