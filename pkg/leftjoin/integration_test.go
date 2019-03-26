package leftjoin

import (
	"math/big"
	"testing"

	"github.com/budden/rlj/pkg/redisclient"
	"github.com/stretchr/testify/assert"
)

func ordersEqual(o, o2 *Order) bool {
	return o.ID == o2.ID &&
		o.Clientid == o2.Clientid &&
		o.Product == o2.Product &&
		o.Total.Cmp(&o2.Total) == 0
}

// Beware that this test flushes and fills the DB!
func TestSetAndGetOneOrder(t *testing.T) {
	rc, err := redisclient.MakeNewClient()
	assert.NoErrorf(t, err, "Failed to connect to redis")
	if err != nil {
		return
	}
	_, err = rc.FlushDB().Result()
	assert.NoErrorf(t, err, "Failed to flushDb")
	if err != nil {
		return
	}
	o := &Order{ID: 1, Clientid: 1, Product: "Mashinka", Total: *big.NewRat(1000, 100)}
	err = o.Save(rc)
	assert.NoErrorf(t, err, "Failed to Save")
	if err != nil {
		return
	}
	var o2 *Order
	o2, err = GetOrderByID(rc, 1)
	assert.NoErrorf(t, err, "Failed to Get")
	if err != nil {
		return
	}
	assert.True(t, ordersEqual(o, o2), "Saved value %v does not match retrieved value %v", o, o2)

	o2, err = GetOrderByID(rc, 2)
	assert.Nilf(t, o2, "Found value %v for non-existing key", o2)

	err = rc.Close()
	assert.NoErrorf(t, err, "Failed to Close")
	return
}
