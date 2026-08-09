package middleware

import (
	"Slink/model"
	"sync"
	"time"
)

// 鉴权中间件的用户查询缓存：JWT/Bearer 校验每个请求都要查 users 表，
// 5 秒短缓存可显著降低热路径数据库压力。权限变更（降级/删除）最多 5 秒生效；
// 管理端修改/删除用户时会主动调用 InvalidateUserCache 立即失效。
type userCacheEntry struct {
	user *model.User
	at   time.Time
}

var (
	userCache    sync.Map // key: uint(userID)
	userCacheTTL = 5 * time.Second
)

// GetCachedUser 带 5 秒 TTL 的用户查询（仅供鉴权中间件使用）
func GetCachedUser(id uint) (*model.User, error) {
	if v, ok := userCache.Load(id); ok {
		e := v.(userCacheEntry)
		if time.Since(e.at) < userCacheTTL {
			return e.user, nil
		}
		userCache.Delete(id)
	}
	u, err := model.GetUserByID(model.DB, id)
	if err != nil {
		return nil, err
	}
	userCache.Store(id, userCacheEntry{user: u, at: time.Now()})
	return u, nil
}

// InvalidateUserCache 使用户缓存立即失效（用户被修改/删除时调用）
func InvalidateUserCache(id uint) {
	userCache.Delete(id)
}
