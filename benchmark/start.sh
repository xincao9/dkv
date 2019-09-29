#!/bin/bash

f=$(which wrk)

echo ""

if [[ -z "$f" ]];then
	echo "请自行安装wrk,或者wrk添加到PATH路径下"
	exit 0
fi

echo ""
echo "put 压力测试"
echo ""

wrk -t1 -c2 -d300s -s put.lua 'http://localhost:9090/kv'

echo ""
echo "get 压力测试"
echo ""

wrk -t1 -c2 -d300s -s get.lua 'http://localhost:9090/kv'
