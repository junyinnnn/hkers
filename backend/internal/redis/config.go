package redis

import (
	"crypto/tls"
	"os"
	"strconv"
	"time"

	redigo "github.com/gomodule/redigo/redis"
)

// Config holds Redis connection configuration.
type Config struct {
	Host                  string
	Port                  string
	Username              string
	Password              string
	DB                    int
	TLSEnabled            bool
	TLSInsecureSkipVerify bool
}

// LoadConfig loads Redis configuration from environment variables.
func LoadConfig() Config {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	redisTLSEnabled := getEnv("REDIS_TLS_ENABLED", "false") == "true"
	redisTLSInsecureSkipVerify := getEnv("REDIS_TLS_INSECURE_SKIP_VERIFY", "false") == "true"

	return Config{
		Host:                  getEnv("REDIS_HOST", "localhost"),
		Port:                  getEnv("REDIS_PORT", "6379"),
		Username:              getEnv("REDIS_USERNAME", ""),
		Password:              getEnv("REDIS_PASSWORD", ""),
		DB:                    redisDB,
		TLSEnabled:            redisTLSEnabled,
		TLSInsecureSkipVerify: redisTLSInsecureSkipVerify,
	}
}

// GetAddr returns the Redis address in host:port format.
func (r *Config) GetAddr() string {
	return r.Host + ":" + r.Port
}

// GetTLSConfig returns a TLS configuration when TLS is enabled, otherwise nil.
func (r *Config) GetTLSConfig() *tls.Config {
	if !r.TLSEnabled {
		return nil
	}

	return &tls.Config{
		InsecureSkipVerify: r.TLSInsecureSkipVerify,
	}
}

// NewRedisPool creates a redigo connection pool with TLS support if enabled.
func (r *Config) NewRedisPool() *redigo.Pool {
	return &redigo.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redigo.Conn, error) {
			opts := []redigo.DialOption{
				redigo.DialPassword(r.Password),
			}

			if r.TLSEnabled {
				opts = append(opts,
					redigo.DialTLSConfig(r.GetTLSConfig()),
					redigo.DialUseTLS(true),
				)
			}

			return redigo.Dial("tcp", r.GetAddr(), opts...)
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

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
