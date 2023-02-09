package sources

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"

	log "github.com/sirupsen/logrus"
)

type (
	RedisClient interface {
		Del(ctx context.Context, key string) error
		Set(ctx context.Context, key string, value interface{}, expire time.Duration) error
		Expire(ctx context.Context, key string, duration time.Duration) (bool, error)
		GetInt64(ctx context.Context, key string) (int64, bool, error)
		GetFloat64(ctx context.Context, key string) (float64, bool, error)
		GetInt(ctx context.Context, key string) (int, bool, error)
		GetString(ctx context.Context, key string) (string, bool, error)
		HIncrBy(ctx context.Context, key, field string, incr int64) error
		HIncrByFloat(ctx context.Context, key, field string, incr float64) error
		HGetString(ctx context.Context, key string, field string) (string, bool, error)
		HGetInt(ctx context.Context, key string, field string) (int, bool, error)
		HGetInt64(ctx context.Context, key string, field string) (int64, bool, error)
		HGetFloat64(ctx context.Context, key string, field string) (float64, bool, error)
		HGetByte(ctx context.Context, key string, field string) ([]byte, bool, error)
		HSet(ctx context.Context, key, field string, value interface{}) error
		HDel(ctx context.Context, key string, fields ...string) error
		HExists(ctx context.Context, key, field string) (bool, error)
		HSetEX(ctx context.Context, key, field string, value interface{}, expire time.Duration) error
		HLen(ctx context.Context, key string) (int64, error)
		HGetAll(ctx context.Context, key string) (map[string]string, bool, error)
		HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, bool, error)
		RPush(ctx context.Context, key string, values ...string) (int64, error)
		LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
		LLen(ctx context.Context, key string) (int64, error)
		LRem(ctx context.Context, key string, count int64, element string) (int64, error)
	}

	redisClient struct {
		redis *redis.Client
	}
)

func NewRedisClient(url string) (RedisClient, CloseHandler) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(opt)

	if err := client.Ping(context.TODO()).Err(); err != nil {
		log.Fatal("Couldn't ping to redis server.", url)
		panic(err)
	}

	rc := &redisClient{
		redis: client,
	}

	return rc, rc.close
}

func (rc *redisClient) close() error {
	return rc.redis.Close()
}

func (rc *redisClient) Del(ctx context.Context, key string) error {
	return rc.redis.Del(ctx, key).Err()
}

func (rc *redisClient) Set(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	return rc.redis.Set(ctx, key, value, expire).Err()
}

func (rc *redisClient) GetInt64(ctx context.Context, key string) (int64, bool, error) {
	val, err := rc.redis.Get(ctx, key).Int64()
	if err == nil {
		return val, true, nil
	}
	if err == redis.Nil {
		return 0, false, nil
	}

	return 0, false, err
}

func (rc *redisClient) GetFloat64(ctx context.Context, key string) (float64, bool, error) {
	val, err := rc.redis.Get(ctx, key).Float64()
	if err == nil {
		return val, true, nil
	}
	if err == redis.Nil {
		return 0, false, nil
	}

	return 0, false, err
}

func (rc *redisClient) GetInt(ctx context.Context, key string) (int, bool, error) {
	val, err := rc.redis.Get(ctx, key).Int()
	if err == nil {
		return val, true, nil
	}
	if err == redis.Nil {
		return 0, false, nil
	}

	return 0, false, err
}

func (rc *redisClient) GetString(ctx context.Context, key string) (string, bool, error) {
	val, err := rc.redis.Get(ctx, key).Result()
	if err == nil {
		return val, true, nil
	}
	if err == redis.Nil {
		return "", false, nil
	}

	return "", false, err
}

func (rc *redisClient) HGetAll(ctx context.Context, key string) (map[string]string, bool, error) {
	val, err := rc.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, false, err
	}
	if len(val) == 0 {
		return nil, false, nil
	}

	return val, true, nil
}

