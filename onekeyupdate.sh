#!/bin/sh

export GO111MODULE=on
export GOPROXY=https://goproxy.io
go build -o bin/go_gateway
ps aux | grep go_gateway | grep -v 'grep' | awk '{print $2}' | xargs kill
nohup ./bin/go_gateway -config=./conf/prod/ -endpoint=dashboard >> logs/dashboard.log 2>&1 &
echo 'nohup ./bin/go_gateway -config=./conf/prod/ -endpoint=dashboard >> logs/dashboard.log 2>&1 &'
nohup ./bin/go_gateway -config=./conf/prod/ -endpoint=server >> logs/server.log 2>&1 &
echo 'nohup ./bin/go_gateway -config=./conf/prod/ -endpoint=server >> logs/server.log 2>&1 &'