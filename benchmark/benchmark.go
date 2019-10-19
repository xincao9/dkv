package main

import (
	"github.com/go-redis/redis/v7"
	"log"
	"math/rand"
	"strconv"
	"time"
)

const maxRequestCount = 1000000

func main () {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6380",
		Password: "",
		DB:       0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		log.Fatalln(err)
	}
	startTime := time.Now()
	val := ""
	for i := 0; i < maxRequestCount; i++ {
		val := strconv.Itoa(i)
		err = client.Set(val, val, 0).Err()
		if err != nil {
			log.Printf("redis set(%s, %s): %v\n", val, val, err)
		}
	}
	log.Printf("redis set TPS: %.2f\n", maxRequestCount/time.Since(startTime).Seconds())
	startTime = time.Now()
	for i := 0; i < maxRequestCount; i++ {
		val = strconv.Itoa(i)
		_, err := client.Get(val).Result()
		if err != nil {
			log.Printf("redis get(%s): %v\n", val, err)
		}
	}
	log.Printf("redis sort get TPS: %.2f\n", maxRequestCount/time.Since(startTime).Seconds())
	startTime = time.Now()
	for i := 0; i < maxRequestCount; i++ {
		val = strconv.Itoa(rand.Intn(maxRequestCount))
		_, err := client.Get(val).Result()
		if err != nil {
			log.Printf("redis get(%s): %v\n", val, err)
		}
	}
	log.Printf("redis random get TPS: %.2f\n", maxRequestCount/time.Since(startTime).Seconds())
}
