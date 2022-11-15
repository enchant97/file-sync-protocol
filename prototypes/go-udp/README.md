# Go UDP
Prototype Go UDP server & client to send a file over the network.

- No error handling
- No rate limiting (packets will be sent even if server cannot keep up)

## Usage
## Server
```
go run . server <ip>:<port> <output-filepath>
```

### Client
```
go run . client <ip>:<port> <input-filepath>
```

- [Source](https://www.linode.com/docs/guides/developing-udp-and-tcp-clients-and-servers-in-go/)
