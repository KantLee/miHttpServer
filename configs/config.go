package configs

// redis的配置项
var RedisConfig = map[string]interface{}{
	// 命名空间前缀（防止与其他数据库冲突）
	"namespace": "mi_",
	// 缓存过期时间（秒，int类型）
	"expire": 3600,
}
