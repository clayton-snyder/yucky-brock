package main

import (
	//	"bytes"
	//	"encoding/binary"
	"context"
	"flag"
	"fmt"
	"strings"

	"net/http"
	"os"
	"os/signal"

	//        "math"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gordonklaus/portaudio"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"gopkg.in/hraban/opus.v2"
)

const redirectURI = "http://localhost:8080/brocktch"

var (
	DiscordToken        string
	GreetingMsg         string
	auth                *spotifyauth.Authenticator
	client              *spotify.Client
	SpotifyClientID     = "e1fae77fde6b40e68f7a27cba74282d5"
	SpotifyClientSecret = "2c3261eafaa24a9a9ebe5d34c9fa3b81"
	spotifyCh           = make(chan *spotify.Client)
	state               = "2323cs"
)

func init() {
	flag.StringVar(&DiscordToken, "t", "", "")
	flag.StringVar(&GreetingMsg, "m", "I believe in rock-hard defense and determination!", "")
	flag.Parse()
	Tonst()

	fmt.Printf("Flags: DiscordToken=%v, GreetingMsg=%v\n", DiscordToken, GreetingMsg)

	os.Setenv("SPOTIFY_ID", SpotifyClientID)
	os.Setenv("SPOTIFY_SECRET", SpotifyClientSecret)

	auth = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopeUserModifyPlaybackState))

	// This will receive the Spotify callback with the token
	http.HandleFunc("/brocktch", doAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Not handling request for: %v", r.URL.String())
	})

	// we have to listen on a separate thread so rest of program can run 8^)
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Printf("Error listening on 8080: %v", err)
		}
	}()
	fmt.Println("Listening on 8080.")

	url := auth.AuthURL(state)
	fmt.Printf("Log in to a Premium account here: %v \n", url)
	client = <-spotifyCh
}

func main() {
	user, spotifyAuthErr := client.CurrentUser(context.Background())
	// queueErr := client.QueueSong(context.Background(), "1goNp8FZSjak6UHYsawniU")
	// playErr := client.Play(context.Background())

	if spotifyAuthErr != nil {
		fmt.Printf("Error logging into Spotify: %v\n", spotifyAuthErr)
		return
	} else if user == nil {
		fmt.Printf("No spotifyAuthErr, but 'user' is nil.\n")
		return
	}

	fmt.Printf("Logged in as %v (%v)\n", user.ID, user.User.DisplayName)

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

	// h, err := portaudio.DefaultHostApi()
	// if err != nil {
	// 	fmt.Printf("Error with portaudio.DefaultHostApi(): '%v'\n", err)
	// 	return
	// }

	// time.Sleep(time.Second)

	// stream, err := portaudio.OpenStream(portaudio.HighLatencyParameters(nil, h.DefaultOutputDevice), func(out []int32) {
	// 	for i := range out {
	// 		out[i] = int32(rand.Uint32() / 10)
	// 	}
	// })
	// if err != nil {
	// 	fmt.Printf("Error with portaudio.OpenStream(): '%v'\n", err)
	// 	return
	// }

	// defer stream.Close()
	// stream.Start()

	time.Sleep(time.Second)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sig

	//	s.Stop()
	//	stream.Stop()
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

	// Input buffer for raw PCM audio data. Note that length*1000/sampleRate must be 2.5, 5, 10, 20, 40, or 60 (it's ms that the frame represents)
	// see hraban/opus.v2 "encoding" for more info
	in := make([]int16, 960)
	/*
		testBytes := make([]byte, 64)
		for x := range testBytes {
			testBytes[x] = byte(50)
		}
	*/

	// Have to encode raw audio with this before sending to Discord
	enc, err := opus.NewEncoder(48000, 1, opus.AppVoIP)
	if err != nil {
		fmt.Printf("Error creating Opus encoder: %v\n", err)
	}

	// This is the default output of our machine, so spotifyd should stream through here
	fmt.Println("*~~~~~~~~~~~~~~~~~~~*~~~~~~~~~~~~~~~~~~~~~~~~~* OPENING THE DAMN DEFAULT STREAM GUY")
	stream, err := portaudio.OpenDefaultStream(1, 0, 48000, len(in), in)

	defer stream.Close()
	stream.Start()

	// Remove these, unnecessary?
	sign := make(chan os.Signal, 1)
	signal.Notify(sign, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		stream.Read()
		data := make([]byte, 1000)
		n, encerr := enc.Encode(in, data)
		if encerr != nil {
			fmt.Printf("Error with enc.Encode(): %v", encerr)
			return
		}
		data = data[:n] // Only the first 'n' bytes are opus data.
		/*		inBytesBuf := new(bytes.Buffer)
				err := binary.Write(inBytesBuf, binary.LittleEndian, in)
				if err != nil {
					fmt.Printf("binary.Write failed: %v", err)
				}
		*/
		vc.OpusSend <- data
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

	cleanMsg := strings.ToLower(strings.Trim(msg.Content, " "))

	if cleanMsg == "playsound" {
		msgChannel, _ := session.State.Channel(msg.ChannelID)
		msgGuild, _ := session.State.Guild(msgChannel.GuildID)

		for _, voiceState := range msgGuild.VoiceStates {
			if voiceState.UserID == msg.Author.ID {
				playSound(session, msgGuild.ID, voiceState.ChannelID)
			}
		}
	}

	if cleanMsg == "bing" {
		fmt.Printf("%v: %v\n", msg.Author.Username, msg.Content)
		session.ChannelMessageSend(msg.ChannelID, "bong")
	}

	if strings.HasPrefix(cleanMsg, "queue ") {
		tokens := strings.SplitN(msg.Content, " ", 2)
		if len(tokens) < 2 {
			session.ChannelMessageSend(msg.ChannelID, "Queue what?")
			return
		}

		track, searchErr := search(client, tokens[1])
		if searchErr != nil {
			session.ChannelMessageSend(msg.ChannelID, searchErr.Error())
			return
		}

		queueErr := queue(client, track.ID)
		if queueErr != nil {
			session.ChannelMessageSend(msg.ChannelID, queueErr.Error())
			return
		}

		session.ChannelMessageSend(
			msg.ChannelID,
			fmt.Sprintf("Added %v by %v to the queue.", track.Name, track.Artists))
	}

	if cleanMsg == "play" {
		client.Play(context.Background())
	}

	if cleanMsg == "stop" || cleanMsg == "pause" {
		client.Pause(context.Background())
	}

	if cleanMsg == "next" || cleanMsg == "skip" {
		client.Next(context.Background())
	}

}

func doAuth(w http.ResponseWriter, r *http.Request) {
	token, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		fmt.Printf("Error getting token: %v\n", err)
	}

	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		fmt.Printf("State mismatch: st=%v, state=%v\n", st, state)
	}

	client := spotify.New(auth.Client(r.Context(), token))
	fmt.Println("Login completed.")
	spotifyCh <- client
}
