package db

import "gorm.io/datatypes"

type ModuleDefinition struct {
	Model interface{}
	Name  string
}

var Registry = []ModuleDefinition{
	{
		Model: &TranslationSettings{},
		Name:  "SettingsTranslation",
	},
}

type Guild struct {
	GuildID string `gorm:"primaryKey"`

	SettingsTranslation TranslationSettings `gorm:"foreignKey:GuildID;constraint:OnDelete:CASCADE;"`
}

type TranslationSettings struct {
	GuildID  string `gorm:"primaryKey"`
	Enabled  bool   `gorm:"default:false"`
	Channels datatypes.JSON
	Lang     string `gorm:"default:'en-US'"`
}
