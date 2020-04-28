package main

import(
	"github.com/go-redis/redis"
)

type redisClient struct{
	Id     string
	Client *redis.Client
}


func (r *redisClient)Info() (string, error){
	return r.Client.Info().Result()
}
