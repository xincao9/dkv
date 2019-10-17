# dkv

**A Log-Structured Hash Table for Fast Key/Value Data**

Installation (currently does not support windows environment)

```
By default, the golang environment has been installed.
git clone https://github.com/xincao9/dkv.git
cd ./dkv
sudo make install
Execute: dkv
Configuration file: vim /usr/local/dkv/config.yaml
Data directory: cd /usr/local/dkv/data
```

Configuration file config.yaml

Placed in the current working directory || /etc/dkv/ || $HOME/.dkv || /usr/local/dkv

```
data:
  dir: /usr/local/dkv/data #data directory
  invalidIndex: false #Whether to rebuild the index when starting
  cache: true #Whether to enable caching
  compress: false #Whether to enable compression
server:
  mode: release
  port: :9090 #port
  sequence: true
logger:
  level: info #log level
```

Interface

```
Add or modify
curl -X PUT -H 'content-type:application/json' 'http://localhost:9090/kv' -d '{"k":"name", "v":"xincao9"}'
Search
curl -X GET 'http://localhost:9090/kv/name'
delete
curl -X DELETE 'http://localhost:9090/kv/name'
```

Management interface

```
View runtime configuration
curl -X GET 'http://localhost:9090/config'
Prometheus indicator
curl -X GET 'http://localhost:9090/metrics'
Pprof interface
curl -X GET 'http://localhost:9090/debug/pprof/'
```

Grafana dashboard resources

```
https://raw.githubusercontent.com/xincao9/dkv/master/prometheus.json
```

Pressure test

```
Execute benchmark/start.sh
```

Reference

* [bitcask-intro](https://github.com/xincao9/dkv/blob/master/bitcask-intro.pdf)