package functions

import (
	"dctest/guards"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func trashMessages(s *discordgo.Session, i *discordgo.InteractionCreate) {

	amount := int(i.ApplicationCommandData().Options[0].IntValue())
	if amount <= 0 {
		return
	}

	if !guards.CanRunTrash(s, i, amount) {
		return
	}

	//  EPHEMERAL ACK (CRITICAL)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Println("ACK failed:", err)
		return
	}

	//  Get response ID so we never delete it
	resp, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		log.Println("Failed to get interaction response:", err)
		return
	}
	responseID := resp.ID

	channelID := i.ChannelID
	deleted := 0
	var beforeID string

	for deleted < amount {

		limit := 100
		if amount-deleted < 100 {
			limit = amount - deleted
		}

		msgs, err := s.ChannelMessages(channelID, limit, beforeID, "", "")
		if err != nil || len(msgs) == 0 {
			break
		}

		var ids []string
		for _, m := range msgs {

			//  never delete our own interaction response
			if m.ID == responseID {
				continue
			}

			ids = append(ids, m.ID)
		}

		if len(ids) == 0 {
			break
		}

		// 🟡 single message → normal delete
		if len(ids) == 1 {
			err := s.ChannelMessageDelete(channelID, ids[0])
			if err == nil {
				deleted++
			}
			beforeID = msgs[len(msgs)-1].ID
			continue
		}

		// 🔥 bulk delete
		err = s.ChannelMessagesBulkDelete(channelID, ids)
		if err != nil {
			log.Println("Bulk delete failed:", err)
			break
		}

		deleted += len(ids)
		beforeID = msgs[len(msgs)-1].ID

		time.Sleep(1200 * time.Millisecond)
	}

	//  Final private response (cannot be deleted)
	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: ptr(fmt.Sprintf(
			"🗑️ **Deleted %d messages successfully.**",
			deleted,
		)),
	})
}

func trash14Messages(s *discordgo.Session, i *discordgo.InteractionCreate) {

	amount := int(i.ApplicationCommandData().Options[0].IntValue())
	if amount <= 0 {
		return
	}

	if !guards.CanRunTrash(s, i, amount) {
		return
	}

	//  EPHEMERAL ACK (CRITICAL)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Println("ACK failed:", err)
		return
	}

	//  Get interaction response ID so we never delete it
	resp, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		log.Println("Failed to get interaction response:", err)
		return
	}
	responseID := resp.ID

	channelID := i.ChannelID
	var beforeID string
	deleted := 0

	cutoff := time.Now().Add(-14 * 24 * time.Hour)

	for deleted < amount {

		msgs, err := s.ChannelMessages(channelID, 100, beforeID, "", "")
		if err != nil || len(msgs) == 0 {
			break
		}

		for _, m := range msgs {

			if deleted >= amount {
				break
			}

			// 🚫 NEVER delete our own interaction response
			if m.ID == responseID {
				continue
			}

			// Only delete 14+ day old messages
			if m.Timestamp.After(cutoff) {
				continue
			}

			err := s.ChannelMessageDelete(channelID, m.ID)
			if err != nil {
				log.Println("Delete failed:", err)
				continue
			}

			deleted++
			time.Sleep(1 * time.Second) // 🚦 Discord rate-limit safe
		}

		beforeID = msgs[len(msgs)-1].ID
	}

	//  Final private response (cannot be deleted)
	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: ptr(fmt.Sprintf(
			"🧹 **Deleted %d messages older than 14 days.**",
			deleted,
		)),
	})
}
