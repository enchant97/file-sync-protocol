# Prototype 1

## Structure
### Packet

```
|----------|
| Type     |
| Header   |
| Metadata |
| Payload  |
|----------|
```

### Packet Types

| Prefix | Note                              |
| :----- | :-------------------------------- |
| SYN    | Open connection                   |
| ACK    | Acknowledge a request/action      |
| REQ    | Request to send or receive PSH    |
| PSH    | Send a payload (sent after a REQ) |
| FIN    | Close connection                  |


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
