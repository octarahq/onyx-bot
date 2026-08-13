package core

import (
	"sort"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"gorm.io/gorm"
)

var RegisteredModules []Module

func Register(m Module) {
	RegisteredModules = append(RegisteredModules, m)
	sort.SliceStable(RegisteredModules, func(i, j int) bool {
		return RegisteredModules[i].Priority() > RegisteredModules[j].Priority()
	})
}

type SubmoduleMeta struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

type Metadata struct {
	Name        string
	Icon        string
	Label       func(locale discord.Locale) string
	Description func(locale discord.Locale) string
	Submodules  func(locale discord.Locale) map[string]SubmoduleMeta
	Loggable    bool
}

type Module interface {
	Metadata() Metadata
	Priority() int
	IsEnabled() bool
	Permissions() []discord.Permissions
}

type DatabaseAware interface {
	Schema() interface{}
	LoadData(db *gorm.DB, guildID string) error
	DataPtr() interface{}
}

type ModuleLogger interface {
	LogInfo(b *Bot, gid string, moduleName string, title string, logs []string)
	LogImportant(b *Bot, gid string, moduleName string, title string, logs []string)
}

type ModuleCommand interface {
	Command() *discord.SlashCommandCreate
	HandleCommand(b *Bot, event *events.ApplicationCommandInteractionCreate) bool
}

type ModuleButtonHandler interface {
	HandleButton(b *Bot, event *events.ComponentInteractionCreate, action string, args []string) bool
}

type ModuleSelectMenuHandler interface {
	HandleSelectMenu(b *Bot, event *events.ComponentInteractionCreate, action string, args []string) bool
}

type ModuleModalHandler interface {
	HandleModal(b *Bot, event *events.ModalSubmitInteractionCreate, action string, args []string) bool
}

type OnApplicationCommandInteractionCreate interface {
	HandleApplicationCommandInteractionCreate(b *Bot, event *events.ApplicationCommandInteractionCreate) bool
}

type OnAutoModerationActionExecution interface {
	HandleAutoModerationActionExecution(b *Bot, event *events.AutoModerationActionExecution) bool
}

type OnAutoModerationRuleCreate interface {
	HandleAutoModerationRuleCreate(b *Bot, event *events.AutoModerationRuleCreate) bool
}

type OnAutoModerationRuleDelete interface {
	HandleAutoModerationRuleDelete(b *Bot, event *events.AutoModerationRuleDelete) bool
}

type OnAutoModerationRuleUpdate interface {
	HandleAutoModerationRuleUpdate(b *Bot, event *events.AutoModerationRuleUpdate) bool
}

type OnAutocompleteInteractionCreate interface {
	HandleAutocompleteInteractionCreate(b *Bot, event *events.AutocompleteInteractionCreate) bool
}

type OnComponentInteractionCreate interface {
	HandleComponentInteractionCreate(b *Bot, event *events.ComponentInteractionCreate) bool
}

type OnDMChannelPinsUpdate interface {
	HandleDMChannelPinsUpdate(b *Bot, event *events.DMChannelPinsUpdate) bool
}

type OnDMMessageCreate interface {
	HandleDMMessageCreate(b *Bot, event *events.DMMessageCreate) bool
}

type OnDMMessageDelete interface {
	HandleDMMessageDelete(b *Bot, event *events.DMMessageDelete) bool
}

type OnDMMessagePollVoteAdd interface {
	HandleDMMessagePollVoteAdd(b *Bot, event *events.DMMessagePollVoteAdd) bool
}

type OnDMMessagePollVoteRemove interface {
	HandleDMMessagePollVoteRemove(b *Bot, event *events.DMMessagePollVoteRemove) bool
}

type OnDMMessageReactionAdd interface {
	HandleDMMessageReactionAdd(b *Bot, event *events.DMMessageReactionAdd) bool
}

type OnDMMessageReactionRemove interface {
	HandleDMMessageReactionRemove(b *Bot, event *events.DMMessageReactionRemove) bool
}

type OnDMMessageReactionRemoveAll interface {
	HandleDMMessageReactionRemoveAll(b *Bot, event *events.DMMessageReactionRemoveAll) bool
}

type OnDMMessageReactionRemoveEmoji interface {
	HandleDMMessageReactionRemoveEmoji(b *Bot, event *events.DMMessageReactionRemoveEmoji) bool
}

type OnDMMessageUpdate interface {
	HandleDMMessageUpdate(b *Bot, event *events.DMMessageUpdate) bool
}

type OnDMUserTypingStart interface {
	HandleDMUserTypingStart(b *Bot, event *events.DMUserTypingStart) bool
}

