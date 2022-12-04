# Prototype 1

## Structure
### Packet

```
|-----------------|----------|
| Type            | uint8    |
| Header Length   | uint64   |
| Header          | protobuf |
| Metadata Length | uint64   |
| Metadata        | protobuf |
| Payload Length  | uint64   |
| Payload         | binary   |
|-----------------|----------|
```

### Packet Types

| Prefix | Value | Note                              |
| :----- | :-----| :-------------------------------- |
| SYN    | 1     | Open connection                   |
| ACK    | 2     | Acknowledge a request/action      |
| REQ    | 3     | Request to send or receive PSH    |
| PSH    | 4     | Send a payload (sent after a REQ) |
| FIN    | 254   | Close connection                  |


## Client File Push

```mermaid
sequenceDiagram
    Note over Client,Server: Init Connection
    Client->>+Server: SYN
    Server-->>-Client: ACK
    Note over Client,Server: Req Push
    Client ->>+ Server: REQ
    Server -->>- Client: ACK
    Note over Client,Server: Req Group
    loop Group Of Chunks
        critical
            alt No More
                Client ->> Server: REQ
            else Next Group
                Client ->> Server: REQ
            end
        option Ready/OK
            Server -->> Client: ACK
        option Resend Chunk(s)
            Server -->> Client: REQ
        end
        Note over Client,Server: Send Chunks
        loop Chunks
            Client ->> Server: PSH
        end
    end
    Note over Client,Server: Close Connection
    Client->>+ Server: FIN
    Server -->>- Client: ACK
```
