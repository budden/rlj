package twohashes

import (
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/go-redis/redis"
)

// Order is used to serialize purchase
type Order struct {
	ID       int
	Clientid int
	Product  string
	// Currency as represented as Rational https://play.golang.org/p/vQ6hT-zcYt
	Total big.Rat
}

// Save writes order to the redis db
func (c *Order) Save(rc *redis.Client) (err error) {
	_, err = rc.HSet("order", strconv.Itoa(c.ID), c).Result()
	return
}

// MarshalBinary implements encoding.BinaryMarshaler
func (c *Order) MarshalBinary() (data []byte, err error) {
	data, err = json.Marshal(c)
	return
}
