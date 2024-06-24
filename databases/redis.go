package databases

import (
	"github.com/gomodule/redigo/redis"
	"log"
	"time"
)

// 存储Redis连接池的实例
var pool *redis.Pool

// 初始化Redis连接池
func InitRedis() {
	pool = &redis.Pool{
		MaxIdle:     16,                // 最大空闲连接数
		MaxActive:   32,                // 最大连接数
		IdleTimeout: 120 * time.Second, // 超时时间
		// 创建与Redis服务器的连接
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(
				"tcp",
				"192.168.96.130:6379",
				redis.DialPassword("admin"),
				redis.DialReadTimeout(10*time.Second),
				redis.DialWriteTimeout(10*time.Second),
			)
			if err != nil {
				log.Fatalf("redisClient dial host:%s,auth:%s error:%s", "192.168.96.130:6379", "admin", err)
				return nil, err
			}
			return conn, nil
		},
		// 从连接池借用连接时检查连接的健康状况
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			if err != nil {
				log.Fatalf("ping redis error:%s", err)
			}
			return err
		},
	}
	log.Println("redis pool init success")
}

// 关闭Redis连接池
func CloseRedis() {
	if pool != nil {
		err := pool.Close()
		if err != nil {
			log.Printf("close redis pool error:%s", err)
		}
	}
}
