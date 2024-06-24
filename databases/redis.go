package databases

import (
	"encoding/json"
	"log"
	"miHttpServer/configs"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

// 商品缓存结构体
type ItemCache struct {
	ItemID int64   `json:"item_id"`
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
}

// 存储Redis连接池的实例
var pool *redis.Pool

// Redis的命名空间
var namespace = configs.RedisConfig["namespace"].(string)

// Redis的缓存过期时间
var expireTime = configs.RedisConfig["expire"].(int)

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
				redis.DialDatabase(1),
				redis.DialReadTimeout(10*time.Second),
				redis.DialWriteTimeout(10*time.Second),
			)
			if err != nil {
				log.Fatalf("Redis连接失败：%s", err)
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
				log.Fatalf("ping Redis 失败：%s", err)
			}
			return err
		},
	}
	log.Println("Redis连接池创建成功")
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

// 增加商品缓存
func AddItemCache(itemID int64, itemCache ItemCache) error {
	itemJson, err := json.Marshal(itemCache)
	if err != nil {
		return err
	}
	// 从连接池中获取一个连接
	conn := pool.Get()
	// 用完后将连接放回连接池
	defer conn.Close()
	key := namespace + strconv.FormatInt(itemID, 10)
	_, err = conn.Do("SET", key, itemJson, "EX", expireTime)
	return err
}

// 获取商品缓存
func QueryItemCache(itemID int64, itemCache *ItemCache) (bool, error) {
	conn := pool.Get()
	defer conn.Close()

	key := namespace + strconv.FormatInt(itemID, 10)
	itemJson, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		return false, err
	}
	err = json.Unmarshal(itemJson, itemCache)
	if err != nil {
		return false, err
	}
	return true, nil
}

// 修改商品缓存
func UpdateItemCache(itemID int64, itemCache ItemCache) error {
	itemJson, err := json.Marshal(itemCache)
	if err != nil {
		return err
	}
	conn := pool.Get()
	defer conn.Close()

	key := namespace + strconv.FormatInt(itemID, 10)
	_, err = conn.Do("SET", key, itemJson, "EX", expireTime)
	return err
}

// 删除商品缓存
func DeleteItemCache(itemID int64) error {
	conn := pool.Get()
	defer conn.Close()

	key := namespace + strconv.FormatInt(itemID, 10)
	_, err := conn.Do("DEL", key)
	return err
}
