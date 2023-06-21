package client

import (
	"bee-search/pkg/config"
	"bee-search/pkg/model/status"
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisService(rc config.RedisConfig) *RedisService {
	rdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    rc.MasterName,
		SentinelAddrs: rc.Host,
		Password:      rc.Password,
		DB:            rc.Database,
	})

	return &RedisService{rdb, context.Background()}
}

func (r *RedisService) Save(key string, value string) {
	_ = r.client.RPush(r.ctx, key, value).Err()
	r.SearchExpire(key)
}

func (r *RedisService) SearchStarted(key string) {
	r.client.HMSet(r.ctx, key, map[string]interface{}{"status": status.STARTED}).Err()
	r.SearchExpire(key)
	r.Increase()
}

func (r *RedisService) SearchFinished(key string, dt int64) {
	r.client.HMSet(r.ctx, key, map[string]interface{}{"status": status.FINISHED, "completedTime": dt}).Err()
	r.SearchExpire(key)
}

func (r *RedisService) SearchException(key string, err error) {
	r.client.HMSet(r.ctx, key, map[string]interface{}{"status": status.ERROR, "error": err.Error()}).Err()
	r.SearchExpire(key)
}

func (r *RedisService) SearchExpire(key string) {
	_ = r.client.Expire(r.ctx, key, time.Minute*60).Err()
}

func (r *RedisService) SearchExpireImmediately(key string) {
	_ = r.client.Expire(r.ctx, key, time.Second).Err()
}

func (r *RedisService) Length(key string) (int64, error) {
	return r.client.LLen(r.ctx, key).Result()
}

func (r *RedisService) GetStateOfSearch(key string) (string, error) {
	res, err := r.client.HMGet(r.ctx, key, "status").Result()
	if err != nil {
		return "", err
	}

	if len(res) == 0 || res[0] == nil {
		return "", errors.New("not found record on Redis")
	}

	return res[0].(string), nil
}

func (r *RedisService) Increase() {
	_ = r.client.Incr(r.ctx, "search_count").Err()
}
