package database

import (
	"encoding/json"
	"log"
	"miHttpServer/config"
	"miHttpServer/models"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

// 存储Redis连接池的实例
var pool *redis.Pool

// Redis的命名空间
var namespace string

// Redis的缓存过期时间
var expireTime int

// 初始化Redis连接池
func InitRedis() {
	// 从配置文件读取Redis的配置信息
	timeout := time.Duration(config.Configs.Redis.Timeout) * time.Second
	maxIdle := config.Configs.Redis.MaxIdle
	idleTimeout := config.Configs.Redis.MaxActive
	protocal := config.Configs.Redis.Protocal
	address := config.Configs.Redis.Address
	password := config.Configs.Redis.Password
	db := config.Configs.Redis.Database
	readTimeout := time.Duration(config.Configs.Redis.ReadTimeout) * time.Second
	writeTimeout := time.Duration(config.Configs.Redis.WriteTimeout) * time.Second
	namespace = config.Configs.Redis.Prefix
	expireTime = config.Configs.Redis.Expire
	pool = &redis.Pool{
		MaxIdle:     maxIdle,     // 最大空闲连接数
		MaxActive:   idleTimeout, // 最大连接数
		IdleTimeout: timeout,     // 超时时间
		// 创建与Redis服务器的连接
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(
				protocal,
				address,
				redis.DialPassword(password),
				redis.DialDatabase(db),
				redis.DialReadTimeout(readTimeout),
				redis.DialWriteTimeout(writeTimeout),
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

// 增加商品
func AddItemCache(itemID int64, itemCache models.ItemCache) error {
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

// 获取商品
func QueryItemCache(itemID int64, itemCache *models.ItemCache) (bool, error) {
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

// 修改商品
func UpdateItemCache(itemID int64, itemCache models.ItemCache) error {
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

// 删除商品
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
			return false, err
		}
		if ok {
			return ok, nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false, nil
}

// Unlock 释放分布式锁
func Unlock(key, value string) error {
	conn := pool.Get()
	defer conn.Close()

	// 先检查是否是自己的锁
	id, err := redis.String(conn.Do("GET", key))
	if err != nil {
		if strings.Contains(err.Error(), "redigo: nil returned") {
			// 锁已经过期，键也会被删除
			return nil
		}
		return err
	}

	if id != value {
		// 不能释放别人的锁（说明锁到期已经被释放过了）
		return nil
	}

	// 执行到这里说明键是存在的，且是自己的锁
	_, err = redis.Int(conn.Do("DEL", key))
	if err != nil {
		return err
	}

	// 释放成功
	return nil
}

// SetNx 设置键值对，如果键不存在
func SetNx(conn redis.Conn, key string, value string, seconds uint64) (bool, error) {
	// "EX" 表示过期时间，"NX" 表示只有键不存在时才设置
	res, err := redis.String(conn.Do("SET", key, value, "EX", seconds, "NX"))
	if err != nil {
		// 设置失败，发生错误
		return false, err
	}
	if res == "OK" {
		// 键不存在，设置成功
		return true, nil
	}
	// 键已存在，说明分布式锁还未被释放
	return false, nil
}
