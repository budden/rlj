// Package leftjoin defines two 'tables' and implements a 'left join' between them
package leftjoin

import (
	"fmt"
	"math/big"

	"github.com/go-redis/redis"

	"github.com/budden/rlj/pkg/redisclient"
	u "github.com/budden/rlj/pkg/rljutil"
)

// Run joins two hashes. Inspired by https://redis.io/topics/indexes
func Run() {
	rc, err := redisclient.MakeNewClient()
	u.FatalIf(err, "Failed to open redis client")

	err = flushAndFillDb(rc)
	u.FatalIf(err, "Failed to fill db")

	PrintJoinedOrders(rc)
}

func flushAndFillDb(rc *redis.Client) (err error) {
	_, err = rc.FlushDB().Result()
	if err != nil {
		return err
	}

	clients := []*Client{
		{ID: 1, Name: "Vasya"},
		{ID: 2, Name: "Маша"}}

	for _, cli := range clients {
		err = cli.Save(rc)
		if err != nil {
			return err
		}
	}

	orders := []*Order{
		{ID: 1, Clientid: 1, Product: "Car", Total: *big.NewRat(10000, 100)},
		{ID: 2, Clientid: 2, Product: "Dress", Total: *big.NewRat(5000, 100)},
		{ID: 3, Clientid: 2, Product: "Туфельки", Total: *big.NewRat(5000, 100)}}

	for _, ord := range orders {
		err = ord.Save(rc)
		if err != nil {
			return err
		}
	}
	return
}

// PrintJoinedOrders prints a sort of left join of clients to orders.
// There is a flaw here. We lock orders once for each client,
// so orders can change during our query. Also orders can change
// between we time we get clients and the time we get orders.
// But this is still an excercise :) In real world, we might lock the entier
// database, or lock both tables, or use multi-version architecture like
// in postgres.
func PrintJoinedOrders(rc *redis.Client) (err error) {
	var clients []*Client
	clients, err = GetAllClients(rc)
	u.FatalIf(err, "Failed to get clients")
	for _, client := range clients {
		var orders []*Order
		orders, err = GetOrdersByClientid(rc, client.ID)
		u.FatalIf(err, "Failed to get orders for client id %d", client.ID)
		for _, order := range orders {
			fmt.Printf("%v <=> %v\n", client, order)
		}
	}
	return
}
