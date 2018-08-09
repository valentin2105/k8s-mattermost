package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"

	"github.com/mattermost/mattermost-server/model"
	"github.com/pelletier/go-toml"
)

var client *model.Client4
var webSocketClient *model.WebSocketClient

var botUser *model.User
var botTeam *model.Team
var debuggingChannel *model.Channel

//Splash is a splash
const Splash = `

┬┌─┬ ┬┌┐ ┌─┐┬─┐┌┐┌┌─┐┌┬┐┌─┐┌─┐
├┴┐│ │├┴┐├┤ ├┬┘│││├┤  │ ├┤ └─┐
┴ ┴└─┘└─┘└─┘┴└─┘└┘└─┘ ┴ └─┘└─┘
┌┬┐┌─┐┌┬┐┌┬┐┌─┐┬─┐┌┬┐┌─┐┌─┐┌┬┐
│││├─┤ │  │ ├┤ ├┬┘││││ │└─┐ │ 
┴ ┴┴ ┴ ┴  ┴ └─┘┴└─┴ ┴└─┘└─┘ ┴ 

`

// Config is a struc of config
type Config struct {
	host         string
	kubectlPath  string
	botName      string
	channelName  string
	teamName     string
	userLogin    string
	userPassword string
}

//LoadConfig load the toml file into the lib
func LoadConfig(ConfigPath string) *toml.Tree {
	config, err := toml.LoadFile(ConfigPath)
	if err != nil {
		fmt.Printf("%s %s\n", ConfigPath, "is unreadable (check the file path or lint it on https://www.tomllint.com)")
		os.Exit(1)
	}
	return config
}

//ParseConfig set the Config object with the actual value of the toml
func ParseConfig(config *toml.Tree) Config {
	keysArray := config.Keys()
	if StringInSlice("general", keysArray) == false {
		fmt.Printf("%s %s\n", "The config file don't get any [general] section in", *configPath)
		os.Exit(1)
	}
	conf := Config{
		botName:      config.Get("general.bot_name").(string),
		kubectlPath:  config.Get("general.kubectl_path").(string),
		host:         config.Get("mattermost.host").(string),
		channelName:  config.Get("mattermost.channel_name").(string),
		teamName:     config.Get("mattermost.team_name").(string),
		userLogin:    config.Get("mattermost.user_login").(string),
		userPassword: config.Get("mattermost.user_password").(string),
	}
	return conf
}

// MakeSureServerIsRunning ensure the server is running
func MakeSureServerIsRunning() {
	if props, resp := client.GetOldClientConfig(""); resp.Error != nil {
		println("There was a problem pinging the Mattermost server.  Are you sure it's running?")
		PrintError(resp.Error)
		os.Exit(1)
	} else {
		println("Server detected and is running version " + props["Version"] + "\n")
	}
}

// LoginAsTheBotUser login as the bot
func LoginAsTheBotUser(email string, password string) {
	if user, resp := client.Login(email, password); resp.Error != nil {
		println("There was a problem logging into the Mattermost server.  Are you sure ran the setup steps from the README.md?")
		PrintError(resp.Error)
		os.Exit(1)
	} else {
		botUser = user
	}
}

// FindBotTeam is use to find the team
func FindBotTeam(teamName string) {
	if team, resp := client.GetTeamByName(teamName, ""); resp.Error != nil {
		println("We failed to get the initial load")
		println("or we do not appear to be a member of the team '" + teamName + "'")
		PrintError(resp.Error)
		os.Exit(1)
	} else {
		botTeam = team
	}
}

// CreateBotDebuggingChannelIfNeeded create the channel
func CreateBotDebuggingChannelIfNeeded(channelName string) {
	if rchannel, resp := client.GetChannelByName(channelName, botTeam.Id, ""); resp.Error != nil {
		println("We failed to get the channels")
		PrintError(resp.Error)
	} else {
		debuggingChannel = rchannel
		return
	}

	// Looks like we need to create the logging channel
	channel := &model.Channel{}
	channel.Name = channelName
	channel.DisplayName = channelName
	channel.Purpose = "This is used as a channel for Kubernetes interactions"
	channel.Type = model.CHANNEL_OPEN
	channel.TeamId = botTeam.Id
	if rchannel, resp := client.CreateChannel(channel); resp.Error != nil {
		println("We failed to create the channel " + channelName)
		PrintError(resp.Error)
	} else {
		debuggingChannel = rchannel
		println("Looks like this might be the first run so we've created the channel " + channelName)
	}
}

// SendMsgToDebuggingChannel send msg if its ok
func SendMsgToDebuggingChannel(msg string, replyToID string) {
	post := &model.Post{}
	post.ChannelId = debuggingChannel.Id
	post.Message = msg

	post.RootId = replyToID

	if _, resp := client.CreatePost(post); resp.Error != nil {
		println("We failed to send a message to the channel")
		PrintError(resp.Error)
	}
}

