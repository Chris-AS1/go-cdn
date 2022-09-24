package main

import (
	"context"

	"github.com/go-redis/redis/v9"
)

var (
	ctx        = context.Background()
	redis_addr = "localhost:6379"
)

func PingRedis() string {
	// log.Print("Connecting to Redis...")
	rdb := redis.NewClient(&redis.Options{
		Addr:     redis_addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return rdb.Ping(ctx).String()

	// err := rdb.Set(ctx, "key", "value", 0).Err()
	// if err != nil {
	// 	panic(err)
	// }

	// val, err := rdb.Get(ctx, "key").Result()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("key", val)

	// val2, err := rdb.Get(ctx, "key2").Result()
	// if err == redis.Nil {
	// 	fmt.Println("key2 does not exist")
	// } else if err != nil {
	// 	panic(err)
	// } else {
	// 	fmt.Println("key2", val2)
	// }
}
