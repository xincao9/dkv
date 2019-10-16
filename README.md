# dkv

**分布式键值系统**

安装（目前暂不支持windows环境）

```
git clone https://github.com/xincao9/dkv.git
cd ./dkv
sudo make install
命令命令: dkv
配置文件: vim /usr/local/dkv/config.yaml
数据目录: cd /usr/local/dkv/data
```

配置文件 config.yaml

放置到当前工作目录下 || /etc/dkv/ || $HOME/.dkv || /usr/local/dkv 下

```
data:
  dir: /usr/local/dkv/data 数据目录
  invalidIndex: false 是否启动时重建索引
  cache: true 是否启用缓存
  compress: false　是否启用压缩
server:
  mode: release
  port: :9090　端口
  sequence: true
logger:
  level: info　日志级别
```

接口

```
增加或者修改
curl -X PUT -H 'content-type:application/json' 'http://localhost:9090/kv' -d '{"k":"name", "v":"xincao9"}'
检索
curl -X GET 'http://localhost:9090/kv/name'
删除
curl -X DELETE  'http://localhost:9090/kv/name'
```

管理接口

```
查看运行时配置
curl -X GET 'http://localhost:9090/config'
普罗米修斯指标
curl -X GET 'http://localhost:9090/metrics'
pprof接口
curl -X GET 'http://localhost:9090/debug/pprof/'
```

Grafana dashboard资源

```
https://raw.githubusercontent.com/xincao9/dkv/master/prometheus.json
```

压力测试

```
代码目录benchmark中文件复制到本地
执行 benchmark/start.sh
```
