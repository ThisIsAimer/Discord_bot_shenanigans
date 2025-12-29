package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// ---------------- COMMAND DEFINITIONS ----------------

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Check bot latency",
	},
	{
		Name:        "search",
		Description: "Searches stuff",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "query",
				Description: "What do you want to search?",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	},
	{
		Name:                     "moveall",
		Description:              "Move users from one VC to another",
		DefaultMemberPermissions: ptr(int64(discordgo.PermissionVoiceMoveMembers)),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:         "from",
				Description:  "Source voice channel",
				Type:         discordgo.ApplicationCommandOptionChannel,
				Required:     true,
				Autocomplete: true,
			},
			{
				Name:         "to",
				Description:  "Target voice channel",
				Type:         discordgo.ApplicationCommandOptionChannel,
				Required:     true,
				Autocomplete: true,
			},
		},
	},
}

func ptr[T any](v T) *T {
	return &v
}

// ---------------- MAIN ----------------

func main() {
	err := godotenv.Load(`server/.env`)
	if err != nil {
		log.Println("Failed to load env")
		return
	}

	token := os.Getenv("BOT_TOKEN")
	guildID := os.Getenv("GUILD_ID")

	if token == "" || guildID == "" {
		log.Fatal("BOT_TOKEN or GUILD_ID missing")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	// Intents (voice states required)
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates

	dg.AddHandler(onInteractionCreate)

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	deleteAllCommands(dg, guildID)
	registerCommands(dg, guildID)

	log.Println("✅ Bot is running")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	dg.Close()
}

// ---------------- COMMAND ddeletion ----------------
func deleteAllCommands(s *discordgo.Session, guildID string) {

	// Delete GUILD commands
	guildCmds, _ := s.ApplicationCommands(s.State.User.ID, guildID)
	for _, c := range guildCmds {
		err := s.ApplicationCommandDelete(s.State.User.ID, guildID, c.ID)
		if err != nil {
			fmt.Println("couldint delete command:", err)
			return
		}
	}
}

// ---------------- COMMAND REGISTRATION ----------------

func registerCommands(s *discordgo.Session, guildid string) {
	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(
			s.State.User.ID,
			guildid,
			cmd,
		)
		if err != nil {
			log.Printf("❌ Failed to register %s: %v", cmd.Name, err)
		} else {
			log.Printf("✅ Registered /%s", cmd.Name)
		}
	}
}

// ---------------- INTERACTION HANDLER ----------------

func onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// 🔹 Autocomplete popup
	if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		handleAutocomplete(s, i)
		return
	}

	// 🔹 Slash command execution
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {

	case "ping":
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "🏓 Pong!",
			},
		})

	case "search":
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "🏓 Pong!",
			},
		})

	case "moveall":
		handleMoveAll(s, i)
	}
}

// ---------------- AUTOCOMPLETE ----------------

func handleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {

	guild, err := s.State.Guild(i.GuildID)
	if err != nil {
		return
	}

	choices := []*discordgo.ApplicationCommandOptionChoice{}

	for _, ch := range guild.Channels {
		if ch.Type == discordgo.ChannelTypeGuildVoice {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  ch.Name,
				Value: ch.ID,
			})
		}
		if len(choices) == 25 {
			break
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
}

// ---------------- MOVE LOGIC ----------------

func handleMoveAll(s *discordgo.Session, i *discordgo.InteractionCreate) {

	opts := i.ApplicationCommandData().Options

	fromID := opts[0].Value.(string)
	toID := opts[1].Value.(string)

	guild, err := s.State.Guild(i.GuildID)
	if err != nil {
		return
	}

	moved := 0

	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == fromID {
			err := s.GuildMemberMove(i.GuildID, vs.UserID, &toID)
			if err == nil {
				moved++
			}
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "✅ Moved " + itoa(moved) + " users",
		},
	})
}

// tiny helper (avoid strconv import)
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
