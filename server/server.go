/**
 *	server.go
 */

package main

import (
	"encoding/json"
	"fmt"
	"mybot/twitchbot"
	"net/http"
	"os"
	"time"
)

func loadFromJSON(filename string, key interface{}) error {
	in, err := os.Open(filename)
	if err != nil {
		return err
	}
	decodeJSON := json.NewDecoder(in)
	err = decodeJSON.Decode(key)
	if err != nil {
		in.Close()
		return err
	}
	in.Close()
	return nil
}

func saveToJSON(filename *os.File, key interface{}) {
	encodeJSON := json.NewEncoder(filename)
	err := encodeJSON.Encode(key)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func createConfig(filename string) {
	// create file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Print(err)
		return
	}

	// fill date to the config file
	myBot := twitchbot.BasicBot{
		ChannelName: "alex_mikas",
		MsgRate:     time.Duration(20/30) * time.Millisecond,
		BotName:     "keks1kbot",
		Port:        "6667",
		PrivatePath: "./configs/oauth.json",
		Server:      "irc.chat.twitch.tv", // irc.twitch.tv
	}
	saveToJSON(file, myBot)
}

func main() {
	// Проверка конфигурационных файлов
	filename := "./configs/config.json"

	if _, err := os.Stat(filename); err == nil {
		// path/to/config.json exists
		fmt.Printf("file %s exists\n", filename)
	} else if os.IsNotExist(err) {
		// path/to/config.json does *not* exist
		fmt.Print(err)
		createConfig(filename)
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		fmt.Print(err)
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
	}

	var twitchBot twitchbot.BasicBot
	err := loadFromJSON(filename, &twitchBot)
	if err == nil {
		fmt.Println(twitchBot)
	} else {
		fmt.Println(err)
	}

	go twitchBot.Start()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	})
	http.ListenAndServe(":80", nil)
}
