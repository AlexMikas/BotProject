package twitchbot

/*
TODO:
1. Аутентификация бота.
2. Подключение к чату.
3. Слушать чат.
4. Отключиться от чата.
5. Выход.

Хэширование пароля?
*/

import (
	// rgb "github.com/foresthoffman/rgblog"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	rgb "mybot/rgblog"
	"net"
	"net/http"
	"net/textproto"
	"regexp"
	"strings"
	"time"
)

const PSTFormat = "Jan 2 15:04:05 PST"

type Command int

const (
	CLEARCHAT Command = iota + 1
	CLEARMSG
	GLOBALUSERSTATE
	PRIVMSG
	ROOMSTATE
	USERNOTICE
	USERSTATE
)

var CommandNames = [...]string{
	"CLEARCHAT",
	"CLEARMSG",
	"GLOBALUSERSTATE",
	"PRIVMSG",
	"ROOMSTATE",
	"USERNOTICE",
	"USERSTATE"}

// Regex for parsing PRIVMSG strings.
//
// First matched group is the user's name and the second matched group is the content of the
// user's message.
// > @badge-info=<badge-info>;badges=<badges>;color=<color>;display-name=<display-name>;emotes=<emotes>;id=<id-of-msg>;mod=<mod>;room-id=<room-id>;subscriber=<subscriber>;tmi-sent-ts=<timestamp>;turbo=<turbo>;user-id=<user-id>;user-type=<user-type> :<user>!<user>@<user>.tmi.twitch.tv PRIVMSG #<channel> :<message>
var MsgRegex *regexp.Regexp = regexp.MustCompile(`^:(\w+)!\w+@\w+\.tmi\.twitch\.tv (PRIVMSG) #\w+(?: :(.*))?$`)

// Universal expression
// @tag-name=<tag-name> :tmi.twitch.tv <command> #<channel> :<user>
// @tag-name-1=<tag-value-1>;tag-name-2=<tag-value-2>;... <rest of the command syntax depends on the particular IRC command used>
var inputRegex *regexp.Regexp = regexp.MustCompile(`^:(\w+)!\w+@\w+\.tmi\.twitch\.tv (PRIVMSG) #\w+(?: :(.*))?$`)

// Regex for parsing user commands, from already parsed PRIVMSG strings.
//
// First matched group is the command name and the second matched group is the argument for the
// command.
var CmdRegex *regexp.Regexp = regexp.MustCompile(`^!(\w+)\s?(\w+)?`)

type OAuthCred struct {

	// The bot account's OAuth password.
	Password string `json:"password,omitempty"`

	// The developer application client ID. Used for API calls to Twitch.
	ClientID string `json:"ClientID,omitempty"`
}

type Bot interface {

	// Opens a connection to the Twitch.tv IRC chat server.
	Connect()

	// Closes a connection to the Twitch.tv IRC chat server.
	Disconnect()

	// Listens to chat messages and PING request from the IRC server.
	HandleChat() error

	// Joins a specific chat channel.
	JoinChannel()

	// Parses credentials needed for authentication.
	ReadCredentials() error

	// Sends a message to the connected channel.
	Say(msg string) error

	// Attempts to keep the bot connected and handling chat.
	Start()
}

type BasicBot struct {

	// The channel that the bot is supposed to join. Note: The name MUST be lowercase, regardless
	// of how the username is displayed on Twitch.tv.
	ChannelName string

	// A reference to the bot's connection to the server.
	conn net.Conn

	headers http.Header
	client  *http.Client
	// The credentials necessary for authentication.
	Credentials *OAuthCred

	// A forced delay between bot responses. This prevents the bot from breaking the message limit
	// rules. A 20/30 millisecond delay is enough for a non-modded bot. If you decrease the delay
	// make sure you're still within the limit!
	//
	// Message Rate Guidelines: https://dev.twitch.tv/docs/irc#irc-command-and-message-limits
	MsgRate time.Duration

	// The name that the bot will use in the chat that it's attempting to join.
	BotName string

	// The port of the IRC server.
	Port string

	// A path to a limited-access directory containing the bot's OAuth credentials.
	PrivatePath string

	// The domain of the IRC server.
	Server string

	// The time at which the bot achieved a connection to the server.
	startTime time.Time
}

// Connects the bot to the Twitch IRC server. The bot will continue to try to connect until it
// succeeds or is forcefully shutdown.
func (bb *BasicBot) Connect() {
	var err error
	rgb.YPrintf("[%s] Connecting to %s...\n", timeStamp(), bb.Server)

	// makes connection to Twitch IRC server
	bb.conn, err = net.Dial("tcp", bb.Server+":"+bb.Port)
	if nil != err {
		rgb.YPrintf("[%s] Cannot connect to %s, retrying.\n", timeStamp(), bb.Server)
		//TODO:: число попыток?
		bb.Connect()
		return
	}
	rgb.YPrintf("[%s] Connected to %s!\n", timeStamp(), bb.Server)
	bb.startTime = time.Now()
}

// Officially disconnects the bot from the Twitch IRC server.
func (bb *BasicBot) Disconnect() {
	bb.conn.Close()
	upTime := time.Now().Sub(bb.startTime).Seconds()
	rgb.YPrintf("[%s] Closed connection from %s! | Live for: %fs\n", timeStamp(), bb.Server, upTime)
}