// HandleWebSocketResponse handle the socket
func HandleWebSocketResponse(event *model.WebSocketEvent) {
	HandleMsgFromDebuggingChannel(event)
}

// HandleMsgFromDebuggingChannel handle the msg
func HandleMsgFromDebuggingChannel(event *model.WebSocketEvent) {
	// If this isn't the debugging channel then lets ingore it
	if event.Broadcast.ChannelId != debuggingChannel.Id {
		return
	}

	// Lets only reponded to messaged posted events
	if event.Event != model.WEBSOCKET_EVENT_POSTED {
		return
	}

	post := model.PostFromJson(strings.NewReader(event.Data["post"].(string)))
	if post != nil {

		// ignore my events
		if post.UserId == botUser.Id {
			return
		}

		if matched, _ := regexp.MatchString(KubeWord, post.Message); matched {
			words := strings.Fields(post.Message)
			cmd := CheckBeforeExec(words, post.Message)
			if len(cmd) > 0 && cmd != "command forbidden" {
				fmt.Printf("responding to -> %s", post.Message)
				cmdOut := ExecKubectl(cmd)
				if cmdOut != "" && len(cmdOut) > 0 {
					SendMsgToDebuggingChannel(cmdOut, post.Id)
					fmt.Printf(" <- Sent. \n")
					return
				}

			}
			if cmd == "command forbidden" {
				SendMsgToDebuggingChannel(cmd, post.Id)
				return

			}
		}

		// if you see any word matching 'alive' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)alive(?:$|\W)`, post.Message); matched {
			SendMsgToDebuggingChannel("Yes I'm running", post.Id)
			return
		}

		if matched, _ := regexp.MatchString(`(?:^|\W)help(?:$|\W)`, post.Message); matched {
			SendMsgToDebuggingChannel("!k [namespace] verb [ressource]", post.Id)
			return
		}

		// if you see any word matching 'up' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)up(?:$|\W)`, post.Message); matched {
			SendMsgToDebuggingChannel("Yes I'm running", post.Id)
			return
		}

		// if you see any word matching 'running' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)running(?:$|\W)`, post.Message); matched {
			SendMsgToDebuggingChannel("Yes I'm running", post.Id)
			return
		}

		// if you see any word matching 'hello' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)Hello(?:$|\W)`, post.Message); matched {
			SendMsgToDebuggingChannel("Hello my friend !", post.Id)
			return
		}
	}

	//SendMsgToDebuggingChannel("I did not understand you!", post.Id)
}

// PrintError print the connexions error
func PrintError(err *model.AppError) {
	println("\tError Details:")
	println("\t\t" + err.Message)
	println("\t\t" + err.Id)
	println("\t\t" + err.DetailedError)
}

// SetupGracefulShutdown preapre the graceful shut
func SetupGracefulShutdown(botName string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			if webSocketClient != nil {
				webSocketClient.Close()
			}

			SendMsgToDebuggingChannel("_"+botName+" has **stopped** running_", "")
			os.Exit(0)
		}
	}()
}

//StringInSlice check if a string is in a []string
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// CheckBeforeExec - Check stuffs before exec.
func CheckBeforeExec(words []string, lastmsg string) string {
	var cmd string
	if words[0] == KubeWord && len(words) >= 3 {

		confToml := LoadConfig(*configPath)
		conf := ParseConfig(confToml)
		kubectlAndNs := fmt.Sprintf(conf.kubectlPath + " -n")
		cmd = strings.Replace(lastmsg, KubeWord, kubectlAndNs, -1)

		// If it contain "all" namespace
		if words[1] == "all" {
			cmd = cmd + " --all-namespaces"
		}

		if !StringInSlice(words[2], ValidVerbs) {
			fmt.Printf("error ->  command unavailable <- %+v \n", lastmsg)
			cmd = "command forbidden"
		}
		// Match TRUSTED words (get, scale ...)
		if words[2] == "logs" && StringInSlice("-f", words) {
			fmt.Printf("error ->  command unavailable <- %+v \n", lastmsg)
			cmd = "command forbidden"
		}
		if words[2] == "exec" && StringInSlice("-it", words) {
			fmt.Printf("error ->  command unavailable <- %+v \n", lastmsg)
			cmd = "command forbidden"
		}
	}
	return cmd
}

// ExecKubectl - Launch and format kubectl cmd.
func ExecKubectl(cmd string) string {
	var cl string
	args := strings.Split(cmd, " ")
	out, err := exec.Command(args[0], args[1:]...).Output()
	if err == nil {
		result := fmt.Sprintf("``` \n %s ```", out)
		cl = strings.Replace(result, "\n\n", "\n", -1)
	}
	return cl
}
