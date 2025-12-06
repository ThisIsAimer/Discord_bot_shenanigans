package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func main() {
	err := godotenv.Load(`server/.env`)
	if err != nil {
		log.Println("Failed to load env")
		return
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Println("empty bot token")
	}

	apiKey := os.Getenv("GEM_API")
	if apiKey == "" {
		log.Println("empty api key")
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + apiKey

	sess, err := discordgo.New("Bot " + token)

	if err != nil {
		log.Fatal("error creating session: " + err.Error())
	}

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

		if m.Author.ID == s.State.User.ID {
			return
		}

		// status ------------------------------------------------------------------------------------------------------------
		if m.Content == "!status" {
			s.ChannelMessageSend(m.ChannelID, "Bot is live and running!")
		}

		// search -----------------------------------------------------------------------------------------------------------
		if strings.HasPrefix(m.Content, "!search ") {

			reqBody := map[string]interface{}{
				"systemInstruction": map[string]interface{}{
					"parts": []map[string]string{
						{"text": "Try to answer Things in 30 words, make answers as short as possible but if answer needs to be long make it."},
					},
				},
				"contents": []map[string]interface{}{
					{
						"parts": []map[string]string{
							{"text": m.Content[8:]},
						},
					},
				},
			}

			jsonData, err := json.Marshal(reqBody)
			if err != nil {
				log.Fatal("error marshaling data: " + err.Error())
			}

			// create POST request
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
			if err != nil {
				log.Fatal("error making request: " + err.Error())
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal("error getting response: " + err.Error())
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal("error reading resp body:", err)
			}

			var result GeminiResponse
			json.Unmarshal(body, &result)

			s.ChannelMessageSend(m.ChannelID, result.Candidates[0].Content.Parts[0].Text)

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
