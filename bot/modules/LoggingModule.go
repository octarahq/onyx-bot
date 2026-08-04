package modules

import (
	"fmt"
	"onyx/bot/core"
	"onyx/bot/locales"
	"onyx/bot/utils"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"gorm.io/gorm"
)

type State struct {
	Label string
	Value string
}

type event bot.Event

func createONLogMessage(oldState []State, newState []State) []discord.TextDisplayComponent {
	var td []discord.TextDisplayComponent

	for i, o := range oldState {
		n := newState[i]

		td = append(td, discord.NewTextDisplayf("%s : %s -> %s %s", o.Label, o.Value, n.Label, n.Value))
	}

	return td
}
func createNLogMessage(newState []State) []discord.TextDisplayComponent {
	var td []discord.TextDisplayComponent

	for _, n := range newState {
		td = append(td, discord.NewTextDisplayf("%s %s", n.Label, n.Value))
	}

	return td
}

func createMessage(color string, title string, components []discord.ContainerSubComponent) discord.MessageCreate {
	builder := discord.NewContainer(
		discord.NewTextDisplayf("## %s", title),
	).WithAccentColor(utils.ParseStrColor(color))

	for _, c := range components {
		builder = builder.AddComponents(c)
	}

	return discord.NewMessageCreateV2(
		builder,
	)
}

type LoggingMainSettings struct {
	Channel string `json:"channel"`
}

type ModuleLogDefaults struct {
	BasicChannel     string `json:"basic_channel"`
	ImportantChannel string `json:"important_channel"`
}

type ModuleLogSetting struct {
	GuildID    string `gorm:"primaryKey" json:"-"`
	ModuleName string `gorm:"primaryKey" json:"module_name"`
	LogInfo    bool   `gorm:"default:true" json:"log_info"`
	LogErrors  bool   `gorm:"default:true" json:"log_errors"`
}

type LoggingSettings struct {
	GuildID        string              `gorm:"primaryKey" json:"guild_id"`
	Enabled        bool                `gorm:"default:false" json:"enabled"`
	Main           LoggingMainSettings `gorm:"embedded;embeddedPrefix:main_" json:"main"`
	ModuleDefaults ModuleLogDefaults   `gorm:"embedded;embeddedPrefix:moddef_" json:"module_defaults"`
	ModuleConfigs  []ModuleLogSetting  `gorm:"foreignKey:GuildID;references:GuildID" json:"module_configs"`
}

type LoggingModule struct {
	Data LoggingSettings
}

func init() {
	Register(&LoggingModule{})
}

func ptrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func formatStateChange(label string, oldVal any, newVal any, multiline bool) string {
	if multiline {
		return fmt.Sprintf("%s:\n%v\n\n-> %v", label, oldVal, newVal)
	}
	return fmt.Sprintf("%s: %v -> %v", label, oldVal, newVal)
}

func formatState(label string, val any) string {
	return fmt.Sprintf("%s: %v", label, val)
}

const (
	ActionAdd    = "#2ecc71"
	ActionUpdate = "#e4ab17"
	ActionDelete = "#e74c3c"
)

func (m *LoggingModule) Metadata() core.Metadata {
	return core.Metadata{
		Name: "LoggingModule",
		Icon: "files",
		Label: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_LoggingModule").Label
		},
		Description: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_LoggingModule").Description
		},
		Submodules: func(locale discord.Locale) map[string]core.SubmoduleMeta {
			meta := locales.GetMeta(locale, "module_LoggingModule")
			subs := make(map[string]core.SubmoduleMeta)
			for k, v := range meta.Submodules {
				subs[k] = core.SubmoduleMeta{
					Label:       v.Label,
					Description: v.Description,
				}
			}
			return subs
		},
	}
}

func (m *LoggingModule) Priority() int   { return 1 }
func (m *LoggingModule) IsEnabled() bool { return m.Data.Enabled }
func (m *LoggingModule) Permissions() []discord.Permissions {
	return []discord.Permissions{
		discord.PermissionSendMessages,
	}
}

func (m *LoggingModule) Schema() interface{}  { return &LoggingSettings{} }
func (m *LoggingModule) DataPtr() interface{} { return &m.Data }
func (m *LoggingModule) LoadData(db *gorm.DB, guildID string) error {
	m.Data = LoggingSettings{GuildID: guildID}
	return db.FirstOrCreate(&m.Data, LoggingSettings{GuildID: guildID}).Error
}

