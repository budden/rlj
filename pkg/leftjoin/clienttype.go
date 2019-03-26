package leftjoin

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

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (c *Client) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// GetAllClients returns a slice of all clients
func GetAllClients(rc *redis.Client) (cArray []*Client, err error) {
	var hashContents map[string]string
	hashContents, err = rc.HGetAll("client").Result()
	if err != nil {
		return
	}
	for _, str := range hashContents {
		c := &Client{}
		binary := []byte(str)
		err = c.UnmarshalBinary(binary)
		if err != nil {
			cArray = []*Client{}
			return
		}
		cArray = append(cArray, c)
	}
	return
}
