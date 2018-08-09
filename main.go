package main

import (
	"flag"
	"fmt"

	"github.com/mattermost/mattermost-server/model"
)

const (
	// KubeWord - The Word that trigger the bot.
	KubeWord = "!k"
	// Version is the app version
	Version = "0.1.0"
)

var (
	// ValidVerbs is the list of the valids verbs of Kubernetes
	ValidVerbs = []string{"get", "scale", "exec", "describe", "label", "annotate", "version", "logs", "rollout", "config"}
	configPath = flag.String("config", "config.toml", "Config file path")
)

// doc at https://godoc.org/github.com/mattermost/platform/model#Client
func main() {
	flag.Parse()
	fmt.Print(Splash)
	confToml := LoadConfig(*configPath)
	conf := ParseConfig(confToml)
	SetupGracefulShutdown(conf.botName)
	url := fmt.Sprintf("https://%s", conf.host)
	client = model.NewAPIv4Client(url)
	// Lets test to see if the mattermost server is up and running
	MakeSureServerIsRunning()
	// lets attempt to login to the Mattermost server as the bot user
	// This will set the token required for all future calls
	// You can get this token with client.AuthToken
	LoginAsTheBotUser(conf.userLogin, conf.userPassword)
	// If the bot user doesn't have the correct information lets update his profile
	//UpdateTheBotUserIfNeeded()
	// Lets find our bot team
	FindBotTeam(conf.teamName)
	// Lets create a bot channel for logging debug messages into
	CreateBotDebuggingChannelIfNeeded(conf.channelName)
	SendMsgToDebuggingChannel("_"+conf.botName+" has **started** running_", "")
	// Lets start listening to some channels via the websocket!
	wssURL := fmt.Sprintf("wss://%s", conf.host)
	webSocketClient, err := model.NewWebSocketClient4(wssURL, client.AuthToken)
	if err != nil {
		println("We failed to connect to the web socket")
		PrintError(err)
	}
	webSocketClient.Listen()
	go func() {
		for {
			select {
			case resp := <-webSocketClient.EventChannel:
				HandleWebSocketResponse(resp)
			}
		}
	}()
	// You can block forever with
	select {}
}
