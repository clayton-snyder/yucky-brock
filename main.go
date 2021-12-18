package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	// "math/rand"
	"syscall"
	// "time"

	"github.com/bwmarrin/discordgo"
	// "github.com/gordonklaus/portaudio"
)

var (
	DiscordToken string
	GreetingMsg  string
)

func init() {
	flag.StringVar(&DiscordToken, "t", "", "")
	flag.StringVar(&GreetingMsg, "m", "I believe in rock-hard defense and determination!", "")
	flag.Parse()

	fmt.Printf("Flags: DiscordToken=%v, GreetingMsg=%v\n", DiscordToken, GreetingMsg)
}

func main() {
	// Create Discord session with token from flag
	dg, err := discordgo.New("Bot " + DiscordToken)

	if err != nil {
		fmt.Printf("Error creating Discord session: '%v'\n", err)
		return
	}

	dg.AddHandler(handleMessage)

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Printf("Error with dg.Open(): '%v'\n", err)
		return
	}

	fmt.Printf("Bot runnin'.")

    // we're gonna blast a welcome message to every channel we can
    for _, guild := range dg.State.Guilds {
        channels, _ := dg.GuildChannels(guild.ID)
        for _, channel := range channels {
            if channel.Type == discordgo.ChannelTypeGuildText {
                dg.ChannelMessageSend(channel.ID, GreetingMsg)
            }
        }
    }

	// portaudio.Initialize()
	// defer portaudio.Terminate()
	// h, err := portaudio.DefaultHostApi()
	// if err != nil {
	//     fmt.Printf("Error with portaudio.DefaultHostApi(): '%v'\n", err)
	//     return
	// }

	// stream, err := portaudio.OpenStream(portaudio.HighLatencyParameters(nil, h.DefaultOutputDevice), func(out []int32) {
	//     for i := range out {
	//         out[i] = int32(rand.Uint32())
	//     }
	// })
	// if err != nil {
	//     fmt.Printf("Error with portaudio.OpenStream(): '%v'\n", err)
	//     return
	// }

	// defer stream.Close()
	// stream.Start()
	// time.Sleep(time.Second)
	// stream.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sig

	dg.Close()
}

func handleMessage(session *discordgo.Session, msg *discordgo.MessageCreate) {
	// Don't talk to yourself
	if msg.Author.ID == session.State.User.ID {
		return
	}

	if msg.Content == "bing" {
		fmt.Printf("%v: %v", msg.Author.Username, msg.Content)
		session.ChannelMessageSend(msg.ChannelID, "bong")
	}
}
