# dkv

**对象存储 - 日志结构哈希表**

[![CodeFactor](https://www.codefactor.io/repository/github/xincao9/dkv/badge)](https://www.codefactor.io/repository/github/xincao9/dkv)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/e062787e83ab41c387e567f5210d4cc4)](https://www.codacy.com/manual/xincao9/dkv?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=xincao9/dkv&amp;utm_campaign=Badge_Grade)

![logo](https://github.com/xincao9/dkv/blob/master/logo.png)

* 读取和写入较低的延迟
* 高吞吐量
* 支持存储海量数据
* 崩溃后数据修复友好
* 支持数据的备份和还原

##安装指导

###本地编译

```
git clone https://github.com/xincao9/dkv.git
cd ./dkv
sudo make install
```

**/usr/local/dkv/conf/config.yaml 配置说明**

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

**目录说明**

```
执行文件: /usr/local/dkv/bin/dkv
配置文件目录: /usr/local/dkv/conf/
数据目录: /usr/local/dkv/data/
日志目录: /usr/local/dkv/log/
```

**执行命令**

```
dkv -d=true -conf=config-prod.yaml
```

###容器化部署

```
docker pull xincao9/dkv
docker run -d -p 9090:9090 -p 6380:6380 dkv:latest
```

##接口说明

###HTTP接口

**键值存储**

1. 增加或修改

    ```
    curl -X PUT -H 'content-type:application/json' 'http://localhost:9090/kv' -d '{"k":"name", "v":"xincao9"}'
    ```
2. 查询

    ```
    curl -X GET 'http://localhost:9090/kv/name'
    ```
3. 删除

    ```
    curl -X DELETE 'http://localhost:9090/kv/name'
    ```

**对象存储**

1. 上传对象，最大64M

    ```
    curl -X POST 'http://localhost:9090/oss' -F "file[]=@config.yaml" -H 'content-type:multipart/form-data'
    ```
2. 读取文件

    ```
    curl -X GET 'http://localhost:9090/oss/116a71ebd837470652f063028127c5cd'
    ```

###REDIS 支持命令


* SET key value
* GET key
* DEL key
* PING

```
go get github.com/go-redis/redis
```

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

###GO SDK接入

```
go get github.com/xincao9/dkv/client
```

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

###管理接口

1. 运行时配置

    ```
    curl -X GET 'http://localhost:9090/config'
    ```
2. 普罗米修斯指示器

    ```
    curl -X GET 'http://localhost:9090/metrics'
    ```
3. pprof 接口

    ```
    curl -X GET 'http://localhost:9090/debug/pprof'
    ```

**Grafana dashboard 资源**

> [prometheus.json](https://github.com/xincao9/dkv/blob/master/resource/prometheus.json)

##压力测试

```
执行: benchmark/start.sh
```

##参考

* [bitcask-intro](https://github.com/xincao9/dkv/blob/master/resource/bitcask-intro.pdf)
