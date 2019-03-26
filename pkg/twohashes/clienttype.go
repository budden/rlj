package twohashes

import (
	"encoding/json"
	"strconv"

	"github.com/go-redis/redis"
)

// Client is used to serialize client (purchaser)
type Client struct {
	ID   int
	Name string
}

// Save writes client to the redis db
func (c *Client) Save(rc *redis.Client) (err error) {
	_, err = rc.HSet("client", strconv.Itoa(c.ID), c).Result()
	return
}

// MarshalBinary implements encoding.BinaryMarshaler
func (c *Client) MarshalBinary() (data []byte, err error) {
	data, err = json.Marshal(c)
	return
}
