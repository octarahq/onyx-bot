package utils

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
)

func GetMemberPermissions(client *bot.Client, member discord.Member) discord.Permissions {
	var botPerms discord.Permissions

	cachedRoles := client.Caches.MemberRoles(member)
	_, everyoneOk := client.Caches.Role(member.GuildID, member.GuildID)

	if everyoneOk && len(cachedRoles) == len(member.RoleIDs) {
		botPerms = client.Caches.MemberPermissions(member)
	} else {
		roles, err := client.Rest.GetRoles(member.GuildID)
		if err != nil {
			botPerms = client.Caches.MemberPermissions(member)
		} else {
			for _, role := range roles {
				if role.ID == member.GuildID {
					botPerms = botPerms.Add(role.Permissions)
					break
				}
			}
			for _, roleID := range member.RoleIDs {
				for _, role := range roles {
					if role.ID == roleID {
						botPerms = botPerms.Add(role.Permissions)
						break
					}
				}
			}
			if guild, ok := client.Caches.Guild(member.GuildID); ok && guild.OwnerID == member.User.ID {
				botPerms = discord.PermissionsAll
			} else if !ok {
				if guild, err := client.Rest.GetGuild(member.GuildID, false); err == nil && guild.OwnerID == member.User.ID {
					botPerms = discord.PermissionsAll
				}
			}
		}
	}

	return botPerms
}
