package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"encoding/json"
	"io/ioutil"
	"time"
)

type DiscordConfig struct {
	BotName string `json:"name"`
	Token string `json:"token"`
	RolePermission string `json:"role"`
}

type BotVoiceState struct {
	VoiceConnection * discordgo.VoiceConnection
	Status bool
}

var AlexaDiscordConfig DiscordConfig
var AlexaVoiceState BotVoiceState

func DiscordBot( configFile string ) ( err error ){
	// read config file
	bytes , _ := ioutil.ReadFile( configFile )
	json.Unmarshal( bytes , &AlexaDiscordConfig )

	// Create a new Discord session using the provided bot token.
	session, err := discordgo.New("Bot " + AlexaDiscordConfig.Token )
	if err != nil {
		return err
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	session.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		return err
	}
	defer session.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	return err
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	// Find the guild for that channel.
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		// Could not find guild.
		return
	}

	// Join to Voice Channel
	if strings.HasPrefix( m.Content , "!join"){
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				member , _ := s.GuildMember( g.ID , vs.UserID )
				if hasMemberRole( s , g , member , AlexaDiscordConfig.RolePermission ) {
					AlexaVoiceState.VoiceConnection , _ = JoinVoiceChannel( s , g.ID , vs.ChannelID )
					AlexaVoiceState.Status = true
					go captureVoice()
				}
			}
		}
		return
	}

	// Leave the Voice Channel
	if strings.HasPrefix( m.Content , "!leave"){
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				member , _ := s.GuildMember( g.ID , vs.UserID )
				if hasMemberRole( s , g , member , AlexaDiscordConfig.RolePermission ) {
					if AlexaVoiceState.VoiceConnection != nil {
						AlexaVoiceState.Status = false
						AlexaVoiceState.VoiceConnection.Disconnect()
					}
				}

			}
		}
		return
	}


	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}

// Capture and Send Voice Example ( funktioniert noch nicht ... )
func captureVoice( ){
	for AlexaVoiceState.Status {
		for packet := range AlexaVoiceState.VoiceConnection.OpusRecv {

			time.Sleep(250 * time.Millisecond)
			// Start speaking.
			AlexaVoiceState.VoiceConnection.Speaking(true)

			AlexaVoiceState.VoiceConnection.OpusSend <- packet.Opus

			// Stop speaking
			AlexaVoiceState.VoiceConnection.Speaking(false)
			time.Sleep(250 * time.Millisecond)
		}
	}
}

// Permission Tests, abfrage ob ein Member die jeweilige Rolle besitzt!
func hasMemberRole( session *discordgo.Session , guild * discordgo.Guild, member *discordgo.Member , permissionRole string ) bool {
	guildRoles , _ := session.GuildRoles( guild.ID )
	for _ , memberRoleID := range member.Roles {
		for _ , guildrole := range guildRoles {
			if guildrole.ID ==  memberRoleID && guildrole.Name == permissionRole {
				return true
			}
		}
	}

	return false
}

// Join the provided voice channel
func JoinVoiceChannel( s *discordgo.Session, guildID, channelID string ) (voice * discordgo.VoiceConnection , err error ){
	return s.ChannelVoiceJoin(guildID, channelID, false, true)
}

func main() {
	DiscordBot("discord-bot-config.json")
}

