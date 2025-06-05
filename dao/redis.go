package dao

import (
	"encoding/json"
	"log"
	"net/url"
	"time"

	"github.com/gomodule/redigo/redis"
)

var RedisPool *redis.Pool
var RedisEnabled bool

func InitRedis(dsn string) {
	u, err := url.Parse(dsn)
	if err != nil {
		log.Fatalf("解析 Redis DSN 失败: %v", err)
	}

	address := u.Host
	password, _ := u.User.Password()
	db := "0" // 默认数据库 0

	if len(u.Path) > 1 {
		db = u.Path[1:] // 去除前导斜杠，例如 "/1" -> "1"
	}

	RedisPool = &redis.Pool{
		MaxIdle:     10,
		MaxActive:   100,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			if db != "0" {
				if _, err := c.Do("SELECT", db); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	conn := RedisPool.Get()
	defer conn.Close()
	_, err = conn.Do("PING")
	if err != nil {
		log.Fatalf("连接 Redis 失败: %v", err)
	}

	RedisEnabled = true
	log.Println("Redis 连接成功")
}

func SaveDataToRedis(key string, data interface{}, ttlSeconds int) error {
	if !RedisEnabled {
		return nil
	}

	conn := RedisPool.Get()
	defer conn.Close()

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", key, jsonBytes, "EX", ttlSeconds)
	return err
}

func LoadDataFromRedis(key string, result interface{}) error {
	if !RedisEnabled {
		return redis.ErrNil
	}

	conn := RedisPool.Get()
	defer conn.Close()

	data, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return err
	}

	return json.Unmarshal(data, result)
}

func LoadDataWithLock(key string, target interface{}, fallback func() (interface{}, error), ttl int) error {
	if !RedisEnabled {
		// 如果没有启用 Redis，直接调用 fallback 函数
		data, err := fallback()
		if err != nil {
			return err
		}
		b, _ := json.Marshal(data)
		return json.Unmarshal(b, target)
	}

	// 1. 先读缓存
	err := LoadDataFromRedis(key, target)
	if err == nil {
		return nil
	}

	lockKey := "lock:" + key
	locked, err := TryLock(lockKey, 5) // 5秒锁
	if err != nil {
		return err
	}

	if locked {
		// 拿锁后再读缓存一次，防止重复读数据库
		err = LoadDataFromRedis(key, target)
		if err == nil {
			_ = Unlock(lockKey)
			return nil
		}

		// 调用回调从数据库加载数据
		data, err := fallback()
		if err != nil {
			_ = Unlock(lockKey)
			return err
		}

		// 将数据赋给target
		b, _ := json.Marshal(data)
		_ = json.Unmarshal(b, target)

		// 写缓存，忽略写缓存错误
		_ = SaveDataToRedis(key, data, ttl)

		_ = Unlock(lockKey)
		return nil
	} else {
		// 没拿到锁，等一会再读缓存
		time.Sleep(100 * time.Millisecond)
		return LoadDataFromRedis(key, target)
	}
}

func TryLock(lockKey string, expire int) (bool, error) {
	conn := RedisPool.Get()
	defer conn.Close()

	reply, err := redis.String(conn.Do("SET", lockKey, "1", "NX", "EX", expire))
	if err != nil && err != redis.ErrNil {
		return false, err
	}
	return reply == "OK", nil
}

// 释放锁
func Unlock(lockKey string) error {
	conn := RedisPool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", lockKey)
	return err
}
