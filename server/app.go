package main

import (
	"dctest/functions"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv"
)

// ---------------- MAIN ----------------

func main() {
	// err := godotenv.Load(`server/.env`)
	// if err != nil {
	// 	log.Println("Failed to load env")
	// 	return
	// }

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

	functions.DeleteAllCommands(dg, guildID)
	functions.RegisterCommands(dg, guildID)

	log.Println("✅ Bot is running")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	dg.Close()
}

// bot handler

func onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// 🔹 Autocomplete popup
	if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		functions.HandleAutocomplete(s, i)
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

		var query string
		if len(i.ApplicationCommandData().Options) > 0 {
			query = i.ApplicationCommandData().Options[0].StringValue()
		}

		resp, err := functions.CallAI(query)

		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: err.Error(),
				},
			})
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: resp,
			},
		})

	case "moveall":
		functions.HandleMoveAll(s, i)
	}
}
