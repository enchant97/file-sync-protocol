# FTP
- Data can be transferred as stream or block mode
- Uses telnet strings for commands
- Simple packet structure
- Requires two ports on a client, one for sending and the other for receiving
- Maintains a stateful connection
- https://datatracker.ietf.org/doc/html/rfc959

```mermaid
%% FTP Connect & Disconnect
sequenceDiagram
    Note over Client: Connect
    Client->Server: TCP Handshake
    Server-->>Client: 220 (Ready)
    Client->>Server: Username
    Server-->>Client: 230 (Logged In)
    Client->>Server: Port Details
    Server-->>Client: 200 (OK)
    Note over Client: Read/Write Messages
    Client->Server: ...
    Note over Client: Disconnect
    Client->>Server: Quit
    Server->Client: TCP Close
```
