package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	otel "github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"

	"gofr.dev/pkg/gofr/config"
	"gofr.dev/pkg/gofr/datasource"
)

const (
	redisPingTimeout = 5 * time.Second
	defaultRedisPort = 6379
)

type Config struct {
	HostName string
	Username string
	Password string
	Port     int
	DB       int
	Options  *redis.Options
}

type Redis struct {
	*redis.Client
	logger datasource.Logger
	config *Config
}

// NewClient return a redis client if connection is successful based on Config.
// In case of error, it returns an error as second parameter.
func NewClient(c config.Config, logger datasource.Logger, metrics Metrics) *Redis {
	redisConfig := getRedisConfig(c)

	// if Hostname is not provided, we won't try to connect to Redis
	if redisConfig.HostName == "" {
		return nil
	}

	logger.Debugf("connecting to redis at '%s:%d' on database %d", redisConfig.HostName, redisConfig.Port, redisConfig.DB)

	rc := redis.NewClient(redisConfig.Options)
	rc.AddHook(&redisHook{config: redisConfig, logger: logger, metrics: metrics})

	ctx, cancel := context.WithTimeout(context.TODO(), redisPingTimeout)
	defer cancel()

	if err := rc.Ping(ctx).Err(); err == nil {
		if err = otel.InstrumentTracing(rc); err != nil {
			logger.Errorf("could not add tracing instrumentation, error: %s", err)
		}

		logger.Infof("connected to redis at %s:%d on database %d", redisConfig.HostName, redisConfig.Port, redisConfig.DB)
	} else {
		logger.Errorf("could not connect to redis at '%s:%d' , error: %s", redisConfig.HostName, redisConfig.Port, err)
	}

	return &Redis{Client: rc, config: redisConfig, logger: logger}
}

// Close shuts down the Redis client, ensuring the current dataset is saved before exiting.
func (r *Redis) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}

	return nil
}

func getRedisConfig(c config.Config) *Config {
	var redisConfig = &Config{}

	redisConfig.HostName = c.Get("REDIS_HOST")

	redisConfig.Username = c.Get("REDIS_USER")

	redisConfig.Password = c.Get("REDIS_PASSWORD")

	port, err := strconv.Atoi(c.Get("REDIS_PORT"))
	if err != nil {
		port = defaultRedisPort
	}

	redisConfig.Port = port

	db, err := strconv.Atoi(c.Get("REDIS_DB"))
	if err != nil {
		db = 0 // default to DB 0 if not specified
	}

	redisConfig.DB = db

	options := new(redis.Options)

	if options.Addr == "" {
		options.Addr = fmt.Sprintf("%s:%d", redisConfig.HostName, redisConfig.Port)
	}

	if options.Username == "" {
		options.Username = redisConfig.Username
	}

	if options.Password == "" {
		options.Password = redisConfig.Password
	}

	options.DB = redisConfig.DB

	redisConfig.Options = options

	return redisConfig
}

// TODO - if we make Redis an interface and expose from container we can avoid c.Redis(c, command) using methods on c and still pass c.
// type Redis interface {
//	Get(string) (string, error)
// }