func (m *LoggingModule) UISchema(locale discord.Locale) core.UISchema {
	meta := locales.GetMeta(locale, "module_LoggingModule")
	_ = meta // Keep it around if needed later, or remove. Let's just keep _ = meta


	schema := core.UISchema{
		SubModules: []core.UISubModule{
			{
				Name:        "main",
				Label:       "Paramètres principaux",
				Description: "Configuration générale du module (logs des évènements du serveur).",
				Components: []core.UIComponent{
					{
						Name:         "channel",
						Label:        "Salon des évènements",
						Description:  "Le salon où les logs du serveur seront envoyés.",
						Type:         core.ComponentTypeChannel,
						Required:     true,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
					},
				},
			},
			{
				Name:        "module_defaults",
				Label:       "Paramètres par défaut",
				Description: "Sélectionnez les salons par défaut pour les logs des modules.",
				Components: []core.UIComponent{
					{
						Name:         "basic_channel",
						Label:        "Salon des Logs Basiques",
						Description:  "Le salon où les logs d'information seront envoyés.",
						Type:         core.ComponentTypeChannel,
						Required:     true,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
					},
					{
						Name:         "important_channel",
						Label:        "Salon des Logs Importants",
						Description:  "Le salon où les logs importants ou d'erreurs seront envoyés.",
						Type:         core.ComponentTypeChannel,
						Required:     true,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
					},
				},
			},
		},
	}

	var moduleGridOptions []core.UISelectOption
	for _, m := range RegisteredModules {
		modMeta := m.Metadata()
		if modMeta.Loggable {
			moduleGridOptions = append(moduleGridOptions, core.UISelectOption{
				Label: modMeta.Label(locale),
				Value: modMeta.Name,
			})
		}
	}

	schema.SubModules = append(schema.SubModules, core.UISubModule{
		Name:        "",
		Label:       "Configuration des modules",
		Description: "Activez ou désactivez les logs pour chaque module.",
		FullWidth:   true,
		Components: []core.UIComponent{
			{
				Name:    "module_configs",
				Type:    core.ComponentTypeModuleGrid,
				Options: moduleGridOptions,
			},
		},
	})

	return schema
}

func (m *LoggingModule) sendLog(b *core.Bot, color, title string, components []discord.ContainerSubComponent) {
	if !m.Data.Enabled || len(components) == 0 {
		return
	}
	if m.Data.Main.Channel == "" {
		return
	}
	b.SendMessage(m.Data.Main.Channel, createMessage(color, title, components))
}

func (m *LoggingModule) LogInfo(b *core.Bot, gid string, moduleName string, title string, logs []string) {
	var settings LoggingSettings
	if err := b.DB.GormDB.Preload("ModuleConfigs").First(&settings, "guild_id = ?", gid).Error; err != nil {
		return
	}
	if !settings.Enabled || settings.ModuleDefaults.BasicChannel == "" {
		return
	}

	canLog := false
	for _, config := range settings.ModuleConfigs {
		if config.ModuleName == moduleName {
			if config.LogInfo {
				canLog = true
			}
			break
		}
	}
	if !canLog {
		return
	}

	var comps []discord.ContainerSubComponent
	for _, c := range logs {
		comps = append(comps, discord.NewTextDisplayf("> %s", c))
	}
	msg := createMessage("#e4ab17", title, comps)
	b.SendMessage(settings.ModuleDefaults.BasicChannel, msg)
}

func (m *LoggingModule) LogImportant(b *core.Bot, gid string, moduleName string, title string, logs []string) {
	var settings LoggingSettings
	if err := b.DB.GormDB.Preload("ModuleConfigs").First(&settings, "guild_id = ?", gid).Error; err != nil {
		return
	}
	if !settings.Enabled || settings.ModuleDefaults.ImportantChannel == "" {
		return
	}

	canLog := false
	for _, config := range settings.ModuleConfigs {
		if config.ModuleName == moduleName {
			if config.LogErrors {
				canLog = true
			}
			break
		}
	}
	if !canLog {
		return
	}

	var comps []discord.ContainerSubComponent
	for _, c := range logs {
		comps = append(comps, discord.NewTextDisplayf("> %s", c))
	}
	msg := createMessage("#e74c3c", title, comps)
	b.SendMessage(settings.ModuleDefaults.ImportantChannel, msg)
}

