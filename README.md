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
