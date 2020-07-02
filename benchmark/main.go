package main

import (
    "dkv/benchmark/client"
    "dkv/benchmark/redis"
    "flag"
)

func main() {
    c := flag.Bool("c", false, "use master and slave mode")
    flag.Parse()
    if *c {
        client.Benchmark()
        return
    }
    redis.Benchmark()
}
