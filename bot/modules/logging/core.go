package logging

import (
	"fmt"
	"onyx/bot/core"
	"onyx/bot/locales"
		"onyx/bot/utils"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
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

	return discord.NewMessageCreateV2(builder)
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
	LogInfo    *bool  `gorm:"default:true" json:"log_info"`
	LogErrors  *bool  `gorm:"default:true" json:"log_errors"`
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
	core.Register(&LoggingModule{})
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

func (m *LoggingModule) Schema() interface{}  { return []interface{}{&LoggingSettings{}, &ModuleLogSetting{}} }
func (m *LoggingModule) DataPtr() interface{} { return &m.Data }
func (m *LoggingModule) LoadData(db *gorm.DB, guildID string) error {
	m.Data = LoggingSettings{GuildID: guildID}
	return db.Preload("ModuleConfigs").FirstOrCreate(&m.Data, LoggingSettings{GuildID: guildID}).Error
}

func (m *LoggingModule) UISchema(locale discord.Locale) core.UISchema {
	meta := locales.GetMeta(locale, "module_LoggingModule")

	schema := core.UISchema{
		SubModules: []core.UISubModule{
			{
				Name:        "main",
				Label:       meta.Submodules["main"].Label,
				Description: meta.Submodules["main"].Description,
				Components: []core.UIComponent{
					{
						Name:         "channel",
						Label:        meta.Submodules["main"].Options["channel"].Label,
						Description:  meta.Submodules["main"].Options["channel"].Description,
						Type:         core.ComponentTypeChannel,
						Required:     true,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
					},
				},
			},
			{
				Name:        "module_defaults",
				Label:       meta.Submodules["module_defaults"].Label,
				Description: meta.Submodules["module_defaults"].Description,
				Components: []core.UIComponent{
					{
						Name:         "basic_channel",
						Label:        meta.Submodules["module_defaults"].Options["basic_channel"].Label,
						Description:  meta.Submodules["module_defaults"].Options["basic_channel"].Description,
						Type:         core.ComponentTypeChannel,
						Required:     true,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
					},
					{
						Name:         "important_channel",
						Label:        meta.Submodules["module_defaults"].Options["important_channel"].Label,
						Description:  meta.Submodules["module_defaults"].Options["important_channel"].Description,
						Type:         core.ComponentTypeChannel,
						Required:     true,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
					},
				},
			},
		},
	}

	var moduleGridOptions []core.UISelectOption
	for _, mod := range core.RegisteredModules {
		modMeta := mod.Metadata()
		if modMeta.Loggable {
			moduleGridOptions = append(moduleGridOptions, core.UISelectOption{
				Label: modMeta.Label(locale),
				Value: modMeta.Name,
			})
		}
	}

	schema.SubModules = append(schema.SubModules, core.UISubModule{
		Name:        "",
		Label:       meta.Submodules["module_configs"].Label,
		Description: meta.Submodules["module_configs"].Description,
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

	canLog := true
	for _, config := range settings.ModuleConfigs {
		if config.ModuleName == moduleName {
			if config.LogInfo != nil {
				canLog = *config.LogInfo
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

	canLog := true
	for _, config := range settings.ModuleConfigs {
		if config.ModuleName == moduleName {
			if config.LogErrors != nil {
				canLog = *config.LogErrors
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
