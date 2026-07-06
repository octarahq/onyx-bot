package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type TranslateParams struct {
	Query        string `json:"q"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	Format       string `json:"format"`
	Alternatives int    `json:"alternatives"`
}

type TranslateReponse struct {
	Alternatives     []string                  `json:"alternatives"`
	DetectedLanguage TranslateDetectedLanguage `json:"detectedLanguage"`
	TranslatedText   string                    `json:"translatedText"`
}

type TranslateDetectedLanguage struct {
	Confidence int    `json:"confidence"`
	Language   string `json:"language"`
}

type Lang struct {
	Name  string
	Value string
	Flag  string
}

var TranslateLangs = map[string]Lang{
	"ar": {Name: "Arabic", Value: "ar", Flag: ":flag_al:"},
	"zh": {Name: "Chinese", Value: "zh", Flag: ":flag_cn:"},
	"en": {Name: "English", Value: "en", Flag: ":flag_us:"},
	"fr": {Name: "French", Value: "fr", Flag: ":flag_fr:"},
	"de": {Name: "German", Value: "de", Flag: ":flag_de:"},
	"it": {Name: "Italian", Value: "it", Flag: ":flag_it:"},
	"ja": {Name: "Japanese", Value: "ja", Flag: ":flag_jp:"},
	"pt": {Name: "Portuguese", Value: "pt", Flag: ":flag_pt:"},
	"ru": {Name: "Russian", Value: "ru", Flag: ":flag_ru:"},
	"es": {Name: "Spanish", Value: "es", Flag: ":flag_es:"},
}

func Translate(params TranslateParams) TranslateReponse {
	url := "https://4027.fr1.orionhost.xyz/translate"
	p, _ := json.Marshal(params)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(p))
	req.Header.Set("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	if res == nil {
		return TranslateReponse{}
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var translated TranslateReponse
	_ = json.Unmarshal(body, &translated)

	return translated
}
