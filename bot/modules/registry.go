package modules

import (
	"onyx/bot/core"
)

var RegisteredModules []core.Module

func Register(m core.Module) {
	RegisteredModules = append(RegisteredModules, m)
}
