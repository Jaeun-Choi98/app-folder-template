package redis

import (
	"fmt"
	"pjt/internal/config"
	redismodel "pjt/internal/redis/redis-model"
	"sync"

	"github.com/Jaeun-Choi98/modules/orm/redisorm"
)

var redisClient *redisorm.RedisClient
var repositories map[redismodel.RedisKey]any
var redisMu sync.RWMutex

func InitRedis(cfg *config.Configuration) error {
	redisClient = redisorm.NewRedisClient(fmt.Sprintf("%s:%s", cfg.RedisIp, cfg.RedisPort),
		cfg.RedisPwd, cfg.RedisDB, cfg.RedisProtocl, cfg.RedisTimeout)
	if redisClient == nil {
		return fmt.Errorf("[REDIS] failed to connect redis")
	}
	repositories = make(map[redismodel.RedisKey]any)
	LoadDefaultRepo()

	return nil
}

func CloseRedisClient() error {
	return redisClient.Close()
}

func AddRepository[T redisorm.Model](key redismodel.RedisKey, model T) error {
	redisMu.Lock()
	defer redisMu.Unlock()
	if _, exists := repositories[key]; !exists {
		repositories[key] = redisorm.NewRepository(redisClient, model)
	}
	return nil
}

func DeleteRepository(key redismodel.RedisKey) error {
	redisMu.Lock()
	defer redisMu.Unlock()
	delete(repositories, key)
	return nil
}

func GetRepository[T redisorm.Model](key redismodel.RedisKey) *redisorm.Repository[T] {
	redisMu.RLock()
	defer redisMu.RUnlock()

	repo, exists := repositories[key]
	if !exists {
		return nil
	}

	// 타입 단언
	typedRepo, ok := repo.(*redisorm.Repository[T])
	if !ok {
		return nil
	}

	return typedRepo
}

func LoadDefaultRepo() {
	AddRepository(redismodel.RedisKeySample, &redismodel.Sample{})
}
