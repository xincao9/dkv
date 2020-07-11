package client

import (
	"errors"
    "hash/crc32"
    "time"
)

type Partition struct {
	Id     int
	Master string
	Slaves []string
}

type cluster struct {
	clients map[int]*Client
}

func NewCluster(partitions []*Partition, timeout time.Duration) (*cluster, error) {
	if len(partitions) <= 0 {
		return nil, errors.New("partitions is required non empty")
	}
	clients := make(map[int]*Client)
	for _, p := range partitions {
		var c *Client
		var err error
		if len(p.Slaves) <= 0 {
			c, err = New(p.Master, timeout)
		} else {
			c, err = NewMS(p.Master, p.Slaves, timeout)
		}
		if err != nil {
			return nil, err
		}
		clients[p.Id] = c
	}
	return &cluster{
		clients: clients,
	}, nil
}

func (c *cluster) Get (key string) (*Result, error) {
    size := len(c.clients)
    hash := (int)(crc32.ChecksumIEEE([]byte(key)))
    return c.clients[hash % size].Get(key)
}

func (c *cluster) Put (key string, value string) (*Result, error) {
    size := len(c.clients)
    hash := (int)(crc32.ChecksumIEEE([]byte(key)))
    return c.clients[hash % size].Put(key, value)
}

func (c *cluster) Delete (key string) (*Result, error) {
    size := len(c.clients)
    hash := (int)(crc32.ChecksumIEEE([]byte(key)))
    return c.clients[hash % size].Delete(key)
}
