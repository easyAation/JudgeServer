package db

import (
	"context"
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"

	"online_judge/talcity/scaffold/criteria/log"
	"online_judge/talcity/scaffold/criteria/merr"
)

// redis config info.
type RedisConfig struct {
	Addr        string
	Password    string
	DB          int
	MaxIdle     int
	MaxActive   int
	DialTimeout duration // eg. "8m03s"
}

// duration time duration.
type duration struct {
	time.Duration
}

// UnmarshalText
// duration type satisfies the `encoding.TextUnmarshaler` interface.
func (d *duration) UnmarshalText(text []byte) (err error) {
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

var (
	pool              *redis.Pool
	withoutInitPool   = errors.New("please Init/Register Pool")
	redisAddrEmpty    = errors.New("init redis instance failed, redis addr is empty")
	alreadyRegistered = errors.New("redis Pool already Registered")
)

// newPool new redis pool
func newPool(conf *RedisConfig, options []redis.DialOption) *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", conf.Addr, options...)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}

			reply, err := c.Do("INFO")
			if err == nil {
				log.Infof("redis info=%s", reply)
			}

			return err
		},
		MaxIdle:   conf.MaxIdle,
		MaxActive: conf.MaxActive,
	}
}

// RegisterRedis
func RegisterRedis(conf *RedisConfig) error {
	var (
		err     error
		timeout time.Duration
		options []redis.DialOption
	)

	if conf == nil || conf.Addr == "" {
		err = redisAddrEmpty
		goto end
	}

	if pool != nil {
		err = alreadyRegistered
		goto end
	}

	options = []redis.DialOption{
		redis.DialDatabase(conf.DB),
		redis.DialPassword(conf.Password),
	}

	timeout = conf.DialTimeout.Duration
	if timeout > 0 {
		options = append(options, redis.DialConnectTimeout(timeout))
	}

	pool = newPool(conf, options)
	if pool.MaxActive > 0 {
		pool.Wait = true
	}

end:
	if err != nil {
		println(err.Error())
	}

	return err
}

// getConnFromPool
// if pool is nil, return err.
func getConnFromPool(ctx context.Context) (redis.Conn, error) {
	if pool == nil {
		return nil, withoutInitPool
	}

	conn, err := pool.GetContext(ctx)
	if err != nil || conn.Err() != nil {
		return nil, err
	}

	return conn, nil
}

// Del del keys from redis
func Del(ctx context.Context, keys ...interface{}) error {
	conn, err := getConnFromPool(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("del", keys...)
	if err != nil {
		return err
	}

	return nil
}

// Set set key:value to redis
func Set(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	conn, err := getConnFromPool(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	switch true {
	case expire < 0:
		_, err = conn.Do("set", key, value)
	case time.Millisecond < expire && expire < time.Second:
		_, err = conn.Do("set", key, value, "px", int64(expire/time.Millisecond))
	default:
		_, err = conn.Do("set", key, value, "ex", int64(expire/time.Second))
	}

	if err != nil {
		return err
	}

	return nil
}

// Get get value by key from redis
func Get(ctx context.Context, key string) (interface{}, error) {
	return RedisCMD(ctx, "get", key)
}

// GetStr get string value by key from redis
func GetStr(ctx context.Context, key string) (string, error) {
	return redis.String(Get(ctx, key))
}

func Expire(ctx context.Context, key string, expire time.Duration) error {
	if expire <= 0 {
		return nil
	}
	conn, err := getConnFromPool(ctx)
	if err != nil {
		return merr.Wrap(err, 0)
	}
	defer conn.Close()

	if time.Millisecond < expire && expire < time.Second {
		_, err = conn.Do("pexpire", key, int64(expire/time.Millisecond))
		if err != nil {
			return merr.Wrap(err, 0)
		}
	} else {
		_, err = conn.Do("expire", key, int64(expire/time.Second))
		if err != nil {
			return merr.Wrap(err, 0)
		}
	}
	return nil
}

func RPush(ctx context.Context, expire time.Duration, key string, value interface{}) error {
	conn, err := getConnFromPool(ctx)
	if err != nil {
		return merr.Wrap(err, 0)
	}
	defer conn.Close()

	_, err = conn.Do("rpush", key, value)
	if err != nil {
		return merr.Wrap(err, 0)
	}

	// 设置过期
	if expire <= 0 {
		return nil
	}
	if time.Millisecond < expire && expire < time.Second {
		_, err = conn.Do("pexpire", key, int64(expire/time.Millisecond))
		if err != nil {
			return merr.Wrap(err, 0)
		}
	} else {
		_, err = conn.Do("expire", key, int64(expire/time.Second))
		if err != nil {
			return merr.Wrap(err, 0)
		}
	}

	return nil
}

func LRange(ctx context.Context, key string, offset, end int) (interface{}, error) {
	return RedisCMD(ctx, "lrange", key, offset, end)
}

func LRangeByteSlice(ctx context.Context, key string, offset, end int) ([][]byte, error) {
	reply, err := LRange(ctx, key, offset, end)
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, nil
	}
	return redis.ByteSlices(reply, nil)
}

func RPOP(ctx context.Context, key string) (interface{}, error) {
	return RedisCMD(ctx, "rpop", key)
}

func RPOPBytes(ctx context.Context, key string) ([]byte, error) {
	reply, err := RPOP(ctx, key)
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, nil
	}
	return redis.Bytes(reply, nil)
}

func RedisCMD(ctx context.Context, cmd string, args ...interface{}) (reply interface{}, err error) {
	conn, err := getConnFromPool(ctx)
	if err != nil {
		return nil, merr.Wrap(err, 0)
	}
	defer conn.Close()

	return wrapErrNil(conn.Do(cmd, args...))
}

func wrapErrNil(reply interface{}, err error) (interface{}, error) {
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, merr.WrapDepth(1, err, 0)
	}
	return reply, nil
}