type OnEmojiCreate interface {
	HandleEmojiCreate(b *Bot, event *events.EmojiCreate) bool
}

type OnEmojiDelete interface {
	HandleEmojiDelete(b *Bot, event *events.EmojiDelete) bool
}

type OnEmojiUpdate interface {
	HandleEmojiUpdate(b *Bot, event *events.EmojiUpdate) bool
}

type OnEmojisUpdate interface {
	HandleEmojisUpdate(b *Bot, event *events.EmojisUpdate) bool
}

type OnEntitlementCreate interface {
	HandleEntitlementCreate(b *Bot, event *events.EntitlementCreate) bool
}

type OnEntitlementDelete interface {
	HandleEntitlementDelete(b *Bot, event *events.EntitlementDelete) bool
}

type OnEntitlementUpdate interface {
	HandleEntitlementUpdate(b *Bot, event *events.EntitlementUpdate) bool
}

type OnGuildApplicationCommandPermissionsUpdate interface {
	HandleGuildApplicationCommandPermissionsUpdate(b *Bot, event *events.GuildApplicationCommandPermissionsUpdate) bool
}

type OnGuildAuditLogEntryCreate interface {
	HandleGuildAuditLogEntryCreate(b *Bot, event *events.GuildAuditLogEntryCreate) bool
}

type OnGuildAvailable interface {
	HandleGuildAvailable(b *Bot, event *events.GuildAvailable) bool
}

type OnGuildBan interface {
	HandleGuildBan(b *Bot, event *events.GuildBan) bool
}

type OnGuildChannelCreate interface {
	HandleGuildChannelCreate(b *Bot, event *events.GuildChannelCreate) bool
}

type OnGuildChannelDelete interface {
	HandleGuildChannelDelete(b *Bot, event *events.GuildChannelDelete) bool
}

type OnGuildChannelPinsUpdate interface {
	HandleGuildChannelPinsUpdate(b *Bot, event *events.GuildChannelPinsUpdate) bool
}

type OnGuildChannelUpdate interface {
	HandleGuildChannelUpdate(b *Bot, event *events.GuildChannelUpdate) bool
}

type OnGuildIntegrationsUpdate interface {
	HandleGuildIntegrationsUpdate(b *Bot, event *events.GuildIntegrationsUpdate) bool
}

type OnGuildJoin interface {
	HandleGuildJoin(b *Bot, event *events.GuildJoin) bool
}

type OnGuildLeave interface {
	HandleGuildLeave(b *Bot, event *events.GuildLeave) bool
}

type OnGuildMemberJoin interface {
	HandleGuildMemberJoin(b *Bot, event *events.GuildMemberJoin) bool
}

type OnGuildMemberLeave interface {
	HandleGuildMemberLeave(b *Bot, event *events.GuildMemberLeave) bool
}

type OnGuildMemberTypingStart interface {
	HandleGuildMemberTypingStart(b *Bot, event *events.GuildMemberTypingStart) bool
}

type OnGuildMemberUpdate interface {
	HandleGuildMemberUpdate(b *Bot, event *events.GuildMemberUpdate) bool
}

type OnGuildMessageCreate interface {
	HandleGuildMessageCreate(b *Bot, event *events.GuildMessageCreate) bool
}

type OnGuildMessageDelete interface {
	HandleGuildMessageDelete(b *Bot, event *events.GuildMessageDelete) bool
}

type OnGuildMessagePollVoteAdd interface {
	HandleGuildMessagePollVoteAdd(b *Bot, event *events.GuildMessagePollVoteAdd) bool
}

type OnGuildMessagePollVoteRemove interface {
	HandleGuildMessagePollVoteRemove(b *Bot, event *events.GuildMessagePollVoteRemove) bool
}

type OnGuildMessageReactionAdd interface {
	HandleGuildMessageReactionAdd(b *Bot, event *events.GuildMessageReactionAdd) bool
}

type OnGuildMessageReactionRemove interface {
	HandleGuildMessageReactionRemove(b *Bot, event *events.GuildMessageReactionRemove) bool
}

type OnGuildMessageReactionRemoveAll interface {
	HandleGuildMessageReactionRemoveAll(b *Bot, event *events.GuildMessageReactionRemoveAll) bool
}

type OnGuildMessageReactionRemoveEmoji interface {
	HandleGuildMessageReactionRemoveEmoji(b *Bot, event *events.GuildMessageReactionRemoveEmoji) bool
}

type OnGuildMessageUpdate interface {
	HandleGuildMessageUpdate(b *Bot, event *events.GuildMessageUpdate) bool
}