func (rc *redisClient) HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, bool, error) {
	val, err := rc.redis.HMGet(ctx, key, fields...).Result()
	if err != nil {
		return nil, false, err
	}
	if len(val) == 0 {
		return nil, false, nil
	}

	return val, true, nil
}

func (rc *redisClient) HSet(ctx context.Context, key, field string, value interface{}) error {
	return rc.redis.HSet(ctx, key, field, value).Err()
}

func (rc *redisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return rc.redis.HDel(ctx, key, fields...).Err()
}

func (rc *redisClient) HIncrBy(ctx context.Context, key, field string, incr int64) error {
	return rc.redis.HIncrBy(ctx, key, field, incr).Err()
}

func (rc *redisClient) HIncrByFloat(ctx context.Context, key, field string, incr float64) error {
	return rc.redis.HIncrByFloat(ctx, key, field, incr).Err()
}

func (rc *redisClient) HGetString(ctx context.Context, key string, field string) (string, bool, error) {
	val, err := rc.redis.HGet(ctx, key, field).Result()
	if err == nil {
		return val, true, nil
	}
	if err == redis.Nil {
		return "", false, nil
	}

	return "", false, err
}

func (rc *redisClient) HGetInt(ctx context.Context, key string, field string) (int, bool, error) {
	val, err := rc.redis.HGet(ctx, key, field).Int()
	if err == nil {
		return val, true, nil
	}
	if err == redis.Nil {
		return 0, false, nil
	}

	return 0, false, err
}

func (rc *redisClient) HGetInt64(ctx context.Context, key string, field string) (int64, bool, error) {
	val, err := rc.redis.HGet(ctx, key, field).Int64()
	if err == nil {
		return val, true, nil
	}
	if err == redis.Nil {
		return 0, false, nil
	}

	return 0, false, err
}

func (rc *redisClient) HGetFloat64(ctx context.Context, key string, field string) (float64, bool, error) {
	val, err := rc.redis.HGet(ctx, key, field).Float64()
	if err == nil {
		return val, true, nil
	}
	if err == redis.Nil {
		return 0, false, nil
	}

	return 0, false, err
}

func (rc *redisClient) HGetByte(ctx context.Context, key string, field string) ([]byte, bool, error) {
	val, err := rc.redis.HGet(ctx, key, field).Bytes()
	if err == nil {
		return val, true, nil
	}
	if err == redis.Nil {
		return nil, false, nil
	}

	return nil, false, err
}

func (rc *redisClient) HExists(ctx context.Context, key, field string) (bool, error) {
	return rc.redis.HExists(ctx, key, field).Result()
}

func (rc *redisClient) Expire(ctx context.Context, key string, duration time.Duration) (bool, error) {
	return rc.redis.Expire(ctx, key, duration).Result()
}

func (rc *redisClient) HSetEX(ctx context.Context, key, field string, value interface{}, expire time.Duration) error {
	err := rc.HSet(ctx, key, field, value)
	if err != nil {
		return err
	}
	_, err = rc.Expire(ctx, key, expire)

	return err
}

func (rc *redisClient) HLen(ctx context.Context, key string) (int64, error) {
	return rc.redis.HLen(ctx, key).Result()
}

func (rc *redisClient) RPush(ctx context.Context, key string, values ...string) (int64, error) {
	return rc.redis.RPush(ctx, key, values).Result()
}

func (rc *redisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return rc.redis.LRange(ctx, key, start, stop).Result()
}

func (rc *redisClient) LLen(ctx context.Context, key string) (int64, error) {
	return rc.redis.LLen(ctx, key).Result()
}

// LRem Removes the first count occurrences of elements equal to element from the list stored at key. The count argument influences the operation in the following ways:
// count > 0: Remove elements equal to element moving from head to tail.
// count < 0: Remove elements equal to element moving from tail to head.
// count = 0: Remove all elements equal to element.
func (rc *redisClient) LRem(ctx context.Context, key string, count int64, element string) (int64, error) {
	return rc.redis.LRem(ctx, key, count, element).Result()
}
