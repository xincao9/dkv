# dkv

**A Log-Structured Hash Table for Fast Key/Value Data** 

[![CodeFactor](https://www.codefactor.io/repository/github/xincao9/dkv/badge)](https://www.codefactor.io/repository/github/xincao9/dkv)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/e062787e83ab41c387e567f5210d4cc4)](https://www.codacy.com/manual/xincao9/dkv?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=xincao9/dkv&amp;utm_campaign=Badge_Grade)

![logo](https://github.com/xincao9/dkv/blob/master/logo.png)

* low latency per item read or written
* high throughput, especially when writing an incoming stream of random items
* ability to handle datasets much larger than RAM w/o degradation
* crash friendliness, both in terms of fast recovery and not losing data
* ease of backup and restore
* a relatively simple, understandable (and thus supportable) code structure and data format • predictable behavior under heavy access load or large volume

**Install**

> By default, the golang environment has been installed.

```
first:
git clone https://github.com/xincao9/dkv.git
cd ./dkv
sudo make install
Execute: dkv -d=true -conf=config-prod.yaml

Bin directory: cd /usr/local/dkv/bin/
Configuration directory: cd /usr/local/dkv/conf/
Data directory: cd /usr/local/dkv/data/
Logger directory: cd /usr/local/dkv/log/

second:
docker pull xincao9/dkv
docker run -d -p 9090:9090 -p 6380:6380 dkv:latest
```

**Configuration file**

> config.yaml Placed in the current working directory or /etc/dkv/ or $HOME/.dkv or /usr/local/dkv

```
data:
    dir: /usr/local/dkv/data
    invalidIndex: false
    cache:
        open: true
        size: 1073741824
    compress:
        open: false
server:
    mode: release
    port: 9090
    sequence: true
    redis:
        port: 6380
logger:
    level: info
    dir: /usr/local/dkv/log
ms:
    role: 0
```

**HTTP interface**

> KV store

```
Add or modify
curl -X PUT -H 'content-type:application/json' 'http://localhost:9090/kv' -d '{"k":"name", "v":"xincao9"}'
Search
curl -X GET 'http://localhost:9090/kv/name'
delete
curl -X DELETE 'http://localhost:9090/kv/name'
```

> OSS (Object Storage Service)

```
Upload file, file max size 64M
curl -X POST 'http://localhost:9090/oss' -F "file[]=@config.yaml" -H 'content-type:multipart/form-data'
Fetch file
curl -X GET 'http://localhost:9090/oss/116a71ebd837470652f063028127c5cd'
```

**Redis command**


* SET key value
* GET key
* DEL key
* PING

> go get github.com/go-redis/redis

```
client := redis.NewClient(&redis.Options{
    Addr:     "localhost:6380",
    Password: "", // no password set
    DB:       0,  // use default DB
})
err := client.Set("name", "xincao9", 0).Err()
if err != nil {
    log.Println(err)
}
val, err := client.Get("name").Result()
if err != nil {
    log.Println(err)
}
log.Println(val)
```

**GO SDK**

> go get github.com/xincao9/dkv/client

```
c, err := client.New("localhost:9090", time.Second)
if err != nil {
    log.Fatalln(err)
}
r, err := c.Put("name", "xincao9")
if err == nil {
    log.Println(r)
}
r, err = c.Get("name")
if err == nil {
    log.Println(r)
}
```

**Management interface**

```
View runtime configuration
curl -X GET 'http://localhost:9090/config'
Prometheus indicator
curl -X GET 'http://localhost:9090/metrics'
Pprof interface
curl -X GET 'http://localhost:9090/debug/pprof'
```

**Grafana dashboard resources**

> [prometheus.json](https://raw.githubusercontent.com/xincao9/dkv/master/prometheus.json)

**Pressure test**

```
Execute benchmark/start.sh
```

**Reference**

* [bitcask-intro](https://github.com/xincao9/dkv/blob/master/bitcask-intro.pdf)
