package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(`server/.env`)
	if err != nil {
		log.Println("Failed to load env")
		return
	}

	token := os.Getenv("BOT_TOKEN")

	sess, err := discordgo.New("Bot " + token)

	if err != nil {
		log.Fatal("error creating session: " + err.Error())
	}

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

		if m.Author.ID == s.State.User.ID {
			return
		}

		if m.Content == "!status" {
			s.ChannelMessageSend(m.ChannelID, "Bot is live and running!")
		}

	})

	// So what does AllWithoutPrivileged include?
	// You still get: Message Create events, Reaction Add/Remove, Channel Create/Delete, Guild Create/Delete, Voice State updates, Typing events, Presence updates (partial)
	// But you do NOT get:
	// ❌ full member lists
	// ❌ full message text access for message content (unless in DMs or tagged commands)
	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()
	if err != nil {
		log.Fatal("error creating session: " + err.Error())
	}

	defer sess.Close()

	log.Println("Bot is live... Press CTRL+C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}
