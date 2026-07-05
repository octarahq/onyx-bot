package db

import (
	"reflect"

	"gorm.io/gorm"
)

func LoadSettings(db *gorm.DB, guildID string) (*Guild, error) {
	guild := &Guild{GuildID: guildID}

	query := db
	for _, mod := range Registry {
		query = query.Preload(mod.Name)
	}

	err := query.First(guild, "guild_id = ?", guildID).Error
	var needsUpdate bool

	if err == gorm.ErrRecordNotFound {
		needsUpdate = true
		err = nil
	} else if err != nil {
		return nil, err
	}

	guildVal := reflect.ValueOf(guild).Elem()
	for _, mod := range Registry {
		field := guildVal.FieldByName(mod.Name)
		if field.IsValid() {
			guildIDField := field.FieldByName("GuildID")
			if guildIDField.IsValid() && guildIDField.String() == "" {
				newStruct := reflect.New(field.Type()).Elem()
				newStruct.FieldByName("GuildID").SetString(guildID)
				field.Set(newStruct)
				needsUpdate = true
			}
		}
	}

	if needsUpdate {
		UpdateSettings(db, guild)
		query.First(guild, "guild_id = ?", guildID)
	}

	return guild, nil
}

func UpdateSettings(db *gorm.DB, guild *Guild) error {
	return db.Session(&gorm.Session{FullSaveAssociations: true}).Save(guild).Error
}
