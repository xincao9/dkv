package client

import (
    "dkv/client/ms"
    "log"
    "math/rand"
    "strconv"
    "time"
)

const maxRequestCount = 1000000

var doc = make([]byte, 1024)

func Benchmark() {
	cli, err := ms.NewMS("localhost:8090", []string{"localhost:8091"}, time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	startTime := time.Now()
	for i := 0; i < maxRequestCount; i++ {
		key := strconv.Itoa(i)
		_, err := cli.Put(key, string(doc))
		if err != nil {
			log.Printf("client put(%s, %s): %v\n", key, doc, err)
		}
	}
	log.Printf("client put TPS: %.2f\n", maxRequestCount/time.Since(startTime).Seconds())
	startTime = time.Now()
	for i := 0; i < maxRequestCount; i++ {
		key := strconv.Itoa(i)
		_, err := cli.Get(key)
		if err != nil {
			log.Printf("client get(%s): %v\n", key, err)
		}
	}
	log.Printf("client sort get TPS: %.2f\n", maxRequestCount/time.Since(startTime).Seconds())
	startTime = time.Now()
	for i := 0; i < maxRequestCount; i++ {
		key := strconv.Itoa(rand.Intn(maxRequestCount))
		_, err := cli.Get(key)
		if err != nil {
			log.Printf("client get(%s): %v\n", key, err)
		}
	}
	log.Printf("client random get TPS: %.2f\n", maxRequestCount/time.Since(startTime).Seconds())
}
