# dkv
Distributed key system

install

```
go get github.com/xincao9/dkv

/home/xincao9/go/bin/dkv
```

test

```
curl -X PUT -H 'content-type:application/json' 'http://localhost:9090/kv' -d '{"k":"name", "v":"xincao9"}'

curl -X GET 'http://localhost:9090/kv/name'

curl -X DELETE 'http://localhost:9090/kv/name'
```

config.yaml 放置到本地目录下，/etc/dkv/，$HOME/.dkv  目录下

```
data:
  dir: /tmp/dkv 数据目录
  invalidIndex: false 是否启动时重建索引
server:
  mode: debug 调试模式
  port: :9090 服务监听端口
  sequence: true 是否序列化执行所有命令
logger:
  level: info 日志输出登记
```

benchmark 

```
. benchmark/start.sh
```
