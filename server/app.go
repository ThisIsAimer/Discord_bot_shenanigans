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

	if token == "" {
		log.Fatal("BOT_TOKEN is missing")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	// Intents (voice states required)
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates

	dg.AddHandler(onGuildCreate)
	dg.AddHandler(onInteractionCreate)

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

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

// for guild ------------------------------------------

func onGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	if g.Guild == nil {
		return
	}

	log.Printf("📌 Syncing commands for guild: %s (%s)", g.Name, g.ID)

	functions.DeleteGlobalCommands(s)
	functions.DeleteAllCommands(s, g.ID)
	functions.RegisterCommands(s, g.ID)
}
