// Package redisclient is a minimal wrapper around github.com/go-redis/redis
package redisclient

import (
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// MakeNewClient makes a redis client. One can pass options. Null options will
// produce default client. Also, MakeNewClient pings the client and informs us
// on any error. Note that redis.NewClient does not ping the client so you will
// not know about error until you start to execute commands
func MakeNewClient(opts ...*redis.Options) (c *redis.Client, err error) {
	if len(opts) == 0 {
		opts = []*redis.Options{{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		}}
	} else if len(opts) > 1 {
		err = errors.New("MakeNewClient expects no more than one redis.Options")
		return
	}

	c = redis.NewClient(opts[0])

	// Ping to check if the client is here
	_, err = c.Ping().Result()

	return c, err
}
