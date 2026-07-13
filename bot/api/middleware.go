package api

import (
	"net/http"
	"sync"
	"time"

	"onyx/bot/core"
	"onyx/bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ExpiringCache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]cacheItem[V]
	ttl   time.Duration
}

type cacheItem[V any] struct {
	value     V
	expiresAt time.Time
}

func NewExpiringCache[K comparable, V any](ttl time.Duration) *ExpiringCache[K, V] {
	return &ExpiringCache[K, V]{
		items: make(map[K]cacheItem[V]),
		ttl:   ttl,
	}
}

func (c *ExpiringCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheItem[V]{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *ExpiringCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found {
		var zero V
		return zero, false
	}
	if time.Now().After(item.expiresAt) {
		var zero V
		return zero, false
	}
	return item.value, true
}

var (
	limiters       = make(map[string]*rate.Limiter)
	mu             sync.Mutex
	apiMemberCache = NewExpiringCache[string, discord.Member](5 * time.Minute)
	apiUserCache   = NewExpiringCache[string, discord.User](5 * time.Minute)
	apiPermCache   = NewExpiringCache[string, discord.Permissions](5 * time.Minute)
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		limiter, exists := limiters[ip]
		if !exists {
			limiter = rate.NewLimiter(10, 20)
			limiters[ip] = limiter
		}
		mu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests.", "error_code": "RATE_LIMITED"})
			return
		}

		c.Next()
	}
}
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing session cookie", "error_code": "AUTH_MISSING"})
			return
		}

		database, exists := c.Get("db")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "User not found. Please relogin", "error_code": "USER_NOT_FOUND"})
			return
		}
		gormDB := database.(*db.DB).GormDB

		var session db.Session
		result := gormDB.Where("session_id = ? AND expires_at > ?", sessionID, time.Now()).First(&session)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid or expired session", "error_code": "AUTH_INVALID"})
			return
		}
		c.Set("session", &session)
		c.Set("user_id", session.UserID)
		bot, exists := c.MustGet("bot").(*core.Bot)
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "bot context not found", "error_code": "INTERNAL_ERROR"})
			return
		}

		uid, err := snowflake.Parse(session.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID", "error_code": "INVALID_USER_ID"})
			return
		}

		var user discord.User
		if cachedUser, ok := apiUserCache.Get(uid.String()); ok {
			user = cachedUser
		} else {
			u, err := bot.Client.Rest.GetUser(uid)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user", "error_code": "INTERNAL_ERROR"})
				return
			}
			user = *u
			apiUserCache.Set(uid.String(), user)
		}

		c.Set("user", &user)

		c.Next()
	}
}

func GuildAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing session cookie", "error_code": "AUTH_MISSING"})
			return
		}

		database, exists := c.Get("db")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "User not found. Please relogin", "error_code": "USER_NOT_FOUND"})
			return
		}
		gormDB := database.(*db.DB).GormDB

		var session db.Session
		result := gormDB.Where("session_id = ? AND expires_at > ?", sessionID, time.Now()).First(&session)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid or expired session", "error_code": "AUTH_INVALID"})
			return
		}

		bot, exists := c.MustGet("bot").(*core.Bot)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "bot context not found", "error_code": "INTERNAL_ERROR"})
			return
		}

		gid, err := snowflake.Parse(c.Param("guildId"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid guild ID", "error_code": "INVALID_GUILD_ID"})
			return
		}

		uid, err := snowflake.Parse(session.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID", "error_code": "INVALID_USER_ID"})
			return
		}

		var member discord.Member
		cacheKey := gid.String() + ":" + uid.String()
		if cachedMember, ok := apiMemberCache.Get(cacheKey); ok {
			member = cachedMember
		} else {
			member, ok = bot.Client.Caches.Member(gid, uid)
			if !ok {
				m, err := bot.Client.Rest.GetMember(gid, uid)
				if err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch member", "error_code": "INTERNAL_ERROR"})
					return
				}
				member = *m
			}
			apiMemberCache.Set(cacheKey, member)
		}

		c.Set("session", &session)
		c.Set("member", &member)
		c.Set("user_id", session.UserID)
		c.Next()
	}
}

func PermissionMiddleware(permission discord.Permissions) gin.HandlerFunc {
	return func(c *gin.Context) {
		bot, exists := c.MustGet("bot").(*core.Bot)
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "bot context not found", "error_code": "INTERNAL_ERROR"})
			return
		}

		member, exists := c.MustGet("member").(*discord.Member)
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "member not found", "error_code": "INTERNAL_ERROR"})
			return
		}

		cacheKey := member.GuildID.String() + ":" + member.User.ID.String()
		if cachedPerms, ok := apiPermCache.Get(cacheKey); ok {
			if cachedPerms.Has(permission) {
				c.Next()
				return
			} else {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You do not have the required permissions", "error_code": "USER_MISSING_PERMISSIONS"})
				return
			}
		}

		if bot.Client.Caches.MemberPermissions(*member).Has(permission) {
			apiPermCache.Set(cacheKey, discord.PermissionsAll)
			c.Next()
			return
		}

		guild, err := bot.Client.Rest.GetGuild(member.GuildID, true)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch guild", "error_code": "INTERNAL_ERROR"})
			return
		}

		if guild.OwnerID == member.User.ID {
			c.Next()
			return
		}

		roles, err := bot.Client.Rest.GetRoles(member.GuildID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch roles", "error_code": "INTERNAL_ERROR"})
			return
		}

		var permissions discord.Permissions
		for _, role := range roles {
			if role.ID == member.GuildID {
				permissions = role.Permissions
				break
			}
		}

		for _, roleID := range member.RoleIDs {
			for _, role := range roles {
				if role.ID == roleID {
					permissions = permissions.Add(role.Permissions)
					break
				}
			}
		}

		if permissions.Has(discord.PermissionAdministrator) {
			permissions = discord.PermissionsAll
		}

		apiPermCache.Set(cacheKey, permissions)

		if !permissions.Has(permission) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You do not have the required permissions", "error_code": "USER_MISSING_PERMISSIONS"})
			return
		}

		c.Next()
	}
}
