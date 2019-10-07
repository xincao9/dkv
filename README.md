# dkv

**分布式键值系统**

安装

```

go get github.com/xincao9/dkv

$GOPATH/bin/dkv
```

配置文件 config.yaml

放置到当前工作目录下 || /etc/dkv/ || $HOME/.dkv 下

```
data:
  dir: /tmp/dkv 数据目录
  invalidIndex: false 是否启动时重建索引
server:
  mode: debug 调试模式
  port: :9090 服务监听端口
  sequence: true 是否序列化执行所有命令
logger:
  level: info 日志输出级别
```

接口

```
增加或者修改
curl -X PUT -H 'content-type:application/json' 'http://localhost:9090/kv' -d '{"k":"name", "v":"xincao9"}'
检索
curl -X GET 'http://localhost:9090/kv/name'
删除
curl -X DELETE 'http://localhost:9090/kv/name'
```

压力测试

```
代码目录benchmark中文件复制到本地
执行 benchmark/start.sh
```
