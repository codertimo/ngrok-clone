# ngrok-clone
clone version of ngrok with @GBS-Skile (GO lang side project)

## Getting Started

```shell
$ go run ngrok-server.go
2021/08/09 03:29:06 Remote Control Address: localhost:5000
2021/08/09 03:29:06 Remote Data Address: localhost:5001
2021/08/09 03:29:06 Local Address: localhost:4000

# 다른 쉘에서
$ go run webserver.go

# 다른 쉘에서
$ go run ngrok-client.go

# 다른 쉘에서
$ curl localhost:4000/chicken
Hi there, I love chicken!
```
