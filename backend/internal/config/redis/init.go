package redis

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/redis/go-redis/v9"

	"hkers-backend/internal/config"
)

// GetTLSConfig returns a TLS configuration when TLS is enabled, otherwise nil.
func GetTLSConfig(redisConfig *config.RedisConfig) *tls.Config {
	if !redisConfig.TLSEnabled {
		return nil
	}

	return &tls.Config{
		InsecureSkipVerify: redisConfig.TLSInsecureSkipVerify,
	}
}

// NewRedisPool creates a redigo connection pool with TLS support if enabled.
func NewRedisPool(redisConfig *config.RedisConfig) *redigo.Pool {
	return &redigo.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redigo.Conn, error) {
			opts := []redigo.DialOption{
				redigo.DialPassword(redisConfig.Password),
			}

			if redisConfig.TLSEnabled {
				opts = append(opts,
					redigo.DialTLSConfig(GetTLSConfig(redisConfig)),
					redigo.DialUseTLS(true),
				)
			}

			return redigo.Dial("tcp", redisConfig.GetAddr(), opts...)
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

// InitRedis creates a Redis client and verifies connectivity with a PING.
func InitRedis(ctx context.Context, redisConfig *config.RedisConfig) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:      redisConfig.GetAddr(),
		Username:  redisConfig.Username,
		Password:  redisConfig.Password,
		DB:        redisConfig.DB,
		TLSConfig: GetTLSConfig(redisConfig),
	}

	log.Printf("Connecting to Redis at %s", opts.Addr)
	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	log.Printf("Redis connection established successfully")
	return client, nil
}
