package redisclient

import (
	"testing"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

func TestSuccessfulConnect(t *testing.T) {
	c, err := MakeNewClient()
	assert.Nilf(t, err, "Unexpected error while connecting to redis: %#v", err)
	err = c.Close()
	assert.Nilf(t, err, "Unexpected error while disconnecting from redis: %#v", err)
}

func TestUnsuccessfulConnect(t *testing.T) {
	c, err := MakeNewClient(&redis.Options{Addr: "example.com:100500"})
	assert.Errorf(t, err, "There must be no redis at example.com")
	err = c.Close()
	assert.Nilf(t, err, "Even if connect was unsuccessful, close should not err")
}

func TestInvalidCall(t *testing.T) {
	c, err := MakeNewClient(
		&redis.Options{Addr: "example.com:100500"},
		&redis.Options{Addr: "example.com:100500"})
	assert.Errorf(t, err, "Must not accept two option sets")
	assert.Containsf(t, err.Error(),
		"expects no more than one",
		"Error must mention that there are multiple option sets")
	assert.Nilf(t, c, "Client must be null")
}