type OnGuildScheduledEventCreate interface {
	HandleGuildScheduledEventCreate(b *Bot, event *events.GuildScheduledEventCreate) bool
}

type OnGuildScheduledEventDelete interface {
	HandleGuildScheduledEventDelete(b *Bot, event *events.GuildScheduledEventDelete) bool
}

type OnGuildScheduledEventUpdate interface {
	HandleGuildScheduledEventUpdate(b *Bot, event *events.GuildScheduledEventUpdate) bool
}

type OnGuildScheduledEventUserAdd interface {
	HandleGuildScheduledEventUserAdd(b *Bot, event *events.GuildScheduledEventUserAdd) bool
}

type OnGuildScheduledEventUserRemove interface {
	HandleGuildScheduledEventUserRemove(b *Bot, event *events.GuildScheduledEventUserRemove) bool
}

type OnGuildSoundboardSoundCreate interface {
	HandleGuildSoundboardSoundCreate(b *Bot, event *events.GuildSoundboardSoundCreate) bool
}

type OnGuildSoundboardSoundDelete interface {
	HandleGuildSoundboardSoundDelete(b *Bot, event *events.GuildSoundboardSoundDelete) bool
}

type OnGuildSoundboardSoundUpdate interface {
	HandleGuildSoundboardSoundUpdate(b *Bot, event *events.GuildSoundboardSoundUpdate) bool
}

type OnGuildSoundboardSoundsUpdate interface {
	HandleGuildSoundboardSoundsUpdate(b *Bot, event *events.GuildSoundboardSoundsUpdate) bool
}

type OnGuildUnavailable interface {
	HandleGuildUnavailable(b *Bot, event *events.GuildUnavailable) bool
}

type OnGuildUnban interface {
	HandleGuildUnban(b *Bot, event *events.GuildUnban) bool
}

type OnGuildUpdate interface {
	HandleGuildUpdate(b *Bot, event *events.GuildUpdate) bool
}

type OnGuildVoiceChannelEffectSend interface {
	HandleGuildVoiceChannelEffectSend(b *Bot, event *events.GuildVoiceChannelEffectSend) bool
}

type OnGuildVoiceJoin interface {
	HandleGuildVoiceJoin(b *Bot, event *events.GuildVoiceJoin) bool
}

type OnGuildVoiceLeave interface {
	HandleGuildVoiceLeave(b *Bot, event *events.GuildVoiceLeave) bool
}

type OnGuildVoiceMove interface {
	HandleGuildVoiceMove(b *Bot, event *events.GuildVoiceMove) bool
}

type OnGuildVoiceStateUpdate interface {
	HandleGuildVoiceStateUpdate(b *Bot, event *events.GuildVoiceStateUpdate) bool
}

type OnIntegrationCreate interface {
	HandleIntegrationCreate(b *Bot, event *events.IntegrationCreate) bool
}

type OnIntegrationDelete interface {
	HandleIntegrationDelete(b *Bot, event *events.IntegrationDelete) bool
}

type OnIntegrationUpdate interface {
	HandleIntegrationUpdate(b *Bot, event *events.IntegrationUpdate) bool
}

type OnInteractionCreate interface {
	HandleInteractionCreate(b *Bot, event *events.InteractionCreate) bool
}

type OnInviteCreate interface {
	HandleInviteCreate(b *Bot, event *events.InviteCreate) bool
}

type OnInviteDelete interface {
	HandleInviteDelete(b *Bot, event *events.InviteDelete) bool
}

type OnMessageCreate interface {
	HandleMessageCreate(b *Bot, event *events.MessageCreate) bool
}

type OnMessageDelete interface {
	HandleMessageDelete(b *Bot, event *events.MessageDelete) bool
}

type OnMessagePollVoteAdd interface {
	HandleMessagePollVoteAdd(b *Bot, event *events.MessagePollVoteAdd) bool
}

type OnMessagePollVoteRemove interface {
	HandleMessagePollVoteRemove(b *Bot, event *events.MessagePollVoteRemove) bool
}

type OnMessageReactionAdd interface {
	HandleMessageReactionAdd(b *Bot, event *events.MessageReactionAdd) bool
}

type OnMessageReactionRemove interface {
	HandleMessageReactionRemove(b *Bot, event *events.MessageReactionRemove) bool
}

type OnMessageReactionRemoveAll interface {
	HandleMessageReactionRemoveAll(b *Bot, event *events.MessageReactionRemoveAll) bool
}

type OnMessageReactionRemoveEmoji interface {
	HandleMessageReactionRemoveEmoji(b *Bot, event *events.MessageReactionRemoveEmoji) bool
}

