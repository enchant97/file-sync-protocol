# Prototype 1

## Discovered Issues
- If a large amount of packets are dropped during a PSH the REQ for resend packet will not be able to contain all chunk id's
  - Send chunks in groups, say 5 chunks at a time then ACK; then another 5?
- Header must be decoded before metadata can be interpreted.
  - Fix by combining header+metadata by combining types e.g. SYN-ACK and having optional fields in protobuf spec?
- Header Length & Metadata length have a reserved uint64 of space. This is wasted as we would never have a header which is 18446744073709551615 bytes long
  - Fix by reserving uint16 instead?

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
| :----- | :---- | :-------------------------------- |
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
    Note over Client,Server: Send Chunks
    loop Until ACK
        loop Send Next Chunk
            break No More Chunks
                Note over Server: Expected Chunk ID's
                Client ->> Server: REQ
            end
            Client ->> Server: PSH
        end
        alt Received All
            Server -->> Client: ACK
        else Resend Chunk(s)
            Note over Client: Missing Chunk ID's
            Server -->> Client: REQ
        end
    end
    Note over Client,Server: Close Connection
    Client->>+ Server: FIN
    Server -->>- Client: ACK
```
