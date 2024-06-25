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

// Lock 尝试获取分布式锁
func Lock(key string, requestID string, expireSec uint64, maxWait time.Duration) (bool, error) {
	conn := pool.Get()
	defer conn.Close()

	for startTime := time.Now(); time.Since(startTime) < maxWait; {

		ok, err := SetNx(conn, key, requestID, expireSec)
		if err != nil {
			// @warn 错误信息向上传递，而不是记录日志。日志没有额外信息透出时，尽量只在处理终止时记录
			// "github.com/pkg/errors"是个不错的库，errors.Wrap(err, "获取分布式锁失败", + ",key:" + key)
			log.Println("获取分布式锁失败", ",key:", key, ",error:", err)
			return false, err
		}
		if ok {
			return ok, nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	log.Println("获取分布式锁超时", ",key:", key)
	return false, nil
}

// Unlock 释放分布式锁
func Unlock(key string) error {
	conn := pool.Get()
	defer conn.Close()
	// @warn 无唯一 ID 校验，可能会造成其他其他协程误释放
	_, err := Del(conn, key)
	if err != nil {
		log.Println("释放锁失败", ",key:", key, ",error:", err)
		return err
	} else {
		log.Println("释放锁成功", ",key:", key)
		return nil
	}
}

// SetNx 设置键值对，如果键不存在
func SetNx(conn redis.Conn, key string, value string, seconds uint64) (bool, error) {
	// "EX" 表示过期时间，"NX" 表示只有键不存在时才设置
	res, err := redis.String(conn.Do("SET", key, value, "EX", seconds, "NX"))
	if err == redis.ErrNil {
		// 键已存在
		return false, nil
	} else if err != nil {
		log.Println("SetNx Error: ", err)
		return false, err
	}
	log.Println("Redis SetNx", ",key:", key, ",seconds:", seconds, ",res:", res)
	return res == "OK", nil
}

// Del 删除键
func Del(conn redis.Conn, key string) (bool, error) {
	_, err := redis.Int(conn.Do("DEL", key))
	if err != nil {
		log.Println("删除键错误，err", "err", err)
		return false, err
	}
	return true, nil
}
