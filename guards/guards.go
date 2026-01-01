package guards

import (
	"dctest/utils"

	"github.com/bwmarrin/discordgo"
)

func CanClaim(i *discordgo.InteractionCreate, supportRoleID string) bool {
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

func CanRunTrash(s *discordgo.Session, i *discordgo.InteractionCreate, amount int) bool {

	// 🔒 Hard permission gate (mods/admins only)
	if i.Member.Permissions&discordgo.PermissionManageMessages == 0 {
		utils.NoPerm(s, i, "❌ You need **Manage Messages** permission.")
		return false
	}

	admin := IsAdmin(i)

	// 🧮 Amount limit for non-admins
	if amount > 500 && !admin {
		utils.NoPerm(s, i, "❌ Only admins can delete more than **500 messages**.")
		return false
	}

	// 👑 Admins can run it anywhere
	if admin {
		return true
	}

	// 🎤 Non-admins → voice text chats only
	ch, err := s.State.Channel(i.ChannelID)
	if err != nil {
		return false
	}

	if ch.Type != discordgo.ChannelTypeGuildVoice {
		utils.NoPerm(
			s,
			i,
			"❌ This command can only be used in **voice channel text chats**.",
		)
		return false
	}

	return true
}

func IsAdmin(i *discordgo.InteractionCreate) bool {
	if i.Member == nil {
		return false
	}

	perms := i.Member.Permissions
	return perms&discordgo.PermissionAdministrator != 0
}

