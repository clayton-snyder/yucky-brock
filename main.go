package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"math/rand"
    "math"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gordonklaus/portaudio"
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

type stereoSine struct {
    *portaudio.Stream
    stepL, phaseL float64
    stepR, phaseR float64
}

func newStereoSine(freqL, freqR, sampleRate float64) *stereoSine {
    s := &stereoSine{nil, freqL / sampleRate, 0, freqR / sampleRate, 0}
    s.Stream, _ = portaudio.OpenDefaultStream(0, 2, 44100, 0, s.processAudio)
    return s
}

func (g *stereoSine) processAudio(out [][]float32) {
    for i := range out[0] {
        out[0][i] = float32(math.Sin(2 * math.Pi * g.phaseL))
        _, g.phaseL = math.Modf(g.phaseL + g.stepL)
        out[1][i] = float32(math.Sin(2 * math.Pi * g.phaseR))
        _, g.phaseR = math.Modf(g.phaseR + g.stepR)
    }
}

func main() {
	// Create Discord session with token from flag
	dg, err := discordgo.New("Bot " + DiscordToken)

	if err != nil {
		fmt.Printf("Error creating Discord session: '%v'\n", err)
		return
	}

	dg.AddHandler(handleMessage)

	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	err = dg.Open()
	if err != nil {
		fmt.Printf("Error with dg.Open(): '%v'\n", err)
		return
	}

	fmt.Printf("*~*~*~*~* Bot runnin'.\n\n")

    // we're gonna blast a welcome message to every channel we can
    for _, guild := range dg.State.Guilds {
        channels, _ := dg.GuildChannels(guild.ID)
        for _, channel := range channels {
            if channel.Type == discordgo.ChannelTypeGuildText {
                dg.ChannelMessageSend(channel.ID, GreetingMsg)
            }
        }
    }

	portaudio.Initialize()
	defer portaudio.Terminate()
	h, err := portaudio.DefaultHostApi()
	if err != nil {
	    fmt.Printf("Error with portaudio.DefaultHostApi(): '%v'\n", err)
	    return
	}

    s := newStereoSine(256, 320, 44100)
    defer s.Close()
    s.Start()
    time.Sleep (5 * time.Second)
    s.Stop()

	stream, err := portaudio.OpenStream(portaudio.HighLatencyParameters(nil, h.DefaultOutputDevice), func(out []int32) {
	    for i := range out {
	        out[i] = int32(rand.Uint32() / 10)
	    }
	})
	if err != nil {
	    fmt.Printf("Error with portaudio.OpenStream(): '%v'\n", err)
	    return
	}

	defer stream.Close()
	stream.Start()
	time.Sleep(time.Second)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sig

	stream.Stop()
	dg.Close()
}

func playSound(session *discordgo.Session, guildID, channelID string) {
	fmt.Printf("*~*~*~*~*  Entered playSound guildID=%v, channelID=%v.\n", guildID, channelID)
	vc, err := session.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		panic("@%@%@%@%@ It all went wrong joining the voice channel.\n")
	}

	time.Sleep(250 * time.Millisecond)
	fmt.Printf("Joined! Now trying to speak...\n")

	vc.Speaking(true)

	in := make([]byte, 64)
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
	defer stream.Close()
	stream.Start()
	sign := make(chan os.Signal, 1)
        signal.Notify(sign, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		stream.Read()
		vc.OpusSend <- in
		select {
			case <-sign:
				return
			default:
		}
	}

	vc.Speaking(false)

	time.Sleep(250 * time.Millisecond)

	vc.Disconnect()
}

func handleMessage(session *discordgo.Session, msg *discordgo.MessageCreate) {
	fmt.Printf("*~*~*~*~*  Entered handleMessage.\n")
	// Don't talk to yourself
	if msg.Author.ID == session.State.User.ID {
		return
	}

	fmt.Printf("msg.ChannelID: %v\n", msg.ChannelID)
	fmt.Printf("msg.GuildID: %v\n", msg.GuildID)
	for _, guild := range session.State.Guilds {
		fmt.Printf("%v, ", guild.ID)
	}
	fmt.Println("end of guilds")

	if msg.Content == "playsound" {
		msgChannel, _ := session.State.Channel(msg.ChannelID)
		msgGuild, _ := session.State.Guild(msgChannel.GuildID)

		for _, voiceState := range msgGuild.VoiceStates {
			if voiceState.UserID == msg.Author.ID {
				playSound(session, msgGuild.ID, voiceState.ChannelID)
			}
		}
	}

	if msg.Content == "bing" {
		fmt.Printf("%v: %v\n", msg.Author.Username, msg.Content)
		session.ChannelMessageSend(msg.ChannelID, "bong")
	}
}
