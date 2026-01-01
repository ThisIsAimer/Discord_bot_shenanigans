package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func NoPerm(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {

	if message == "" {
		message = "⛔ You don’t have permission to do that."
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func Sanitize(name string) string {
	name = strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9-]`)
	return re.ReplaceAllString(name, "")
}

func FetchAllMessages(s *discordgo.Session, channelID string) ([]*discordgo.Message, error) {
	var all []*discordgo.Message
	var beforeID string

	for {
		msgs, err := s.ChannelMessages(channelID, 100, beforeID, "", "")
		if err != nil {
			return nil, err
		}
		if len(msgs) == 0 {
			break
		}
		all = append(all, msgs...)
		beforeID = msgs[len(msgs)-1].ID
	}

	// reverse to oldest → newest
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}

	return all, nil
}

func BuildTimelineEmbeds(s *discordgo.Session, mediaChannelID string, msgs []*discordgo.Message) ([]*discordgo.MessageEmbed, error) {

	ist, _ := time.LoadLocation("Asia/Kolkata")
	var timeline []*discordgo.MessageEmbed

	for _, m := range msgs {
		if m.Author.Bot {
			continue
		}

		t := m.Timestamp.In(ist)
		timeStr := t.Format("2006-01-02 15:04")

		footer := &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("%s • %s", m.Author.Username, timeStr),
		}

		// 🔹 TEXT EMBED (if any)
		if strings.TrimSpace(m.Content) != "" {
			timeline = append(timeline, &discordgo.MessageEmbed{
				Description: m.Content,
				Color:       0x5865F2,
				Footer:      footer,
				Timestamp:   m.Timestamp.Format(time.RFC3339),
			})
		}

		// 🔹 ATTACHMENTS → reupload → embed
		for _, a := range m.Attachments {
			data, err := DownloadFile(a.URL)
			if err != nil {
				continue
			}

			newURL, err := ReuploadFile(
				s,
				mediaChannelID, // 👈 upload to STORAGE, not transcript
				a.Filename,
				a.ContentType,
				data,
			)

			data = nil

			if err != nil {
				continue
			}

			em := &discordgo.MessageEmbed{
				Footer:    footer,
				Timestamp: m.Timestamp.Format(time.RFC3339),
				Color:     0xFAA61A,
			}

			if strings.HasPrefix(a.ContentType, "image/") {
				em.Title = "🖼️ Image"
				em.Image = &discordgo.MessageEmbedImage{
					URL: newURL,
				}
			} else {
				em.Title = "📎 Attachment"
				em.Description = fmt.Sprintf("[%s](%s)", a.Filename, newURL)
			}

			timeline = append(timeline, em)
		}
	}

	return timeline, nil
}

func DownloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func ReuploadFile(
	s *discordgo.Session,
	channelID, filename, contentType string,
	data []byte,
) (string, error) {

	msg, err := s.ChannelMessageSendComplex(
		channelID,
		&discordgo.MessageSend{
			Files: []*discordgo.File{
				{
					Name:        filename,
					ContentType: contentType,
					Reader:      bytes.NewReader(data),
				},
			},
		},
	)
	if err != nil {
		return "", err
	}

	return msg.Attachments[0].URL, nil
}
