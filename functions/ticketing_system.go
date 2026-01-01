package functions

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

var TICKET_ADMIN_ROLE_ID string
var TICKET_ROLE_ID string
var LOG_CHANNEL_ID string
var IMAGE_DUMP_ID string
var CATAGORY_ID string

var ticket_number = 1
var mu sync.Mutex

func ticketCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "🎟️ Help/Complaint Ticket",
					Description: `Hey member create your help/complaint ticket by just clicking on the button below "📩 Create a Ticket".`,
					Color:       0x5865F2, // Discord blurple
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "📌 NOTE:",
							Value: "• Tag <@&" + TICKET_ADMIN_ROLE_ID + "> in your ticket channel for further process.\n• Do not tag mods more than once.",
						},
					},

					Footer: &discordgo.MessageEmbedFooter{
						IconURL: "https://cdn.discordapp.com/attachments/1455598749123346616/1455918006801535088/Sananatani_Sena_Logo_2.png?ex=695678ce&is=6955274e&hm=c238f2f2418b26f60e6ae90dd8673a79751492ae4b1d0295bebdb1180ecc0082&",
						Text:    "Sanatani Sena Ticket Panel",
					},
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Create a Ticket",
							Style:    discordgo.SuccessButton,
							CustomID: "ticket_create",
							Emoji: &discordgo.ComponentEmoji{
								Name: "📩",
							},
						},
					},
				},
			},
		},
	})
}

func ticketCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {

	mu.Lock()
	defer mu.Unlock()

	guildID := i.GuildID
	user := i.Member.User

	username := sanitize(user.Username)

	channel, err := s.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name:     "ticket-" + strconv.Itoa(ticket_number) + "-" + username,
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: CATAGORY_ID,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:   guildID,
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: discordgo.PermissionViewChannel,
			},
			{
				ID:   user.ID,
				Type: discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionViewChannel |
					discordgo.PermissionAttachFiles |
					discordgo.PermissionSendMessages |
					discordgo.PermissionReadMessageHistory,
			},
			{
				ID:   s.State.User.ID, // Bot's user ID
				Type: discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionViewChannel |
					discordgo.PermissionSendMessages |
					discordgo.PermissionReadMessageHistory,
			},
			// {
			// 	ID:   modRoleID, // Moderator role ID
			// 	Type: discordgo.PermissionOverwriteTypeRole,
			// 	Allow: discordgo.PermissionViewChannel |
			// 		discordgo.PermissionSendMessages |
			// 		discordgo.PermissionReadMessageHistory,
			// },
		},
	})

	ticket_number++

	ticketOwner[channel.ID] = i.Member.User.ID

	if err != nil {
		log.Println(err)
		return
	}

	err = s.GuildMemberRoleAdd(
		i.GuildID,
		i.Member.User.ID,
		TICKET_ROLE_ID,
	)
	if err != nil {
		log.Println("failed to add ticket role:", err)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "✅ Ticket created: <#" + channel.ID + ">",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	_, err = s.ChannelMessageSend(channel.ID,
		fmt.Sprintf("💜 **<@%s> Your complaint ticket has been created**", user.ID),
	)
	if err != nil {
		log.Println("failed to send close message:", err)
	}

	// Welcome message in ticket
	s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "Welcome " + user.Username + "!",
				Description: "Please describe the reasoning for opening this ticket, include any information you think may be relevant such as proof, other third parties and so on.\n\n• Tag <@&" + TICKET_ADMIN_ROLE_ID + "> to look at your complaint.",
				Color:       0xED4245, // red accent

				Image: &discordgo.MessageEmbedImage{
					URL: "https://cdn-longterm.mee6.xyz/plugins/embeds/images/1192481727185158144/fd936bf0632d64f1d7cb4b63aa83bd9b77cc2f491b587698373cd675c00b96a0.png", // upload image somewhere & paste link
				},

				Footer: &discordgo.MessageEmbedFooter{
					IconURL: "https://cdn.discordapp.com/attachments/1455598749123346616/1455918006801535088/Sananatani_Sena_Logo_2.png?ex=695678ce&is=6955274e&hm=c238f2f2418b26f60e6ae90dd8673a79751492ae4b1d0295bebdb1180ecc0082&",
					Text:    "सनातनी सेना",
				},
			},
		},
		Components: *ticketButtons(true),
	})

}

