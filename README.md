# dkv

**A Log-Structured Hash Table for Fast Key/Value Data**

Installation (currently does not support windows environment)

```
By default, the golang environment has been installed.
Git clone https://github.com/xincao9/dkv.git
Cd ./dkv
Sudo make install
Command command: dkv
Configuration file: vim /usr/local/dkv/config.yaml
Data directory: cd /usr/local/dkv/data
```

Configuration file config.yaml

Placed in the current working directory || /etc/dkv/ || $HOME/.dkv || /usr/local/dkv

```
Data:
  Dir: /usr/local/dkv/data data directory
  invalidIndex: false Whether to rebuild the index when starting
  Cache: true Whether to enable caching
  Compress: false Whether to enable compression
Server:
  Mode: release
  Port: :9090 port
  Sequence: true
Logger:
  Level: info log level
```

interface

```
Add or modify
Curl -X PUT -H 'content-type:application/json' 'http://localhost:9090/kv' -d '{"k":"name", "v":"xincao9"}'
Search
Curl -X GET 'http://localhost:9090/kv/name'
delete
Curl -X DELETE 'http://localhost:9090/kv/name'
```

Management interface

```
View runtime configuration
Curl -X GET 'http://localhost:9090/config'
Prometheus indicator
Curl -X GET 'http://localhost:9090/metrics'
Pprof interface
Curl -X GET 'http://localhost:9090/debug/pprof/'
```

Grafana dashboard resources

```
Https://raw.githubusercontent.com/xincao9/dkv/master/prometheus.json
```

pressure test

```
Execute benchmark/start.sh
```
