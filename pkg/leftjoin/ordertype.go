package leftjoin

import (
	"encoding/json"
	"fmt"
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
func (o *Order) Save(rc *redis.Client) (err error) {
	var binary []byte
	binary, err = o.MarshalBinary()
	if err != nil {
		err = errors.Wrapf(err, "Failed to marshal an Order")
		return
	}

	// We modify several coordinated structures, so we need a sort of lock
	err = WithNxLock(rc, "order", func() (err1 error) {

		_, err1 = rc.HSet("order", strconv.Itoa(o.ID), binary).Result()
		if err1 != nil {
			return
		}

		// each order-by-clientid.$Clientid hash is a bucket storing
		// copies of Order indexed by Order.ID
		bucketKey := fmt.Sprintf("order-by-clientid.%d", o.Clientid)
		_, err1 = rc.HSet(bucketKey, strconv.Itoa(o.ID), binary).Result()
		return
	})
	return
}

// MarshalBinary implements encoding.BinaryMarshaler
func (o *Order) MarshalBinary() (data []byte, err error) {
	data, err = json.Marshal(o)
	return
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *Order) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, o)
}

// GetOrderByID loads an order from database, using ID. If no order found, error is returned
func GetOrderByID(rc *redis.Client, ID int) (o *Order, err error) {
	// To keep data consistent, we always need a lock
	err = WithNxLock(rc, "order", func() (err1 error) {
		var exists bool
		IDAsString := strconv.Itoa(ID)
		exists, err1 = rc.HExists("order", IDAsString).Result()
		if err1 != nil {
			return
		}
		if !exists {
			err1 = fmt.Errorf("Order «%s» not found", IDAsString)
			return
		}
		var str string
		str, err1 = rc.HGet("order", IDAsString).Result()
		if err1 != nil {
			return
		}
		binary := []byte(str)
		o = &Order{}
		err1 = o.UnmarshalBinary(binary)
		if err1 != nil {
			o = nil
			return
		}
		return
	})
	return
}

// GetOrdersByClientid uses an index to get all clients
func GetOrdersByClientid(rc *redis.Client, Clientid int) (oArray []*Order, err error) {
	// To keep data consistent, we always need a lock
	err = WithNxLock(rc, "order", func() (err1 error) {
		bucketKey := fmt.Sprintf("order-by-clientid.%d", Clientid)
		var bucketContents map[string]string
		bucketContents, err1 = rc.HGetAll(bucketKey).Result()
		if err1 != nil {
			return
		}
		for _, str := range bucketContents {
			o := &Order{}
			binary := []byte(str)
			err1 = o.UnmarshalBinary(binary)
			if err1 != nil {
				oArray = []*Order{}
				return
			}
			oArray = append(oArray, o)
		}
		return
	})
	return
}
