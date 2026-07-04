package locales

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/disgoorg/disgo/discord"
)

//go:generate go run generator/main.go

var localizations = make(map[discord.Locale]map[string]any)

type CommandMeta struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Options     map[string]CommandMeta `json:"options,omitempty"`
}

var Metas = make(map[discord.Locale]map[string]CommandMeta)

type fileFormat struct {
	Meta CommandMeta     `json:"meta"`
	Trad json.RawMessage `json:"trad"`
}

func Load(localesDir string) error {
	return filepath.Walk(localesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		relPath, err := filepath.Rel(localesDir, path)
		if err != nil {
			return err
		}

		parts := strings.Split(filepath.ToSlash(relPath), "/")
		if len(parts) < 3 {
			// locales/<locale>/commands/<category>/<cmd>.json
			return nil
		}

		localeStr := parts[0]
		locale := discord.Locale(localeStr)
		cmdName := strings.TrimSuffix(parts[len(parts)-1], ".json")

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var ff fileFormat
		if err := json.Unmarshal(content, &ff); err != nil {
			return err
		}

		if _, ok := localizations[locale]; !ok {
			localizations[locale] = make(map[string]any)
			Metas[locale] = make(map[string]CommandMeta)
		}

		Metas[locale][cmdName] = ff.Meta

		trad, err := unmarshalTrad(cmdName, ff.Trad)
		if err != nil {
			return err
		}

		if trad != nil {
			localizations[locale][cmdName] = trad
		}

		return nil
	})
}
