# SMB2
- Block-Level transfer, (however block size is not limited unlike SMB1)
- Utilises pipelining, sending other requests before a response is received
- Can send multiple actions in a single request
- Only 19 commands
- When running over TCP requires one port making use of TCP's full-duplex capability
- Maintains stateful  connection
- Complex connection negotiation, requiring multiple message exchanges before a transfer can be accomplished
    1.  Negotiate (establish which protocol version is being used)
    2. Session Setup (auth tokens and such)
    3. Tree Connect (connect to a share)
    4. ...
    5. Tree Disconnect
    6. Logoff
- [Source](https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/5606ad47-5ee0-437a-817e-70c366052962)

```mermaid
%% SMB2 Connect & Disconnect
sequenceDiagram
    Note over Client: Establish Connection
    Client->>Server: Negotiate Request
    Server-->>Client: Negotiate Response
    Client->>Server: Session Setup Request
    Server-->>Client: Session Setup Response
    Client->>Server: Tree Connect Request
    Server-->>Client: Tree Connect Response
    Note over Client: Read/Write Messages
    Client->>Server: ...
    Server-->>Client: ...
    Note over Client: Disconnect & Logoff
    Client->>Server: Tree Disconnect Request
    Server-->>Client: Tree Disconnect Response
    Client->>Server: Logoff Request
    Server-->>Client: Logoff Response
```
