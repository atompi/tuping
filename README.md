# TUPing

A network connectivity testing tool. Supports testing via ICMP, TCP, and UDP.

## Build

```
go build -o ./tuping
```

## Usage

```
./tuping --help
A network connectivity testing tool. Supports testing via ICMP, TCP, and UDP.

Usage:
  tuping [flags]

Flags:
  -c, --count int         stop after <count> replies
  -d, --dns string        specify the dns server instead of using the system default dns server, tcp/udp protocol only
  -h, --help              help for tuping
  -i, --interval int      millisecond between sending each packet (default 1000)
  -p, --protocol string   specify protocol to use (default "icmp")
  -s, --size int          use <size> as number of data bytes to be sent (default 64)
  -t, --ttl int           define time to live (default 64)
  -v, --version           version for tuping
  -w, --wait              whether to wait for server response, should be set in udp protocol only, default false
```

- ping via icmp

```
./tuping gitee.com
```

- ping via tcp

```
./tuping -d 114.114.114.114 -p tcp gitee.com 443
```

- ping via udp

```
./tuping -p udp -w 192.168.15.128 53
```
