package handlers

import (
	"onyx/bot/locales"

	"github.com/disgoorg/disgo/discord"
)

func localizeOptions(options []discord.ApplicationCommandOption, locale discord.Locale, meta locales.CommandMeta) []discord.ApplicationCommandOption {
	var newOptions []discord.ApplicationCommandOption

	for _, opt := range options {
		optMeta, exists := meta.Options[opt.OptionName()]
		if !exists {
			newOptions = append(newOptions, opt)
			continue
		}

		switch o := opt.(type) {
		case discord.ApplicationCommandOptionSubCommand:
			if o.NameLocalizations == nil {
				o.NameLocalizations = make(map[discord.Locale]string)
			}
			if o.DescriptionLocalizations == nil {
				o.DescriptionLocalizations = make(map[discord.Locale]string)
			}
			if optMeta.Name != "" {
				o.NameLocalizations[locale] = optMeta.Name
			}
			if optMeta.Description != "" {
				o.DescriptionLocalizations[locale] = optMeta.Description
			}
			if len(optMeta.Options) > 0 && len(o.Options) > 0 {
				o.Options = localizeOptions(o.Options, locale, optMeta)
			}
			newOptions = append(newOptions, o)

		case discord.ApplicationCommandOptionSubCommandGroup:
			if o.NameLocalizations == nil {
				o.NameLocalizations = make(map[discord.Locale]string)
			}
			if o.DescriptionLocalizations == nil {
				o.DescriptionLocalizations = make(map[discord.Locale]string)
			}
			if optMeta.Name != "" {
				o.NameLocalizations[locale] = optMeta.Name
			}
			if optMeta.Description != "" {
				o.DescriptionLocalizations[locale] = optMeta.Description
			}
			if len(optMeta.Options) > 0 && len(o.Options) > 0 {
				var subCmds []discord.ApplicationCommandOption
				for _, subOpt := range o.Options {
					subCmds = append(subCmds, subOpt)
				}
				subCmds = localizeOptions(subCmds, locale, optMeta)
				var newSubCmds []discord.ApplicationCommandOptionSubCommand
				for _, subOpt := range subCmds {
					newSubCmds = append(newSubCmds, subOpt.(discord.ApplicationCommandOptionSubCommand))
				}
				o.Options = newSubCmds
			}
			newOptions = append(newOptions, o)

		case discord.ApplicationCommandOptionString:
			if o.NameLocalizations == nil {
				o.NameLocalizations = make(map[discord.Locale]string)
			}
			if o.DescriptionLocalizations == nil {
				o.DescriptionLocalizations = make(map[discord.Locale]string)
			}
			if optMeta.Name != "" {
				o.NameLocalizations[locale] = optMeta.Name
			}
			if optMeta.Description != "" {
				o.DescriptionLocalizations[locale] = optMeta.Description
			}
			newOptions = append(newOptions, o)

		case discord.ApplicationCommandOptionInt:
			if o.NameLocalizations == nil {
				o.NameLocalizations = make(map[discord.Locale]string)
			}
			if o.DescriptionLocalizations == nil {
				o.DescriptionLocalizations = make(map[discord.Locale]string)
			}
			if optMeta.Name != "" {
				o.NameLocalizations[locale] = optMeta.Name
			}
			if optMeta.Description != "" {
				o.DescriptionLocalizations[locale] = optMeta.Description
			}
			newOptions = append(newOptions, o)

		case discord.ApplicationCommandOptionBool:
			if o.NameLocalizations == nil {
				o.NameLocalizations = make(map[discord.Locale]string)
			}
			if o.DescriptionLocalizations == nil {
				o.DescriptionLocalizations = make(map[discord.Locale]string)
			}
			if optMeta.Name != "" {
				o.NameLocalizations[locale] = optMeta.Name
			}
			if optMeta.Description != "" {
				o.DescriptionLocalizations[locale] = optMeta.Description
			}
			newOptions = append(newOptions, o)

		case discord.ApplicationCommandOptionUser:
			if o.NameLocalizations == nil {
				o.NameLocalizations = make(map[discord.Locale]string)
			}
			if o.DescriptionLocalizations == nil {
				o.DescriptionLocalizations = make(map[discord.Locale]string)
			}
			if optMeta.Name != "" {
				o.NameLocalizations[locale] = optMeta.Name
			}
			if optMeta.Description != "" {
				o.DescriptionLocalizations[locale] = optMeta.Description
			}
			newOptions = append(newOptions, o)

		case discord.ApplicationCommandOptionChannel:
			if o.NameLocalizations == nil {
				o.NameLocalizations = make(map[discord.Locale]string)
			}
			if o.DescriptionLocalizations == nil {
				o.DescriptionLocalizations = make(map[discord.Locale]string)
			}
			if optMeta.Name != "" {
				o.NameLocalizations[locale] = optMeta.Name
			}
			if optMeta.Description != "" {
				o.DescriptionLocalizations[locale] = optMeta.Description
			}
			newOptions = append(newOptions, o)

		case discord.ApplicationCommandOptionRole:
			if o.NameLocalizations == nil {
				o.NameLocalizations = make(map[discord.Locale]string)
			}
			if o.DescriptionLocalizations == nil {
				o.DescriptionLocalizations = make(map[discord.Locale]string)
			}
			if optMeta.Name != "" {
				o.NameLocalizations[locale] = optMeta.Name
			}
			if optMeta.Description != "" {
				o.DescriptionLocalizations[locale] = optMeta.Description
			}
			newOptions = append(newOptions, o)

		case discord.ApplicationCommandOptionMentionable:
			if o.NameLocalizations == nil {
				o.NameLocalizations = make(map[discord.Locale]string)
			}
			if o.DescriptionLocalizations == nil {
				o.DescriptionLocalizations = make(map[discord.Locale]string)
			}
			if optMeta.Name != "" {
				o.NameLocalizations[locale] = optMeta.Name
			}
			if optMeta.Description != "" {
				o.DescriptionLocalizations[locale] = optMeta.Description
			}
			newOptions = append(newOptions, o)

		case discord.ApplicationCommandOptionFloat:
			if o.NameLocalizations == nil {
				o.NameLocalizations = make(map[discord.Locale]string)
			}
			if o.DescriptionLocalizations == nil {
				o.DescriptionLocalizations = make(map[discord.Locale]string)
			}
			if optMeta.Name != "" {
				o.NameLocalizations[locale] = optMeta.Name
			}
			if optMeta.Description != "" {
				o.DescriptionLocalizations[locale] = optMeta.Description
			}
			newOptions = append(newOptions, o)

		case discord.ApplicationCommandOptionAttachment:
			if o.NameLocalizations == nil {
				o.NameLocalizations = make(map[discord.Locale]string)
			}
			if o.DescriptionLocalizations == nil {
				o.DescriptionLocalizations = make(map[discord.Locale]string)
			}
			if optMeta.Name != "" {
				o.NameLocalizations[locale] = optMeta.Name
			}
			if optMeta.Description != "" {
				o.DescriptionLocalizations[locale] = optMeta.Description
			}
			newOptions = append(newOptions, o)

		default:
			newOptions = append(newOptions, opt)
		}
	}
	return newOptions
}
