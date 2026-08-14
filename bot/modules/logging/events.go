package logging

import (
	"fmt"
	"onyx/bot/core"
	"onyx/bot/locales"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

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
