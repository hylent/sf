package clients

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

type Redis interface {
	Do(key string, cmd string, args ...interface{}) (interface{}, error)
	Get(key string) (string, error)
	Set(key string, value string) error
	Del(key string) error
}

type RedisClient struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Password     string `yaml:"password"`
	DbNo         int    `yaml:"db_no"`
	TimeoutMilli int    `yaml:"timeout_milli"`

	pool *redis.Pool
}

func (x *RedisClient) Init() error {
	addr := fmt.Sprintf("%s:%d", x.Host, x.Port)
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				addr,
				redis.DialDatabase(x.DbNo),
				redis.DialPassword(x.Password),
				redis.DialConnectTimeout(time.Millisecond*time.Duration(x.TimeoutMilli)),
				redis.DialReadTimeout(time.Millisecond*time.Duration(x.TimeoutMilli)),
				redis.DialWriteTimeout(time.Millisecond*time.Duration(x.TimeoutMilli)),
			)
		},
	}

	conn := pool.Get()
	defer conn.Close()

	if err := conn.Err(); err != nil {
		return err
	}

	if reply, err := redis.String(conn.Do("PING")); err != nil || reply != "PONG" {
		return fmt.Errorf("redis_ping_fail: err=%v reply=%s", err, reply)
	}

	x.pool = pool
	return nil
}

func (x *RedisClient) Close() error {
	return x.pool.Close()
}

func (x *RedisClient) Do(cmd string, args ...interface{}) (interface{}, error) {
	conn := x.pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		return nil, err
	}
	return conn.Do(cmd, args...)
}

func (x *RedisClient) Get(key string) (string, error) {
	return redis.String(x.Do(key, "GET", key))
}

func (x *RedisClient) Set(key string, value string) error {
	_, err := x.Do(key, "SET", key, value)
	return err
}

func (x *RedisClient) Del(key string) error {
	_, err := x.Do(key, "DEL", key)
	return err
}
