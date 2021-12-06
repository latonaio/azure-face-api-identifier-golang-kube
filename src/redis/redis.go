package redis

import (
	"azure-face-api-identifier-golang-kube/src/model"

	"github.com/gomodule/redigo/redis"
)

// Connection
func Connection(redisDSN string) redis.Conn {
	Addr := redisDSN

	c, err := redis.Dial("tcp", Addr)
	if err != nil {
		panic(err)
	}
	return c
}

// データの登録(Redis: SET key value)
func SetData(key string, value *model.InsertToRedis, c redis.Conn) error {
	_, err := redis.String(c.Do("HMSET", key, "status", value.Status, "customer", value.Customer, "guest_id", value.GuestID, "failed_ms", value.FailedMS, "image_path", value.ImagePath))
	if err != nil {
		panic(err)
	}
	return err
}
