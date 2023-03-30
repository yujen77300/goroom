package models

import (
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
)

var RedisDefaultPool *redis.Pool

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}

	REDISENDPOINT := viper.GetString("REDISENDPOINT")
	RedisDefaultPool = newPool(REDISENDPOINT)
}
