package utils

import (
	"miHttpServer/config"
	"miHttpServer/database"
	"time"
)

// 获取分布式锁
func GetLock(key, value string) (bool, error) {
	locked, err := database.Lock(
		key,
		value,
		config.Configs.Lock.ExpireSec,
		time.Duration(config.Configs.Lock.WaitSec)*time.Second,
	)
	return locked, err
}

// 释放分布式锁
func ReleaseLock(key, value string) {
	for i := 0; i < 3; i++ {
		err := database.Unlock(key, value)
		if err == nil {
			break
		}
	}
}
