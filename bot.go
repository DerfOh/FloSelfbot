package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/bwmarrin/discordgo"
)

var conf *Config

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

type Config struct {
	Token  string
	Prefix string
}

func createConfig() *Config {
	var tempprefix string
	var temptoken string

	fmt.Println("No config file found, so let's make one!")
	fmt.Println("\nTo find your User Token. In browser or desktop Discord, type Ctrl-Shift-I. Go to the Console section, and type localStorage.token. Your user token will appear. Do not share this token with anyone! This token provides complete access to your Discord account, so never share it!")
	fmt.Println("\nInput your User Token here: ")
	fmt.Scanln(&temptoken)
	fmt.Println("Input your desired prefix here: ")
	fmt.Scanln(&tempprefix)

	buf := new(bytes.Buffer)
	tempconfig := &Config{temptoken, tempprefix}

	check(toml.NewEncoder(buf).Encode(tempconfig))

	f, err := os.Create("config.toml")
	check(err)
	defer f.Close()

	_, err = f.Write(buf.Bytes())

	return tempconfig
}

func main() {
	_, err := toml.DecodeFile("config.toml", &conf)
	if os.IsNotExist(err) {
		conf = createConfig()
	}

	dg, err := discordgo.New(conf.Token)

	dg.AddHandler(messageCreate)

	err = dg.Open()

	check(err)

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	fmt.Println("Your prefix is", conf.Prefix)

	<-make(chan struct{})
	return
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created others
	if m.Author.ID != s.State.User.ID {
		return
	}

	// If the message is "ping" reply with "Pong!"
	if strings.HasPrefix(m.Content, conf.Prefix) {
		// Setting values for the commands
		args := strings.Split(m.Content[len(conf.Prefix):len(m.Content)], " ")
		invoked := args[0]
		args = args[1:]
		// channel, _ := s.State.Channel(m.ChannelID)

		if invoked == "ping" {
			start := time.Now()
			msg, _ := s.ChannelMessageSend(m.ChannelID, "Pong!")
			elapsed := time.Since(start)
			s.ChannelMessageEdit(m.ChannelID, msg.ID, fmt.Sprintf("Pong! `%s`", elapsed))
		}
	}
}
