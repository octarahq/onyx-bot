package core

import "github.com/disgoorg/disgo/discord"

type ComponentType string

const (
	ComponentTypeString   ComponentType = "string"
	ComponentTypeNumber   ComponentType = "number"
	ComponentTypeBoolean  ComponentType = "boolean"
	ComponentTypeSelect   ComponentType = "select"
	ComponentTypeChannel  ComponentType = "channel"
	ComponentTypeRole     ComponentType = "role"
	ComponentTypeTextarea ComponentType = "textarea"
)

type UISelectOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type Variables struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	MaxLength   int    `json:"maxLength,omitempty"`
	Length      int    `json:"length,omitempty"`
}

type UIComponent struct {
	Name         string               `json:"name"`
	Label        string               `json:"label,omitempty"`
	Description  string               `json:"description,omitempty"`
	Placeholder  string               `json:"placeholder,omitempty"`
	Type         ComponentType        `json:"type"`
	Required     bool                 `json:"required"`
	Min          *int                 `json:"min,omitempty"`
	Max          *int                 `json:"max,omitempty"`
	Multiple     bool                 `json:"multiple,omitempty"`
	Options      []UISelectOption     `json:"options,omitempty"`
	ChannelTypes []discord.ChannelTag `json:"channelTypes,omitempty"`
	Variables    []Variables          `json:"variables,omitempty"`
}

type UISubModule struct {
	Name        string        `json:"name"`
	Label       string        `json:"label,omitempty"`
	Description string        `json:"description,omitempty"`
	Components  []UIComponent `json:"components"`
}

type UISchema struct {
	SubModules []UISubModule `json:"submodules"`
}

type UIProvider interface {
	UISchema(locale discord.Locale) UISchema
}
