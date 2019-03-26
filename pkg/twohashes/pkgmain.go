// Package twohashes will demonstrate how to left join two data sets in hashes
package twohashes

import (
	"fmt"
	"log"
	"math/big"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"

	"github.com/budden/rlj/pkg/redisclient"
)

// FatalIf rasies fatal error if err is not nil
func FatalIf(err error, format string, args ...interface{}) {
	if err != nil {
		log.Fatal(errors.Wrapf(err, format, args...))
	}
}

// Run joins two hashes. Inspired by https://redis.io/topics/indexes
func Run() {
	rc, err := redisclient.MakeNewClient()
	FatalIf(err, "Failed to open redis client")

	err = flushAndFillDb(rc)
	FatalIf(err, "Failed to fill db")

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

// PrintJoinedOrders prints a join of clients to orders
func PrintJoinedOrders(rc *redis.Client) (err error) {
	var clientsStrings map[string]string
	clientsStrings, err = rc.HGetAll("client").Result()
	if err != nil {
		return
	}
	fmt.Print(clientsStrings)
	return
	//for clientID, clientString = range {
	//
	//}

}