func ticketButtons(open bool) *[]discordgo.MessageComponent {
	return &[]discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Claim",
					Style:    discordgo.PrimaryButton,
					CustomID: "ticket_claim",
					Emoji:    &discordgo.ComponentEmoji{Name: "🎟️"},
					Disabled: !open,
				},
				discordgo.Button{
					Label:    "Close",
					Style:    discordgo.SecondaryButton,
					CustomID: "ticket_close",
					Emoji:    &discordgo.ComponentEmoji{Name: "🔒"},
					Disabled: !open, // ✅ enabled initially
				},
				discordgo.Button{
					Label:    "Reopen",
					Style:    discordgo.SuccessButton,
					CustomID: "ticket_reopen",
					Emoji:    &discordgo.ComponentEmoji{Name: "🔓"},
					Disabled: open, // ❌ disabled initially
				},
				discordgo.Button{
					Label:    "Delete",
					Style:    discordgo.DangerButton,
					CustomID: "ticket_delete",
					Emoji:    &discordgo.ComponentEmoji{Name: "🗑️"},
				},
			},
		},
	}
}

func claimTicket(s *discordgo.Session, i *discordgo.InteractionCreate) {

	if !canClaim(i, TICKET_ADMIN_ROLE_ID) {
		noPerm(s, i)
		return
	}

	ch, err := s.State.Channel(i.ChannelID)
	if err != nil {
		noPerm(s, i)
		return
	}

	if strings.Contains(ch.Name, "-claimed-") {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ This ticket has already been claimed.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// ✅ ACK FIRST (this is critical)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Println("ack failed:", err)
		return
	}

	// 🔁 now it is SAFE to do slow operations
	claimer := sanitize(i.Member.User.Username)
	newName := ch.Name + "-claimed-" + claimer

	_, err = s.ChannelEdit(i.ChannelID, &discordgo.ChannelEdit{
		Name: newName,
	})
	if err != nil {
		log.Println("rename failed:", err)
	}

	_, err = s.ChannelMessageSend(i.ChannelID,
		fmt.Sprintf("🎟️ **Ticket Has now been claimed by <@%s>**", i.Member.User.ID),
	)
	if err != nil {
		log.Println("failed to send close message:", err)
	}
	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{
			{
				Description: fmt.Sprintf(
					"🙌 %s, **you claimed the ticket successfully!**\n\n"+
						"📁 The ticket has been moved to **Help/Complaint** category.",
					i.Member.User.Mention(),
				),
				Color: 0x57F287,
			},
		},
		Flags: discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		log.Println("followup failed:", err)
	}

}

func closeTicket(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ownerID, ok := ticketOwner[i.ChannelID]
	if !ok {
		noPerm(s, i)
		return
	}

	userID := i.Member.User.ID
	isMod := i.Member.Permissions&discordgo.PermissionModerateMembers != 0
	isOwner := userID == ownerID

	// only mods OR owner can close
	if !isMod && !isOwner {
		noPerm(s, i)
		return
	}

	// ✅ 1. ACK the interaction FIRST
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Println("interaction ack failed:", err)
		return
	}

	_, err = s.ChannelMessageSend(i.ChannelID,
		fmt.Sprintf("🔒 **Ticket closed by <@%s>**", userID),
	)
	if err != nil {
		log.Println("failed to send close message:", err)
	}

	// ✅ 2. lock user from typing
	err = s.ChannelPermissionSet(
		i.ChannelID,
		ownerID,
		discordgo.PermissionOverwriteTypeMember,
		discordgo.PermissionViewChannel|
			discordgo.PermissionReadMessageHistory,
		discordgo.PermissionSendMessages,
	)
	if err != nil {
		log.Println("perm close failed:", err)
	}

	err = s.GuildMemberRoleRemove(
		i.GuildID,
		ownerID,
		TICKET_ROLE_ID,
	)
	if err != nil {
		log.Println("failed to remove ticket role:", err)
	}

	// ✅ 3. update buttons (close disabled, reopen enabled)
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Components: ticketButtons(false),
	})
	if err != nil {
		log.Println("edit failed:", err)
	}
}

func reOpenTicket(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ownerID, ok := ticketOwner[i.ChannelID]
	if !ok {
		noPerm(s, i)
		return
	}

	userID := i.Member.User.ID
	isMod := i.Member.Permissions&discordgo.PermissionModerateMembers != 0
	isOwner := userID == ownerID

	if !isMod && !isOwner {
		noPerm(s, i)
		return
	}

	// ✅ ACK first
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	if err != nil {
		log.Println("interaction ack failed:", err)
		return
	}

	_, err = s.ChannelMessageSend(i.ChannelID,
		fmt.Sprintf("🔒 **Ticket reopened by <@%s>**", userID),
	)
	if err != nil {
		log.Println("failed to send close message:", err)
	}

	// ✅ allow typing again
	err = s.ChannelPermissionSet(
		i.ChannelID,
		ownerID,
		discordgo.PermissionOverwriteTypeMember,
		discordgo.PermissionViewChannel|
			discordgo.PermissionSendMessages|
			discordgo.PermissionReadMessageHistory,
		0,
	)

	err = s.GuildMemberRoleAdd(
		i.GuildID,
		i.Member.User.ID,
		TICKET_ROLE_ID,
	)

	if err != nil {
		log.Println("failed to add ticket role:", err)
	}

	if err != nil {
		log.Println("perm close failed:", err)
	}

	// ✅ toggle buttons back
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Components: ticketButtons(true),
	})
}

