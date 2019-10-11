package cache

import (
	"github.com/dgraph-io/ristretto"
	"log"
)

var (
	C *ristretto.Cache
)

func init() {
	var err error
	C, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		log.Fatalf("Fatal error cache: %v\n", err)
	}
}

func Close () {
	C.Close()
}