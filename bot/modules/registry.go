package modules

import (
	"onyx/bot/core"
	"sort"
)

var RegisteredModules []core.Module

func Register(m core.Module) {
	RegisteredModules = append(RegisteredModules, m)
	sort.SliceStable(RegisteredModules, func(i, j int) bool {
		return RegisteredModules[i].Priority() > RegisteredModules[j].Priority()
	})
}
