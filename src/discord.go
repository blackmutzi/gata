package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"time"
)

func main() {
	DiscordBot( "my_token" )
}

/*

 */
func DiscordBot( token string ) ( err error ){
	// Create a new Discord session using the provided bot token.
	session, err := discordgo.New("Bot " + token)
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

	if strings.HasPrefix( m.Content , "!join"){
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

		// Look for the message sender in that guild's current voice states.
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {

				member , _ := s.GuildMember( g.ID , vs.UserID )
				if hasMemberRole( s , g , member , "Master of Disaster") {
					s.ChannelMessageSend( m.ChannelID, "Hallo " + member.User.Username )
					voiceConnection , _ := JoinVoiceChannel( s , g.ID , vs.ChannelID )
					time.Sleep(2000 * time.Millisecond)
					voiceConnection.Disconnect()
				}

			}
		}
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


func JoinVoiceChannel( s *discordgo.Session, guildID, channelID string ) (voice * discordgo.VoiceConnection , err error ){
	// Join the provided voice channel.
	return s.ChannelVoiceJoin(guildID, channelID, false, true)
}




