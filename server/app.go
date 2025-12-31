package main

import (
	"dctest/functions"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// ---------------- MAIN ----------------

func main() {
	err := godotenv.Load(`server/.env`)
	if err != nil {
		log.Println("Failed to load env")
		return
	}

	log.SetOutput(os.Stdout)

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

	dg.AddHandler(functions.OnGuildCreate)
	dg.AddHandler(functions.OnInteractionCreate)

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
