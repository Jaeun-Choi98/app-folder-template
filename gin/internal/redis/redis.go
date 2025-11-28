package redis

import (
	"fmt"
	"pjt/internal/config"
	"sync"

	"github.com/Jaeun-Choi98/modules/orm/redisorm"
)

var redisClient *redisorm.RedisClient
var repositoies map[string]*redisorm.Repository[redisorm.Model]
var redisMu sync.RWMutex

func InitRedis(cfg *config.Configuration) error {
	redisClient = redisorm.NewRedisClient(fmt.Sprintf("%s:%s", cfg.RedisIp, cfg.RedisPort),
		cfg.RedisPwd, cfg.RedisDB, cfg.RedisProtocl, cfg.RedisTimeout)
	if redisClient == nil {
		return fmt.Errorf("[REDIS] failed to connect redis")
	}
	return nil
}

func CloseRedisClient() error {
	return redisClient.Close()
}

func AddRepository(key string, model redisorm.Model) error {
	redisMu.Lock()
	defer redisMu.Unlock()
	if _, exists := repositoies[key]; !exists {
		repositoies[key] = redisorm.NewRepository[redisorm.Model](redisClient, model)
	}
	return nil
}

func DeleteRepository(key string) error {
	redisMu.Lock()
	defer redisMu.Unlock()
	if _, exists := repositoies[key]; exists {
		delete(repositoies, key)
	}
	return nil
}