// Listens for and logs messages from chat. Responds to commands from the channel owner. The bot
// continues until it gets disconnected, told to shutdown, or forcefully shutdown.
func (bb *BasicBot) HandleChat() error {
	rgb.YPrintf("[%s] Watching #%s...\n", timeStamp(), bb.ChannelName)

	// reads from connection
	tp := textproto.NewReader(bufio.NewReader(bb.conn))

	// listens for chat messages
	for {
		line, err := tp.ReadLine()
		if nil != err {

			// officially disconnects the bot from the server
			bb.Disconnect()

			return errors.New("bb.Bot.HandleChat: Failed to read line from channel. Disconnected.")
		}

		// logs the response from the IRC server
		rgb.YPrintf("[%s] %s\n", timeStamp(), line)

		if "PING :tmi.twitch.tv" == line {

			// respond to PING message with a PONG message, to maintain the connection
			bb.conn.Write([]byte("PONG :tmi.twitch.tv\r\n"))
			continue
		} else {

			// handle a PRIVMSG message
			matches := MsgRegex.FindStringSubmatch(line)
			if nil != matches {
				userName := matches[1]
				msgType := matches[2]

				bb.GetUserByName(userName)

				switch msgType {
				case "PRIVMSG":
					msg := matches[3]
					rgb.GPrintf("[%s] %s: %s\n", timeStamp(), userName, msg)

					// parse commands from user message
					cmdMatches := CmdRegex.FindStringSubmatch(msg)
					if nil != cmdMatches {
						cmd := cmdMatches[1]

						// channel-owner specific commands
						if userName == bb.ChannelName {
							switch cmd {
							case "tbdown":
								rgb.CPrintf(
									"[%s] Shutdown command received. Shutting down now...\n",
									timeStamp(),
								)

								bb.Disconnect()
								return nil
							default:
								// do nothing
							}
						}
					}
				default:
					// do nothing
				}
			}
		}
		time.Sleep(bb.MsgRate)
	}
}

// Makes the bot join its pre-specified channel.
func (bb *BasicBot) JoinChannel() {
	rgb.YPrintf("[%s] Joining #%s...\n", timeStamp(), bb.ChannelName)
	bb.conn.Write([]byte(fmt.Sprintf("USER %s 8 * :%s\r\n", bb.BotName, bb.BotName)))
	bb.conn.Write([]byte(fmt.Sprintf("PASS %s\r\n", bb.Credentials.Password)))
	bb.conn.Write([]byte(fmt.Sprintf("NICK %s\r\n", bb.BotName)))
	bb.conn.Write([]byte(fmt.Sprintf("JOIN #%s\r\n", strings.ToLower(bb.ChannelName))))

	rgb.YPrintf("[%s] Joined #%s as @%s!\n", timeStamp(), bb.ChannelName, bb.BotName)
}

// Reads from the private credentials file and stores the data in the bot's Credentials field.
func (bb *BasicBot) ReadCredentials() error {

	// reads from the file
	credFile, err := ioutil.ReadFile(bb.PrivatePath)
	if nil != err {
		return err
	}

	bb.Credentials = &OAuthCred{}

	// parses the file contents
	dec := json.NewDecoder(strings.NewReader(string(credFile)))
	if err = dec.Decode(bb.Credentials); nil != err && io.EOF != err {
		return err
	}

	return nil
}

func (bb *BasicBot) CreateHeaders(v5 bool) http.Header {
	bb.headers = make(http.Header)
	bb.headers.Set("Client-ID", bb.Credentials.ClientID)
	bb.headers.Set("Content-Type", "application/json")

	if v5 {
		bb.headers.Set("Accept", "application/vnd.twitchtv.v5+json")
	}

	return bb.headers
}

// Makes the bot send a message to the chat channel.
func (bb *BasicBot) Say(msg string) error {
	if "" == msg {
		return errors.New("BasicBot.Say: msg was empty.")
	}

	// check if message is too large for IRC
	if len(msg) > 512 {
		return errors.New("BasicBot.Say: msg exceeded 512 bytes")
	}

	_, err := bb.conn.Write([]byte(fmt.Sprintf("PRIVMSG #%s :%s\r\n", bb.ChannelName, msg)))
	if nil != err {
		return err
	}
	return nil
}

// Starts a loop where the bot will attempt to connect to the Twitch IRC server, then connect to the
// pre-specified channel, and then handle the chat. It will attempt to reconnect until it is told to
// shut down, or is forcefully shutdown.
func (bb *BasicBot) Start() {
	err := bb.ReadCredentials()
	if nil != err {
		fmt.Println(err)
		fmt.Println("Aborting...")
		return
	}
	bb.CreateHeaders(true)
	bb.client = &http.Client{}
	// Мониторинг чата go ...
	for {
		bb.Connect()
		bb.JoinChannel()
		err = bb.HandleChat()
		if nil != err {

			// attempts to reconnect upon unexpected chat error
			time.Sleep(1000 * time.Millisecond)
			fmt.Println(err)
			fmt.Println("Starting bot again...")
		} else {
			return
		}
	}
	// Мониторинг статистики go ...
}

func timeStamp() string {
	return TimeStamp(PSTFormat)
}

func TimeStamp(format string) string {
	return time.Now().Format(format)
}