type OnMessageUpdate interface {
	HandleMessageUpdate(b *Bot, event *events.MessageUpdate) bool
}

type OnModalSubmitInteractionCreate interface {
	HandleModalSubmitInteractionCreate(b *Bot, event *events.ModalSubmitInteractionCreate) bool
}

type OnPresenceUpdate interface {
	HandlePresenceUpdate(b *Bot, event *events.PresenceUpdate) bool
}

type OnReady interface {
	HandleReady(b *Bot, event *events.Ready) bool
}

type OnRoleCreate interface {
	HandleRoleCreate(b *Bot, event *events.RoleCreate) bool
}

type OnRoleDelete interface {
	HandleRoleDelete(b *Bot, event *events.RoleDelete) bool
}

type OnRoleUpdate interface {
	HandleRoleUpdate(b *Bot, event *events.RoleUpdate) bool
}

type OnSoundboardSounds interface {
	HandleSoundboardSounds(b *Bot, event *events.SoundboardSounds) bool
}

type OnStageInstanceCreate interface {
	HandleStageInstanceCreate(b *Bot, event *events.StageInstanceCreate) bool
}

type OnStageInstanceDelete interface {
	HandleStageInstanceDelete(b *Bot, event *events.StageInstanceDelete) bool
}

type OnStageInstanceUpdate interface {
	HandleStageInstanceUpdate(b *Bot, event *events.StageInstanceUpdate) bool
}

type OnStickerCreate interface {
	HandleStickerCreate(b *Bot, event *events.StickerCreate) bool
}

type OnStickerDelete interface {
	HandleStickerDelete(b *Bot, event *events.StickerDelete) bool
}

type OnStickerUpdate interface {
	HandleStickerUpdate(b *Bot, event *events.StickerUpdate) bool
}

type OnStickersUpdate interface {
	HandleStickersUpdate(b *Bot, event *events.StickersUpdate) bool
}

type OnSubscriptionCreate interface {
	HandleSubscriptionCreate(b *Bot, event *events.SubscriptionCreate) bool
}

type OnSubscriptionDelete interface {
	HandleSubscriptionDelete(b *Bot, event *events.SubscriptionDelete) bool
}

type OnSubscriptionUpdate interface {
	HandleSubscriptionUpdate(b *Bot, event *events.SubscriptionUpdate) bool
}

type OnThreadCreate interface {
	HandleThreadCreate(b *Bot, event *events.ThreadCreate) bool
}

type OnThreadDelete interface {
	HandleThreadDelete(b *Bot, event *events.ThreadDelete) bool
}

type OnThreadHide interface {
	HandleThreadHide(b *Bot, event *events.ThreadHide) bool
}

type OnThreadMemberAdd interface {
	HandleThreadMemberAdd(b *Bot, event *events.ThreadMemberAdd) bool
}

type OnThreadMemberRemove interface {
	HandleThreadMemberRemove(b *Bot, event *events.ThreadMemberRemove) bool
}

type OnThreadMemberUpdate interface {
	HandleThreadMemberUpdate(b *Bot, event *events.ThreadMemberUpdate) bool
}

type OnThreadShow interface {
	HandleThreadShow(b *Bot, event *events.ThreadShow) bool
}

type OnThreadUpdate interface {
	HandleThreadUpdate(b *Bot, event *events.ThreadUpdate) bool
}

type OnUserActivityStart interface {
	HandleUserActivityStart(b *Bot, event *events.UserActivityStart) bool
}

type OnUserActivityStop interface {
	HandleUserActivityStop(b *Bot, event *events.UserActivityStop) bool
}

type OnUserActivityUpdate interface {
	HandleUserActivityUpdate(b *Bot, event *events.UserActivityUpdate) bool
}

type OnUserClientStatusUpdate interface {
	HandleUserClientStatusUpdate(b *Bot, event *events.UserClientStatusUpdate) bool
}

type OnUserStatusUpdate interface {
	HandleUserStatusUpdate(b *Bot, event *events.UserStatusUpdate) bool
}

type OnUserTypingStart interface {
	HandleUserTypingStart(b *Bot, event *events.UserTypingStart) bool
}

type OnUserUpdate interface {
	HandleUserUpdate(b *Bot, event *events.UserUpdate) bool
}

type OnVoiceServerUpdate interface {
	HandleVoiceServerUpdate(b *Bot, event *events.VoiceServerUpdate) bool
}

type OnWebhooksUpdate interface {
	HandleWebhooksUpdate(b *Bot, event *events.WebhooksUpdate) bool
}
