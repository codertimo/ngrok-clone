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

## TLS Support

```shell
# Reference: https://xshine.tistory.com/316

$ openssl genrsa -out server.key 2048
Generating RSA private key, 2048 bit long modulus (2 primes)
............+++++
.........+++++
e is 65537 (0x010001)

$ openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
Country Name (2 letter code) [AU]:KR
State or Province Name (full name) [Some-State]:.
Locality Name (eg, city) []:.
Organization Name (eg, company) [Internet Widgits Pty Ltd]:.
Organizational Unit Name (eg, section) []:.
Common Name (e.g. server FQDN or YOUR name) []:.
Email Address []:.

$ go run ngrok-server.go -cert=server.crt -key=server.key

# 다른 쉘에서 (-k 옵션은 insecure)
$ go run ngrok-client.go -t -k
```