func (m *LoggingModule) HandleGuildChannelCreate(b *core.Bot, e *events.GuildChannelCreate) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).ChannelCreate
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Name, e.Channel.Name())))
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Type, e.Channel.Type())))
	m.sendLog(b, ActionAdd, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleGuildChannelUpdate(b *core.Bot, e *events.GuildChannelUpdate) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).ChannelUpdate
	var comps []discord.ContainerSubComponent

	if e.OldChannel.Name() != e.Channel.Name() {
		comps = append(comps, discord.NewTextDisplay(formatStateChange(t.States.Name, e.OldChannel.Name(), e.Channel.Name(), false)))
	}

	m.sendLog(b, ActionUpdate, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleGuildChannelDelete(b *core.Bot, e *events.GuildChannelDelete) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).ChannelDelete
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Name, e.Channel.Name())))
	m.sendLog(b, ActionDelete, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleRoleCreate(b *core.Bot, e *events.RoleCreate) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).RoleCreate
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Name, e.Role.Name)))
	m.sendLog(b, ActionAdd, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleRoleUpdate(b *core.Bot, e *events.RoleUpdate) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).RoleUpdate
	var comps []discord.ContainerSubComponent

	if e.OldRole.Name != e.Role.Name {
		comps = append(comps, discord.NewTextDisplay(formatStateChange(t.States.Name, e.OldRole.Name, e.Role.Name, false)))
	}
	if e.OldRole.Color != e.Role.Color {
		comps = append(comps, discord.NewTextDisplay(formatStateChange(t.States.Color, e.OldRole.Color, e.Role.Color, false)))
	}

	m.sendLog(b, ActionUpdate, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleRoleDelete(b *core.Bot, e *events.RoleDelete) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).RoleDelete
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Name, e.Role.Name)))
	m.sendLog(b, ActionDelete, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleGuildMemberJoin(b *core.Bot, e *events.GuildMemberJoin) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).MemberJoin
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.User, e.Member.User.Tag())))
	m.sendLog(b, ActionAdd, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleGuildMemberUpdate(b *core.Bot, e *events.GuildMemberUpdate) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).MemberUpdate
	var comps []discord.ContainerSubComponent

	oldNick := ptrToStr(e.OldMember.Nick)
	newNick := ptrToStr(e.Member.Nick)
	if oldNick != newNick {
		comps = append(comps, discord.NewTextDisplay(formatStateChange(t.States.Nickname, oldNick, newNick, false)))
	}

	m.sendLog(b, ActionUpdate, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleGuildMemberLeave(b *core.Bot, e *events.GuildMemberLeave) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).MemberLeave
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.User, e.User.Tag())))
	m.sendLog(b, ActionDelete, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleGuildUpdate(b *core.Bot, e *events.GuildUpdate) bool {
	if m.Data.Enabled {
		var components []discord.ContainerSubComponent

		guild, ok := e.Client().Caches.Guild(e.GuildID)
		if !ok {
			return false
		}
		trad := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale))
		t := trad.GuildUpdate
		states := t.States
		title := t.Title

		if e.OldGuild.Name != e.Guild.Name {
			components = append(components, discord.NewTextDisplay(formatStateChange(states.Name, e.OldGuild.Name, e.Guild.Name, false)))
		}

		oldDesc := ptrToStr(e.OldGuild.Description)
		newDesc := ptrToStr(e.Guild.Description)
		if oldDesc != newDesc {
			components = append(components, discord.NewTextDisplay(formatStateChange(states.Description, oldDesc, newDesc, true)))
		}

		if e.OldGuild.VerificationLevel != e.Guild.VerificationLevel {
			components = append(components, discord.NewTextDisplay(formatStateChange(states.VerificationLevel, e.OldGuild.VerificationLevel, e.Guild.VerificationLevel, false)))
		}

		oldIcon := ptrToStr(e.OldGuild.IconURL())
		newIcon := ptrToStr(e.Guild.IconURL())
		if oldIcon != newIcon {
			components = append(components, discord.NewTextDisplay(states.Icon))
			if newIcon != "" {
				components = append(components, discord.NewMediaGallery(discord.MediaGalleryItem{Media: discord.UnfurledMediaItem{URL: newIcon}}))
			}
		}

		oldBanner := ptrToStr(e.OldGuild.BannerURL())
		newBanner := ptrToStr(e.Guild.BannerURL())
		if oldBanner != newBanner {
			components = append(components, discord.NewTextDisplay(states.Banner))
			if newBanner != "" {
				components = append(components, discord.NewMediaGallery(discord.MediaGalleryItem{Media: discord.UnfurledMediaItem{URL: newBanner}}))
			}
		}

		if len(components) > 0 {
			b.SendMessage(
				m.Data.Main.Channel,
				createMessage(
					ActionUpdate,
					title,
					components,
				),
			)
		}
	}
	return false
}