func warnTicketDelete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Description: fmt.Sprintf(
						"👋 %s, **are you sure you want to delete this ticket?**\n\n"+
							"⚠️ The channel will be **deleted** and a **transcript will be generated**.",
						i.Member.User.Mention(),
					),
					Color: 0xED4245, // Red (danger)
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Confirm",
							Style:    discordgo.DangerButton,
							CustomID: "ticket_confirm_delete",
							Emoji:    &discordgo.ComponentEmoji{Name: "🗑️"},
						},
					},
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func ticketConfirmDelete(s *discordgo.Session, i *discordgo.InteractionCreate) {

	ownerID, ok := ticketOwner[i.ChannelID]
	if !ok {
		noPerm(s, i)
		return
	}

	member := i.Member
	if member == nil {
		noPerm(s, i)
		return
	}

	// ✅ Check admin permission
	isAdmin := member.Permissions&discordgo.PermissionAdministrator != 0

	// ✅ Check specific role
	hasRole := false
	for _, roleID := range member.Roles {
		if roleID == TICKET_ADMIN_ROLE_ID {
			hasRole = true
			break
		}
	}

	if !isAdmin && !hasRole {
		noPerm(s, i)
		return
	}

	// ✅ ACK interaction first
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	// Optional: system message
	s.ChannelMessageSend(
		i.ChannelID,
		fmt.Sprintf("🗑️ **Ticket deleted by <@%s>**", member.User.ID),
	)

	err := s.GuildMemberRoleRemove(
		i.GuildID,
		ownerID,
		TICKET_ROLE_ID,
	)
	if err != nil {
		log.Println("failed to remove ticket role:", err)
	}

	msgs, err := fetchAllMessages(s, i.ChannelID)
	if err != nil {
		log.Println("failed to fetch messages:", err)
		return
	}

	ch, err := s.State.Channel(i.ChannelID)
	if err != nil {
		ch, err = s.Channel(i.ChannelID) // fallback if not cached
		if err != nil {
			log.Println("failed to get channel:", err)
			return
		}
	}
	channelName := ch.Name

	embeds, err := buildTimelineEmbeds(s, IMAGE_DUMP_ID, msgs)
	if err != nil {
		log.Println(err)
		return
	}

	// header message
	s.ChannelMessageSend(
		LOG_CHANNEL_ID,
		fmt.Sprintf("📄 **Transcript for <#%s>**", channelName),
	)

	// send embeds in rows
	for i := 0; i < len(embeds); i += 10 {
		end := i + 10
		if end > len(embeds) {
			end = len(embeds)
		}

		s.ChannelMessageSendComplex(
			LOG_CHANNEL_ID,
			&discordgo.MessageSend{
				Embeds: embeds[i:end],
			},
		)
	}

	// ✅ Delete channel
	_, err = s.ChannelDelete(i.ChannelID)
	if err != nil {
		log.Println("failed to delete channel:", err)
	}

	delete(ticketOwner, i.ChannelID)

}

// util functions ------------------------------------------------------------------------------------------------------------------------

func sanitize(name string) string {
	name = strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9-]`)
	return re.ReplaceAllString(name, "")
}

func canClaim(i *discordgo.InteractionCreate, supportRoleID string) bool {
	if i.Member == nil {
		return false
	}

	// 1 Admins can always claim
	if i.Member.Permissions&discordgo.PermissionAdministrator != 0 {
		return true
	}

	// 2 Members with support role can claim
	for _, roleID := range i.Member.Roles {
		if roleID == supportRoleID {
			return true
		}
	}

	return false
}

func fetchAllMessages(s *discordgo.Session, channelID string) ([]*discordgo.Message, error) {
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

func buildTimelineEmbeds(s *discordgo.Session, mediaChannelID string, msgs []*discordgo.Message) ([]*discordgo.MessageEmbed, error) {

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
			data, err := downloadFile(a.URL)
			if err != nil {
				continue
			}

			newURL, err := reuploadFile(
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

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func reuploadFile(
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
