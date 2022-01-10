#!/bin/bash
# 当接收到EXIT消息的时候，执行 rm server;kill 0命令，删掉server二进制文件，结束子进程，kill 0 检查进程是否存在，存在返回0，不存在返回1，EXIT信号也会被子进程捕获到
trap "rm server;kill 0" EXIT

# 编译为server的二级制文件
go build -o server

./server -port=8001 &
./server -port=8002 &
# 让第三个节点启动api服务器
./server -port=8003 -api=1 &

sleep 2

echo ">>> start test"
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &

# 等待所有子进程退出
wait