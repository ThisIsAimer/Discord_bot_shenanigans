package functions

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

var ticketOwner = map[string]string{}

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
		Name:                     "ticket",
		Description:              "Create a support ticket",
		DefaultMemberPermissions: ptr(int64(discordgo.PermissionAdministrator)),
	},
	{
		Name:                     "trash",
		Description:              "Delete a number of messages in this channel",
		DefaultMemberPermissions: ptr(int64(discordgo.PermissionManageMessages)),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "amount",
				Description: "Number of messages to delete",
				Type:        discordgo.ApplicationCommandOptionInteger,
				Required:    true,
			},
		},
	},
	{
		Name:                     "trash14",
		Description:              "Delete messages older than 14 days (slow mode)",
		DefaultMemberPermissions: ptr(int64(discordgo.PermissionManageMessages)),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "amount",
				Description: "Number of old messages to delete",
				Type:        discordgo.ApplicationCommandOptionInteger,
				Required:    true,
				MinValue:    ptr(1.),
				MaxValue:    20000,
			},
		},
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
		DefaultMemberPermissions: ptr(int64(discordgo.PermissionKickMembers)),
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

func OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {

	switch i.Type {

	// 🔹 AUTOCOMPLETE
	case discordgo.InteractionApplicationCommandAutocomplete:
		handleAutocomplete(s, i)

	// 🔹 SLASH COMMANDS
	case discordgo.InteractionApplicationCommand:
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

			resp, err := callAI(query)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: resp,
				},
			})

		case "trash":
			trashMessages(s, i)

		case "trash14":
			trash14Messages(s, i)

		case "moveall":
			handleMoveAll(s, i)

		case "ticket":
			ticketCommand(s, i)
		}

	// 🔹 BUTTONS / COMPONENTS
	case discordgo.InteractionMessageComponent:
		switch i.MessageComponentData().CustomID {

		case "ticket_claim":
			claimTicket(s, i)

		case "ticket_create":
			ticketCreate(s, i)

		case "ticket_close":
			closeTicket(s, i)

		case "ticket_reopen":
			reOpenTicket(s, i)

		case "ticket_delete":
			warnTicketDelete(s, i)

		case "ticket_confirm_delete":
			ticketConfirmDelete(s, i)
		}

	default:
		return
	}

}

// for guild ------------------------------------------

func OnGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	if g.Guild == nil {
		return
	}

	log.Printf("📌 Syncing commands for guild: %s (%s)", g.Name, g.ID)

	deleteGlobalCommands(s)
	deleteAllCommands(s, g.ID)
	registerCommands(s, g.ID)
}
