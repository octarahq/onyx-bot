package handlers

import (
	"onyx/bot/core"
	"onyx/bot/utils"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func ExecModulesEvent(b *core.Bot, event bot.Event) bool {
	if e, ok := event.(*events.Ready); ok {
		for _, mod := range b.Modules {
			if handler, ok := mod.(core.OnReady); ok {
				if handler.HandleReady(b, e) {
					return true
				}
			}
		}
		return false
	}

	var guildID *snowflake.ID

	switch e := event.(type) {
	case *events.MessageCreate:
		guildID = e.GuildID
	case *events.GuildMemberJoin:
		guildID = &e.GuildID
	case *events.GuildMemberLeave:
		guildID = &e.GuildID

	}

	if guildID == nil {
		return false
	}

	guildIDStr := guildID.String()

	me, ok := event.Client().Caches.Member(*guildID, event.Client().ID())
	if !ok {
		m, err := event.Client().Rest.GetMember(*guildID, event.Client().ID())
		if err != nil {
			return false
		}
		me = *m
	}

	for _, mod := range b.Modules {
		if dbAware, ok := mod.(core.DatabaseAware); ok {
			if err := dbAware.LoadData(b.DB.GormDB, guildIDStr); err != nil {
				continue
			}
		}

		if !mod.IsEnabled() {
			continue
		}

		if !CheckPerms(b, mod, me) {
			continue
		}

		switch e := event.(type) {
		case *events.ApplicationCommandInteractionCreate:
			if handler, ok := mod.(core.OnApplicationCommandInteractionCreate); ok {
				if handler.HandleApplicationCommandInteractionCreate(b, e) {
					return true
				}
			}
		case *events.AutoModerationActionExecution:
			if handler, ok := mod.(core.OnAutoModerationActionExecution); ok {
				if handler.HandleAutoModerationActionExecution(b, e) {
					return true
				}
			}
		case *events.AutoModerationRuleCreate:
			if handler, ok := mod.(core.OnAutoModerationRuleCreate); ok {
				if handler.HandleAutoModerationRuleCreate(b, e) {
					return true
				}
			}
		case *events.AutoModerationRuleDelete:
			if handler, ok := mod.(core.OnAutoModerationRuleDelete); ok {
				if handler.HandleAutoModerationRuleDelete(b, e) {
					return true
				}
			}
		case *events.AutoModerationRuleUpdate:
			if handler, ok := mod.(core.OnAutoModerationRuleUpdate); ok {
				if handler.HandleAutoModerationRuleUpdate(b, e) {
					return true
				}
			}
		case *events.AutocompleteInteractionCreate:
			if handler, ok := mod.(core.OnAutocompleteInteractionCreate); ok {
				if handler.HandleAutocompleteInteractionCreate(b, e) {
					return true
				}
			}
		case *events.ComponentInteractionCreate:
			if handler, ok := mod.(core.OnComponentInteractionCreate); ok {
				if handler.HandleComponentInteractionCreate(b, e) {
					return true
				}
			}
		case *events.DMChannelPinsUpdate:
			if handler, ok := mod.(core.OnDMChannelPinsUpdate); ok {
				if handler.HandleDMChannelPinsUpdate(b, e) {
					return true
				}
			}
		case *events.DMMessageCreate:
			if handler, ok := mod.(core.OnDMMessageCreate); ok {
				if handler.HandleDMMessageCreate(b, e) {
					return true
				}
			}
		case *events.DMMessageDelete:
			if handler, ok := mod.(core.OnDMMessageDelete); ok {
				if handler.HandleDMMessageDelete(b, e) {
					return true
				}
			}
		case *events.DMMessagePollVoteAdd:
			if handler, ok := mod.(core.OnDMMessagePollVoteAdd); ok {
				if handler.HandleDMMessagePollVoteAdd(b, e) {
					return true
				}
			}
		case *events.DMMessagePollVoteRemove:
			if handler, ok := mod.(core.OnDMMessagePollVoteRemove); ok {
				if handler.HandleDMMessagePollVoteRemove(b, e) {
					return true
				}
			}
		case *events.DMMessageReactionAdd:
			if handler, ok := mod.(core.OnDMMessageReactionAdd); ok {
				if handler.HandleDMMessageReactionAdd(b, e) {
					return true
				}
			}
		case *events.DMMessageReactionRemove:
			if handler, ok := mod.(core.OnDMMessageReactionRemove); ok {
				if handler.HandleDMMessageReactionRemove(b, e) {
					return true
				}
			}
		case *events.DMMessageReactionRemoveAll:
			if handler, ok := mod.(core.OnDMMessageReactionRemoveAll); ok {
				if handler.HandleDMMessageReactionRemoveAll(b, e) {
					return true
				}
			}
		case *events.DMMessageReactionRemoveEmoji:
			if handler, ok := mod.(core.OnDMMessageReactionRemoveEmoji); ok {
				if handler.HandleDMMessageReactionRemoveEmoji(b, e) {
					return true
				}
			}
		case *events.DMMessageUpdate:
			if handler, ok := mod.(core.OnDMMessageUpdate); ok {
				if handler.HandleDMMessageUpdate(b, e) {
					return true
				}
			}
		case *events.DMUserTypingStart:
			if handler, ok := mod.(core.OnDMUserTypingStart); ok {
				if handler.HandleDMUserTypingStart(b, e) {
					return true
				}
			}
		case *events.EmojiCreate:
			if handler, ok := mod.(core.OnEmojiCreate); ok {
				if handler.HandleEmojiCreate(b, e) {
					return true
				}
			}
		case *events.EmojiDelete:
			if handler, ok := mod.(core.OnEmojiDelete); ok {
				if handler.HandleEmojiDelete(b, e) {
					return true
				}
			}
		case *events.EmojiUpdate:
			if handler, ok := mod.(core.OnEmojiUpdate); ok {
				if handler.HandleEmojiUpdate(b, e) {
					return true
				}
			}
		case *events.EmojisUpdate:
			if handler, ok := mod.(core.OnEmojisUpdate); ok {
				if handler.HandleEmojisUpdate(b, e) {
					return true
				}
			}
		case *events.EntitlementCreate:
			if handler, ok := mod.(core.OnEntitlementCreate); ok {
				if handler.HandleEntitlementCreate(b, e) {
					return true
				}
			}
		case *events.EntitlementDelete:
			if handler, ok := mod.(core.OnEntitlementDelete); ok {
				if handler.HandleEntitlementDelete(b, e) {
					return true
				}
			}
		case *events.EntitlementUpdate:
			if handler, ok := mod.(core.OnEntitlementUpdate); ok {
				if handler.HandleEntitlementUpdate(b, e) {
					return true
				}
			}
		case *events.GuildApplicationCommandPermissionsUpdate:
			if handler, ok := mod.(core.OnGuildApplicationCommandPermissionsUpdate); ok {
				if handler.HandleGuildApplicationCommandPermissionsUpdate(b, e) {
					return true
				}
			}
		case *events.GuildAuditLogEntryCreate:
			if handler, ok := mod.(core.OnGuildAuditLogEntryCreate); ok {
				if handler.HandleGuildAuditLogEntryCreate(b, e) {
					return true
				}
			}
		case *events.GuildAvailable:
			if handler, ok := mod.(core.OnGuildAvailable); ok {
				if handler.HandleGuildAvailable(b, e) {
					return true
				}
			}
		case *events.GuildBan:
			if handler, ok := mod.(core.OnGuildBan); ok {
				if handler.HandleGuildBan(b, e) {
					return true
				}
			}
		case *events.GuildChannelCreate:
			if handler, ok := mod.(core.OnGuildChannelCreate); ok {
				if handler.HandleGuildChannelCreate(b, e) {
					return true
				}
			}
		case *events.GuildChannelDelete:
			if handler, ok := mod.(core.OnGuildChannelDelete); ok {
				if handler.HandleGuildChannelDelete(b, e) {
					return true
				}
			}
		case *events.GuildChannelPinsUpdate:
			if handler, ok := mod.(core.OnGuildChannelPinsUpdate); ok {
				if handler.HandleGuildChannelPinsUpdate(b, e) {
					return true
				}
			}
		case *events.GuildChannelUpdate:
			if handler, ok := mod.(core.OnGuildChannelUpdate); ok {
				if handler.HandleGuildChannelUpdate(b, e) {
					return true
				}
			}
		case *events.GuildIntegrationsUpdate:
			if handler, ok := mod.(core.OnGuildIntegrationsUpdate); ok {
				if handler.HandleGuildIntegrationsUpdate(b, e) {
					return true
				}
			}
		case *events.GuildJoin:
			if handler, ok := mod.(core.OnGuildJoin); ok {
				if handler.HandleGuildJoin(b, e) {
					return true
				}
			}
		case *events.GuildLeave:
			if handler, ok := mod.(core.OnGuildLeave); ok {
				if handler.HandleGuildLeave(b, e) {
					return true
				}
			}
		case *events.GuildMemberJoin:
			if handler, ok := mod.(core.OnGuildMemberJoin); ok {
				if handler.HandleGuildMemberJoin(b, e) {
					return true
				}
			}
		case *events.GuildMemberLeave:
			if handler, ok := mod.(core.OnGuildMemberLeave); ok {
				if handler.HandleGuildMemberLeave(b, e) {
					return true
				}
			}
		case *events.GuildMemberTypingStart:
			if handler, ok := mod.(core.OnGuildMemberTypingStart); ok {
				if handler.HandleGuildMemberTypingStart(b, e) {
					return true
				}
			}
		case *events.GuildMemberUpdate:
			if handler, ok := mod.(core.OnGuildMemberUpdate); ok {
				if handler.HandleGuildMemberUpdate(b, e) {
					return true
				}
			}
		case *events.GuildMessageCreate:
			if handler, ok := mod.(core.OnGuildMessageCreate); ok {
				if handler.HandleGuildMessageCreate(b, e) {
					return true
				}
			}
		case *events.GuildMessageDelete:
			if handler, ok := mod.(core.OnGuildMessageDelete); ok {
				if handler.HandleGuildMessageDelete(b, e) {
					return true
				}
			}
		case *events.GuildMessagePollVoteAdd:
			if handler, ok := mod.(core.OnGuildMessagePollVoteAdd); ok {
				if handler.HandleGuildMessagePollVoteAdd(b, e) {
					return true
				}
			}
		case *events.GuildMessagePollVoteRemove:
			if handler, ok := mod.(core.OnGuildMessagePollVoteRemove); ok {
				if handler.HandleGuildMessagePollVoteRemove(b, e) {
					return true
				}
			}
		case *events.GuildMessageReactionAdd:
			if handler, ok := mod.(core.OnGuildMessageReactionAdd); ok {
				if handler.HandleGuildMessageReactionAdd(b, e) {
					return true
				}
			}
		case *events.GuildMessageReactionRemove:
			if handler, ok := mod.(core.OnGuildMessageReactionRemove); ok {
				if handler.HandleGuildMessageReactionRemove(b, e) {
					return true
				}
			}
		case *events.GuildMessageReactionRemoveAll:
			if handler, ok := mod.(core.OnGuildMessageReactionRemoveAll); ok {
				if handler.HandleGuildMessageReactionRemoveAll(b, e) {
					return true
				}
			}
		case *events.GuildMessageReactionRemoveEmoji:
			if handler, ok := mod.(core.OnGuildMessageReactionRemoveEmoji); ok {
				if handler.HandleGuildMessageReactionRemoveEmoji(b, e) {
					return true
				}
			}
		case *events.GuildMessageUpdate:
			if handler, ok := mod.(core.OnGuildMessageUpdate); ok {
				if handler.HandleGuildMessageUpdate(b, e) {
					return true
				}
			}
		case *events.GuildScheduledEventCreate:
			if handler, ok := mod.(core.OnGuildScheduledEventCreate); ok {
				if handler.HandleGuildScheduledEventCreate(b, e) {
					return true
				}
			}
		case *events.GuildScheduledEventDelete:
			if handler, ok := mod.(core.OnGuildScheduledEventDelete); ok {
				if handler.HandleGuildScheduledEventDelete(b, e) {
					return true
				}
			}
		case *events.GuildScheduledEventUpdate:
			if handler, ok := mod.(core.OnGuildScheduledEventUpdate); ok {
				if handler.HandleGuildScheduledEventUpdate(b, e) {
					return true
				}
			}
		case *events.GuildScheduledEventUserAdd:
			if handler, ok := mod.(core.OnGuildScheduledEventUserAdd); ok {
				if handler.HandleGuildScheduledEventUserAdd(b, e) {
					return true
				}
			}
		case *events.GuildScheduledEventUserRemove:
			if handler, ok := mod.(core.OnGuildScheduledEventUserRemove); ok {
				if handler.HandleGuildScheduledEventUserRemove(b, e) {
					return true
				}
			}
		case *events.GuildSoundboardSoundCreate:
			if handler, ok := mod.(core.OnGuildSoundboardSoundCreate); ok {
				if handler.HandleGuildSoundboardSoundCreate(b, e) {
					return true
				}
			}
		case *events.GuildSoundboardSoundDelete:
			if handler, ok := mod.(core.OnGuildSoundboardSoundDelete); ok {
				if handler.HandleGuildSoundboardSoundDelete(b, e) {
					return true
				}
			}
		case *events.GuildSoundboardSoundUpdate:
			if handler, ok := mod.(core.OnGuildSoundboardSoundUpdate); ok {
				if handler.HandleGuildSoundboardSoundUpdate(b, e) {
					return true
				}
			}
		case *events.GuildSoundboardSoundsUpdate:
			if handler, ok := mod.(core.OnGuildSoundboardSoundsUpdate); ok {
				if handler.HandleGuildSoundboardSoundsUpdate(b, e) {
					return true
				}
			}
		case *events.GuildUnavailable:
			if handler, ok := mod.(core.OnGuildUnavailable); ok {
				if handler.HandleGuildUnavailable(b, e) {
					return true
				}
			}
		case *events.GuildUnban:
			if handler, ok := mod.(core.OnGuildUnban); ok {
				if handler.HandleGuildUnban(b, e) {
					return true
				}
			}
		case *events.GuildUpdate:
			if handler, ok := mod.(core.OnGuildUpdate); ok {
				if handler.HandleGuildUpdate(b, e) {
					return true
				}
			}
		case *events.GuildVoiceChannelEffectSend:
			if handler, ok := mod.(core.OnGuildVoiceChannelEffectSend); ok {
				if handler.HandleGuildVoiceChannelEffectSend(b, e) {
					return true
				}
			}
		case *events.GuildVoiceJoin:
			if handler, ok := mod.(core.OnGuildVoiceJoin); ok {
				if handler.HandleGuildVoiceJoin(b, e) {
					return true
				}
			}
		case *events.GuildVoiceLeave:
			if handler, ok := mod.(core.OnGuildVoiceLeave); ok {
				if handler.HandleGuildVoiceLeave(b, e) {
					return true
				}
			}
		case *events.GuildVoiceMove:
			if handler, ok := mod.(core.OnGuildVoiceMove); ok {
				if handler.HandleGuildVoiceMove(b, e) {
					return true
				}
			}
		case *events.GuildVoiceStateUpdate:
			if handler, ok := mod.(core.OnGuildVoiceStateUpdate); ok {
				if handler.HandleGuildVoiceStateUpdate(b, e) {
					return true
				}
			}
		case *events.IntegrationCreate:
			if handler, ok := mod.(core.OnIntegrationCreate); ok {
				if handler.HandleIntegrationCreate(b, e) {
					return true
				}
			}
		case *events.IntegrationDelete:
			if handler, ok := mod.(core.OnIntegrationDelete); ok {
				if handler.HandleIntegrationDelete(b, e) {
					return true
				}
			}
		case *events.IntegrationUpdate:
			if handler, ok := mod.(core.OnIntegrationUpdate); ok {
				if handler.HandleIntegrationUpdate(b, e) {
					return true
				}
			}
		case *events.InteractionCreate:
			if handler, ok := mod.(core.OnInteractionCreate); ok {
				if handler.HandleInteractionCreate(b, e) {
					return true
				}
			}
		case *events.InviteCreate:
			if handler, ok := mod.(core.OnInviteCreate); ok {
				if handler.HandleInviteCreate(b, e) {
					return true
				}
			}
		case *events.InviteDelete:
			if handler, ok := mod.(core.OnInviteDelete); ok {
				if handler.HandleInviteDelete(b, e) {
					return true
				}
			}
		case *events.MessageCreate:
			if handler, ok := mod.(core.OnMessageCreate); ok {
				if handler.HandleMessageCreate(b, e) {
					return true
				}
			}
		case *events.MessageDelete:
			if handler, ok := mod.(core.OnMessageDelete); ok {
				if handler.HandleMessageDelete(b, e) {
					return true
				}
			}
		case *events.MessagePollVoteAdd:
			if handler, ok := mod.(core.OnMessagePollVoteAdd); ok {
				if handler.HandleMessagePollVoteAdd(b, e) {
					return true
				}
			}
		case *events.MessagePollVoteRemove:
			if handler, ok := mod.(core.OnMessagePollVoteRemove); ok {
				if handler.HandleMessagePollVoteRemove(b, e) {
					return true
				}
			}
		case *events.MessageReactionAdd:
			if handler, ok := mod.(core.OnMessageReactionAdd); ok {
				if handler.HandleMessageReactionAdd(b, e) {
					return true
				}
			}
		case *events.MessageReactionRemove:
			if handler, ok := mod.(core.OnMessageReactionRemove); ok {
				if handler.HandleMessageReactionRemove(b, e) {
					return true
				}
			}
		case *events.MessageReactionRemoveAll:
			if handler, ok := mod.(core.OnMessageReactionRemoveAll); ok {
				if handler.HandleMessageReactionRemoveAll(b, e) {
					return true
				}
			}
		case *events.MessageReactionRemoveEmoji:
			if handler, ok := mod.(core.OnMessageReactionRemoveEmoji); ok {
				if handler.HandleMessageReactionRemoveEmoji(b, e) {
					return true
				}
			}
		case *events.MessageUpdate:
			if handler, ok := mod.(core.OnMessageUpdate); ok {
				if handler.HandleMessageUpdate(b, e) {
					return true
				}
			}
		case *events.ModalSubmitInteractionCreate:
			if handler, ok := mod.(core.OnModalSubmitInteractionCreate); ok {
				if handler.HandleModalSubmitInteractionCreate(b, e) {
					return true
				}
			}
		case *events.PresenceUpdate:
			if handler, ok := mod.(core.OnPresenceUpdate); ok {
				if handler.HandlePresenceUpdate(b, e) {
					return true
				}
			}
		case *events.RoleCreate:
			if handler, ok := mod.(core.OnRoleCreate); ok {
				if handler.HandleRoleCreate(b, e) {
					return true
				}
			}
		case *events.RoleDelete:
			if handler, ok := mod.(core.OnRoleDelete); ok {
				if handler.HandleRoleDelete(b, e) {
					return true
				}
			}
		case *events.RoleUpdate:
			if handler, ok := mod.(core.OnRoleUpdate); ok {
				if handler.HandleRoleUpdate(b, e) {
					return true
				}
			}
		case *events.SoundboardSounds:
			if handler, ok := mod.(core.OnSoundboardSounds); ok {
				if handler.HandleSoundboardSounds(b, e) {
					return true
				}
			}
		case *events.StageInstanceCreate:
			if handler, ok := mod.(core.OnStageInstanceCreate); ok {
				if handler.HandleStageInstanceCreate(b, e) {
					return true
				}
			}
		case *events.StageInstanceDelete:
			if handler, ok := mod.(core.OnStageInstanceDelete); ok {
				if handler.HandleStageInstanceDelete(b, e) {
					return true
				}
			}
		case *events.StageInstanceUpdate:
			if handler, ok := mod.(core.OnStageInstanceUpdate); ok {
				if handler.HandleStageInstanceUpdate(b, e) {
					return true
				}
			}
		case *events.StickerCreate:
			if handler, ok := mod.(core.OnStickerCreate); ok {
				if handler.HandleStickerCreate(b, e) {
					return true
				}
			}
		case *events.StickerDelete:
			if handler, ok := mod.(core.OnStickerDelete); ok {
				if handler.HandleStickerDelete(b, e) {
					return true
				}
			}
		case *events.StickerUpdate:
			if handler, ok := mod.(core.OnStickerUpdate); ok {
				if handler.HandleStickerUpdate(b, e) {
					return true
				}
			}
		case *events.StickersUpdate:
			if handler, ok := mod.(core.OnStickersUpdate); ok {
				if handler.HandleStickersUpdate(b, e) {
					return true
				}
			}
		case *events.SubscriptionCreate:
			if handler, ok := mod.(core.OnSubscriptionCreate); ok {
				if handler.HandleSubscriptionCreate(b, e) {
					return true
				}
			}
		case *events.SubscriptionDelete:
			if handler, ok := mod.(core.OnSubscriptionDelete); ok {
				if handler.HandleSubscriptionDelete(b, e) {
					return true
				}
			}
		case *events.SubscriptionUpdate:
			if handler, ok := mod.(core.OnSubscriptionUpdate); ok {
				if handler.HandleSubscriptionUpdate(b, e) {
					return true
				}
			}
		case *events.ThreadCreate:
			if handler, ok := mod.(core.OnThreadCreate); ok {
				if handler.HandleThreadCreate(b, e) {
					return true
				}
			}
		case *events.ThreadDelete:
			if handler, ok := mod.(core.OnThreadDelete); ok {
				if handler.HandleThreadDelete(b, e) {
					return true
				}
			}
		case *events.ThreadHide:
			if handler, ok := mod.(core.OnThreadHide); ok {
				if handler.HandleThreadHide(b, e) {
					return true
				}
			}
		case *events.ThreadMemberAdd:
			if handler, ok := mod.(core.OnThreadMemberAdd); ok {
				if handler.HandleThreadMemberAdd(b, e) {
					return true
				}
			}
		case *events.ThreadMemberRemove:
			if handler, ok := mod.(core.OnThreadMemberRemove); ok {
				if handler.HandleThreadMemberRemove(b, e) {
					return true
				}
			}
		case *events.ThreadMemberUpdate:
			if handler, ok := mod.(core.OnThreadMemberUpdate); ok {
				if handler.HandleThreadMemberUpdate(b, e) {
					return true
				}
			}
		case *events.ThreadShow:
			if handler, ok := mod.(core.OnThreadShow); ok {
				if handler.HandleThreadShow(b, e) {
					return true
				}
			}
		case *events.ThreadUpdate:
			if handler, ok := mod.(core.OnThreadUpdate); ok {
				if handler.HandleThreadUpdate(b, e) {
					return true
				}
			}
		case *events.UserActivityStart:
			if handler, ok := mod.(core.OnUserActivityStart); ok {
				if handler.HandleUserActivityStart(b, e) {
					return true
				}
			}
		case *events.UserActivityStop:
			if handler, ok := mod.(core.OnUserActivityStop); ok {
				if handler.HandleUserActivityStop(b, e) {
					return true
				}
			}
		case *events.UserActivityUpdate:
			if handler, ok := mod.(core.OnUserActivityUpdate); ok {
				if handler.HandleUserActivityUpdate(b, e) {
					return true
				}
			}
		case *events.UserClientStatusUpdate:
			if handler, ok := mod.(core.OnUserClientStatusUpdate); ok {
				if handler.HandleUserClientStatusUpdate(b, e) {
					return true
				}
			}
		case *events.UserStatusUpdate:
			if handler, ok := mod.(core.OnUserStatusUpdate); ok {
				if handler.HandleUserStatusUpdate(b, e) {
					return true
				}
			}
		case *events.UserTypingStart:
			if handler, ok := mod.(core.OnUserTypingStart); ok {
				if handler.HandleUserTypingStart(b, e) {
					return true
				}
			}
		case *events.UserUpdate:
			if handler, ok := mod.(core.OnUserUpdate); ok {
				if handler.HandleUserUpdate(b, e) {
					return true
				}
			}
		case *events.VoiceServerUpdate:
			if handler, ok := mod.(core.OnVoiceServerUpdate); ok {
				if handler.HandleVoiceServerUpdate(b, e) {
					return true
				}
			}
		case *events.WebhooksUpdate:
			if handler, ok := mod.(core.OnWebhooksUpdate); ok {
				if handler.HandleWebhooksUpdate(b, e) {
					return true
				}
			}
	}
	}

	return false
}

func CheckPerms(b *core.Bot, module core.Module, me discord.Member) bool {
	var missing []string
	botPerms := utils.GetMemberPermissions(b.Client, me)

	if botPerms.Has(discord.PermissionAdministrator) {
		return true
	}

	permissions := module.Permissions()
	for _, p := range permissions {
		if !botPerms.Has(p) {
			missing = append(missing, p.String())
		}
	}

	if len(missing) == 0 {
		return true
	}
	return false
}