func (m *LoggingModule) HandleGuildBan(b *core.Bot, e *events.GuildBan) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).BanAdd
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.User, e.User.Tag())))
	m.sendLog(b, ActionDelete, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleGuildUnban(b *core.Bot, e *events.GuildUnban) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).BanRemove
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.User, e.User.Tag())))
	m.sendLog(b, ActionUpdate, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleEmojiCreate(b *core.Bot, e *events.EmojiCreate) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).EmojiCreate
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Name, e.Emoji.Name)))
	m.sendLog(b, ActionAdd, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleEmojiUpdate(b *core.Bot, e *events.EmojiUpdate) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).EmojiUpdate
	var comps []discord.ContainerSubComponent
	if e.OldEmoji.Name != e.Emoji.Name {
		comps = append(comps, discord.NewTextDisplay(formatStateChange(t.States.Name, e.OldEmoji.Name, e.Emoji.Name, false)))
	}
	m.sendLog(b, ActionUpdate, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleEmojiDelete(b *core.Bot, e *events.EmojiDelete) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).EmojiDelete
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Name, e.Emoji.Name)))
	m.sendLog(b, ActionDelete, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleStickerCreate(b *core.Bot, e *events.StickerCreate) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).StickerCreate
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Name, e.Sticker.Name)))
	m.sendLog(b, ActionAdd, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleStickerUpdate(b *core.Bot, e *events.StickerUpdate) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).StickerUpdate
	var comps []discord.ContainerSubComponent
	if e.OldSticker.Name != e.Sticker.Name {
		comps = append(comps, discord.NewTextDisplay(formatStateChange(t.States.Name, e.OldSticker.Name, e.Sticker.Name, false)))
	}
	m.sendLog(b, ActionUpdate, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleStickerDelete(b *core.Bot, e *events.StickerDelete) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).StickerDelete
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Name, e.Sticker.Name)))
	m.sendLog(b, ActionDelete, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleInviteCreate(b *core.Bot, e *events.InviteCreate) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(*e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).InviteCreate
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Code, e.Code)))
	m.sendLog(b, ActionAdd, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleInviteDelete(b *core.Bot, e *events.InviteDelete) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(*e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).InviteDelete
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Code, e.Code)))
	m.sendLog(b, ActionDelete, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleThreadCreate(b *core.Bot, e *events.ThreadCreate) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).ThreadCreate
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Name, e.Thread.Name())))
	m.sendLog(b, ActionAdd, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleThreadUpdate(b *core.Bot, e *events.ThreadUpdate) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).ThreadUpdate
	var comps []discord.ContainerSubComponent
	if e.OldThread.Name() != e.Thread.Name() {
		comps = append(comps, discord.NewTextDisplay(formatStateChange(t.States.Name, e.OldThread.Name(), e.Thread.Name(), false)))
	}
	m.sendLog(b, ActionUpdate, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleThreadDelete(b *core.Bot, e *events.ThreadDelete) bool {
	if !m.Data.Enabled {
		return false
	}
	guild, ok := e.Client().Caches.Guild(e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).ThreadDelete
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Name, e.ThreadID)))
	m.sendLog(b, ActionDelete, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleMessageUpdate(b *core.Bot, e *events.MessageUpdate) bool {
	if !m.Data.Enabled || e.GuildID == nil {
		return false
	}
	if e.Message.Author.Bot || e.OldMessage.ID == 0 {
		return false
	}
	if e.OldMessage.Content == e.Message.Content {
		return false
	}
	guild, ok := e.Client().Caches.Guild(*e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).MessageUpdate
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Author, e.Message.Author.Tag())))
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Channel, fmt.Sprintf("<#%s>", e.Message.ChannelID.String()))))
	oldContent := e.OldMessage.Content
	if oldContent == "" {
		oldContent = "*No Content*"
	}
	newContent := e.Message.Content
	if newContent == "" {
		newContent = "*No Content*"
	}
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.OldContent, "\n"+oldContent)))
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.NewContent, "\n"+newContent)))
	m.sendLog(b, ActionUpdate, t.Title, comps)
	return false
}

func (m *LoggingModule) HandleMessageDelete(b *core.Bot, e *events.MessageDelete) bool {
	if !m.Data.Enabled || e.GuildID == nil {
		return false
	}
	if e.Message.Author.ID == 0 || e.Message.Author.Bot {
		return false
	}
	guild, ok := e.Client().Caches.Guild(*e.GuildID)
	if !ok {
		return false
	}
	t := locales.GetModule_LoggingModule(discord.Locale(guild.PreferredLocale)).MessageDelete
	var comps []discord.ContainerSubComponent
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Author, e.Message.Author.Tag())))
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Channel, fmt.Sprintf("<#%s>", e.ChannelID.String()))))
	content := e.Message.Content
	if content == "" {
		content = "*No content*"
	}
	comps = append(comps, discord.NewTextDisplay(formatState(t.States.Content, "\n"+content)))
	m.sendLog(b, ActionDelete, t.Title, comps)
	return false
}
