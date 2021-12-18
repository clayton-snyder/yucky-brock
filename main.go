package main

import (
    "flag"
    "os/signal"
    "os"
    "fmt"
    "math/rand"
    "syscall"
    "time"

    "github.com/bwmarrin/discordgo"
//    "github.com/zmb3/spotify/v2"
    "github.com/gordonklaus/portaudio"
)

var (
    DiscordToken        string
    SpotifyClientID     string
    SpotifyClientSecret string
    GreetingMsg         string
    DEBUG_AudioFilePath string
)

func init() {
    flag.StringVar(&DiscordToken, "discord-token", "OTE5NDU1ODQyNTkwNDI1MTA4.YbWD-w.cnrThGTgu3E7V0Bztjs0DnUag8E", "")
    flag.StringVar(&SpotifyClientID, "spotify-client-id", "", "")
    flag.StringVar(&SpotifyClientSecret, "spotify-client-secret", "", "")

    flag.StringVar(&DEBUG_AudioFilePath, "audio-file", "", "")
    flag.Parse()

    fmt.Printf("Flags: DiscordToken=%v, SpotifyClientID=%v, SpotifyClientSecret=%v\n",
        DiscordToken, SpotifyClientID, SpotifyClientSecret)
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


    //f, err := os.Open(DEBUG_AudioFilePath)
    //if err != null {
    //    fmt.Printf("Error reading audio file: '%v' (DEBUG_AudioFilePath=%v)\n",
    //        err, DEBUG_AudioFilePath)
    //}

    portaudio.Initialize()
    defer portaudio.Terminate()
    h, err := portaudio.DefaultHostApi()
    if err != nil {
        fmt.Printf("Error with portaudio.DefaultHostApi(): '%v'\n", err)
        return
    }

    stream, err := portaudio.OpenStream(portaudio.HighLatencyParameters(nil, h.DefaultOutputDevice), func(out []int32) {
        for i := range out {
            out[i] = int32(rand.Uint32())
        }
    })
    if err != nil {
        fmt.Printf("Error with portaudio.OpenStream(): '%v'\n", err)
        return
    }

    defer stream.Close()
    stream.Start()
    time.Sleep(time.Second)
    stream.Stop()

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
