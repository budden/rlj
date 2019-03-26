package twohashes

import (
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
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
	var binary []byte
	binary, err = c.MarshalBinary()
	if err != nil {
		err = errors.Wrapf(err, "Failed to marshal an Order")
		return
	}

	// We modify several coordinated structures, so we need a sort of lock
	err = WithNxLock(rc, "order", func() (err1 error) {

		_, err1 = rc.HSet("order", strconv.Itoa(c.ID), binary).Result()
		if err1 != nil {
			return
		}

		_, err1 = rc.HSet("order-by-clientid", strconv.Itoa(c.Clientid), binary).Result()
		return
	})
	return
}

// MarshalBinary implements encoding.BinaryMarshaler
func (c *Order) MarshalBinary() (data []byte, err error) {
	data, err = json.Marshal(c)
	return
}
