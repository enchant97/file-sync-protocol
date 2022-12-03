# Prototype 1

## Structure
### Packet

```
|--------------|----------|
| Type         | uint8    |
| Header Size  | uint64   |
| Header       | protobuf |
| Metadata     | protobuf |
| Payload Size | uint64   |
| Payload      | binary   |
|--------------|----------|
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
    Note over Client,Server: Send Groups
    loop Group Of Chunks
        Note over Client,Server: Send Chunks
        loop Chunks
            Client ->> Server: PSH
        end
        critical Proceed to Next Group
            Server -->> Client: ACK
        option Missing Chunk(s)
            Server -->> Client: REQ
        end
    end
    Note over Client,Server: Close Connection
    Client->>+ Server: FIN
    Server -->>- Client: ACK
```
