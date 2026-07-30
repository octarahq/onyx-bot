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

func CheckDangerousPermissions(client *bot.Client, member discord.Member) (bool, []discord.Permissions) {
	perms := GetMemberPermissions(client, member)
	var dangerousPerms []discord.Permissions

	for _, p := range []discord.Permissions{
		discord.PermissionAdministrator,
		discord.PermissionManageGuild,
		discord.PermissionManageRoles,
		discord.PermissionManageChannels,
		discord.PermissionKickMembers,
		discord.PermissionBanMembers,
		discord.PermissionManageNicknames,
		discord.PermissionManageWebhooks,
	} {
		if perms&p != 0 {
			dangerousPerms = append(dangerousPerms, p)
		}
	}

	return len(dangerousPerms) > 0, dangerousPerms
}

func RemoveMemberPerms(client *bot.Client, member discord.Member, perms []discord.Permissions) {
	for _, roleID := range member.RoleIDs {
		if role, ok := client.Caches.Role(member.GuildID, roleID); ok {
			for _, p := range perms {
				if role.Permissions.Has(p) {
					client.Rest.RemoveMemberRole(member.GuildID, member.User.ID, roleID)
				}
			}
		}
	}
}
