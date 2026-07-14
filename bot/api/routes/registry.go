package routes

import (
	_ "onyx/bot/api/routes/auth/discord"
	_ "onyx/bot/api/routes/commands"
	_ "onyx/bot/api/routes/dash/guilds/guildId"
	_ "onyx/bot/api/routes/dash/guilds/guildId/modules"
	_ "onyx/bot/api/routes/discord"
	_ "onyx/bot/api/routes/invite"
	_ "onyx/bot/api/routes/status"
)
