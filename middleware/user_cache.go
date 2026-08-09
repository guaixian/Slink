package middleware

import (
	"Slink/cache"
	"Slink/model"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// 鉴权中间件的用户查询缓存（两级）：
//   L1 进程内 sync.Map(5s TTL)——单实例部署下零网络开销;
//   L2 全局 cache.GlobalCache——配置 Redis 时数据落 Redis,多实例共享,
//      L1 未命中时回源 L2,仍远快于数据库往返。
// 权限变更（降级/删除）由管理端调用 InvalidateUserCache 主动失效两级缓存;
// 未主动失效的变更最多 5 秒（L1 TTL）生效。
const userCacheTTL = 5 * time.Second

type userCacheEntry struct {
	user *model.User
	at   time.Time
}

var userL1 sync.Map // key: uint(userID)

func userCacheKey(id uint) string {
	return fmt.Sprintf("slink:user:%d", id)
}

// GetCachedUser 带 5 秒 TTL 的用户查询（仅供鉴权中间件使用）
func GetCachedUser(id uint) (*model.User, error) {
	// L1:进程内
	if v, ok := userL1.Load(id); ok {
		e := v.(userCacheEntry)
		if time.Since(e.at) < userCacheTTL {
			return e.user, nil
		}
		userL1.Delete(id)
	}

	// L2:全局缓存（Redis/文件/内存）
	key := userCacheKey(id)
	if cache.GlobalCache != nil {
		if v, err := cache.GlobalCache.Get(key); err == nil && v != nil {
			var u *model.User
			switch t := v.(type) {
			case *model.User:
				u = t
			case model.User:
				u = &t
			default:
				// Redis 取回的是 JSON 反序列化后的 map，转回结构体
				if b, err := json.Marshal(v); err == nil {
					var u2 model.User
					if json.Unmarshal(b, &u2) == nil && u2.ID == id {
						u = &u2
					}
				}
			}
			if u != nil {
				userL1.Store(id, userCacheEntry{user: u, at: time.Now()})
				return u, nil
			}
		}
	}

	u, err := model.GetUserByID(model.DB, id)
	if err != nil {
		return nil, err
	}
	// 缓存副本剔除密码哈希，避免敏感信息写入外部缓存（Redis）
	uc := *u
	uc.Password = ""
	if cache.GlobalCache != nil {
		_ = cache.GlobalCache.Set(key, &uc, userCacheTTL)
	}
	userL1.Store(id, userCacheEntry{user: &uc, at: time.Now()})
	return u, nil
}

// InvalidateUserCache 使用户缓存立即失效（用户被修改/删除时调用）
func InvalidateUserCache(id uint) {
	userL1.Delete(id)
	if cache.GlobalCache != nil {
		_ = cache.GlobalCache.Delete(userCacheKey(id))
	}
}
