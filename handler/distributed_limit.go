package handler

import (
	"github.com/e421083458/go_gateway/public"
	"github.com/garyburd/redigo/redis"
	"math"
	"time"
)

type DistributedLimiter struct {
	Name     string
	Dtype    int //0=qps 1=qpm 2=qph
	Rate     int64
	Capacity int64
}

func NewDistributedLimiter(name string, dtype int, rate, capacity int64) *DistributedLimiter {
	if dtype == 1 {
		rate = capacity / 60
	}
	if dtype == 2 {
		rate = capacity / 3600
	}
	if rate < 1 {
		rate = 1
	}
	return &DistributedLimiter{
		Name:     name,
		Dtype:    dtype,
		Rate:     rate,
		Capacity: capacity,
	}
}

func (d *DistributedLimiter) Allow() bool {
	fillTime := float64(d.Capacity) / float64(d.Rate)
	ttl := math.Floor(fillTime * 2)
	redisKey := public.DistributedLimiterPrefix + d.Name
	redisMap, _ := redis.Int64Map(public.RedisConfDo("HGETALL", redisKey))
	lastTokens, ok := redisMap["tokens"]
	if !ok {
		lastTokens = lastTokens
	}
	lastRefreshed, ok := redisMap["timestamp"]
	if !ok {
		lastRefreshed = 0
	}
	delta := math.Max(0, float64(time.Now().Unix()-lastRefreshed))
	filledTokens := math.Min(float64(d.Capacity), float64(lastTokens)+(delta*float64(d.Rate)))
	allowed := false
	newTokens := filledTokens
	if filledTokens >= 1 {
		allowed = true
		newTokens = filledTokens - 1
	}
	public.RedisConfDo("HMSET", redisKey, "tokens", newTokens, "timestamp", )
	public.RedisConfDo("EXPIRE", redisKey, ttl)
	return allowed
}
