## v9.1 客户端实现-建立连接
//编译运行
go build -o server server.go main.go
./server

go build client.go
./client
```
```go
//测试连接
nc 127.0.0.1 8888//ubuntu
telnet 192.168.56.105 8888//windows
```
![](./v9.1.png)