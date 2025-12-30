package functions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
)

func ptr[T any](v T) *T {
	return &v
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

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
// delete global cmds
func DeleteGlobalCommands(s *discordgo.Session) {
	cmds, err := s.ApplicationCommands(s.State.User.ID, "")
	if err != nil {
		log.Println("Failed to fetch global commands:", err)
		return
	}

	for _, c := range cmds {
		log.Println("🗑 Deleting GLOBAL command:", c.Name)
		_ = s.ApplicationCommandDelete(s.State.User.ID, "", c.ID)
	}
}

// ---------------- COMMAND ddeletion ----------------
func DeleteAllCommands(s *discordgo.Session, guildID string) {

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

func RegisterCommands(s *discordgo.Session, guildid string) {
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

// ---------------- AUTOCOMPLETE ----------------

func HandleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {

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

func HandleMoveAll(s *discordgo.Session, i *discordgo.InteractionCreate) {

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

// call ai -----------------------------------------------

func CallAI(query string) (string, error) {
	url := "https://api.groq.com/openai/v1/chat/completions"

	// JSON payload for Groq
	payload := map[string]interface{}{
		"model": "llama-3.1-8b-instant",
		"messages": []map[string]string{
			{"role": "system", "content": "Keep responses short, maybe 20-30 words. but if answer needs to be longer or user want longer responses make them longer. Also use a few emojis"},
			{"role": "user", "content": query},
		},
	}

	bodyJson, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyJson))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("GROQ_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("groq API error (%d): %s", resp.StatusCode, string(b))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	return result.Choices[0].Message.Content, nil
}
