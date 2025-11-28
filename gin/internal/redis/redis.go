package redis

import (
	"fmt"
	"pjt/internal/config"
	redismodel "pjt/internal/redis/redis-model"
	"sync"

	"github.com/Jaeun-Choi98/modules/orm/redisorm"
)

var redisClient *redisorm.RedisClient
var repositoies map[redismodel.RedisKey]*redisorm.Repository[redisorm.Model]
var redisMu sync.RWMutex

func InitRedis(cfg *config.Configuration) error {
	redisClient = redisorm.NewRedisClient(fmt.Sprintf("%s:%s", cfg.RedisIp, cfg.RedisPort),
		cfg.RedisPwd, cfg.RedisDB, cfg.RedisProtocl, cfg.RedisTimeout)
	if redisClient == nil {
		return fmt.Errorf("[REDIS] failed to connect redis")
	}
	LoadDefaultRepo()

	return nil
}

func CloseRedisClient() error {
	return redisClient.Close()
}

func AddRepository(key redismodel.RedisKey, model redisorm.Model) error {
	redisMu.Lock()
	defer redisMu.Unlock()
	if _, exists := repositoies[key]; !exists {
		repositoies[key] = redisorm.NewRepository[redisorm.Model](redisClient, model)
	}
	return nil
}

func DeleteRepository(key redismodel.RedisKey) error {
	redisMu.Lock()
	defer redisMu.Unlock()
	if _, exists := repositoies[key]; exists {
		delete(repositoies, key)
	}
	return nil
}

func GetRepository(key redismodel.RedisKey) *redisorm.Repository[redisorm.Model] {
	redisMu.RLock()
	defer redisMu.RUnlock()
	return repositoies[key]
}

func LoadDefaultRepo() {
	AddRepository(redismodel.RedisKeySample, &redismodel.Sample{})
}
